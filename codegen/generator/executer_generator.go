package generator

import (
	"go/types"
	"os"
	"path"
	"text/template"

	"github.com/schartey/dgraph-lambda-go/codegen/config"
	"github.com/schartey/dgraph-lambda-go/codegen/parser"
	"github.com/schartey/dgraph-lambda-go/codegen/rewriter"
)

func generateExecuter(c *config.Config, r *rewriter.Rewriter) error {
	f, err := os.Create(c.Exec.Filename)
	if err != nil {
		return err
	}
	defer f.Close()

	pkgs := make(map[string]*types.Package)

	for _, m := range c.ParsedTree.ResolverTree.FieldResolvers {
		if m.Field.TypeName.Exported() {
			pkgs[m.Field.TypeName.Pkg().Name()] = m.Field.TypeName.Pkg()
		}
		if m.Parent.TypeName.Exported() {
			pkgs[m.Parent.TypeName.Pkg().Name()] = m.Parent.TypeName.Pkg()
		}
	}

	for _, m := range c.ParsedTree.ResolverTree.Queries {
		if m.Return.TypeName.Exported() {
			//pkgs[m.Return.TypeName.Pkg().Name()] = m.Return.TypeName.Pkg()
		}

		for _, f := range m.Arguments {
			if f.TypeName.Exported() {
				pkgs[f.GoType.TypeName.Pkg().Name()] = f.GoType.TypeName.Pkg()
			}
		}
	}

	for _, m := range c.ParsedTree.ResolverTree.Mutations {
		if m.Return.TypeName.Exported() {
			pkgs[m.Return.TypeName.Pkg().Name()] = m.Return.TypeName.Pkg()
		}

		for _, f := range m.Arguments {
			if f.TypeName.Exported() {
				pkgs[f.GoType.TypeName.Pkg().Name()] = f.GoType.TypeName.Pkg()
			}
		}
	}

	pkgs["context"] = types.NewPackage("context", "context")
	pkgs["errors"] = types.NewPackage("errors", "errors")
	pkgs["api"] = types.NewPackage("github.com/schartey/dgraph-lambda-go/api", "api")
	pkgs["api"] = types.NewPackage("github.com/schartey/dgraph-lambda-go/api", "api")
	pkgs["json"] = types.NewPackage("encoding/json", "json")

	pkgs[c.Resolver.Package] = types.NewPackage(path.Join(c.Root, c.Resolver.Dir), c.Resolver.Package)

	err = executerTemplate.Execute(f, struct {
		FieldResolvers      map[string]*parser.FieldResolver
		Queries             map[string]*parser.Query
		Mutations           map[string]*parser.Mutation
		Middleware          map[string]string
		Models              map[string]*parser.Model
		Packages            map[string]*types.Package
		PackageName         string
		ResolverPackageName string
	}{
		FieldResolvers:      c.ParsedTree.ResolverTree.FieldResolvers,
		Queries:             c.ParsedTree.ResolverTree.Queries,
		Mutations:           c.ParsedTree.ResolverTree.Mutations,
		Middleware:          c.ParsedTree.Middleware,
		Models:              c.ParsedTree.ModelTree.Models,
		Packages:            pkgs,
		PackageName:         c.Exec.Package,
		ResolverPackageName: c.Resolver.Package,
	})
	if err != nil {
		return err
	}
	return nil
}

