package internal

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"io/ioutil"
	"strings"

	"golang.org/x/tools/go/packages"
)

var mode = packages.NeedName |
	packages.NeedFiles |
	packages.NeedImports |
	packages.NeedTypes |
	packages.NeedSyntax |
	packages.NeedTypesInfo

type Packages struct {
	packages map[string]*packages.Package
	files    map[string]string
}

func (p *Packages) Load(importPath string) (*packages.Package, error) {
	fmt.Println("load")
	fmt.Println(importPath)
	if p.packages == nil {
		p.packages = map[string]*packages.Package{}
	}
	if p.files == nil {
		p.files = map[string]string{}
	}
	pkgs, err := packages.Load(&packages.Config{Mode: mode}, importPath)
	if err != nil {
		return nil, err
	}

	for _, pkg := range pkgs {
		if len(pkg.Errors) != 0 {
			return nil, pkg.Errors[0]
		}
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

func (p *Packages) GetFileNameType(importPath string, t string) (string, error) {
	if pkg, err := p.PackageFromPath(importPath); err != nil {
		return "", err
	} else {
		for _, f := range pkg.Syntax {
			for _, d := range f.Decls {
				d, isDecl := d.(*ast.GenDecl)
				if !isDecl {
					continue
				}

				fileName, body := p.GetSource(pkg, d.Pos(), d.End())
				if strings.Contains(body, "type "+t) {
					return fileName, nil
				}
			}
		}
		return "", errors.New("not found")
	}
}

func (p *Packages) GetSource(pkg *packages.Package, start, end token.Pos) (string, string) {
	startPos := pkg.Fset.Position(start)
	endPos := pkg.Fset.Position(end)

	if startPos.Filename != endPos.Filename {
		panic("cant get source spanning multiple files")
	}

	file := p.getFile(startPos.Filename)
	return startPos.Filename, file[startPos.Offset:endPos.Offset]
}

func (p *Packages) getFile(filename string) string {

	if file, ok := p.files[filename]; !ok {
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			panic(fmt.Errorf("unable to load file, already exists: %s", err.Error()))
		}
		p.files[filename] = string(b)
		return p.files[filename]
	} else {
		return file
	}
}
