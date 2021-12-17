package wasm

import (
	"go/types"
	"os"
	"text/template"

	"github.com/schartey/dgraph-lambda-go/codegen/generator/wasm/templates"
	"github.com/schartey/dgraph-lambda-go/codegen/parser"
	"github.com/schartey/dgraph-lambda-go/config"
)

func generateModel(c *config.Config, parsedTree *parser.Tree, pkgs map[string]*types.Package) error {

	f, err := os.Create(c.ConfigFile.DGraph.Model.Filename)
	if err != nil {
		return err
	}
	defer f.Close()

	var t *template.Template
	if c.ConfigFile.Wasm.Language == config.GOLANG {
		pkgs["fmt"] = types.NewPackage("fmt", "fmt")
		pkgs["fastjson"] = types.NewPackage("github.com/valyala/fastjson", "fastjson")
		pkgs["wasm"] = types.NewPackage("github.com/schartey/dgraph-lambda-go/wasm", "wasm")

		t = templates.GolangModelTemplate
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
		PackageName: c.ConfigFile.DGraph.Model.Package,
	})
	if err != nil {
		return err
	}
	return nil
}
