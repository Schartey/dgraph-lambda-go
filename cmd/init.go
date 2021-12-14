package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/schartey/dgraph-lambda-go/codegen/generator"
	"github.com/schartey/dgraph-lambda-go/codegen/generator/gogen"
	"github.com/schartey/dgraph-lambda-go/codegen/generator/wasm"
	"github.com/schartey/dgraph-lambda-go/config"
	"github.com/urfave/cli/v2"
)

type Template string

const (
	WASM_ONLY     Template = "wasm-only"
	WASM_Server   Template = "wasm-server"
	NATIVE_SERVER Template = "native-server"
)

// Questions
var qs = []*survey.Question{
	{
		Name: "template",
		Prompt: &survey.Select{
			Message: "Select a template:",
			Options: []string{string(WASM_ONLY), string(WASM_Server), string(NATIVE_SERVER)},
			Default: "wasm-only",
		},
	},
	{
		Name: "schema",
		Prompt: &survey.Multiline{
			Message: "Enter Schema Path or Host",
		},
	},
	{
		Name: "models",
		Prompt: &survey.Confirm{
			Message: "Do you have any pre-existing models?",
		},
	},
}

/*
Run generate aferwards.
*/
var initCmd = &cli.Command{
	Name:        "init",
	Usage:       "init -c \"lambda.yaml\"",
	Description: "generates the config.",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "config", Aliases: []string{"c"}, Usage: "optional path to lambda config"},
	},
	Action: func(ctx *cli.Context) error {
		configFilePath := ctx.String("config")

		if configFilePath == "" {
			configFilePath = "lambda.yaml"
		}

		if t, err := os.Open(configFilePath); err == nil {
			t.Close()
			overwrite := false
			survey.AskOne(&survey.Confirm{
				Message: "Config already exists. Overwrite?",
			}, &overwrite)
			if !overwrite {
				fmt.Println("Cancelled initialization!")
				return nil
			}
		}

		answers := struct {
			Template string
			Schema   string
			Models   bool
		}{}

		err := survey.Ask(qs, &answers)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		configFile := config.DefaultConfigFile

		if answers.Template == string(WASM_ONLY) {
			configFile.DGraph.Generator = config.WASM
			configFile.Lambda.Generate = false
		} else if answers.Template == string(WASM_Server) {
			configFile.DGraph.Generator = config.WASM
			configFile.Lambda.Generate = true
		} else if answers.Template == string(NATIVE_SERVER) {
			configFile.DGraph.Generator = config.NATIVE
			configFile.Lambda.Generate = true
		}

		if schemaFileNames := strings.Split(answers.Schema, "\n"); len(schemaFileNames) > 0 && schemaFileNames[0] != "" {
			configFile.DGraph.SchemaFileName = schemaFileNames
		}

		if answers.Models {
			models := ""
			prompt := &survey.Multiline{
				Message: "List your models",
			}
			survey.AskOne(prompt, &models)

			configFile.DGraph.Model.AutoBind = strings.Split(models, "\n")
		}

		err = configFile.Generate(configFilePath)
		if err != nil {
			return err
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

		fmt.Println(c)
		return nil
	},
}
