package templates

import (
	"text/template"

	"github.com/schartey/dgraph-lambda-go/codegen/generator/tools"
)

var GolangModelTemplate = template.Must(template.New("model").Funcs(template.FuncMap{
	"path":       tools.PkgPath,
	"title":      tools.Title,
	"ref":        tools.ModelRef,
	"jsonRef":    tools.JsonRef,
	"jsonRefVal": tools.JsonRefVal,
	"jsonVal":    tools.JsonVal,
	"untitle":    tools.Untitle,
}).Option().Parse(`package model

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
	{{ $field.Name | title }} {{ ref $field.GoType $field.IsArray }} ` + "`{{$field.Tag}}`" + `
	{{- end }}
}
func ({{ untitle .Name }} *{{ .Name }}) Marshal() []byte {
	if {{ untitle .Name }} == nil {
		return []byte("null")
	}
	return []byte(fmt.Sprintf(` + "`" + `{
	{{- range $field := .Fields }}
	"{{ $field.Name }}": {{ jsonRef $field.GoType $field.IsArray }},
	{{- end }}
}` + "`" + `, 
	{{- range $field := .Fields }}
	{{ jsonRefVal $field $model.GoType }},
	{{- end }}))
}
func Unmarshal{{ .Name }}(v *fastjson.Value) {{ ref $model.GoType false }} {
	if v == nil {
		return nil
	}
	{{ untitle .Name }} := &{{ .Name }}{
		{{- range $field := .Fields }}
		{{ title $field.Name }}: {{ jsonVal $field }},
		{{- end }}
	}
	return {{ untitle .Name }}
}
{{ end }}

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

var GolangFieldResolverTemplate = template.Must(template.New("field-resolver").Funcs(template.FuncMap{
	"path":    tools.PkgPath,
	"pointer": tools.Pointer,
	"body":    tools.FieldResolverBody,
	"is":      tools.Is,
}).Parse(`
package {{.PackageName}}

import(
	{{- range $pkg := .Packages }}
	"{{ $pkg | path }}"{{- end}}
)

type FieldResolverInterface interface {
{{- range $fieldResolver := .FieldResolvers}}
	{{$fieldResolver.Parent.Name }}_{{$fieldResolver.Field.Name}}(parents []{{ pointer $fieldResolver.Parent.GoType false }}, authHeader wasm.AuthHeader) ([]{{ pointer $fieldResolver.Field.GoType $fieldResolver.Field.IsArray }}, error){{ end }}
}

type FieldResolver struct {
	*Resolver
}

{{- range $fieldResolver := .FieldResolvers}}
func (f *FieldResolver) {{$fieldResolver.Parent.Name }}_{{$fieldResolver.Field.Name}}(parents []{{ pointer $fieldResolver.Parent.GoType false }}, authHeader wasm.AuthHeader) ([]{{ pointer $fieldResolver.Field.GoType $fieldResolver.Field.IsArray }}, error) { 
{{- body (printf "%s_%s" $fieldResolver.Parent.Name $fieldResolver.Field.Name) $.Rewriter }}}
{{ end }}

{{- range $key, $depBody := .Rewriter.DeprecatedBodies }}
{{ if and (not (is $key "Query_")) (not (is $key "Mutation_")) (not (is $key "Middleware_")) }}*/
/* {{ $depBody }} */ /*
{{ end }}
{{ end }}
`))

var GolangExecuterTemplate = template.Must(template.New("executer").Funcs(template.FuncMap{
	"pointer":    tools.Pointer,
	"marshal":    tools.Marshal,
	"jsonRefVal": tools.JsonRefVal,
}).Parse(`package generated

import (
	"github.com/schartey/dgraph-lambda-go/wasm"
	"github.com/valyala/fastjson"
)

type Executor struct {
	fieldResolver *resolvers.FieldResolver
}

func NewExecutor(resolver *resolvers.Resolver) *Executor {
	return &Executor{&resolvers.FieldResolver{Resolver: resolver}}
}

func (e *Executor) Resolve(request *wasm.Request) ([]byte, error) {
	switch request.Resolver {
	default:
		return e.resolveField(request)
	}
	return nil, nil
}

func (e *Executor) resolveField(request *wasm.Request) ([]byte, error) {
	switch request.Resolver {
		{{- range $fieldResolver := .FieldResolvers}}
		case "{{$fieldResolver.Parent.Name }}.{{$fieldResolver.Field.Name}}":
			{
				var parents []{{ pointer $fieldResolver.Parent.GoType false }}
				
				for _, e := range request.Parents {
					parents = append(parents, model.Unmarshal{{$fieldResolver.Parent.Name }}(e))
				}

				result, err := e.fieldResolver.{{$fieldResolver.Parent.Name }}_{{$fieldResolver.Field.Name}}( parents, request.AuthHeader)
				if err != nil {
					return nil, err
				}
				// Marshal result...
				return {{ marshal $fieldResolver.Field.GoType true }}, nil
			}
		{{- end }}
	}

	return nil, errors.New("resolver not found")
}

`))
