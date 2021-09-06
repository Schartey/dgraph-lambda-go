package generator

import (
	"errors"
	"go/types"
	"os"
	"path"
	"text/template"

	"github.com/schartey/dgraph-lambda-go/codegen/config"
	"github.com/schartey/dgraph-lambda-go/codegen/rewriter"
)

func generateMiddleware(c *config.Config, r *rewriter.Rewriter) error {
	if c.ResolverFilename == "resolver" {

		fileName := path.Join(c.Resolver.Dir, "middleware.resolver.go")
		f, err := os.Create(fileName)
		if err != nil {
			return err
		}
		defer f.Close()

		pkgs := make(map[string]*types.Package)

		if len(c.ParsedTree.Middleware) > 0 {
			pkgs["api"] = types.NewPackage("github.com/schartey/dgraph-lambda-go/api", "api")
		}

		err = middlewareResolverTemplate.Execute(f, struct {
			Middleware  map[string]string
			Rewriter    *rewriter.Rewriter
			Packages    map[string]*types.Package
			PackageName string
		}{
			Middleware:  c.ParsedTree.Middleware,
			Rewriter:    r,
			Packages:    pkgs,
			PackageName: c.Resolver.Package,
		})
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("Resolver file pattern invalid")
}

var middlewareResolverTemplate = template.Must(template.New("middleware-resolver").Funcs(template.FuncMap{
	"path": pkgPath,
	"body": middlewareBody,
	"is":   is,
}).Parse(`
package {{.PackageName}}

import(
	{{- range $pkg := .Packages }}
	"{{ $pkg | path }}"{{- end}}
)

/** Put these into resolvers.go  or similar **/
type MiddlewareResolver struct {
	*Resolver
}

{{ range $middleware := .Middleware}}
func (m *MiddlewareResolver) Middleware_{{$middleware}}(md *api.MiddlewareData) *api.LambdaError { {{ body (printf "Middleware_%s" $middleware) $.Rewriter }}}
{{ end }}

{{- range $key, $depBody := .Rewriter.DeprecatedBodies }}
{{ if is $key "Middleware_" }}
/* {{ $depBody }} */
{{ end }}
{{ end }}
`))
