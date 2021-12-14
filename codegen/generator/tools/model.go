package tools

import (
	"go/types"

	"github.com/schartey/dgraph-lambda-go/codegen/parser"
)

// Tools isn't really the right spot for this
func GetDefaultPackageTree(defaultPackagePath string, parsedTree *parser.Tree) (*parser.Tree, map[string]*types.Package) {
	// These packages should be collected earlier within the Tree
	var pkgs = make(map[string]*types.Package)
	var models = make(map[string]*parser.Model)
	var enums = make(map[string]*parser.Enum)
	var interfaces = make(map[string]*parser.Interface)
	var scalars = make(map[string]*parser.Scalar)

	for _, m := range parsedTree.ModelTree.Models {
		if m.GoType.TypeName.Pkg().Path() == defaultPackagePath && !m.GoType.Autobind {
			models[m.Name] = m
		}
		for _, f := range m.Fields {
			if f.TypeName.Exported() && f.GoType.TypeName.Pkg().Path() != defaultPackagePath {
				pkgs[f.GoType.TypeName.Pkg().Name()] = f.GoType.TypeName.Pkg()
			}
		}
	}
	for _, m := range parsedTree.ModelTree.Enums {
		if m.TypeName.Exported() && m.GoType.TypeName.Pkg().Path() == defaultPackagePath && !m.GoType.Autobind {
			enums[m.Name] = m
		}
	}
	for _, m := range parsedTree.ModelTree.Interfaces {
		if m.TypeName.Exported() && m.GoType.TypeName.Pkg().Path() == defaultPackagePath && !m.GoType.Autobind {
			interfaces[m.Name] = m
		}
	}
	for _, m := range parsedTree.ModelTree.Scalars {
		if m.TypeName.Exported() && m.GoType.TypeName.Pkg().Path() == defaultPackagePath && !m.GoType.Autobind {
			scalars[m.Name] = m
		}
	}
	if len(enums) > 0 {
		pkgs["fmt"] = types.NewPackage("fmt", "fmt")
		pkgs["strconv"] = types.NewPackage("strconv", "strconv")
		pkgs["io"] = types.NewPackage("io", "io")
	}

	defaultPackageTree := &parser.Tree{
		ModelTree:    &parser.ModelTree{Models: models, Interfaces: interfaces, Enums: enums, Scalars: scalars},
		ResolverTree: parsedTree.ResolverTree,
		Middleware:   parsedTree.Middleware,
	}
	return defaultPackageTree, pkgs
}
