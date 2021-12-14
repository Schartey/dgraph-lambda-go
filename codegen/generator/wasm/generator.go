package wasm

import (
	"path"

	"github.com/pkg/errors"
	"github.com/schartey/dgraph-lambda-go/codegen/autobind"
	"github.com/schartey/dgraph-lambda-go/codegen/generator"
	"github.com/schartey/dgraph-lambda-go/codegen/generator/tools"
	"github.com/schartey/dgraph-lambda-go/codegen/parser"
	"github.com/schartey/dgraph-lambda-go/config"
	"github.com/schartey/dgraph-lambda-go/internal"
)

type WasmGenerator struct {
	config   *config.Config
	packages *internal.Packages
	//parsedTree    *parser.Tree
	//modelPackages map[string]*types.Package
	//rewriter      *rewriter.Rewriter
}

func NewGenerator(c *config.Config) generator.Generator {
	return &WasmGenerator{config: c, packages: &internal.Packages{}}
}

func (g *WasmGenerator) Generate() error {

	if err := generateWorkspace(g.config, g.packages); err != nil {
		return err
	}

	parser := parser.NewParser(g.config.Schema, g.packages, g.config.ConfigFile.DGraph.Model.Force)
	parsedTree, err := parser.Parse()
	if err != nil {
		return err
	}

	packagePath := path.Join(g.config.ConfigFile.DGraph.Model.Filename)
	defaultModelPackage, err := g.packages.Load(packagePath)
	if err != nil {
		return err
	}

	autobinder := autobind.New(g.packages, defaultModelPackage, g.config.ConfigFile.DGraph.Model.Filename)
	err = autobinder.Bind(g.config.ConfigFile.DGraph.Model.AutoBind, parsedTree)
	if err != nil {
		return err
	}

	modelTree, pkgs := tools.GetDefaultPackageTree(defaultModelPackage.PkgPath, parsedTree)

	if err := generateModel(g.config, modelTree, pkgs); err != nil {
		return errors.Wrap(err, "Could not generate model")
	}

	/*if err := generateExecuter(g.config, g.parsedTree, g.pkgs); err != nil {
		return errors.Wrap(err, "Could not generate model")
	}*/

	return nil
}
