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
	"{{ $field.Name }}": {{ jsonRef $field.GoType }},
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

func MarshalTime(t *time.Time) string {
	if t == nil {
		return "null"
	}
	if v, err := t.MarshalJSON(); err != nil {
		fmt.Println(err)
		return "null"
	} else {
		return string(v)
	}
}

func UnmarshalTime(data []byte) *time.Time {
	if len(data) == 0 {
		return nil
	}
	var t time.Time
	if err := t.UnmarshalText(data); err != nil {
		fmt.Println(err)
		return nil
	} else {
		return &t
	}
}

`))

var GolangExecuterTemplate = template.Must(template.New("executer").Funcs(template.FuncMap{}).Parse(`package main

import (
	"unsafe"

	"github.com/schartey/dgraph-lambda-go/api"
	"github.com/schartey/dgraph-lambda-go/wasm"
	"github.com/valyala/fastjson"
)

/** This is the buffer that the host needs to write parameters into and read results from that are not int and float **/
var buf [2048]byte

//export getBuffer
func getBuffer() *byte {
	return &buf[0]
}

var resBuf [2048]byte

//export getResult
func getResult() *byte {
	return &resBuf[0]
}

// This is called on startup
func main() { 
	// resolver.start()
}

/** Run resolver - This should be generated **/
//export execute
func execute(ptr *byte, length int, res *byte) int {
	requestBuffer := buf[:length]

	request, err := unmarshalRequest(requestBuffer)
	if err != nil {
		wasm.Log(err.Error())
	}
	
	// Generate all resolvers
	return 0
}

func unmarshalRequest(requestBuffer [] byte) (*api.Request, error) {
	t := api.Request{}

	var p fastjson.Parser
	v, err := p.Parse(string(requestBuffer))
	if err != nil {
		return nil, err
	}
	t.AccessToken = string(v.GetStringBytes("X-Dgraph-AccessToken"))
	t.Args = make(map[string][]byte)
	t.Args["name"] = v.Get("args").GetStringBytes("name")

	t.AuthHeader = api.AuthHeader{
		Key:   string(v.Get("authHeader").GetStringBytes("key")),
		Value: string(v.Get("authHeader").GetStringBytes("value")),
	}

	t.Event = &api.Event{}

	t.Info = api.InfoField{}
	t.Info.Field.Alias = string(v.Get("info").Get("field").GetStringBytes("alias"))
	t.Info.Field.Name = string(v.Get("info").Get("field").GetStringBytes("name"))
	t.Info.Field.Arguments = v.Get("info").Get("field").GetStringBytes("arguments")
	t.Info.Field.Directives = []api.Directive{}
	t.Info.Field.SelectionSet = []api.SelectionField{}

	t.Parents = v.GetStringBytes("parents")

	t.Resolver = string(v.GetStringBytes("resolver"))

	return nil, nil
}
`))
