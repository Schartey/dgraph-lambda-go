package parser

import (
	"encoding/json"
	"errors"
	"go/types"
	"regexp"

	"github.com/schartey/dgraph-lambda-go/codegen/graphql"
	"github.com/schartey/dgraph-lambda-go/internal"
	"github.com/vektah/gqlparser/v2/ast"
	"golang.org/x/tools/go/packages"
)

var middlewareRegex = regexp.MustCompile(`@middleware\(([^)]+)\)`)

type GoType struct {
	TypeName *types.TypeName
}

type Scalar struct {
	*GoType
	Name        string
	Description string
}

type Enum struct {
	*GoType
	Name        string
	Description string
	Values      []*EnumValue
}

type EnumValue struct {
	Name        string
	Description string
}

type Interface struct {
	*GoType
	Name        string
	Description string
}

type Field2 struct {
	*GoType
	Name           string
	Description    string
	Tag            string
	ParentTypeName string
}

type Model struct {
	*GoType
	Name        string
	Description string
	Fields      []*Field2
	Implements  []*GoType
}

type Argument2 struct {
	*GoType
	Name string
}

type Query2 struct {
	Name        string
	Description string
	Arguments   []*Argument2
	Return      *GoType
	Middleware  []string
}

type Mutation2 struct {
	Name        string
	Description string
	Arguments   []*Argument2
	Return      *GoType
	Middleware  []string
}

type FieldResolver2 struct {
	Field      *Field2
	Middleware []string
}

type Tree struct {
	ModelTree    *ModelTree
	ResolverTree *ResolverTree
	Middleware   map[string]string
}

type ModelTree struct {
	Interfaces map[string]*Interface
	Models     map[string]*Model
	Enums      map[string]*Enum
	Scalars    map[string]*Scalar
}

type ResolverTree struct {
	FieldResolvers map[string]*FieldResolver2
	Queries        map[string]*Query2
	Mutations      map[string]*Mutation2
}

type Parser struct {
	schema         *ast.Schema
	tree           *Tree
	packages       *internal.Packages
	defaultPackage *packages.Package
}

func NewParser(schema *ast.Schema, packages *internal.Packages, defaultPackage *packages.Package) *Parser {
	return &Parser{schema: schema, tree: &Tree{
		ModelTree: &ModelTree{
			Interfaces: make(map[string]*Interface),
			Models:     make(map[string]*Model),
			Enums:      make(map[string]*Enum),
			Scalars:    make(map[string]*Scalar),
		},
		ResolverTree: &ResolverTree{
			FieldResolvers: make(map[string]*FieldResolver2),
			Queries:        make(map[string]*Query2),
			Mutations:      make(map[string]*Mutation2),
		},
		Middleware: make(map[string]string),
	},
		packages:       packages,
		defaultPackage: defaultPackage,
	}
}

func (p *Parser) Parse() (*Tree, error) {
	for _, schemaType := range p.schema.Types {
		p.parseType(schemaType, true)
	}
	return p.tree, nil
}

