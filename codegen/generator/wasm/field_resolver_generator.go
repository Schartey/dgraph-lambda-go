package wasm

import (
	"errors"
	"go/types"
	"os"
	"path"

	"github.com/schartey/dgraph-lambda-go/codegen/generator/wasm/templates"
	"github.com/schartey/dgraph-lambda-go/codegen/parser"
	"github.com/schartey/dgraph-lambda-go/codegen/rewriter"
	"github.com/schartey/dgraph-lambda-go/config"
)

func generateFieldResolvers(c *config.Config, parsedTree *parser.Tree, r *rewriter.Rewriter) error {

	if c.ResolverFilename == "resolver" {
		fileName := path.Join(c.ConfigFile.DGraph.Resolver.Dir, "field.resolver.go")
		f, err := os.Create(fileName)
		if err != nil {
			return err
		}
		defer f.Close()

		var pkgs = make(map[string]*types.Package)

		for _, m := range parsedTree.ResolverTree.FieldResolvers {
			if m.Field.TypeName.Exported() {
				pkgs[m.Field.TypeName.Pkg().Name()] = m.Field.TypeName.Pkg()
			}
			if m.Parent.TypeName.Exported() {
				pkgs[m.Parent.TypeName.Pkg().Name()] = m.Parent.TypeName.Pkg()
			}
		}
		if len(parsedTree.ResolverTree.FieldResolvers) > 0 {
			pkgs["wasm"] = types.NewPackage("github.com/schartey/dgraph-lambda-go/wasm", "wasm")
		}

		err = templates.GolangFieldResolverTemplate.Execute(f, struct {
			FieldResolvers map[string]*parser.FieldResolver
			Rewriter       *rewriter.Rewriter
			Packages       map[string]*types.Package
			PackageName    string
		}{
			FieldResolvers: parsedTree.ResolverTree.FieldResolvers,
			Rewriter:       r,
			Packages:       pkgs,
			PackageName:    c.ConfigFile.DGraph.Resolver.Package,
		})
		if err != nil {
			return err
		}
	} else {
		return errors.New("resolver file template not supported")
	}
	return nil
}
