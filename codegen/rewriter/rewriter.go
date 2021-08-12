package rewriter

import (
	"fmt"
	"go/ast"
	"go/token"
	"io/ioutil"
	"path"
	"strings"

	"github.com/schartey/dgraph-lambda-go/codegen/config"
	"github.com/schartey/dgraph-lambda-go/internal"
	"golang.org/x/tools/go/packages"
)

type Rewriter struct {
	config           *config.Config
	files            map[string]string
	RewriteBodies    map[string]string
	DeprecatedBodies map[string]string
}

func New(config *config.Config) *Rewriter {
	files := make(map[string]string)
	rewriteBodies := make(map[string]string)
	deprecatedBodies := make(map[string]string)
	return &Rewriter{config: config, files: files, RewriteBodies: rewriteBodies, DeprecatedBodies: deprecatedBodies}
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

					for _, query := range r.config.ParsedTree.ResolverTree.Queries {
						if query.Name == queryName {
							r.RewriteBodies[d.Name.Name] = r.getSource(pkg, d.Body.Pos()+1, d.Body.End()-1)
							found = true
							break
						}
					}
				}

				if strings.HasPrefix(d.Name.Name, "Mutation_") {
					mutationName := strings.TrimPrefix(d.Name.Name, "Mutation_")

					for _, mutation := range r.config.ParsedTree.ResolverTree.Mutations {
						if mutation.Name == mutationName {
							r.RewriteBodies[d.Name.Name] = r.getSource(pkg, d.Body.Pos()+1, d.Body.End()-1)
							found = true
							break
						}
					}
				}

				if strings.HasPrefix(d.Name.Name, "Middleware_") {
					middlewareName := strings.TrimPrefix(d.Name.Name, "Middleware_")

					for _, middleware := range r.config.ParsedTree.Middleware {
						if middleware == middlewareName {
							r.RewriteBodies[d.Name.Name] = r.getSource(pkg, d.Body.Pos()+1, d.Body.End()-1)
							found = true
							break
						}
					}
				}

				if strings.HasPrefix(d.Name.Name, "Webhook_") {
					webhookName := strings.TrimPrefix(d.Name.Name, "Webhook_")

					for _, model := range r.config.ParsedTree.ModelTree.Models {
						if model.TypeName.Name() == webhookName {
							r.RewriteBodies[d.Name.Name] = r.getSource(pkg, d.Body.Pos()+1, d.Body.End()-1)
							found = true
							break
						}
					}
				}

				for _, fieldResolver := range r.config.ParsedTree.ResolverTree.FieldResolvers {
					splitName := strings.Split(d.Name.Name, "_")

					if splitName[0] == fieldResolver.Field.ParentTypeName && splitName[1] == fieldResolver.Field.Name {
						r.RewriteBodies[d.Name.Name] = r.getSource(pkg, d.Body.Pos()+1, d.Body.End()-1)
						found = true
						break
					}
				}

				if !found {
					r.DeprecatedBodies[d.Name.Name] = r.getSource(pkg, d.Pos(), d.End())
				}
			}
		}
	}
	return nil
}

func (r *Rewriter) getSource(pkg *packages.Package, start, end token.Pos) string {
	startPos := pkg.Fset.Position(start)
	endPos := pkg.Fset.Position(end)

	if startPos.Filename != endPos.Filename {
		panic("cant get source spanning multiple files")
	}

	file := r.getFile(startPos.Filename)
	return file[startPos.Offset:endPos.Offset]
}

func (r *Rewriter) getFile(filename string) string {
	if _, ok := r.files[filename]; !ok {
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			panic(fmt.Errorf("unable to load file, already exists: %s", err.Error()))
		}

		r.files[filename] = string(b)

	}

	return r.files[filename]
}
