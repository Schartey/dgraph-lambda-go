package cmd

import (
	"github.com/schartey/dgraph-lambda-go/codegen"
	"github.com/urfave/cli/v2"
)

var initCmd = &cli.Command{
	Name:  "init",
	Usage: "generate a basic server",
	Action: func(ctx *cli.Context) error {

		return codegen.GenerateServer()
	},
}
