package gogen

import (
	"errors"
	"text/template"

	"github.com/schartey/dgraph-lambda-go/codegen/generator/tools"
	"github.com/schartey/dgraph-lambda-go/codegen/parser"
	"github.com/schartey/dgraph-lambda-go/codegen/rewriter"
	"github.com/schartey/dgraph-lambda-go/config"
)

func generateMutationResolvers(c *config.Config, parsedTree *parser.Tree, r *rewriter.Rewriter) error {
	/*if c.ResolverFilename == "resolver" {

		fileName := path.Join(c.Resolver.Dir, "mutation.resolver.go")
		f, err := os.Create(fileName)
		if err != nil {
			fmt.Println(err.Error())
		}
		defer f.Close()

		pkgs := make(map[string]*types.Package)

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
		if len(parsedTree.ResolverTree.Mutations) > 0 {
			pkgs["context"] = types.NewPackage("context", "context")
			pkgs["api"] = types.NewPackage("github.com/schartey/dgraph-lambda-go/api", "api")
		}

		err = mutationResolverTemplate.Execute(f, struct {
			MutationResolvers map[string]*parser.Mutation
			Rewriter          *rewriter.Rewriter
			Packages          map[string]*types.Package
			PackageName       string
		}{
			MutationResolvers: parsedTree.ResolverTree.Mutations,
			Rewriter:          r,
			Packages:          pkgs,
			PackageName:       c.Resolver.Package,
		})
		if err != nil {
			return err
		}
		return nil
	}*/
	return errors.New("Resolver file pattern invalid")
}

var mutationResolverTemplate = template.Must(template.New("mutation-resolver").Funcs(template.FuncMap{
	"ref":     returnRef,
	"path":    tools.PkgPath,
	"pointer": tools.Pointer,
	"argsW":   tools.ArgsW,
	"body":    tools.Body,
	"is":      tools.Is,
}).Parse(`
package {{.PackageName}}

import(
	{{- range $pkg := .Packages }}
	"{{ $pkg | path }}"{{- end}}
)

type MutationResolverInterface interface {
{{- range $mutationResolver := .MutationResolvers}}
	Mutation_{{$mutationResolver.Name}}(ctx context.Context{{ if ne (len $mutationResolver.Arguments) 0}}, {{ $mutationResolver.Arguments | argsW }}{{ end }}, authHeader api.AuthHeader) ({{ ref $mutationResolver.Return.GoType $mutationResolver.Return.IsArray }}, *api.LambdaError){{ end }}
}

type MutationResolver struct {
	*Resolver
}

{{- range $mutationResolver := .MutationResolvers}}
func (q *MutationResolver) Mutation_{{$mutationResolver.Name}}(ctx context.Context{{ if ne (len $mutationResolver.Arguments) 0}}, {{ $mutationResolver.Arguments | argsW }}{{ end }}, authHeader api.AuthHeader) ({{ ref $mutationResolver.Return.GoType $mutationResolver.Return.IsArray }}, *api.LambdaError) { {{ body $mutationResolver.Return.GoType $mutationResolver.Return.IsArray (printf "Mutation_%s" $mutationResolver.Name) $.Rewriter }}}
{{ end }}

{{- range $key, $depBody := .Rewriter.DeprecatedBodies }}
{{ if is $key "Mutation_" }}
/* {{ $depBody }} */
{{ end }}
{{ end }}
`))
