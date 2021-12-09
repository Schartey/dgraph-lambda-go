package gogen

import (
	"errors"
	"fmt"
	"go/types"
	"os"
	"path"
	"text/template"

	"github.com/schartey/dgraph-lambda-go/codegen/config"
	"github.com/schartey/dgraph-lambda-go/codegen/generator"
	"github.com/schartey/dgraph-lambda-go/codegen/parser"
	"github.com/schartey/dgraph-lambda-go/codegen/rewriter"
)

func generateQueryResolvers(c *config.Config, parsedTree *parser.Tree, r *rewriter.Rewriter) error {

	if c.ResolverFilename == "resolver" {

		fileName := path.Join(c.Resolver.Dir, "query.resolver.go")
		f, err := os.Create(fileName)
		if err != nil {
			return err
		}
		defer f.Close()

		pkgs := make(map[string]*types.Package)

		for _, m := range parsedTree.ResolverTree.Queries {
			if m.Return.TypeName.Exported() {
				pkgs[m.Return.TypeName.Pkg().Name()] = m.Return.TypeName.Pkg()
			}

			for _, f := range m.Arguments {
				if f.TypeName.Exported() {
					pkgs[f.GoType.TypeName.Pkg().Name()] = f.GoType.TypeName.Pkg()
				}
			}
		}
		if len(parsedTree.ResolverTree.Queries) > 0 {
			pkgs["context"] = types.NewPackage("context", "context")
			pkgs["api"] = types.NewPackage("github.com/schartey/dgraph-lambda-go/api", "api")
		}

		err = queryResolverTemplate.Execute(f, struct {
			QueryResolvers map[string]*parser.Query
			Rewriter       *rewriter.Rewriter
			Packages       map[string]*types.Package
			PackageName    string
		}{
			QueryResolvers: parsedTree.ResolverTree.Queries,
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
	"path":    generator.PkgPath,
	"pointer": generator.Pointer,
	"argsW":   generator.ArgsW,
	"body":    generator.Body,
	"is":      generator.Is,
}).Parse(`
package {{.PackageName}}

import(
	{{- range $pkg := .Packages }}
	"{{ $pkg | path }}"{{- end}}
)

type QueryResolverInterface interface {
{{- range $queryResolver := .QueryResolvers}}
	Query_{{$queryResolver.Name}}(ctx context.Context{{ if ne (len $queryResolver.Arguments) 0}}, {{ $queryResolver.Arguments | argsW }}{{ end }}, authHeader api.AuthHeader) ({{ ref $queryResolver.Return.GoType $queryResolver.Return.IsArray }}, *api.LambdaError){{ end }}
}

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
