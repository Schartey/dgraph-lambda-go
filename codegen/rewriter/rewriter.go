package rewriter

import (
	"go/ast"
	"path"
	"strings"

	"github.com/miko/dgraph-lambda-go/codegen/config"
	"github.com/miko/dgraph-lambda-go/codegen/parser"
	"github.com/miko/dgraph-lambda-go/internal"
)

type Rewriter struct {
	config           *config.Config
	parsedTree       *parser.Tree
	RewriteBodies    map[string]string
	DeprecatedBodies map[string]string
}

func New(config *config.Config, parsedTree *parser.Tree) *Rewriter {
	rewriteBodies := make(map[string]string)
	deprecatedBodies := make(map[string]string)
	return &Rewriter{config: config, parsedTree: parsedTree, RewriteBodies: rewriteBodies, DeprecatedBodies: deprecatedBodies}
}

func (r *Rewriter) Load() error {
	r.RewriteBodies = make(map[string]string)
	r.DeprecatedBodies = make(map[string]string)

	pkgs := &internal.Packages{}
	// field resolvers
	if r.config.ResolverFilename == "resolver" {
		pkg, err := pkgs.Load(path.Join(r.config.Root, r.config.Resolver.Dir))
		if err != nil {
			return err
		}
		for _, f := range pkg.Syntax {
			for _, d := range f.Decls {
				found := false

				d, isFunc := d.(*ast.FuncDecl)
				if !isFunc {
					continue
				}

				if strings.HasPrefix(d.Name.Name, "Query_") {
					queryName := strings.TrimPrefix(d.Name.Name, "Query_")

					for _, query := range r.parsedTree.ResolverTree.Queries {
						if query.Name == queryName {
							_, r.RewriteBodies[d.Name.Name] = r.config.Packages.GetSource(pkg, d.Body.Pos()+1, d.Body.End()-1)
							found = true
							break
						}
					}
				}

				if strings.HasPrefix(d.Name.Name, "Mutation_") {
					mutationName := strings.TrimPrefix(d.Name.Name, "Mutation_")

					for _, mutation := range r.parsedTree.ResolverTree.Mutations {
						if mutation.Name == mutationName {
							_, r.RewriteBodies[d.Name.Name] = r.config.Packages.GetSource(pkg, d.Body.Pos()+1, d.Body.End()-1)
							found = true
							break
						}
					}
				}

				if strings.HasPrefix(d.Name.Name, "Middleware_") {
					middlewareName := strings.TrimPrefix(d.Name.Name, "Middleware_")

					for _, middleware := range r.parsedTree.Middleware {
						if middleware == middlewareName {
							_, r.RewriteBodies[d.Name.Name] = r.config.Packages.GetSource(pkg, d.Body.Pos()+1, d.Body.End()-1)
							found = true
							break
						}
					}
				}

				if strings.HasPrefix(d.Name.Name, "Webhook_") {
					webhookName := strings.TrimPrefix(d.Name.Name, "Webhook_")

					for _, model := range r.parsedTree.ModelTree.Models {
						if model.TypeName.Name() == webhookName {
							_, r.RewriteBodies[d.Name.Name] = r.config.Packages.GetSource(pkg, d.Body.Pos()+1, d.Body.End()-1)
							found = true
							break
						}
					}
				}

				for _, fieldResolver := range r.parsedTree.ResolverTree.FieldResolvers {
					splitName := strings.Split(d.Name.Name, "_")

					if splitName[0] == fieldResolver.Parent.Name && splitName[1] == fieldResolver.Field.Name {
						_, r.RewriteBodies[d.Name.Name] = r.config.Packages.GetSource(pkg, d.Body.Pos()+1, d.Body.End()-1)
						found = true
						break
					}
				}

				if !found {
					_, r.DeprecatedBodies[d.Name.Name] = r.config.Packages.GetSource(pkg, d.Pos(), d.End())
				}
			}
		}
	}
	return nil
}
