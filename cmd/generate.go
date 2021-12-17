package cmd

import (
	"github.com/schartey/dgraph-lambda-go/codegen/generator"
	"github.com/schartey/dgraph-lambda-go/codegen/generator/gogen"
	"github.com/schartey/dgraph-lambda-go/codegen/generator/wasm"
	"github.com/schartey/dgraph-lambda-go/config"
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
		configFilePath := ctx.String("config")

		if configFilePath == "" {
			configFilePath = "lambda.yaml"
		}

		c, err := config.LoadConfig(configFilePath)
		if err != nil {
			return err
		}

		var gen generator.Generator

		// If no language is selected, we generate a pure golang server
		if c.ConfigFile.DGraph.Generator == config.NATIVE {
			gen = gogen.NewGenerator(c)
		} else {
			gen = wasm.NewGenerator(c)
		}
		if err := gen.Generate(); err != nil {
			return err
		}

		if err := internal.FixImports(); err != nil {
			return err
		}

		return nil
	},
}
