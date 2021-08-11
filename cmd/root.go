package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

// TODO: Add model generation from graphql introspection
func Execute() {
	app := cli.NewApp()
	app.Name = "dgraph-lambda-go"
	app.Usage = initCmd.Usage
	app.Description = "This project implements the dgraph-lambda server based on go."
	app.HideVersion = true
	app.Flags = initCmd.Flags
	app.Version = "0.5.0"
	app.Before = func(context *cli.Context) error {
		if context.Bool("verbose") {
			log.SetFlags(0)
		} else {
			log.SetOutput(ioutil.Discard)
		}
		return nil
	}

	app.Action = initCmd.Action
	app.Commands = []*cli.Command{
		initCmd,
		runCmd,
		generateCmd,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
}