func (p *Parser) parseType(schemaType *ast.Definition, mustLambda bool) (*GoType, error) {
	if mustLambda && !p.hasLambda(schemaType) {
		return nil, errors.New("Type has no lambda field")
	}

	var goType *GoType
	var pkg *packages.Package
	var err error

	pkgPath, typeName, err := graphql.SchemaDefToGoDef(schemaType)
	if err != nil {
		// Keep pkg as null, we check later with autobind
		pkg = p.defaultPackage
		typeName = schemaType.Name

		goType = &GoType{
			TypeName: types.NewTypeName(0, types.NewPackage(pkg.PkgPath, pkg.Name), typeName, nil),
		}
	} else {
		if pkgPath == "" {
			goType = &GoType{
				TypeName: types.NewTypeName(0, nil, typeName, nil),
			}
		} else {
			pkg, err = p.packages.PackageFromPath(pkgPath)
			if err != nil {
				pkg, err = p.packages.Load(pkgPath)
			}

			goType = &GoType{
				TypeName: types.NewTypeName(0, types.NewPackage(pkg.PkgPath, pkg.Name), typeName, nil),
			}
		}
	}

	switch schemaType.Kind {
	case ast.Interface, ast.Union:
		if it, ok := p.tree.ModelTree.Interfaces[schemaType.Name]; ok {
			return it.GoType, nil
		}
		it := &Interface{
			Description: schemaType.Description,
			Name:        schemaType.Name,
			GoType:      goType,
		}

		p.tree.ModelTree.Interfaces[it.Name] = it
		return it.GoType, nil

	case ast.Object, ast.InputObject:
		if schemaType == p.schema.Subscription {
			return nil, errors.New("Subscription not supported")
		}
		if schemaType == p.schema.Query || schemaType == p.schema.Mutation {
			if it, ok := p.tree.ResolverTree.Queries[schemaType.Name]; ok {
				return it.Return, nil
			}
			if it, ok := p.tree.ResolverTree.Mutations[schemaType.Name]; ok {
				return it.Return, nil
			}
			for _, field := range schemaType.Fields {
				lambdaDirective := field.Directives.ForName("lambda")

				if lambdaDirective == nil {
					continue
				}

				returnType := p.schema.Types[field.Type.Name()]
				returnGoType, err := p.parseType(returnType, false)
				if err != nil {
					return nil, err
				}

				var args []*Argument2
				for _, arg := range field.Arguments {
					argType := p.schema.Types[arg.Type.Name()]
					argGoType, err := p.parseType(argType, false)
					if err != nil {
						return nil, err
					}

					args = append(args, &Argument2{Name: arg.Name, GoType: argGoType})
				}
				out := middlewareRegex.FindAllStringSubmatch(field.Description, -1)

				var fieldMiddleware []string
				for _, i := range out {
					rawMiddleware := i[1]
					json.Unmarshal([]byte(rawMiddleware), &fieldMiddleware)
				}
				for _, m := range fieldMiddleware {
					p.tree.Middleware[m] = m
				}

				if schemaType == p.schema.Query {
					p.tree.ResolverTree.Queries[field.Name] = &Query2{Name: field.Name, Description: field.Description, Arguments: args, Return: returnGoType, Middleware: fieldMiddleware}
				}

				if schemaType == p.schema.Mutation {
					p.tree.ResolverTree.Mutations[field.Name] = &Mutation2{Name: field.Name, Description: field.Description, Arguments: args, Return: returnGoType, Middleware: fieldMiddleware}
				}
			}
		} else {
			// Model
			if it, ok := p.tree.ModelTree.Models[schemaType.Name]; ok {
				return it.GoType, nil
			}
			it := &Model{
				Name:        schemaType.Name,
				Description: schemaType.Description,
				GoType:      goType,
			}
			for _, implementor := range p.schema.GetImplements(schemaType) {
				interfaceType, err := p.parseType(implementor, false)
				if err != nil {
					return nil, err
				}
				it.Implements = append(it.Implements, interfaceType)
			}

			for _, field := range schemaType.Fields {
				fieldType := p.schema.Types[field.Type.Name()]

				fieldGoType, err := p.parseType(fieldType, false)
				if err != nil {
					return nil, err
				}

				modelField := &Field2{
					Name:           field.Name,
					Description:    field.Description,
					Tag:            `json:"` + field.Name + `"`,
					GoType:         fieldGoType,
					ParentTypeName: schemaType.Name,
				}
				it.Fields = append(it.Fields, modelField)

				lambdaDirective := field.Directives.ForName("lambda")

				if lambdaDirective != nil {
					out := middlewareRegex.FindAllStringSubmatch(field.Description, -1)

					var fieldMiddleware []string
					for _, i := range out {
						rawMiddleware := i[1]
						json.Unmarshal([]byte(rawMiddleware), &fieldMiddleware)
					}
					for _, m := range fieldMiddleware {
						p.tree.Middleware[m] = m
					}
					p.tree.ResolverTree.FieldResolvers[field.Name] = &FieldResolver2{Field: modelField, Middleware: fieldMiddleware}
				}
			}
			p.tree.ModelTree.Models[it.Name] = it
			return it.GoType, nil
		}

	case ast.Enum:
		if it, ok := p.tree.ModelTree.Enums[schemaType.Name]; ok {
			return it.GoType, nil
		}
		it := &Enum{
			Name:        schemaType.Name,
			Description: schemaType.Description,
			GoType:      goType,
		}

		for _, v := range schemaType.EnumValues {
			it.Values = append(it.Values, &EnumValue{
				Name:        v.Name,
				Description: v.Description,
			})
		}

		p.tree.ModelTree.Enums[it.Name] = it
		return it.GoType, nil
	case ast.Scalar:
		if it, ok := p.tree.ModelTree.Scalars[schemaType.Name]; ok {
			return it.GoType, nil
		}
		it := &Scalar{
			Name:        schemaType.Name,
			Description: schemaType.Description,
			GoType:      goType,
		}
		p.tree.ModelTree.Scalars[schemaType.Name] = it
		return it.GoType, nil
	}

	return nil, nil
}

func (p *Parser) hasLambda(def *ast.Definition) bool {
	for _, field := range def.Fields {
		if field.Directives.ForName("lambda") != nil {
			return true
		}
	}
	return false
}
