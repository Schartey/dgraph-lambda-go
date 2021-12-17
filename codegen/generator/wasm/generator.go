package wasm

import (
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/pkg/errors"
	"github.com/schartey/dgraph-lambda-go/codegen/autobind"
	"github.com/schartey/dgraph-lambda-go/codegen/generator"
	"github.com/schartey/dgraph-lambda-go/codegen/generator/tools"
	"github.com/schartey/dgraph-lambda-go/codegen/parser"
	"github.com/schartey/dgraph-lambda-go/codegen/rewriter"
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

	packagePath := path.Dir(path.Join(g.config.Root, g.config.ConfigFile.DGraph.Model.Filename))
	defaultModelPackage, err := g.packages.Load(packagePath)
	if err != nil {
		return err
	}

	autobinder := autobind.New(g.packages, defaultModelPackage, g.config.ConfigFile.DGraph.Model.Filename)
	err = autobinder.Bind(g.config.ConfigFile.DGraph.Model.AutoBind, parsedTree)
	if err != nil {
		return err
	}

	rewriter := rewriter.New(g.config, g.packages, parsedTree)
	err = rewriter.Load()
	if err != nil {
		return err
	}

	modelTree, pkgs := tools.GetDefaultPackageTree(defaultModelPackage.PkgPath, parsedTree)

	if err := generateModel(g.config, modelTree, pkgs); err != nil {
		return errors.Wrap(err, "Could not generate model")
	}

	if err := generateFieldResolvers(g.config, parsedTree, rewriter); err != nil {
		return errors.Wrap(err, "Could not generate field resolvers")
	}

	if err := generateExecutor(g.config, parsedTree, rewriter); err != nil {
		return errors.Wrap(err, "Could not generate model")
	}

	return nil
}

func generateWorkspace(c *config.Config, packages *internal.Packages) error {
	modelPath := c.ConfigFile.DGraph.Model.Filename

	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		f, err := internal.CreateFile(modelPath)
		if err != nil {
			return err
		}
		template.Must(template.New("model").Parse("package "+c.ConfigFile.DGraph.Model.Package)).Execute(f, struct{}{})
		f.Close()
	}

	resolverPath := path.Join(c.ConfigFile.DGraph.Resolver.Dir, "resolver.go")

	if _, err := os.Stat(resolverPath); os.IsNotExist(err) {
		f, err := internal.CreateFile(resolverPath)
		if err != nil {
			return err
		}
		template.Must(template.New("resolver").Parse(fmt.Sprintf(`package %s

// Add objects to your desire
type Resolver struct {
}`, c.ConfigFile.DGraph.Resolver.Package))).Execute(f, struct{}{})
		f.Close()
	}

	return nil
}
