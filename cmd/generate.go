package cmd

import (
	"fmt"

	"github.com/schartey/dgraph-lambda-go/codegen/config"
	"github.com/schartey/dgraph-lambda-go/codegen/generator"
	"github.com/schartey/dgraph-lambda-go/codegen/rewriter"
	"github.com/urfave/cli/v2"
)

var generateCmd = &cli.Command{
	Name:  "generate",
	Usage: "generate resolvers and types from schema",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "schema, s", Usage: "the schema filename"},
	},
	Action: func(ctx *cli.Context) error {
		config, err := config.LoadConfig("lambda.yaml")
		if err != nil {
			return err
		}

		if config.Init(); err != nil {
			fmt.Println(err.Error())
		}

		rewriter := rewriter.New(config)
		rewriter.Load()

		err = generator.Generate(config, rewriter)
		if err != nil {
			return err
		}

		/*schemaFile := ctx.String("schema")

		fmt.Println(schemaFile)

		schema, err := graphql.SchemaLoaderFromFile(schemaFile)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		generator := modelgen.NewGenerator()

		generator.Parse(schema)*/
		return nil
	},
}
