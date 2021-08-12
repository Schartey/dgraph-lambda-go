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
		if m.Field.TypeName.Exported() && m.Field.TypeName.Pkg().Path() != c.DefaultModelPackage.PkgPath {
			pkgs[m.Field.TypeName.Pkg().Name()] = m.Field.TypeName.Pkg()
		}
	}

	for _, m := range c.ParsedTree.ResolverTree.Queries {
		if m.Return.TypeName.Exported() {
			pkgs[m.Return.TypeName.Pkg().Name()] = m.Return.TypeName.Pkg()
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
	return Executer{fieldResolver: {{.ResolverPackageName}}.FieldResolver{Resolver: resolver}, queryResolver: {{.ResolverPackageName}}.QueryResolver{Resolver: resolver}, middlewareResolver: {{.ResolverPackageName}}.MiddlewareResolver{Resolver: resolver}, webhookResolver: {{.ResolverPackageName}}.WebhookResolver{Resolver: resolver}}
}

func (e *Executer) Middleware(md *api.MiddlewareData) error {
	var err error
	switch md.Dbody.Resolver {
		{{- range $fieldResolver := .FieldResolvers}}
		case "{{$fieldResolver.Field.TypeName | typeName }}.{{$fieldResolver.Field.Name}}":
			{
				{{- range $middleware := $fieldResolver.Middleware}}
				if err = e.middlewareResolver.Middleware_{{$middleware}}(md); err != nil {
					return err
				}
				{{- end}}
				break
			}
		{{- end }}

		{{- range $query := .Queries}}
		case "Query.{{$query.Name}}":
			{
				{{- range $middleware := $query.Middleware}}
				if err = e.middlewareResolver.Middleware_{{$middleware}}(md); err != nil {
					return err
				}
				{{- end}}
				break
			}
		{{- end }}
		{{- range $mutation := .Mutations}}
		case "Mutation.{{$mutation.Name}}":
			{
				{{- range $middleware := $mutation.Middleware}}
				if err = e.middlewareResolver.Middleware_{{$middleware}}(md); err != nil {
					return err
				}
				{{- end}}
				break
			}
		{{- end }}
	}
	return nil
}

func (e Executer) Resolve(ctx context.Context, dbody api.DBody) ([]byte, error) {
	if &dbody.Event != nil {
		var err error
		switch dbody.Event.TypeName {
			{{- range $model := .Models}} {{ if ne (len $model.LambdaOnMutate) 0 }}
			case "{{ $model.TypeName | typeName }}":
				err = e.webhookResolver.Webhook_{{ $model.TypeName | typeName }}(ctx, dbody.Event)
				return nil, err
			{{ end }} {{ end }}
		}
	} else {
		parentsBytes, err := dbody.Parents.MarshalJSON()
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
			case "{{$fieldResolver.Field.ParentTypeName }}.{{$fieldResolver.Field.Name}}":
				{
					var parents []{{$fieldResolver.Field.GoType | pointer }}
					json.Unmarshal(parentsBytes, &parents)

					// Dependent on generation loop or just direct
					/*var {{$fieldResolver.Field.Name}}s []{{$fieldResolver.Field.GoType | ref}}
					for _, parent := range parents {
						{{$fieldResolver.Field.Name}}s = fullnames.append(e.fieldResolver.{{$fieldResolver.Field.GoType | ref }}_{{$fieldResolver.Field.Name}}(ctx, parent))
					}*/
					{{$fieldResolver.Field.Name | untitle }}s, err := e.fieldResolver.{{$fieldResolver.Field.ParentTypeName }}_{{$fieldResolver.Field.Name}}(ctx, parents, dbody.AuthHeader)
					if err != nil {
						return nil, err
					}

					response, err = json.Marshal({{$fieldResolver.Field.Name | untitle }}s)
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
					var {{ $arg.Name }} {{ $arg.GoType | pointer }} 
					json.Unmarshal(dbody.Args["{{$arg.Name}}"], &{{$arg.Name}})
					{{- end }}	
					{{ $query.Return.TypeName | typeName | untitle }}, err := e.queryResolver.Query_{{$query.Name}}(ctx, {{$query.Arguments | args}}, dbody.AuthHeader)
					if err != nil {
						return nil, err
					}
		
					response, err = json.Marshal({{ $query.Return.TypeName | typeName | untitle }})
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
					var {{ $arg.Name }} {{ $arg.GoType | pointer }} 
					json.Unmarshal(dbody.Args["{{$arg.Name}}"], &{{$arg.Name}})
					{{- end }}	
					{{ $mutation.Return.TypeName | typeName | untitle }}, err := e.mutationResolver.Mutation_{{$mutation.Name}}(ctx, {{$mutation.Arguments | args}}, dbody.AuthHeader)
					if err != nil {
						return nil, err
					}
		
					response, err = json.Marshal({{ $mutation.Return.TypeName | typeName | untitle }})
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
