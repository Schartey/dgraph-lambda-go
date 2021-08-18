package cmd

import (
	"github.com/schartey/dgraph-lambda-go/codegen/config"
	"github.com/schartey/dgraph-lambda-go/codegen/generator"
	"github.com/schartey/dgraph-lambda-go/internal"
	"github.com/urfave/cli/v2"
)

var initCmd = &cli.Command{
	Name:        "init",
	Usage:       "init -c \"lambda.yaml\"",
	Description: "generates folder structure and lambda-server. Call generate command afterwards",
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

		if err := generator.GenerateConfig(configFile); err != nil {
			return err
		}

		config, err := config.LoadConfig(moduleName, configFile)
		if err != nil {
			return err
		}
		config.LoadSchema()
		if err != nil {
			return err
		}

		return generator.Init(config)
	},
}
