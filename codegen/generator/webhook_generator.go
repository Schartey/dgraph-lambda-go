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

func generateWebhook(c *config.Config, parsedTree *parser.Tree, r *rewriter.Rewriter) error {
	if c.ResolverFilename == "resolver" {

		fileName := path.Join(c.Resolver.Dir, "webhook.resolver.go")
		f, err := os.Create(fileName)
		if err != nil {
			return err
		}
		defer f.Close()

		pkgs := make(map[string]*types.Package)

		var models = make(map[string]*parser.Model)

		for _, m := range parsedTree.ModelTree.Models {
			if len(m.LambdaOnMutate) > 0 {
				models[m.Name] = m
				//pkgs[m.TypeName.Pkg().Name()] = m.TypeName.Pkg()
			}
		}

		if len(models) > 0 {
			pkgs["context"] = types.NewPackage("context", "context")
			pkgs["api"] = types.NewPackage("github.com/schartey/dgraph-lambda-go/api", "api")
		}

		err = webhookResolverTemplate.Execute(f, struct {
			Models      map[string]*parser.Model
			Rewriter    *rewriter.Rewriter
			Packages    map[string]*types.Package
			PackageName string
		}{
			Models:      models,
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

var webhookResolverTemplate = template.Must(template.New("webhook-resolver").Funcs(template.FuncMap{
	"path":     pkgPath,
	"body":     middlewareBody,
	"typeName": typeName,
	"is":       is,
}).Parse(`
package {{.PackageName}}

import(
	{{- range $pkg := .Packages }}
	"{{ $pkg | path }}"{{- end}}
)

type WebhookResolverInterface interface {
{{- range $model := .Models}}
	Webhook_{{ $model.TypeName | typeName }}(ctx context.Context, event api.Event) *api.LambdaError{{ end }}
}

type WebhookResolver struct {
	*Resolver
}

{{ range $model := .Models}}
func (w *WebhookResolver) Webhook_{{ $model.TypeName | typeName }}(ctx context.Context, event api.Event) *api.LambdaError { {{ body (printf "Webhook_%s" ($model.TypeName | typeName)) $.Rewriter }}}
{{ end }}
`))
