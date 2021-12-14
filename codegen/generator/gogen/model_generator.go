package gogen

import (
	"fmt"
	"go/types"
	"text/template"

	"github.com/schartey/dgraph-lambda-go/codegen/generator/tools"
	"github.com/schartey/dgraph-lambda-go/codegen/parser"
	"github.com/schartey/dgraph-lambda-go/config"
)

func generateModel(c *config.Config, parsedTree *parser.Tree, pkgs map[string]*types.Package) error {
	/*
		f, err := os.Create(c.Model.Filename)
		if err != nil {
			return err
		}
		defer f.Close()

		err = modelTemplate.Execute(f, struct {
			Interfaces  map[string]*parser.Interface
			Enums       map[string]*parser.Enum
			Scalars     map[string]*parser.Scalar
			Models      map[string]*parser.Model
			Packages    map[string]*types.Package
			PackageName string
		}{
			Interfaces:  parsedTree.ModelTree.Interfaces,
			Enums:       parsedTree.ModelTree.Enums,
			Scalars:     parsedTree.ModelTree.Scalars,
			Models:      parsedTree.ModelTree.Models,
			Packages:    pkgs,
			PackageName: c.Model.Package,
		})
		if err != nil {
			return err
		}*/
	return nil
}

func modelRef(t *parser.GoType, isArray bool) string {
	if t.TypeName.Exported() && !t.IsDefaultPackage {
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
	"path":  tools.PkgPath,
	"title": tools.Title,
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
