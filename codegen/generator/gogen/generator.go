package gogen

import (
	"go/types"

	"github.com/schartey/dgraph-lambda-go/codegen/generator"
	"github.com/schartey/dgraph-lambda-go/codegen/parser"
	"github.com/schartey/dgraph-lambda-go/codegen/rewriter"
	"github.com/schartey/dgraph-lambda-go/config"
	"github.com/schartey/dgraph-lambda-go/internal"
)

type GoGenerator struct {
	config        *config.Config
	packages      *internal.Packages
	parsedTree    *parser.Tree
	modelPackages map[string]*types.Package
	rewriter      *rewriter.Rewriter
}

func NewGenerator(c *config.Config) generator.Generator {
	return &GoGenerator{config: c, packages: &internal.Packages{}}
}

func (g *GoGenerator) Generate() error {

	/*	if err := generateModel(g.config, g.parsedTree, g.pkgs); err != nil {
			return errors.Wrap(err, "Could not generate model")
		}
		if err := generateFieldResolvers(g.config, g.parsedTree, g.rewriter); err != nil {
			return errors.Wrap(err, "Could not generate field resolvers")
		}
		if err := generateQueryResolvers(g.config, g.parsedTree, g.rewriter); err != nil {
			return errors.Wrap(err, "Could not generate query resolvers")
		}
		if err := generateMutationResolvers(g.config, g.parsedTree, g.rewriter); err != nil {
			return errors.Wrap(err, "Could not generate mutation resolvers")
		}
		if err := generateMiddleware(g.config, g.parsedTree, g.rewriter); err != nil {
			return errors.Wrap(err, "Could not generate middleware resolvers")
		}
		if err := generateWebhook(g.config, g.parsedTree, g.rewriter); err != nil {
			return errors.Wrap(err, "Could not generate webhook resolvers")
		}
		if err := generateExecuter(g.config, g.parsedTree, g.rewriter); err != nil {
			return errors.Wrap(err, "Could not generate executer")
		}*/
	return nil
}
