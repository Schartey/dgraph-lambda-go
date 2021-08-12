package cmd

import (
	"github.com/schartey/dgraph-lambda-go/codegen/config"
	"github.com/schartey/dgraph-lambda-go/codegen/generator"
	"github.com/urfave/cli/v2"
)

var initCmd = &cli.Command{
	Name:        "init",
	Usage:       "init",
	Description: "generates folder structure and lambda-server. Call generate command afterwards",
	Action: func(ctx *cli.Context) error {
		generator.GenerateConfig()

		config, err := config.LoadConfig("lambda.yaml")
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
