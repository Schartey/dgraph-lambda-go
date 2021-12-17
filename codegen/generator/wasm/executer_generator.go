package wasm

import (
	"go/types"
	"os"
	"path"

	"github.com/schartey/dgraph-lambda-go/codegen/generator/wasm/templates"
	"github.com/schartey/dgraph-lambda-go/codegen/parser"
	"github.com/schartey/dgraph-lambda-go/codegen/rewriter"
	"github.com/schartey/dgraph-lambda-go/config"
)

func generateExecutor(c *config.Config, parsedTree *parser.Tree, r *rewriter.Rewriter) error {

	fileName := path.Join(path.Dir(c.ConfigFile.DGraph.Resolver.Dir), "executor.go")
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
		pkgs["api"] = types.NewPackage("github.com/schartey/dgraph-lambda-go/api", "api")
	}

	err = templates.GolangExecuterTemplate.Execute(f, struct {
		FieldResolvers map[string]*parser.FieldResolver
		Rewriter       *rewriter.Rewriter
		Packages       map[string]*types.Package
	}{
		FieldResolvers: parsedTree.ResolverTree.FieldResolvers,
		Rewriter:       r,
		Packages:       pkgs,
	})
	if err != nil {
		return err
	}
	return nil
}

/*func generateExecuter(c *config.Config, parsedTree *parser.Tree, pkgs map[string]*types.Package) error {

	f, err := os.Create(c.Exec.Filename)
	if err != nil {
		return err
	}
	defer f.Close()

	var t *template.Template
	if c.Wasm.Lang == config.GOLANG {
		t = templates.GolangExecuterTemplate
	}

	err = t.Execute(f, struct {
		Interfaces  map[string]*parser.Interface
		Enums       map[string]*parser.Enum
		Scalars     map[string]*parser.Scalar
		Models      map[string]*parser.Model
		Packages    map[string]*types.Package
		PackageName string
	}{
		Interfaces:  parsedTree.ModelTree.Interfaces,
		Enums:       parsedTree.ModelTree.Enums,
		Scalars:     parsedTree.ModelTree.Scalars,
		Models:      parsedTree.ModelTree.Models,
		Packages:    pkgs,
		PackageName: c.Model.Package,
	})
	if err != nil {
		return err
	}
	return nil
}*/
