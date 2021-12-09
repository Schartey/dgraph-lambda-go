package wasm

import (
	"github.com/schartey/dgraph-lambda-go/codegen/config"
	"github.com/schartey/dgraph-lambda-go/codegen/generator"
	"github.com/schartey/dgraph-lambda-go/codegen/parser"
	"github.com/schartey/dgraph-lambda-go/codegen/rewriter"
)

type WasmGenerator struct {
	config   *config.Config
	parser   *parser.Tree
	rewriter *rewriter.Rewriter
}

func NewGenerator(c *config.Config, p *parser.Tree, r *rewriter.Rewriter) generator.Generator {
	return &WasmGenerator{config: c, parser: p, rewriter: r}
}

func (g *WasmGenerator) Generate() error {
	return nil
}
