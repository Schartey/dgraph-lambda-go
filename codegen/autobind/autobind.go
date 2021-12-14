package autobind

import (
	"fmt"
	"go/types"
	"strings"

	"github.com/pkg/errors"
	"github.com/schartey/dgraph-lambda-go/codegen/parser"
	"github.com/schartey/dgraph-lambda-go/internal"
	"golang.org/x/tools/go/packages"
)

type Autobinder struct {
	pkgs                   *internal.Packages
	defaultPkg             *packages.Package
	generatedModelFileName string
}

func New(pkgs *internal.Packages, defaultPkg *packages.Package, generatedModelFileName string) *Autobinder {
	return &Autobinder{pkgs: pkgs, defaultPkg: defaultPkg, generatedModelFileName: generatedModelFileName}
}

func (a *Autobinder) Bind(bindPaths []string, parsedTree *parser.Tree) error {
	if len(bindPaths) == 0 {
		for _, it := range parsedTree.ModelTree.Models {
			a.bindPackage(nil, it.GoType, it.Name)
		}

		for _, it := range parsedTree.ModelTree.Interfaces {
			a.bindPackage(nil, it.GoType, it.Name)
		}

		for _, it := range parsedTree.ModelTree.Enums {
			a.bindPackage(nil, it.GoType, it.Name)
		}

		for _, it := range parsedTree.ModelTree.Scalars {
			a.bindPackage(nil, it.GoType, it.Name)
		}
	}

	for _, autobind := range bindPaths {
		var pkg *packages.Package
		pkg, err := a.pkgs.PackageFromPath(autobind)
		if err != nil {
			pkg, err = a.pkgs.Load(autobind)
			if err != nil {
				return errors.Wrap(err, "Could not load package")
			}
		}

		for _, it := range parsedTree.ModelTree.Models {
			if it.GoType.TypeName.Exported() {
				if it.GoType.TypeName.Pkg() == nil {
					a.bindPackage(pkg, it.GoType, it.Name)
				}
			}
		}

		for _, it := range parsedTree.ModelTree.Interfaces {
			if it.GoType.TypeName.Exported() {
				if it.GoType.TypeName.Pkg() == nil {
					a.bindPackage(pkg, it.GoType, it.Name)
				}
			}
		}

		for _, it := range parsedTree.ModelTree.Enums {
			if it.GoType.TypeName.Exported() {
				if it.GoType.TypeName.Pkg() == nil {
					a.bindPackage(pkg, it.GoType, it.Name)
				}
			}
		}

		for _, it := range parsedTree.ModelTree.Scalars {
			if it.GoType.TypeName.Exported() {
				if it.GoType.TypeName.Pkg() == nil {
					a.bindPackage(pkg, it.GoType, it.Name)
				}
			}
		}
	}

	return nil
}

func (a *Autobinder) isCustomInDefaultPkg(pkg *packages.Package, name string) bool {
	if fileName, err := a.pkgs.GetFileNameType(pkg.PkgPath, name); err == nil && !strings.Contains(fileName, a.generatedModelFileName) {
		return true
	}
	return false
}

func (a *Autobinder) bindPackage(pkg *packages.Package, t *parser.GoType, name string) {
	if t.TypeName.Exported() {
		if a.isCustomInDefaultPkg(a.defaultPkg, name) {
			t.TypeName = types.NewTypeName(0, types.NewPackage(a.defaultPkg.PkgPath, a.defaultPkg.Name), name, nil)
			t.Autobind = true
			t.IsDefaultPackage = false
		} else if pkg != nil && pkg.PkgPath != a.defaultPkg.PkgPath && pkgHasType(pkg, name) {
			t.TypeName = types.NewTypeName(0, types.NewPackage(pkg.PkgPath, pkg.Name), name, nil)
			t.Autobind = true
			t.IsDefaultPackage = false
			fmt.Printf("Autobind: %s -> %s\n", name, t.TypeName.Pkg().Name())
		} else {
			t.TypeName = types.NewTypeName(0, types.NewPackage(a.defaultPkg.PkgPath, a.defaultPkg.Name), name, nil)
			t.IsDefaultPackage = true
		}
	}
}

func pkgHasType(pkg *packages.Package, name string) bool {
	for _, typeName := range pkg.Types.Scope().Names() {
		if name == typeName {
			return true
		}
	}
	return false
}
