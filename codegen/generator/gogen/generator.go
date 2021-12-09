package gogen

import (
	"github.com/pkg/errors"
	"github.com/schartey/dgraph-lambda-go/codegen/config"
	"github.com/schartey/dgraph-lambda-go/codegen/generator"
	"github.com/schartey/dgraph-lambda-go/codegen/parser"
	"github.com/schartey/dgraph-lambda-go/codegen/rewriter"
)

type GoGenerator struct {
	config   *config.Config
	parser   *parser.Tree
	rewriter *rewriter.Rewriter
}

func NewGenerator(c *config.Config, p *parser.Tree, r *rewriter.Rewriter) generator.Generator {
	return &GoGenerator{config: c, parser: p, rewriter: r}
}

func (g *GoGenerator) Generate() error {

	if err := generateModel(g.config, g.parser); err != nil {
		return errors.Wrap(err, "Could not generate model")
	}
	if err := generateFieldResolvers(g.config, g.parser, g.rewriter); err != nil {
		return errors.Wrap(err, "Could not generate field resolvers")
	}
	if err := generateQueryResolvers(g.config, g.parser, g.rewriter); err != nil {
		return errors.Wrap(err, "Could not generate query resolvers")
	}
	if err := generateMutationResolvers(g.config, g.parser, g.rewriter); err != nil {
		return errors.Wrap(err, "Could not generate mutation resolvers")
	}
	if err := generateMiddleware(g.config, g.parser, g.rewriter); err != nil {
		return errors.Wrap(err, "Could not generate middleware resolvers")
	}
	if err := generateWebhook(g.config, g.parser, g.rewriter); err != nil {
		return errors.Wrap(err, "Could not generate webhook resolvers")
	}
	if err := generateExecuter(g.config, g.parser, g.rewriter); err != nil {
		return errors.Wrap(err, "Could not generate executer")
	}
	return nil
}
