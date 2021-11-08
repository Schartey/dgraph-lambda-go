package generator

import (
	"fmt"
	"go/types"
	"os"
	"text/template"

	"github.com/schartey/dgraph-lambda-go/codegen/config"
	"github.com/schartey/dgraph-lambda-go/codegen/parser"
	"golang.org/x/tools/go/packages"
)

var defaultPackage *packages.Package

func generateModel(c *config.Config, parsedTree *parser.Tree) error {

	defaultPackage = c.DefaultModelPackage

	f, err := os.Create(c.Model.Filename)
	if err != nil {
		return err
	}
	defer f.Close()

	var pkgs = make(map[string]*types.Package)
	var models = make(map[string]*parser.Model)
	var enums = make(map[string]*parser.Enum)
	var interfaces = make(map[string]*parser.Interface)
	var scalars = make(map[string]*parser.Scalar)

	for _, m := range parsedTree.ModelTree.Models {
		if m.GoType.TypeName.Pkg().Path() == c.DefaultModelPackage.PkgPath && !m.GoType.Autobind {
			models[m.Name] = m
		}
		for _, f := range m.Fields {
			if f.TypeName.Exported() && f.GoType.TypeName.Pkg().Path() != c.DefaultModelPackage.PkgPath {
				pkgs[f.GoType.TypeName.Pkg().Name()] = f.GoType.TypeName.Pkg()
			}
		}
	}
	for _, m := range parsedTree.ModelTree.Enums {
		if m.TypeName.Exported() && m.GoType.TypeName.Pkg().Path() == c.DefaultModelPackage.PkgPath && !m.GoType.Autobind {
			enums[m.Name] = m
		}
	}
	for _, m := range parsedTree.ModelTree.Interfaces {
		if m.TypeName.Exported() && m.GoType.TypeName.Pkg().Path() == c.DefaultModelPackage.PkgPath && !m.GoType.Autobind {
			interfaces[m.Name] = m
		}
	}
	for _, m := range parsedTree.ModelTree.Scalars {
		if m.TypeName.Exported() && m.GoType.TypeName.Pkg().Path() == c.DefaultModelPackage.PkgPath && !m.GoType.Autobind {
			scalars[m.Name] = m
		}
	}
	if len(enums) > 0 {
		pkgs["fmt"] = types.NewPackage("fmt", "fmt")
		pkgs["strconv"] = types.NewPackage("strconv", "strconv")
		pkgs["io"] = types.NewPackage("io", "io")
	}

	err = modelTemplate.Execute(f, struct {
		Interfaces  map[string]*parser.Interface
		Enums       map[string]*parser.Enum
		Scalars     map[string]*parser.Scalar
		Models      map[string]*parser.Model
		Packages    map[string]*types.Package
		PackageName string
	}{
		Interfaces:  interfaces,
		Enums:       enums,
		Scalars:     parsedTree.ModelTree.Scalars,
		Models:      models,
		Packages:    pkgs,
		PackageName: c.Model.Package,
	})
	if err != nil {
		return err
	}
	return nil
}

func modelRef(t *parser.GoType, isArray bool) string {
	if t.TypeName.Exported() && t.TypeName.Pkg().Path() != defaultPackage.PkgPath {
		if isArray {
			return fmt.Sprintf("[]*%s.%s", t.TypeName.Pkg().Name(), t.TypeName.Name())
		} else {
			return fmt.Sprintf("*%s.%s", t.TypeName.Pkg().Name(), t.TypeName.Name())
		}
	}
	if t.TypeName.Exported() {
		if isArray {
			return fmt.Sprintf("[]*%s", t.TypeName.Name())
		} else {
			return fmt.Sprintf("*%s", t.TypeName.Name())
		}
	}
	if isArray {
		return fmt.Sprintf("[]%s", t.TypeName.Name())
	} else {
		return t.TypeName.Name()
	}
}

var modelTemplate = template.Must(template.New("model").Funcs(template.FuncMap{
	"ref":   modelRef,
	"path":  pkgPath,
	"title": title,
}).Parse(`package {{.PackageName}}

import(
	{{- range $pkg := .Packages }}
	"{{ $pkg | path }}"{{- end}}
)

{{- range $model := .Interfaces }}
type {{.Name }} interface {
	Is{{.Name }}()
}
{{- end }}
{{- range $model := .Models }}
type {{ .Name }} struct {
	{{- range $field := .Fields }}
		{{- with .Description }}
		{{- end}}
		{{ $field.Name | title }} {{ ref $field.GoType $field.IsArray }} ` + "`{{$field.Tag}}`" + `
	{{- end }}
}
{{- end }}
{{ range $enum := .Enums }}
type {{$enum.Name }} string
const (
{{- range $value := $enum.Values}}
	{{- with $value.Description}}
	{{- end}}
	{{ $enum.Name }}{{ $value.Name }} {{$enum.Name }} = "{{$value.Name}}"
{{- end }}
)

var All{{$enum.Name }} = []{{ $enum.Name }}{
{{- range $value := $enum.Values}}
	{{$enum.Name }}{{ .Name }},
{{- end }}
}

func (e {{$enum.Name }}) IsValid() bool {
	switch e {
	case {{ range $index, $element := $enum.Values}}{{if $index}},{{end}}{{ $enum.Name }}{{ $element.Name }}{{end}}:
		return true
	}
	return false
}

func (e {{$enum.Name }}) String() string {
	return string(e)
}

func (e *{{$enum.Name }}) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = {{ $enum.Name }}(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid {{ $enum.Name }}", str)
	}
	return nil
}

func (e {{$enum.Name }}) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

{{- end }}
`))
