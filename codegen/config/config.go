package config

import (
	"fmt"
	"go/types"
	"io/ioutil"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/schartey/dgraph-lambda-go/codegen/graphql"
	"github.com/schartey/dgraph-lambda-go/codegen/parser"
	"github.com/schartey/dgraph-lambda-go/internal"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"golang.org/x/tools/go/packages"
	"gopkg.in/yaml.v2"
)

var resolverTemplateRegex = regexp.MustCompile(`\{([^)]+)\}.resolver.*`)

type PackageConfig struct {
	Filename string
	Package  string
}

type ResolverConfig struct {
	Layout           string `yaml:"layout"`
	Dir              string `yaml:"dir"`
	Package          string `yaml:"package"`
	FilenameTemplate string `yaml:"filename_template"`
}

type Config struct {
	SchemaFilename []string       `yaml:"schema"`
	Exec           PackageConfig  `yaml:"exec"`
	Model          PackageConfig  `yaml:"model"`
	Resolver       ResolverConfig `yaml:"resolver"`
	Force          []string       `yaml:"force"`
	AutoBind       []string       `yaml:"autobind"`
	Server         struct {
		Standalone bool `yaml:"standalone"`
	} `yaml:"server"`

	Sources             []*ast.Source      `yaml:"-"`
	Packages            *internal.Packages `yaml:"-"`
	Schema              *ast.Schema        `yaml:"-"`
	Root                string             `yaml:"-"`
	DefaultModelPackage *packages.Package  `yaml:"-"`
	ResolverFilename    string             `yaml:"-"`
}

func LoadConfigFile(moduleName string, filename string) (*Config, error) {
	config := &Config{}

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read config")
	}

	if err := yaml.UnmarshalStrict(b, config); err != nil {
		return nil, errors.Wrap(err, "unable to parse config")
	}

	if config.Exec.Package == "" || config.Model.Package == "" || config.Resolver.Package == "" {
		return nil, errors.New("package name must be set in lambda config")
	}

	if config.Exec.Filename == "" || config.Model.Filename == "" {
		return nil, errors.New("file names for generated executer and model must be set in lambda config")
	}

	if config.Resolver.Dir == "" {
		return nil, errors.New("resovler target direcotry must be set in lambda config")
	}

	config.Root = moduleName

	resolverTemplateSub := resolverTemplateRegex.FindStringSubmatch(config.Resolver.FilenameTemplate)
	if len(resolverTemplateSub) > 1 {
		if resolverTemplateSub[1] != "resolver" {
			return nil, errors.New("Currently only {resolver}.resolver.go is supported as resolver filename template")
		} else {
			config.ResolverFilename = resolverTemplateSub[1]
		}
	} else {
		return nil, errors.New("Could not find match name for filename template")
	}

	return config, nil
}

func (config *Config) LoadConfig(filename string) error {
	preGlobbing := config.SchemaFilename

	abs, err := filepath.Abs(filename)
	if err != nil {
		return errors.Wrap(err, "unable to detect config folder")
	}
	abs = filepath.Dir(abs)

	var schemaFiles []string
	for _, f := range preGlobbing {
		var matches []string

		fp := filepath.Join(abs, f)

		matches, err = filepath.Glob(fp)
		if err != nil {
			return errors.Wrapf(err, "failed to glob schema filename %s", f)
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
			errors.Wrap(err, "unable to open schema")
		}

		config.Sources = append(config.Sources, &ast.Source{Input: graphql.SchemaInputs + graphql.DirectiveDefs})
		config.Sources = append(config.Sources, &ast.Source{Input: graphql.ApolloSchemaQueries + graphql.ApolloSchemaExtras})
		config.Sources = append(config.Sources, &ast.Source{Name: filename, Input: string(schemaRaw)})
	}

	if config.Packages == nil {
		config.Packages = &internal.Packages{}

		defaultModelPath := config.Root + "/" + path.Dir(config.Model.Filename)

		defaultPackage, err := config.Packages.Load(defaultModelPath)
		if err != nil {
			return errors.Wrap(err, "Could not load generated model package")
		}
		config.DefaultModelPackage = defaultPackage
	}

	return nil
}

func (c *Config) LoadSchema() error {
	if c.Schema == nil {
		if err := c.loadSchema(); err != nil {
			return errors.Wrap(err, "Could not load schema")
		}
	}
	return nil
}

