package generator

import (
	"errors"
	"fmt"
	"go/types"
	"os"
	"path"
	"text/template"

	"github.com/schartey/dgraph-lambda-go/codegen/config"
	"github.com/schartey/dgraph-lambda-go/codegen/parser"
	"github.com/schartey/dgraph-lambda-go/codegen/rewriter"
)

func generateMutationResolvers(c *config.Config, r *rewriter.Rewriter) error {
	if c.ResolverFilename == "resolver" {

		fileName := path.Join(c.Resolver.Dir, "mutation.resolver.go")
		f, err := os.Create(fileName)
		if err != nil {
			fmt.Println(err.Error())
		}
		defer f.Close()

		pkgs := make(map[string]*types.Package)

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
		if len(c.ParsedTree.ResolverTree.Mutations) > 0 {
			pkgs["context"] = types.NewPackage("context", "context")
			pkgs["api"] = types.NewPackage("github.com/schartey/dgraph-lambda-go/api", "api")
		}

		err = mutationResolverTemplate.Execute(f, struct {
			MutationResolvers map[string]*parser.Mutation
			Rewriter          *rewriter.Rewriter
			Packages          map[string]*types.Package
			PackageName       string
		}{
			MutationResolvers: c.ParsedTree.ResolverTree.Mutations,
			Rewriter:          r,
			Packages:          pkgs,
			PackageName:       c.Resolver.Package,
		})
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("Resolver file pattern invalid")
}

var mutationResolverTemplate = template.Must(template.New("mutation-resolver").Funcs(template.FuncMap{
	"ref":     resolverRef,
	"path":    pkgPath,
	"pointer": pointer,
	"argsW":   argsW,
	"body":    body,
	"is":      is,
}).Parse(`
package {{.PackageName}}

import(
	{{- range $pkg := .Packages }}
	"{{ $pkg | path }}"{{- end}}
)

/** Put these into resolvers.go  or similar **/
type MutationResolver struct {
	*Resolver
}

{{- range $mutationResolver := .MutationResolvers}}
func (q *MutationResolver) Mutation_{{$mutationResolver.Name}}(ctx context.Context, {{ $mutationResolver.Arguments | argsW }}, authHeader api.AuthHeader) ({{$mutationResolver.Return | pointer}}, error) { {{ body (printf "Mutation_%s" $mutationResolver.Name) $.Rewriter }}}
{{ end }}

{{- range $key, $depBody := .Rewriter.DeprecatedBodies }}
{{ if is $key "Mutation_" }}
/* {{ $depBody }} */
{{ end }}
{{ end }}
`))
