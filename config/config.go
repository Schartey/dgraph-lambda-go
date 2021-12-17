package config

import (
	"io/ioutil"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"
	"github.com/schartey/dgraph-lambda-go/codegen/graphql"
	"github.com/schartey/dgraph-lambda-go/internal"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"gopkg.in/yaml.v2"
)

var resolverTemplateRegex = regexp.MustCompile(`\{([^)]+)\}.resolver.*`)

type ResolverTemplate string

const (
	RESOLVER ResolverTemplate = "resolver"
)

type Generator string

const (
	WASM   Generator = "wasm"
	NATIVE Generator = "native"
)

type Language string

const (
	GOLANG         Language = "go"
	RUST           Language = "rust"
	ASSEMBLYSCRIPT Language = "asc"
)

type Model struct {
	Filename string   `yaml:"filename"`
	Package  string   `yaml:"package"`
	AutoBind []string `yaml:"autobind"`
	Force    []string `yaml:"force"`
}
type Resolver struct {
	Executer         string `yaml:"executer"`
	Dir              string `yaml:"dir"`
	Package          string `yaml:"package"`
	FilenameTemplate string `yaml:"filename_template"`
}
type DGraph struct {
	Generator      Generator `yaml:"generator"`
	SchemaFileName []string  `yaml:"schema"`
	Model          Model     `yaml:"model"`
	Resolver       Resolver  `yaml:"resolver"`
}

type Wasm struct {
	Dir      string   `yaml:"dir"`
	Language Language `yaml:"language"`
}

type Native struct {
	Dir string `yaml:"dir"`
}

type Lambda struct {
	Generate bool `yaml:"generate"`
	Router   bool `yaml:"router"`
}

type ConfigFile struct {
	DGraph *DGraph `yaml:"dgraph"`
	Wasm   *Wasm   `yaml:"wasm"`
	Native *Native `yaml:"native"`
	Lambda *Lambda `yaml:"lambda"`
}

type Config struct {
	ConfigFile       *ConfigFile
	ConfigPath       string
	Schema           *ast.Schema
	Root             string
	ResolverFilename ResolverTemplate
}

var DefaultConfigFile = &ConfigFile{
	DGraph: &DGraph{
		Generator:      WASM,
		SchemaFileName: []string{"./*.graphql"},
		Model: Model{
			Filename: "generated/model/models_gen.go",
			Package:  "model",
			AutoBind: []string{},
			Force:    []string{},
		},
		Resolver: Resolver{
			Executer:         "generated/executer.go",
			Dir:              "generated/resolvers",
			Package:          "resolvers",
			FilenameTemplate: "{resolver}.resolver.go",
		},
	},
	Wasm: &Wasm{
		Dir:      "wasm",
		Language: GOLANG,
	},
	Lambda: &Lambda{
		Generate: true,
		Router:   true,
	},
}

func LoadConfig(filename string) (*Config, error) {
	configFile := &ConfigFile{}

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read config")
	}

	if err := yaml.UnmarshalStrict(b, configFile); err != nil {
		return nil, errors.Wrap(err, "unable to parse config")
	}

	if err = validateConfigFile(configFile); err != nil {
		return nil, err
	}

	// Append default autobinds
	configFile.DGraph.Model.AutoBind = append(configFile.DGraph.Model.AutoBind, "time")

	moduleName, err := internal.GetModuleName()
	if err != nil {
		return nil, err
	}

	var resolverFilename ResolverTemplate
	resolverTemplateSub := resolverTemplateRegex.FindStringSubmatch(configFile.DGraph.Resolver.FilenameTemplate)
	if len(resolverTemplateSub) > 1 {
		if resolverTemplateSub[1] == string(RESOLVER) {
			resolverFilename = RESOLVER
		} else {
			return nil, errors.New("Currently only {resolver}.resolver.go is supported as resolver filename template")
		}
	} else {
		return nil, errors.New("Could not find match name for filename template")
	}

	schema, err := loadSchema(filename, configFile.DGraph.SchemaFileName)
	if err != nil {
		return nil, err
	}

	config := &Config{
		ConfigFile:       configFile,
		ConfigPath:       filename,
		Root:             moduleName,
		Schema:           schema,
		ResolverFilename: resolverFilename,
	}

	return config, nil
}

func validateConfigFile(configFile *ConfigFile) error {
	if configFile.DGraph.Model.Package == "" || configFile.DGraph.Resolver.Package == "" {
		return errors.New("package name must be set for model, resolver and wasm in config")
	}

	if configFile.DGraph.Model.Filename == "" {
		return errors.New("file names for generated executer and model must be set in config")
	}

	if configFile.DGraph.Resolver.Dir == "" {
		return errors.New("resolver target directory must be set in lambda config")
	}
	return nil
}

func loadSchema(configPath string, schemaFilename []string) (*ast.Schema, error) {
	sources, err := loadSources(configPath, schemaFilename)
	if err != nil {
		return nil, err
	}
	schema, err := gqlparser.LoadSchema(sources...)

	if schema.Query == nil {
		schema.Query = &ast.Definition{
			Kind: ast.Object,
			Name: "Query",
		}
		schema.Types["Query"] = schema.Query
	}

	return schema, nil
}

func loadSources(configPath string, schemaFilename []string) ([]*ast.Source, error) {
	var sources []*ast.Source

	abs, err := filepath.Abs(configPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to detect config folder")
	}
	abs = filepath.Dir(abs)

	// Globbing
	var schemaFiles []string
	for _, f := range schemaFilename {
		var matches []string

		fp := filepath.Join(abs, f)

		matches, err = filepath.Glob(fp)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to glob schema filename %s", f)
		}

		for _, m := range matches {
			exists := false
			for _, s := range schemaFiles {
				if s == m {
					exists = true
					break
				}
			}
			if !exists {
				schemaFiles = append(schemaFiles, m)
			}
		}
	}

	// Combine schema files
	for _, filename := range schemaFiles {
		filename = filepath.ToSlash(filename)
		var err error
		var schemaRaw []byte
		schemaRaw, err = ioutil.ReadFile(filename)
		if err != nil {
			errors.Wrap(err, "unable to open schema")
		}

		sources = append(sources, &ast.Source{Input: graphql.SchemaInputs + graphql.DirectiveDefs})
		sources = append(sources, &ast.Source{Input: graphql.ApolloSchemaQueries + graphql.ApolloSchemaExtras})
		sources = append(sources, &ast.Source{Name: filename, Input: string(schemaRaw)})
	}
	return sources, nil
}
