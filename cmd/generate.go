package cmd

import (
	"github.com/schartey/dgraph-lambda-go/codegen/config"
	"github.com/schartey/dgraph-lambda-go/codegen/generator"
	"github.com/schartey/dgraph-lambda-go/codegen/rewriter"
	"github.com/urfave/cli/v2"
)

var generateCmd = &cli.Command{
	Name:        "generate",
	Usage:       "generate",
	Description: "generates types, resolvers and middleware from schema in lambda.yaml",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "schema, s", Usage: "the schema filename"},
	},
	Action: func(ctx *cli.Context) error {
		config, err := config.LoadConfig("lambda.yaml")
		if err != nil {
			return err
		}

		if err := config.LoadSchema(); err != nil {
			return err
		}

		rewriter := rewriter.New(config)

		if err := rewriter.Load(); err != nil {
			return err
		}

		if err := generator.Generate(config, rewriter); err != nil {
			return err
		}

		return nil
	},
}
