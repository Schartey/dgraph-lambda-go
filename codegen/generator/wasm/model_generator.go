package wasm

import (
	"go/types"
	"os"
	"text/template"

	"github.com/schartey/dgraph-lambda-go/codegen/generator/wasm/templates"
	"github.com/schartey/dgraph-lambda-go/codegen/parser"
	"github.com/schartey/dgraph-lambda-go/config"
	"github.com/schartey/dgraph-lambda-go/internal"
)

func generateWorkspace(c *config.Config, packages *internal.Packages) error {
	modelPath := c.ConfigFile.DGraph.Model.Filename

	if t, err := os.Open(modelPath); os.IsNotExist(err) {
		f, err := internal.CreateFile(modelPath)
		if err != nil {
			return err
		}
		template.Must(template.New("model").Parse("package "+c.ConfigFile.DGraph.Model.Package)).Execute(f, struct{}{})
		f.Close()
	} else {
		t.Close()
	}

	if _, err := packages.Load(modelPath); err != nil {
		return err
	}
	return nil
}

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