var executerTemplate = template.Must(template.New("executer").Funcs(template.FuncMap{
	"path":     pkgPath,
	"typeName": typeName,
	"ref":      resolverRef,
	"untitle":  untitle,
	"args":     args,
	"pointer":  pointer,
}).Parse(`
package {{.PackageName}}

import(
	{{- range $pkg := .Packages }}
	"{{ $pkg | path }}"{{- end}}
)

type Executer struct {
	fieldResolver    	{{.ResolverPackageName}}.FieldResolver
	queryResolver    	{{.ResolverPackageName}}.QueryResolver
	mutationResolver 	{{.ResolverPackageName}}.MutationResolver
	middlewareResolver 	{{.ResolverPackageName}}.MiddlewareResolver
	webhookResolver 	{{.ResolverPackageName}}.WebhookResolver
}

func NewExecuter(resolver *{{.ResolverPackageName}}.Resolver) api.ExecuterInterface {
	return Executer{fieldResolver: {{.ResolverPackageName}}.FieldResolver{Resolver: resolver}, queryResolver: {{.ResolverPackageName}}.QueryResolver{Resolver: resolver}, mutationResolver: {{.ResolverPackageName}}.MutationResolver{Resolver: resolver}, middlewareResolver: {{.ResolverPackageName}}.MiddlewareResolver{Resolver: resolver}, webhookResolver: {{.ResolverPackageName}}.WebhookResolver{Resolver: resolver}}
}

func (e *Executer) Middleware(md *api.MiddlewareData) (err error) {
	switch md.Dbody.Resolver {
		{{- range $fieldResolver := .FieldResolvers}}{{ if ne (len $fieldResolver.Middleware) 0 }}
		case "{{$fieldResolver.Parent.Name }}.{{$fieldResolver.Field.Name}}":
			{
				{{- range $middleware := $fieldResolver.Middleware}}
				if err = e.middlewareResolver.Middleware_{{$middleware}}(md); err != nil {
					return err
				}
				{{- end}}
				break
			}
		{{ end }}{{- end }}

		{{- range $query := .Queries}}{{ if ne (len $query.Middleware) 0 }}
		case "Query.{{$query.Name}}":
			{
				{{- range $middleware := $query.Middleware}}
				if err = e.middlewareResolver.Middleware_{{$middleware}}(md); err != nil {
					return err
				}
				{{- end}}
				break
			}
		{{ end }}{{- end }}
		{{- range $mutation := .Mutations}}{{ if ne (len $mutation.Middleware) 0 }}
		case "Mutation.{{$mutation.Name}}":
			{
				{{- range $middleware := $mutation.Middleware}}
				if err = e.middlewareResolver.Middleware_{{$middleware}}(md); err != nil {
					return err
				}
				{{- end}}
				break
			}
		{{ end }}{{- end }}
	}
	return nil
}

func (e Executer) Resolve(ctx context.Context, dbody api.DBody) (response []byte, err error) {
	if dbody.Event.Operation != "" {
		switch dbody.Event.TypeName {
			{{- range $model := .Models}} {{ if ne (len $model.LambdaOnMutate) 0 }}
			case "{{ $model.TypeName | typeName }}":
				err = e.webhookResolver.Webhook_{{ $model.TypeName | typeName }}(ctx, dbody.Event)
				return nil, err
			{{ end }} {{ end }}
		}
	} else {
		{{ if ne (len .FieldResolvers) 0}}parentsBytes, err := dbody.Parents.MarshalJSON(){{ end }}
		if err != nil {
			return nil, err
		}

		md := &api.MiddlewareData{Ctx: ctx, Dbody: dbody}
		if err = e.Middleware(md); err != nil {
			return nil, err
		}
		ctx = md.Ctx
		dbody = md.Dbody

		response := []byte{}

		switch dbody.Resolver {
			{{- range $fieldResolver := .FieldResolvers}}
			case "{{$fieldResolver.Parent.Name }}.{{$fieldResolver.Field.Name}}":
				{
					var parents []{{ pointer $fieldResolver.Parent.GoType false }}
					json.Unmarshal(parentsBytes, &parents)

					// Dependent on generation loop or just direct
					/*var {{$fieldResolver.Field.Name}}s []{{$fieldResolver.Field.GoType | ref}}
					for _, parent := range parents {
						{{$fieldResolver.Field.Name}}s = fullnames.append(e.fieldResolver.{{$fieldResolver.Field.GoType | ref }}_{{$fieldResolver.Field.Name}}(ctx, parent))
					}*/
					result, err := e.fieldResolver.{{$fieldResolver.Parent.Name }}_{{$fieldResolver.Field.Name}}(ctx, parents, dbody.AuthHeader)
					if err != nil {
						return nil, err
					}

					response, err = json.Marshal(result)
					if err != nil {
						return nil, err
					}
					break
				}
			{{- end }}

			{{- range $query := .Queries}}
			case "Query.{{$query.Name}}":
				{
					{{- range $arg := $query.Arguments }}
					var {{ $arg.Name }} {{ pointer $arg.GoType $arg.IsArray }} 
					json.Unmarshal(dbody.Args["{{$arg.Name}}"], &{{$arg.Name}})
					{{- end }}	
					result, err := e.queryResolver.Query_{{$query.Name}}(ctx{{ if ne (len $query.Arguments) 0}}, {{$query.Arguments | args}}{{end}}, dbody.AuthHeader)
					if err != nil {
						return nil, err
					}
		
					response, err = json.Marshal(result)
					if err != nil {
						return nil, err
					}
					break
				}
			{{- end }}
			{{- range $mutation := .Mutations}}
			case "Mutation.{{$mutation.Name}}":
				{
					{{- range $arg := $mutation.Arguments }}
					var {{ $arg.Name }} {{ pointer $arg.GoType $arg.IsArray }} 
					json.Unmarshal(dbody.Args["{{$arg.Name}}"], &{{$arg.Name}})
					{{- end }}	
					result, err := e.mutationResolver.Mutation_{{$mutation.Name}}(ctx{{ if ne (len $mutation.Arguments) 0}}, {{$mutation.Arguments | args}}{{ end }}, dbody.AuthHeader)
					if err != nil {
						return nil, err
					}
		
					response, err = json.Marshal(result)
					if err != nil {
						return nil, err
					}
					break
				}
			{{- end }}
		}
		return response, nil
	}
	return nil, errors.New("No resolver found")
}
`))