func (c *Config) loadSchema() error {
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

func (c *Config) Bind(parsedTree *parser.Tree) error {

	if len(c.AutoBind) == 0 {
		for _, model := range parsedTree.ModelTree.Models {
			if model.GoType.TypeName.Exported() {
				if model.GoType.TypeName.Pkg() == nil {
					model.GoType.TypeName = types.NewTypeName(0, types.NewPackage(c.DefaultModelPackage.PkgPath, c.DefaultModelPackage.Name), model.Name, nil)
				}
			}
		}

		for _, it := range parsedTree.ModelTree.Interfaces {
			if it.GoType.TypeName.Exported() {
				if it.GoType.TypeName.Pkg() == nil {
					it.GoType.TypeName = types.NewTypeName(0, types.NewPackage(c.DefaultModelPackage.PkgPath, c.DefaultModelPackage.Name), it.Name, nil)
				}
			}
		}

		for _, it := range parsedTree.ModelTree.Enums {
			if it.GoType.TypeName.Exported() {
				if it.GoType.TypeName.Pkg() == nil {
					it.GoType.TypeName = types.NewTypeName(0, types.NewPackage(c.DefaultModelPackage.PkgPath, c.DefaultModelPackage.Name), it.Name, nil)
				}
			}
		}

		for _, it := range parsedTree.ModelTree.Scalars {
			if it.GoType.TypeName.Exported() {
				if it.GoType.TypeName.Pkg() == nil {
					it.GoType.TypeName = types.NewTypeName(0, types.NewPackage(c.DefaultModelPackage.PkgPath, c.DefaultModelPackage.Name), it.Name, nil)
				}
			}
		}
	}

	for _, autobind := range c.AutoBind {
		var pkg *packages.Package
		pkg, err := c.Packages.PackageFromPath(autobind)
		if err != nil {
			pkg, err = c.Packages.Load(autobind)
			if err != nil {
				return errors.Wrap(err, "Could not load package")
			}
		}

		for _, model := range parsedTree.ModelTree.Models {
			if model.GoType.TypeName.Exported() {
				if model.GoType.TypeName.Pkg() == nil {
					if c.pkgHasType(pkg, model.Name) {
						model.GoType.TypeName = types.NewTypeName(0, types.NewPackage(pkg.PkgPath, pkg.Name), model.Name, nil)
						model.GoType.Autobind = true
						fmt.Printf("Autobind: %s -> %s\n", model.Name, model.GoType.TypeName.Pkg().Name())
					} else if c.isCustomInDefaultPkg(c.DefaultModelPackage, model.Name) {
						model.GoType.TypeName = types.NewTypeName(0, types.NewPackage(c.DefaultModelPackage.PkgPath, c.DefaultModelPackage.Name), model.Name, nil)
						model.GoType.Autobind = true
					} else {
						model.GoType.TypeName = types.NewTypeName(0, types.NewPackage(c.DefaultModelPackage.PkgPath, c.DefaultModelPackage.Name), model.Name, nil)
					}
				}
			}
		}

		for _, it := range parsedTree.ModelTree.Interfaces {
			if it.GoType.TypeName.Exported() {
				if it.GoType.TypeName.Pkg() == nil {
					if c.pkgHasType(pkg, it.Name) {
						it.GoType.TypeName = types.NewTypeName(0, types.NewPackage(pkg.PkgPath, pkg.Name), it.Name, nil)
						it.GoType.Autobind = true
						fmt.Printf("Autobind: %s -> %s\n", it.Name, it.GoType.TypeName.Pkg().Name())
					} else if c.isCustomInDefaultPkg(c.DefaultModelPackage, it.Name) {
						it.GoType.TypeName = types.NewTypeName(0, types.NewPackage(c.DefaultModelPackage.PkgPath, c.DefaultModelPackage.Name), it.Name, nil)
						it.GoType.Autobind = true
					} else {
						it.GoType.TypeName = types.NewTypeName(0, types.NewPackage(c.DefaultModelPackage.PkgPath, c.DefaultModelPackage.Name), it.Name, nil)
					}
				}
			}
		}

		for _, it := range parsedTree.ModelTree.Enums {
			if it.GoType.TypeName.Exported() {
				if it.GoType.TypeName.Pkg() == nil {
					if c.pkgHasType(pkg, it.Name) {
						it.GoType.TypeName = types.NewTypeName(0, types.NewPackage(pkg.PkgPath, pkg.Name), it.Name, nil)
						it.GoType.Autobind = true
						fmt.Printf("Autobind: %s -> %s\n", it.Name, it.GoType.TypeName.Pkg().Name())
					} else if c.isCustomInDefaultPkg(c.DefaultModelPackage, it.Name) {
						it.GoType.TypeName = types.NewTypeName(0, types.NewPackage(c.DefaultModelPackage.PkgPath, c.DefaultModelPackage.Name), it.Name, nil)
						it.GoType.Autobind = true
					} else {
						it.GoType.TypeName = types.NewTypeName(0, types.NewPackage(c.DefaultModelPackage.PkgPath, c.DefaultModelPackage.Name), it.Name, nil)
					}
				}
			}
		}

		for _, it := range parsedTree.ModelTree.Scalars {
			if it.GoType.TypeName.Exported() {
				if it.GoType.TypeName.Pkg() == nil {
					if c.pkgHasType(pkg, it.Name) {
						it.GoType.TypeName = types.NewTypeName(0, types.NewPackage(pkg.PkgPath, pkg.Name), it.Name, nil)
						it.GoType.Autobind = true
						fmt.Printf("Autobind: %s -> %s\n", it.Name, it.GoType.TypeName.Pkg().Name())
					} else if c.isCustomInDefaultPkg(c.DefaultModelPackage, it.Name) {
						it.GoType.TypeName = types.NewTypeName(0, types.NewPackage(c.DefaultModelPackage.PkgPath, c.DefaultModelPackage.Name), it.Name, nil)
						it.GoType.Autobind = true
					} else {
						it.GoType.TypeName = types.NewTypeName(0, types.NewPackage(c.DefaultModelPackage.PkgPath, c.DefaultModelPackage.Name), it.Name, nil)
					}
				}
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

func (c *Config) isCustomInDefaultPkg(pkg *packages.Package, name string) bool {
	if fileName, err := c.Packages.GetFileNameType(pkg.PkgPath, name); err == nil && !strings.Contains(fileName, c.Model.Filename) {
		return true
	}
	return false
}
