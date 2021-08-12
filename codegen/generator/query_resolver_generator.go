package generator

import (
	"errors"
	"go/types"
	"os"
	"path"
	"text/template"

	"github.com/schartey/dgraph-lambda-go/codegen/config"
	"github.com/schartey/dgraph-lambda-go/codegen/parser"
	"github.com/schartey/dgraph-lambda-go/codegen/rewriter"
)

func generateQueryResolvers(c *config.Config, r *rewriter.Rewriter) error {

	if c.ResolverFilename == "resolver" {

		fileName := path.Join(c.Resolver.Dir, "query.resolver.go")
		f, err := os.Create(fileName)
		if err != nil {
			return err
		}
		f.Close()

		pkgs := make(map[string]*types.Package)

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
		if len(c.ParsedTree.ResolverTree.Queries) > 0 {
			pkgs["context"] = types.NewPackage("context", "context")
			pkgs["api"] = types.NewPackage("github.com/schartey/dgraph-lambda-go/api", "api")
		}

		err = queryResolverTemplate.Execute(f, struct {
			QueryResolvers map[string]*parser.Query
			Rewriter       *rewriter.Rewriter
			Packages       map[string]*types.Package
			PackageName    string
		}{
			QueryResolvers: c.ParsedTree.ResolverTree.Queries,
			Rewriter:       r,
			Packages:       pkgs,
			PackageName:    c.Resolver.Package,
		})
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("Resolver file pattern invalid")
}

var queryResolverTemplate = template.Must(template.New("query-resolver").Funcs(template.FuncMap{
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
type QueryResolver struct {
	*Resolver
}

{{- range $queryResolver := .QueryResolvers}}
func (q *QueryResolver) Query_{{$queryResolver.Name}}(ctx context.Context, {{ $queryResolver.Arguments | argsW }}, authHeader api.AuthHeader) ({{$queryResolver.Return | pointer}}, error) { {{ body (printf "Query_%s" $queryResolver.Name) $.Rewriter }}}
{{ end }}

{{- range $key, $depBody := .Rewriter.DeprecatedBodies }}
{{ if is $key "Query_" }}
/* {{ $depBody }} */
{{ end }}
{{ end }}
`))
