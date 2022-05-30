package generator

import (
	"errors"
	"go/types"
	"os"
	"path"
	"text/template"

	"github.com/miko/dgraph-lambda-go/codegen/config"
	"github.com/miko/dgraph-lambda-go/codegen/parser"
	"github.com/miko/dgraph-lambda-go/codegen/rewriter"
)

func generateMiddleware(c *config.Config, parsedTree *parser.Tree, r *rewriter.Rewriter) error {
	if c.ResolverFilename == "resolver" {

		fileName := path.Join(c.Resolver.Dir, "middleware.resolver.go")
		f, err := os.Create(fileName)
		if err != nil {
			return err
		}
		defer f.Close()

		pkgs := make(map[string]*types.Package)

		if len(parsedTree.Middleware) > 0 {
			pkgs["api"] = types.NewPackage("github.com/miko/dgraph-lambda-go/api", "api")
		}

		err = middlewareResolverTemplate.Execute(f, struct {
			Middleware  map[string]string
			Rewriter    *rewriter.Rewriter
			Packages    map[string]*types.Package
			PackageName string
		}{
			Middleware:  parsedTree.Middleware,
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

type MiddlewareResolverInterface interface {
{{- range $middleware := .Middleware}}
	Middleware_{{$middleware}}(mc *api.MiddlewareContext) *api.LambdaError{{ end }}
}

type MiddlewareResolver struct {
	*Resolver
}

{{ range $middleware := .Middleware}}
func (m *MiddlewareResolver) Middleware_{{$middleware}}(mc *api.MiddlewareContext) *api.LambdaError { {{ body (printf "Middleware_%s" $middleware) $.Rewriter }}}
{{ end }}

{{- range $key, $depBody := .Rewriter.DeprecatedBodies }}
{{ if is $key "Middleware_" }}
/* {{ $depBody }} */
{{ end }}
{{ end }}
`))
