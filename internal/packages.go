package internal

import (
	"errors"

	"golang.org/x/tools/go/packages"
)

var mode = packages.NeedName |
	packages.NeedFiles |
	packages.NeedImports |
	packages.NeedTypes |
	packages.NeedSyntax |
	packages.NeedTypesInfo

type Packages struct {
	packages     map[string]*packages.Package
	importToName map[string]string
	loadErrors   []error
}

func (p *Packages) Load(importPath string) (*packages.Package, error) {
	if p.packages == nil {
		p.packages = map[string]*packages.Package{}
	}
	pkgs, err := packages.Load(&packages.Config{Mode: mode}, importPath)
	if err != nil {
		return nil, err
	}

	for _, pkg := range pkgs {
		p.packages[importPath] = pkg
	}

	return p.packages[importPath], nil
}

func (p *Packages) PackageFromPath(importPath string) (*packages.Package, error) {

	if pkg, found := p.packages[importPath]; !found {
		return nil, errors.New("package not found")
	} else {
		return pkg, nil
	}
}
