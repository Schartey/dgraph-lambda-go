package gogen

import (
	"go/types"
	"os"
	"path"
	"text/template"

	"github.com/schartey/dgraph-lambda-go/codegen/config"
	"github.com/schartey/dgraph-lambda-go/codegen/generator"
	"github.com/schartey/dgraph-lambda-go/codegen/parser"
	"github.com/schartey/dgraph-lambda-go/codegen/rewriter"
)

func generateExecuter(c *config.Config, parsedTree *parser.Tree, r *rewriter.Rewriter) error {
	f, err := os.Create(c.Exec.Filename)
	if err != nil {
		return err
	}
	defer f.Close()

	pkgs := make(map[string]*types.Package)
	var lambdaOnMutate []string

	for _, m := range parsedTree.ModelTree.Models {
		if len(m.LambdaOnMutate) > 0 {
			lambdaOnMutate = append(lambdaOnMutate, m.Name)
		}
	}

	for _, m := range parsedTree.ResolverTree.FieldResolvers {
		if m.Field.TypeName.Exported() {
			pkgs[m.Field.TypeName.Pkg().Name()] = m.Field.TypeName.Pkg()
		}
		if m.Parent.TypeName.Exported() {
			pkgs[m.Parent.TypeName.Pkg().Name()] = m.Parent.TypeName.Pkg()
		}
	}

	for _, m := range parsedTree.ResolverTree.Queries {
		if m.Return.TypeName.Exported() {
			//pkgs[m.Return.TypeName.Pkg().Name()] = m.Return.TypeName.Pkg()
		}

		for _, f := range m.Arguments {
			if f.TypeName.Exported() {
				pkgs[f.GoType.TypeName.Pkg().Name()] = f.GoType.TypeName.Pkg()
			}
		}
	}

	for _, m := range parsedTree.ResolverTree.Mutations {
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
	pkgs["http"] = types.NewPackage("net/http", "http")
	pkgs["strings"] = types.NewPackage("strings", "strings")
	pkgs["api"] = types.NewPackage("github.com/schartey/dgraph-lambda-go/api", "api")

	if len(parsedTree.ResolverTree.FieldResolvers) > 0 ||
		len(parsedTree.ResolverTree.Queries) > 0 ||
		len(parsedTree.ResolverTree.Mutations) > 0 {
		pkgs["json"] = types.NewPackage("encoding/json", "json")
	}

	pkgs[c.Resolver.Package] = types.NewPackage(path.Join(c.Root, c.Resolver.Dir), c.Resolver.Package)

	err = executerTemplate.Execute(f, struct {
		FieldResolvers      map[string]*parser.FieldResolver
		Queries             map[string]*parser.Query
		Mutations           map[string]*parser.Mutation
		Middleware          map[string]string
		Models              map[string]*parser.Model
		LambdaOnMutate      []string
		Packages            map[string]*types.Package
		PackageName         string
		ResolverPackageName string
	}{
		FieldResolvers:      parsedTree.ResolverTree.FieldResolvers,
		Queries:             parsedTree.ResolverTree.Queries,
		Mutations:           parsedTree.ResolverTree.Mutations,
		Middleware:          parsedTree.Middleware,
		Models:              parsedTree.ModelTree.Models,
		LambdaOnMutate:      lambdaOnMutate,
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
	"path":     generator.PkgPath,
	"typeName": generator.TypeName,
	"ref":      generator.ResolverRef,
	"untitle":  generator.Untitle,
	"args":     generator.Args,
	"pointer":  generator.Pointer,
}).Parse(`
package {{.PackageName}}

import(
	{{- range $pkg := .Packages }}
	"{{ $pkg | path }}"{{- end}}
)

type Executer struct {
	api.ExecuterInterface
	fieldResolver    	{{.ResolverPackageName}}.FieldResolver
	queryResolver    	{{.ResolverPackageName}}.QueryResolver
	mutationResolver 	{{.ResolverPackageName}}.MutationResolver
	middlewareResolver 	{{.ResolverPackageName}}.MiddlewareResolver
	webhookResolver 	{{.ResolverPackageName}}.WebhookResolver
}

func NewExecuter(resolver *{{.ResolverPackageName}}.Resolver) api.ExecuterInterface {
	return Executer{fieldResolver: {{.ResolverPackageName}}.FieldResolver{Resolver: resolver}, queryResolver: {{.ResolverPackageName}}.QueryResolver{Resolver: resolver}, mutationResolver: {{.ResolverPackageName}}.MutationResolver{Resolver: resolver}, middlewareResolver: {{.ResolverPackageName}}.MiddlewareResolver{Resolver: resolver}, webhookResolver: {{.ResolverPackageName}}.WebhookResolver{Resolver: resolver}}
}

func (e Executer) Resolve(ctx context.Context, request *api.Request) (response []byte, err *api.LambdaError) {
	if request.Resolver == "$webhook" {
		return nil, e.resolveWebhook(ctx, request)
	} else {
		parentsBytes, underlyingError := request.Parents.MarshalJSON()
		if underlyingError != nil {
			return nil, &api.LambdaError{Underlying: underlyingError, Status: http.StatusInternalServerError}
		}

		mc := &api.MiddlewareContext{Ctx: ctx, Request: request}
		if err = e.middleware(mc); err != nil {
			return nil, err
		}
		ctx = mc.Ctx
		request = mc.Request

		if strings.HasPrefix(request.Resolver, "Query.") {
			return e.resolveQuery(ctx, request)
		} else if strings.HasPrefix(request.Resolver, "Mutation.") {
			return e.resolveMutation(ctx, request)
		} else {
			return e.resolveField(ctx, request, parentsBytes)
		}
	}
}

func (e Executer) middleware(mc *api.MiddlewareContext) (err *api.LambdaError) {
	switch mc.Request.Resolver {
		{{- range $fieldResolver := .FieldResolvers}}{{ if ne (len $fieldResolver.Middleware) 0 }}
		case "{{$fieldResolver.Parent.Name }}.{{$fieldResolver.Field.Name}}":
			{
				{{- range $middleware := $fieldResolver.Middleware}}
				if err = e.middlewareResolver.Middleware_{{$middleware}}(mc); err != nil {
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
				if err = e.middlewareResolver.Middleware_{{$middleware}}(mc); err != nil {
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
				if err = e.middlewareResolver.Middleware_{{$middleware}}(mc); err != nil {
					return err
				}
				{{- end}}
				break
			}
		{{ end }}{{- end }}
	}
	return nil
}

func (e Executer) resolveField(ctx context.Context, request *api.Request, parentsBytes []byte) (response []byte, err *api.LambdaError) {
	switch request.Resolver {
		{{- range $fieldResolver := .FieldResolvers}}
		case "{{$fieldResolver.Parent.Name }}.{{$fieldResolver.Field.Name}}":
			{
				var parents []{{ pointer $fieldResolver.Parent.GoType false }}
				json.Unmarshal(parentsBytes, &parents)

				result, err := e.fieldResolver.{{$fieldResolver.Parent.Name }}_{{$fieldResolver.Field.Name}}(ctx, parents, request.AuthHeader)
				if err != nil {
					return nil, err
				}

				var underlyingError error
				response, underlyingError = json.Marshal(result)
				if underlyingError != nil {
					return nil, &api.LambdaError{Underlying: underlyingError, Status: http.StatusInternalServerError}
				} else {
					return response, nil
				}
				break
			}
		{{- end }}
	}

	return nil, &api.LambdaError{Underlying: errors.New("could not find query resolver"), Status: http.StatusNotFound}
}

func (e Executer) resolveQuery(ctx context.Context, request *api.Request) (response []byte, err *api.LambdaError) {
	switch request.Resolver {
	{{- range $query := .Queries}}
         case "Query.{{$query.Name}}":
	{
		{{- range $arg := $query.Arguments }}
		var {{ $arg.Name }} {{ pointer $arg.GoType $arg.IsArray }} 
		json.Unmarshal(request.Args["{{$arg.Name}}"], &{{$arg.Name}})
		{{- end }}	
		result, err := e.queryResolver.Query_{{$query.Name}}(ctx{{ if ne (len $query.Arguments) 0}}, {{$query.Arguments | args}}{{end}}, request.AuthHeader)
		if err != nil {
			return nil, err
		}

		var underlyingError error
		response, underlyingError = json.Marshal(result)
		if underlyingError != nil {
			return nil, &api.LambdaError{Underlying: underlyingError, Status: http.StatusInternalServerError}
		} else {
			return response, nil
		}
		break
	}
{{- end }}
    }

	return nil, &api.LambdaError{Underlying: errors.New("could not find query resolver"), Status: http.StatusNotFound}
}

func (e Executer) resolveMutation(ctx context.Context, request *api.Request) (response []byte, err *api.LambdaError) {
	switch request.Resolver {
		{{- range $mutation := .Mutations}}
		case "Mutation.{{$mutation.Name}}":
			{
				{{- range $arg := $mutation.Arguments }}
				var {{ $arg.Name }} {{ pointer $arg.GoType $arg.IsArray }} 
				json.Unmarshal(request.Args["{{$arg.Name}}"], &{{$arg.Name}})
				{{- end }}	
				result, err := e.mutationResolver.Mutation_{{$mutation.Name}}(ctx{{ if ne (len $mutation.Arguments) 0}}, {{$mutation.Arguments | args}}{{ end }}, request.AuthHeader)
				if err != nil {
					return nil, err
				}

				var underlyingError error
				response, underlyingError = json.Marshal(result)
				if underlyingError != nil {
					return nil, &api.LambdaError{Underlying: underlyingError, Status: http.StatusInternalServerError}
				} else {
					return response, nil
				}
				break
			}
		{{- end }}
    }

	return nil, &api.LambdaError{Underlying: errors.New("could not find query resolver"), Status: http.StatusNotFound}
}

func (e Executer) resolveWebhook(ctx context.Context, request *api.Request) (err *api.LambdaError) {
	switch request.Event.TypeName {
		{{- range $name := .LambdaOnMutate}}
	case "{{$name}}":
		err = e.webhookResolver.Webhook_{{$name}}(ctx, request.Event)
		return err
		{{- end }}
	}
	
	return &api.LambdaError{Underlying: errors.New("could not find webhook resolver"), Status: http.StatusNotFound}
}

`))
