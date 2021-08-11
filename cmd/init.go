package cmd

import (
	"github.com/schartey/dgraph-lambda-go/codegen/config"
	"github.com/schartey/dgraph-lambda-go/codegen/generator"
	"github.com/urfave/cli/v2"
)

var initCmd = &cli.Command{
	Name:  "init",
	Usage: "generate a basic server",
	Action: func(ctx *cli.Context) error {
		generator.GenerateConfig()

		config, err := config.LoadConfig("lambda.yaml")
		if err != nil {
			return err
		}
		config.Init()
		if err != nil {
			return err
		}

		return generator.Init(config)
	},
}
