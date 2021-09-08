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

func generateQueryResolvers(c *config.Config, r *rewriter.Rewriter) error {

	if c.ResolverFilename == "resolver" {

		fileName := path.Join(c.Resolver.Dir, "query.resolver.go")
		f, err := os.Create(fileName)
		if err != nil {
			return err
		}
		defer f.Close()

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

func returnRef(t *parser.GoType, isArray bool) string {
	if t.TypeName.Pkg() != nil {
		for _, te := range autobind {
			if te == t.TypeName.Pkg().Path() {
				if isArray {
					return fmt.Sprintf("[]*%s.%s", t.TypeName.Pkg().Name(), t.TypeName.Name())
				} else {
					return fmt.Sprintf("*%s.%s", t.TypeName.Pkg().Name(), t.TypeName.Name())
				}
			}
		}
		if t.TypeName.Exported() {
			if isArray {
				return fmt.Sprintf("[]*%s.%s", t.TypeName.Pkg().Name(), t.TypeName.Name())
			} else {
				return fmt.Sprintf("*%s.%s", t.TypeName.Pkg().Name(), t.TypeName.Name())
			}
		}
	}
	if isArray {
		return fmt.Sprintf("[]%s", t.TypeName.Name())
	} else {
		return t.TypeName.Name()
	}
}

var queryResolverTemplate = template.Must(template.New("query-resolver").Funcs(template.FuncMap{
	"ref":     returnRef,
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
func (q *QueryResolver) Query_{{$queryResolver.Name}}(ctx context.Context{{ if ne (len $queryResolver.Arguments) 0}}, {{ $queryResolver.Arguments | argsW }}{{ end }}, authHeader api.AuthHeader) ({{ ref $queryResolver.Return.GoType $queryResolver.Return.IsArray }}, *api.LambdaError) { {{ body $queryResolver.Return.GoType $queryResolver.Return.IsArray (printf "Query_%s" $queryResolver.Name) $.Rewriter }}}
{{ end }}

{{- range $key, $depBody := .Rewriter.DeprecatedBodies }}
{{ if is $key "Query_" }}
/* {{ $depBody }} */
{{ end }}
{{ end }}
`))
