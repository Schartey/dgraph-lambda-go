package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/types"
	"regexp"

	"github.com/schartey/dgraph-lambda-go/codegen/graphql"
	"github.com/schartey/dgraph-lambda-go/internal"
	"github.com/vektah/gqlparser/v2/ast"
)

var middlewareRegex = regexp.MustCompile(`@middleware\(([^)]+)\)`)

type LambdaOnMutateEvent string

const (
	ADD    LambdaOnMutateEvent = "add"
	UPDATE LambdaOnMutateEvent = "update"
	DELETE LambdaOnMutateEvent = "delete"
)

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

type Field struct {
	*GoType
	Name        string
	Description string
	Tag         string
	IsArray     bool
}

type Model struct {
	*GoType
	Name           string
	Description    string
	Fields         []*Field
	Implements     []*GoType
	LambdaOnMutate []LambdaOnMutateEvent
}

type Argument struct {
	*GoType
	Name    string
	IsArray bool
}

type Return struct {
	*GoType
	IsArray bool
}

type Query struct {
	Name        string
	Description string
	Arguments   []*Argument
	Return      *Return
	Middleware  []string
}

type Mutation struct {
	Name        string
	Description string
	Arguments   []*Argument
	Return      *Return
	Middleware  []string
}

type Parent struct {
	*GoType
	Name string
}

type FieldResolver struct {
	Field      *Field
	Parent     *Parent
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
	FieldResolvers map[string]*FieldResolver
	Queries        map[string]*Query
	Mutations      map[string]*Mutation
}

type Parser struct {
	schema   *ast.Schema
	tree     *Tree
	packages *internal.Packages
}

func NewParser(schema *ast.Schema, packages *internal.Packages) *Parser {
	return &Parser{schema: schema, tree: &Tree{
		ModelTree: &ModelTree{
			Interfaces: make(map[string]*Interface),
			Models:     make(map[string]*Model),
			Enums:      make(map[string]*Enum),
			Scalars:    make(map[string]*Scalar),
		},
		ResolverTree: &ResolverTree{
			FieldResolvers: make(map[string]*FieldResolver),
			Queries:        make(map[string]*Query),
			Mutations:      make(map[string]*Mutation),
		},
		Middleware: make(map[string]string),
	},
		packages: packages,
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
		return nil, errors.New("type has no lambda field")
	}

	var goType *GoType
	var err error

	pkgPath, typeName, err := graphql.SchemaDefToGoDef(schemaType)
	if err != nil {
		goType = &GoType{
			TypeName: types.NewTypeName(0, nil, schemaType.Name, nil),
		}
	} else {
		if pkgPath == "" {
			goType = &GoType{
				TypeName: types.NewTypeName(0, nil, typeName, nil),
			}
		} else {
			pkg, err := p.packages.PackageFromPath(pkgPath)
			if err != nil {
				pkg, err = p.packages.Load(pkgPath)
				if err != nil {
					fmt.Println("Could not load package")
				}
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

		for _, field := range schemaType.Fields {
			fieldType := p.schema.Types[field.Type.Name()]

			fieldGoType, err := p.parseType(fieldType, false)
			if err != nil {
				return nil, err
			}

			tag := `json:"` + field.Name + `"`
			if field.Name == "id" {
				tag += ` dql:"uid"`
			} else {
				tag += ` dql:"` + it.Name + "." + field.Name + `"`
			}

			modelField := &Field{
				Name:        field.Name,
				Description: field.Description,
				Tag:         tag,
				GoType:      fieldGoType,
				IsArray:     graphql.IsArray(field.Type.String()),
			}

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
				p.tree.ResolverTree.FieldResolvers[field.Name] = &FieldResolver{Field: modelField, Parent: &Parent{Name: schemaType.Name, GoType: it.GoType}, Middleware: fieldMiddleware}
			}
		}

		return it.GoType, nil

	case ast.Object, ast.InputObject:
		if schemaType == p.schema.Subscription {
			fmt.Println("Subscription not supported - skipping")
			//return nil, errors.New("subscription not supported")
		}
		if schemaType == p.schema.Query || schemaType == p.schema.Mutation {
			if it, ok := p.tree.ResolverTree.Queries[schemaType.Name]; ok {
				return it.Return.GoType, nil
			}
			if it, ok := p.tree.ResolverTree.Mutations[schemaType.Name]; ok {
				return it.Return.GoType, nil
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
				returnField := &Return{
					GoType:  returnGoType,
					IsArray: graphql.IsArray(field.Type.String()),
				}

				var args []*Argument
				for _, arg := range field.Arguments {
					argType := p.schema.Types[arg.Type.Name()]
					argGoType, err := p.parseType(argType, false)
					if err != nil {
						return nil, err
					}

					args = append(args, &Argument{Name: arg.Name,
						GoType:  argGoType,
						IsArray: graphql.IsArray(field.Type.String()),
					})
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
					p.tree.ResolverTree.Queries[field.Name] = &Query{Name: field.Name, Description: field.Description, Arguments: args, Return: returnField, Middleware: fieldMiddleware}
				}

				if schemaType == p.schema.Mutation {
					p.tree.ResolverTree.Mutations[field.Name] = &Mutation{Name: field.Name, Description: field.Description, Arguments: args, Return: returnField, Middleware: fieldMiddleware}
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

			p.tree.ModelTree.Models[it.Name] = it

			lambdaOnMutate := schemaType.Directives.ForName("lambdaOnMutate")
			if lambdaOnMutate != nil {
				if lambdaOnMutate.Arguments.ForName("add") != nil {
					it.LambdaOnMutate = append(it.LambdaOnMutate, ADD)
				}
				if lambdaOnMutate.Arguments.ForName("update") != nil {
					it.LambdaOnMutate = append(it.LambdaOnMutate, UPDATE)
				}
				if lambdaOnMutate.Arguments.ForName("delete") != nil {
					it.LambdaOnMutate = append(it.LambdaOnMutate, DELETE)
				}
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

				tag := `json:"` + field.Name + `"`
				if field.Name == "id" {
					tag += ` dql:"uid"`
				} else {
					tag += ` dql:"` + it.Name + "." + field.Name + `"`
				}

				modelField := &Field{
					Name:        field.Name,
					Description: field.Description,
					Tag:         tag,
					GoType:      fieldGoType,
					IsArray:     graphql.IsArray(field.Type.String()),
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
					p.tree.ResolverTree.FieldResolvers[field.Name] = &FieldResolver{Field: modelField, Parent: &Parent{Name: schemaType.Name, GoType: it.GoType}, Middleware: fieldMiddleware}
				}
			}
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
	if def.Directives.ForName("lambdaOnMutate") != nil {
		return true
	}
	for _, field := range def.Fields {
		if field.Directives.ForName("lambda") != nil {
			return true
		}
	}
	return false
}
