package config

import (
	"fmt"
	"go/types"
	"io/ioutil"
	"path"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"
	"github.com/schartey/dgraph-lambda-go/codegen/graphql"
	"github.com/schartey/dgraph-lambda-go/codegen/parser"
	"github.com/schartey/dgraph-lambda-go/internal"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"golang.org/x/tools/go/packages"
	"gopkg.in/yaml.v2"
)

var ResolverTemplateRegex = regexp.MustCompile(`\{([^)]+)\}.resolver.*`)

type PackageConfig struct {
	Filename string
	Package  string
}

type ResolverConfig struct {
	Layout           string `yaml:"layout,omitempty"`
	Dir              string `yaml:"dir,omitempty"`
	Package          string `yaml:"package,omitempty"`
	FilenameTemplate string `yaml:"filename_template,omitempty"`
}

type Config struct {
	SchemaFilename []string       `yaml:"schema,omitempty"`
	Exec           PackageConfig  `yaml:"exec"`
	Model          PackageConfig  `yaml:"model,omitempty"`
	Resolver       ResolverConfig `yaml:"resolver,omitempty"`
	AutoBind       []string       `yaml:"autobind"`
	Server         struct {
		Standalone bool `yaml:"standalone"`
	} `yaml:"server"`
	Sources             []*ast.Source      `yaml:"-"`
	Packages            *internal.Packages `yaml:"-"`
	Schema              *ast.Schema        `yaml:"-"`
	DefaultModelPackage *packages.Package  `yaml:"-"`

	ParsedTree *parser.Tree
	Root       string
}

// DefaultConfig creates a copy of the default config
func DefaultConfig() *Config {
	return &Config{
		SchemaFilename: []string{"schema.graphql"},
		Model:          PackageConfig{Filename: "models_gen.go"},
		Exec:           PackageConfig{Filename: "generated.go"},
	}
}

// LoadConfig reads the lambda.yaml config file
func LoadConfig(filename string) (*Config, error) {
	config := DefaultConfig()

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read config")
	}

	if err := yaml.UnmarshalStrict(b, config); err != nil {
		return nil, errors.Wrap(err, "unable to parse config")
	}

	preGlobbing := config.SchemaFilename

	var schemaFiles []string
	for _, f := range preGlobbing {
		var matches []string

		matches, err = filepath.Glob(f)
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
	config.SchemaFilename = schemaFiles

	for _, filename := range config.SchemaFilename {
		filename = filepath.ToSlash(filename)
		var err error
		var schemaRaw []byte
		schemaRaw, err = ioutil.ReadFile(filename)
		if err != nil {
			return nil, errors.Wrap(err, "unable to open schema")
		}

		config.Sources = append(config.Sources, &ast.Source{Input: graphql.SchemaInputs + graphql.DirectiveDefs})

		config.Sources = append(config.Sources, &ast.Source{Name: filename, Input: string(schemaRaw)})
	}

	return config, nil
}

func (c *Config) Init() error {
	if c.Packages == nil {
		c.Packages = &internal.Packages{}

		root, err := internal.GetModuleName()
		if err != nil {
			print(err.Error())
		}
		defaultModelPath := root + "/" + path.Dir(c.Model.Filename)

		// Load Default Model Package
		fmt.Println(defaultModelPath)
		defaultPackage, err := c.Packages.Load(defaultModelPath)
		if err != nil {
			print(err.Error())
			return err
		}
		fmt.Println(defaultPackage.Name)
		c.Root = root
		c.DefaultModelPackage = defaultPackage

		// Load packages from yaml
		for _, bind := range c.AutoBind {
			c.Packages.Load(bind)
		}
	}

	if c.Schema == nil {
		if err := c.LoadSchema(); err != nil {
			return err
		}
	}
	parser := parser.NewParser(c.Schema, c.Packages, c.DefaultModelPackage)

	parsedTree, err := parser.Parse()
	if err != nil {
		return err
	}

	c.ParsedTree = parsedTree

	err = c.Autobind()
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) LoadSchema() error {

	schema, err := gqlparser.LoadSchema(c.Sources...)
	if err != nil {
		return err
	}

	if schema.Query == nil {
		schema.Query = &ast.Definition{
			Kind: ast.Object,
			Name: "Query",
		}
		schema.Types["Query"] = schema.Query
	}

	c.Schema = schema
	return nil
}

func (c *Config) Autobind() error {

	for _, autobind := range c.AutoBind {
		pkg, err := c.Packages.PackageFromPath(autobind)
		if err != nil {
			fmt.Println(err)
		}

		for _, model := range c.ParsedTree.ModelTree.Models {
			if model.GoType.TypeName.Pkg() == nil {
				if c.pkgHasType(pkg, model.Name) {
					model.GoType.TypeName = types.NewTypeName(0, types.NewPackage(pkg.PkgPath, pkg.Name), model.Name, nil)
				} else {
					model.GoType.TypeName = types.NewTypeName(0, types.NewPackage(c.DefaultModelPackage.PkgPath, c.DefaultModelPackage.Name), model.Name, nil)
				}
				fmt.Printf("%s -> %s\n", model.Name, model.GoType.TypeName.Pkg().Name())
			}
		}

		for _, it := range c.ParsedTree.ModelTree.Interfaces {
			if it.GoType.TypeName.Pkg() == nil {
				if c.pkgHasType(pkg, it.Name) {
					it.GoType.TypeName = types.NewTypeName(0, types.NewPackage(pkg.PkgPath, pkg.Name), it.Name, nil)
				} else {
					it.GoType.TypeName = types.NewTypeName(0, types.NewPackage(c.DefaultModelPackage.PkgPath, c.DefaultModelPackage.Name), it.Name, nil)
				}
				fmt.Printf("%s -> %s\n", it.Name, it.GoType.TypeName.Pkg().Name())
			}
		}

		for _, it := range c.ParsedTree.ModelTree.Enums {
			if it.GoType.TypeName.Pkg() == nil {
				if c.pkgHasType(pkg, it.Name) {
					it.GoType.TypeName = types.NewTypeName(0, types.NewPackage(pkg.PkgPath, pkg.Name), it.Name, nil)
				} else {
					it.GoType.TypeName = types.NewTypeName(0, types.NewPackage(c.DefaultModelPackage.PkgPath, c.DefaultModelPackage.Name), it.Name, nil)
				}
				fmt.Printf("%s -> %s\n", it.Name, it.GoType.TypeName.Pkg().Name())
			}
		}

		for _, it := range c.ParsedTree.ModelTree.Scalars {
			if it.GoType.TypeName.Pkg() == nil {
				if c.pkgHasType(pkg, it.Name) {
					it.GoType.TypeName = types.NewTypeName(0, types.NewPackage(pkg.PkgPath, pkg.Name), it.Name, nil)
				} else {
					it.GoType.TypeName = types.NewTypeName(0, types.NewPackage(c.DefaultModelPackage.PkgPath, c.DefaultModelPackage.Name), it.Name, nil)
				}
				fmt.Printf("%s -> %s\n", it.Name, it.GoType.TypeName.Pkg().Name())
			}
		}
	}

	return nil
}

func (c *Config) pkgHasType(pkg *packages.Package, name string) bool {
	for _, typeName := range pkg.Types.Scope().Names() {
		if name == typeName {
			return true
		}
	}
	return false
}
