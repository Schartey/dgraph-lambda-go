package cmd

import (
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

		/*moduleName, err := internal.GetModuleName()
		if err != nil {
			return err
		}

		c, err := config.LoadConfigFile(moduleName, configFile)
		if err != nil {
			return err
		}
		err = c.LoadConfig()
		if err != nil {
			return err
		}

		if err := c.LoadSchema(); err != nil {
			return err
		}

		parser := parser.NewParser(c.Schema, c.Packages, c.Force)
		parsedTree, err := parser.Parse()
		if err != nil {
			return err
		}

		if err := c.Bind(parsedTree); err != nil {
			return err
		}

		rewriter := rewriter.New(c, parsedTree)

		if err := rewriter.Load(); err != nil {
			return err
		}

		parsedTree, pkgs := generator.GetDefaultPackageTree(c.DefaultModelPackage.PkgPath, parsedTree)

		var gen generator.Generator

		// If no language is selected, we generate a pure golang server
		if c.Wasm.Lang == "" {
			gen = gogen.NewGenerator(c, parsedTree, pkgs, rewriter)
		} else {
			gen = wasm.NewGenerator(c, parsedTree, pkgs, rewriter)
		}
		if err := gen.Generate(); err != nil {
			return err
		}

		// Run go mod tidy
		if err := internal.FixImports(); err != nil {
			return err
		}*/

		return nil
	},
}
