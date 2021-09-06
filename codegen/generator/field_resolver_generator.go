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

func generateFieldResolvers(c *config.Config, r *rewriter.Rewriter) error {

	if c.ResolverFilename == "resolver" {

		fileName := path.Join(c.Resolver.Dir, "field.resolver.go")
		f, err := os.Create(fileName)
		if err != nil {
			return err
		}
		defer f.Close()

		var pkgs = make(map[string]*types.Package)

		for _, m := range c.ParsedTree.ResolverTree.FieldResolvers {
			if m.Field.TypeName.Exported() {
				pkgs[m.Field.TypeName.Pkg().Name()] = m.Field.TypeName.Pkg()
			}
			if m.Parent.TypeName.Exported() {
				pkgs[m.Parent.TypeName.Pkg().Name()] = m.Parent.TypeName.Pkg()
			}
		}
		if len(c.ParsedTree.ResolverTree.FieldResolvers) > 0 {
			pkgs["context"] = types.NewPackage("context", "context")
			pkgs["api"] = types.NewPackage("github.com/schartey/dgraph-lambda-go/api", "api")
		}

		err = fieldResolverTemplate.Execute(f, struct {
			FieldResolvers map[string]*parser.FieldResolver
			Rewriter       *rewriter.Rewriter
			Packages       map[string]*types.Package
			PackageName    string
		}{
			FieldResolvers: c.ParsedTree.ResolverTree.FieldResolvers,
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

var fieldResolverTemplate = template.Must(template.New("field-resolver").Funcs(template.FuncMap{
	"ref":     modelRef,
	"path":    pkgPath,
	"pointer": pointer,
	"body":    body,
	"is":      is,
}).Parse(`
package {{.PackageName}}

import(
	{{- range $pkg := .Packages }}
	"{{ $pkg | path }}"{{- end}}
)

/** Put these into resolvers.go  or similar **/
type FieldResolver struct {
	*Resolver
}

{{- range $fieldResolver := .FieldResolvers}}
func (f *FieldResolver) {{$fieldResolver.Parent.Name }}_{{$fieldResolver.Field.Name}}(ctx context.Context, parents []{{ pointer $fieldResolver.Parent.GoType false }}, authHeader api.AuthHeader) ([]{{ pointer $fieldResolver.Field.GoType $fieldResolver.Field.IsArray }}, *api.LambdaError) { {{ body (printf "%s_%s" $fieldResolver.Parent.Name $fieldResolver.Field.Name) $.Rewriter }}}
{{ end }}

{{- range $key, $depBody := .Rewriter.DeprecatedBodies }}
{{ if and (not (is $key "Query_")) (not (is $key "Mutation_")) (not (is $key "Middleware_")) }}
/* {{ $depBody }} */
{{ end }}
{{ end }}
`))
