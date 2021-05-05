package cmd

import (
	"github.com/schartey/dgraph-lambda-go/examples"
	"github.com/urfave/cli/v2"
)

var runCmd = &cli.Command{
	Name:  "run",
	Usage: "run lambda server",
	Action: func(ctx *cli.Context) error {
		examples.Run()
		return nil
	},
}
