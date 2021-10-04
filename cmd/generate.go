package cmd

import (
	"github.com/schartey/dgraph-lambda-go/codegen/config"
	"github.com/schartey/dgraph-lambda-go/codegen/generator"
	"github.com/schartey/dgraph-lambda-go/codegen/parser"
	"github.com/schartey/dgraph-lambda-go/codegen/rewriter"
	"github.com/schartey/dgraph-lambda-go/internal"
	"github.com/urfave/cli/v2"
)

var generateCmd = &cli.Command{
	Name:        "generate",
	Usage:       "generate -c \"lambda.yaml\"",
	Description: "generates types, resolvers and middleware from schema in lambda.yaml",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "config", Aliases: []string{"c"}, Usage: "the lambda config file"},
	},
	Action: func(ctx *cli.Context) error {
		configFile := ctx.String("config")

		if configFile == "" {
			configFile = "lambda.yaml"
		}

		moduleName, err := internal.GetModuleName()
		if err != nil {
			return err
		}

		config, err := config.LoadConfigFile(moduleName, configFile)
		if err != nil {
			return err
		}
		err = config.LoadConfig(configFile)
		if err != nil {
			return err
		}

		if err := config.LoadSchema(); err != nil {
			return err
		}

		parser := parser.NewParser(config.Schema, config.Packages)
		parsedTree, err := parser.Parse()
		if err != nil {
			return err
		}

		if err := config.Bind(parsedTree); err != nil {
			return err
		}

		rewriter := rewriter.New(config, parsedTree)

		if err := rewriter.Load(); err != nil {
			return err
		}

		if err := generator.Generate(config, parsedTree, rewriter); err != nil {
			return err
		}

		// Run go mod tidy
		if err := internal.Tidy(); err != nil {
			return err
		}

		return nil
	},
}
