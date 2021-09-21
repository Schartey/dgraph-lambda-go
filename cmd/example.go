package cmd

import (
	"github.com/schartey/dgraph-lambda-go/examples"
	"github.com/urfave/cli/v2"
)

var exampleCmd = &cli.Command{
	Name:        "example",
	Usage:       "example",
	Description: "Runs example server",
	Action: func(ctx *cli.Context) error {
		examples.RunWithServer()
		return nil
	},
}
