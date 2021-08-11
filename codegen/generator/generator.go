package generator

import (
	"fmt"
	"go/types"
	"os"
	"path"
	"strings"
	"text/template"
	"unicode"

	"github.com/schartey/dgraph-lambda-go/codegen/config"
	"github.com/schartey/dgraph-lambda-go/codegen/parser"
	"github.com/schartey/dgraph-lambda-go/codegen/rewriter"
	"golang.org/x/tools/go/packages"
)

func Generate(c *config.Config, r *rewriter.Rewriter) error {

	test = c.AutoBind
	defaultPackage = c.DefaultModelPackage

	f, err := os.Create(c.Model.Filename)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer f.Close()

	var pkgs = make(map[string]*types.Package)
	var models = make(map[string]*parser.Model)
	var enums = make(map[string]*parser.Enum)
	var interfaces = make(map[string]*parser.Interface)
	var scalars = make(map[string]*parser.Scalar)

	for _, m := range c.ParsedTree.ModelTree.Models {
		if m.GoType.TypeName.Pkg().Path() == c.DefaultModelPackage.PkgPath {
			models[m.Name] = m
		}
		for _, f := range m.Fields {
			if f.TypeName.Exported() && f.GoType.TypeName.Pkg().Path() != c.DefaultModelPackage.PkgPath {
				pkgs[f.GoType.TypeName.Pkg().Name()] = f.GoType.TypeName.Pkg()
			}
		}
	}
	for _, m := range c.ParsedTree.ModelTree.Enums {
		if m.TypeName.Exported() && m.GoType.TypeName.Pkg().Path() == c.DefaultModelPackage.PkgPath {
			enums[m.Name] = m
		}
	}
	for _, m := range c.ParsedTree.ModelTree.Interfaces {
		if m.TypeName.Exported() && m.GoType.TypeName.Pkg().Path() == c.DefaultModelPackage.PkgPath {
			interfaces[m.Name] = m
		}
	}
	for _, m := range c.ParsedTree.ModelTree.Scalars {
		if m.TypeName.Exported() && m.GoType.TypeName.Pkg().Path() == c.DefaultModelPackage.PkgPath {
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
		Scalars:     c.ParsedTree.ModelTree.Scalars,
		Models:      models,
		Packages:    pkgs,
		PackageName: c.Model.Package,
	})
	if err != nil {
		print(err.Error())
	}

	resolverFileTemplate := config.ResolverTemplateRegex.FindStringSubmatch(c.Resolver.FilenameTemplate)

	if resolverFileTemplate[1] == "resolver" {

		fileName := path.Join(c.Resolver.Dir, "field.resolver.go")
		resolverFile, err := os.Create(fileName)
		if err != nil {
			fmt.Println(err.Error())
		}

		var pkgs = make(map[string]*types.Package)

		for _, m := range c.ParsedTree.ResolverTree.FieldResolvers {
			if c.DefaultModelPackage.PkgPath != path.Join(c.Root, c.Resolver.Dir) {
				if m.Field.TypeName.Exported() {
					pkgs[m.Field.TypeName.Pkg().Name()] = m.Field.TypeName.Pkg()
				}
			} else {
				if m.Field.TypeName.Exported() && m.Field.TypeName.Pkg().Path() != c.DefaultModelPackage.PkgPath {
					pkgs[m.Field.TypeName.Pkg().Name()] = m.Field.TypeName.Pkg()
				}
			}
		}
		if len(c.ParsedTree.ResolverTree.FieldResolvers) > 0 {
			pkgs["context"] = types.NewPackage("context", "context")
			pkgs["api"] = types.NewPackage("github.com/schartey/dgraph-lambda-go/api", "api")
		}

		// TOOD: Fields are not unique I think
		err = fieldResolverTemplate.Execute(resolverFile, struct {
			FieldResolvers map[string]*parser.FieldResolver2
			Rewriter       *rewriter.Rewriter
			Packages       map[string]*types.Package
			PackageName    string
		}{
			FieldResolvers: c.ParsedTree.ResolverTree.FieldResolvers,
			Rewriter:       r,
			Packages:       pkgs,
			PackageName:    c.Resolver.Package,
		})
		if err != nil {
			fmt.Println(err.Error())
		}
		resolverFile.Close()

		fileName = path.Join(c.Resolver.Dir, "query.resolver.go")
		resolverFile, err = os.Create(fileName)
		if err != nil {
			fmt.Println(err.Error())
		}

		pkgs = make(map[string]*types.Package)

		for _, m := range c.ParsedTree.ResolverTree.Queries {
			if m.Return.TypeName.Exported() {
				pkgs[m.Return.TypeName.Pkg().Name()] = m.Return.TypeName.Pkg()
			}

			for _, f := range m.Arguments {
				if f.TypeName.Exported() {
					pkgs[f.GoType.TypeName.Pkg().Name()] = f.GoType.TypeName.Pkg()
				}
			}
		}
		if len(c.ParsedTree.ResolverTree.Queries) > 0 {
			pkgs["context"] = types.NewPackage("context", "context")
			pkgs["api"] = types.NewPackage("github.com/schartey/dgraph-lambda-go/api", "api")
		}

		err = queryResolverTemplate.Execute(resolverFile, struct {
			QueryResolvers map[string]*parser.Query2
			Rewriter       *rewriter.Rewriter
			Packages       map[string]*types.Package
			PackageName    string
		}{
			QueryResolvers: c.ParsedTree.ResolverTree.Queries,
			Rewriter:       r,
			Packages:       pkgs,
			PackageName:    c.Resolver.Package,
		})
		if err != nil {
			fmt.Println(err.Error())
		}
		resolverFile.Close()

		fileName = path.Join(c.Resolver.Dir, "mutation.resolver.go")
		resolverFile, err = os.Create(fileName)
		if err != nil {
			fmt.Println(err.Error())
		}

		pkgs = make(map[string]*types.Package)

		for _, m := range c.ParsedTree.ResolverTree.Mutations {
			if m.Return.TypeName.Exported() {
				pkgs[m.Return.TypeName.Pkg().Name()] = m.Return.TypeName.Pkg()
			}

			for _, f := range m.Arguments {
				if f.TypeName.Exported() {
					pkgs[f.GoType.TypeName.Pkg().Name()] = f.GoType.TypeName.Pkg()
				}
			}
		}
		if len(c.ParsedTree.ResolverTree.Mutations) > 0 {
			pkgs["context"] = types.NewPackage("context", "context")
			pkgs["api"] = types.NewPackage("github.com/schartey/dgraph-lambda-go/api", "api")
		}

		err = mutationResolverTemplate.Execute(resolverFile, struct {
			MutationResolvers map[string]*parser.Mutation2
			Rewriter          *rewriter.Rewriter
			Packages          map[string]*types.Package
			PackageName       string
		}{
			MutationResolvers: c.ParsedTree.ResolverTree.Mutations,
			Rewriter:          r,
			Packages:          pkgs,
			PackageName:       c.Resolver.Package,
		})
		if err != nil {
			fmt.Println(err.Error())
		}
		resolverFile.Close()

		fileName = path.Join(c.Resolver.Dir, "middleware.resolver.go")
		resolverFile, err = os.Create(fileName)
		if err != nil {
			fmt.Println(err.Error())
		}

		pkgs = make(map[string]*types.Package)

		if len(c.ParsedTree.Middleware) > 0 {
			pkgs["api"] = types.NewPackage("github.com/schartey/dgraph-lambda-go/api", "api")
		}

		err = middlewareResolverTemplate.Execute(resolverFile, struct {
			Middleware  map[string]string
			Rewriter    *rewriter.Rewriter
			Packages    map[string]*types.Package
			PackageName string
		}{
			Middleware:  c.ParsedTree.Middleware,
			Rewriter:    r,
			Packages:    pkgs,
			PackageName: c.Resolver.Package,
		})
		if err != nil {
			fmt.Println(err.Error())
		}
		resolverFile.Close()

		fileName = path.Join(c.Resolver.Dir, "webhook.resolver.go")
		resolverFile, err = os.Create(fileName)
		if err != nil {
			fmt.Println(err.Error())
		}

		pkgs = make(map[string]*types.Package)

		var models = make(map[string]*parser.Model)

		for _, m := range c.ParsedTree.ModelTree.Models {
			if len(m.LambdaOnMutate) > 0 {
				models[m.Name] = m
				//pkgs[m.TypeName.Pkg().Name()] = m.TypeName.Pkg()
			}
		}

		if len(models) > 0 {
			pkgs["context"] = types.NewPackage("context", "context")
			pkgs["api"] = types.NewPackage("github.com/schartey/dgraph-lambda-go/api", "api")
		}

		err = webhookResolverTemplate.Execute(resolverFile, struct {
			Models      map[string]*parser.Model
			Rewriter    *rewriter.Rewriter
			Packages    map[string]*types.Package
			PackageName string
		}{
			Models:      models,
			Rewriter:    r,
			Packages:    pkgs,
			PackageName: c.Resolver.Package,
		})
		if err != nil {
			fmt.Println(err.Error())
		}
		f.Close()

		resolverFile, err = os.Create(c.Exec.Filename)
		if err != nil {
			fmt.Println(err.Error())
		}

		pkgs = make(map[string]*types.Package)

		for _, m := range c.ParsedTree.ResolverTree.FieldResolvers {
			if m.Field.TypeName.Exported() && m.Field.TypeName.Pkg().Path() != c.DefaultModelPackage.PkgPath {
				pkgs[m.Field.TypeName.Pkg().Name()] = m.Field.TypeName.Pkg()
			}
		}

		for _, m := range c.ParsedTree.ResolverTree.Queries {
			if m.Return.TypeName.Exported() {
				pkgs[m.Return.TypeName.Pkg().Name()] = m.Return.TypeName.Pkg()
			}

			for _, f := range m.Arguments {
				if f.TypeName.Exported() {
					pkgs[f.GoType.TypeName.Pkg().Name()] = f.GoType.TypeName.Pkg()
				}
			}
		}

		for _, m := range c.ParsedTree.ResolverTree.Mutations {
			if m.Return.TypeName.Exported() {
				pkgs[m.Return.TypeName.Pkg().Name()] = m.Return.TypeName.Pkg()
			}

			for _, f := range m.Arguments {
				if f.TypeName.Exported() {
					pkgs[f.GoType.TypeName.Pkg().Name()] = f.GoType.TypeName.Pkg()
				}
			}
		}

		pkgs["context"] = types.NewPackage("context", "context")
		pkgs["errors"] = types.NewPackage("errors", "errors")
		pkgs["api"] = types.NewPackage("github.com/schartey/dgraph-lambda-go/api", "api")
		pkgs["api"] = types.NewPackage("github.com/schartey/dgraph-lambda-go/api", "api")
		pkgs["json"] = types.NewPackage("encoding/json", "json")

		pkgs[c.Resolver.Package] = types.NewPackage(path.Join(c.Root, c.Resolver.Dir), c.Resolver.Package)

		for _, model := range c.ParsedTree.ModelTree.Models {
			fmt.Println(len(model.LambdaOnMutate))
		}

		err = executerTemplate.Execute(resolverFile, struct {
			FieldResolvers      map[string]*parser.FieldResolver2
			Queries             map[string]*parser.Query2
			Mutations           map[string]*parser.Mutation2
			Middleware          map[string]string
			Models              map[string]*parser.Model
			Packages            map[string]*types.Package
			PackageName         string
			ResolverPackageName string
		}{
			FieldResolvers:      c.ParsedTree.ResolverTree.FieldResolvers,
			Queries:             c.ParsedTree.ResolverTree.Queries,
			Mutations:           c.ParsedTree.ResolverTree.Mutations,
			Middleware:          c.ParsedTree.Middleware,
			Models:              c.ParsedTree.ModelTree.Models,
			Packages:            pkgs,
			PackageName:         c.Exec.Package,
			ResolverPackageName: c.Resolver.Package,
		})
		if err != nil {
			fmt.Println(err.Error())
		}
		resolverFile.Close()
	} else {
		/*for _, fieldResolver := range c.ParsedTree.ResolverTree.FieldResolvers {
			fileName := path.Join(c.Resolver.Dir, fieldResolver.Field.Name+"Resolver.go")
			resolverFile, err := os.Create(fileName)
			if err != nil {
				fmt.Println(err.Error())
			}

			err = fieldResolverTemplate.Execute(resolverFile, struct {
				FieldResolvers *parser.FieldResolver2
			}{
				FieldResolvers: fieldResolver,
			})
			resolverFile.Close()
		}*/

	}

	return nil
}

func Model_Ref(t *parser.GoType) string {
	for _, te := range test {
		if te == t.TypeName.Pkg().Path() {
			return fmt.Sprintf("%s.%s", t.TypeName.Pkg().Name(), t.TypeName.Name())
		}
	}
	if t.TypeName.Exported() && t.TypeName.Pkg().Path() != defaultPackage.PkgPath {
		return fmt.Sprintf("%s.%s", t.TypeName.Pkg().Name(), t.TypeName.Name())
	}
	return t.TypeName.Name()
}

func Resolver_Ref(t *parser.GoType) string {
	for _, te := range test {
		if te == t.TypeName.Pkg().Path() {
			return fmt.Sprintf("%s.%s", t.TypeName.Pkg().Name(), t.TypeName.Name())
		}
	}
	if t.TypeName.Exported() {
		return fmt.Sprintf("%s.%s", t.TypeName.Pkg().Name(), t.TypeName.Name())
	}
	return t.TypeName.Name()
}

func Path(t *types.Package) string {
	return t.Path()
}

func Title(t string) string {
	return strings.Title(t)
}

func Untitle(s string) string {
	if len(s) == 0 {
		return s
	}

	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func TypeName(t *types.TypeName) string {
	return t.Name()
}

func Pointer(t *parser.GoType) string {
	if !t.TypeName.Exported() {
		return t.TypeName.Name()
	} else {
		return fmt.Sprintf("*%s", Resolver_Ref(t))
	}
}

func Args(args []*parser.Argument2) string {
	var arglist []string

	for _, arg := range args {
		arglist = append(arglist, fmt.Sprintf("%s", arg.Name))
	}
	return strings.Join(arglist, ",")
}

func ArgsW(args []*parser.Argument2) string {
	var arglist []string

	for _, arg := range args {
		arglist = append(arglist, fmt.Sprintf("%s %s", arg.Name, Pointer(arg.GoType)))
	}
	return strings.Join(arglist, ",")
}

func Body(key string, rewriter *rewriter.Rewriter) string {
	if val, ok := rewriter.RewriteBodies[key]; ok {
		return val
	} else {
		return `
	return nil, nil
`
	}
}

func MiddlewareBody(key string, rewriter *rewriter.Rewriter) string {
	if val, ok := rewriter.RewriteBodies[key]; ok {
		return val
	} else {
		return `
	return nil
`
	}
}

func Is(key string, resolverType string) bool {
	return strings.HasPrefix(key, resolverType)
}

var test []string
var defaultPackage *packages.Package
var modelTemplate = template.Must(template.New("model").Funcs(template.FuncMap{
	"ref":   Model_Ref,
	"path":  Path,
	"title": Title,
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
		{{ $field.Name | title }} {{$field.GoType | ref }} ` + "`{{$field.Tag}}`" + `
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

var fieldResolverTemplate = template.Must(template.New("field-resolver").Funcs(template.FuncMap{
	"ref":     Model_Ref,
	"path":    Path,
	"pointer": Pointer,
	"body":    Body,
	"is":      Is,
}).Parse(`
package {{.PackageName}}

import(
	{{- range $pkg := .Packages }}
	"{{ $pkg | path }}"{{- end}}
)

/** Put these into resolvers.go  or similar **/
type FieldResolver struct {
	*Resolver
}

{{- range $fieldResolver := .FieldResolvers}}
func (f *FieldResolver) {{$fieldResolver.Field.ParentTypeName }}_{{$fieldResolver.Field.Name}}(ctx context.Context, parents []{{$fieldResolver.Field.GoType | pointer}}, authHeader api.AuthHeader) ([]{{$fieldResolver.Field.GoType | pointer}}, error) { {{ body (printf "%s_%s" $fieldResolver.Field.ParentTypeName $fieldResolver.Field.Name) $.Rewriter }}}
{{ end }}

{{- range $key, $depBody := .Rewriter.DeprecatedBodies }}
{{ if and (not (is $key "Query_")) (not (is $key "Mutation_")) (not (is $key "Middleware_")) }}
/* {{ $depBody }} */
{{ end }}
{{ end }}
`))

var queryResolverTemplate = template.Must(template.New("query-resolver").Funcs(template.FuncMap{
	"ref":     Resolver_Ref,
	"path":    Path,
	"pointer": Pointer,
	"argsW":   ArgsW,
	"body":    Body,
	"is":      Is,
}).Parse(`
package {{.PackageName}}

import(
	{{- range $pkg := .Packages }}
	"{{ $pkg | path }}"{{- end}}
)

/** Put these into resolvers.go  or similar **/
type QueryResolver struct {
	*Resolver
}

{{- range $queryResolver := .QueryResolvers}}
func (q *QueryResolver) Query_{{$queryResolver.Name}}(ctx context.Context, {{ $queryResolver.Arguments | argsW }}, authHeader api.AuthHeader) ({{$queryResolver.Return | pointer}}, error) { {{ body (printf "Query_%s" $queryResolver.Name) $.Rewriter }}}
{{ end }}

{{- range $key, $depBody := .Rewriter.DeprecatedBodies }}
{{ if is $key "Query_" }}
/* {{ $depBody }} */
{{ end }}
{{ end }}
`))

var mutationResolverTemplate = template.Must(template.New("mutation-resolver").Funcs(template.FuncMap{
	"ref":     Resolver_Ref,
	"path":    Path,
	"pointer": Pointer,
	"argsW":   ArgsW,
	"body":    Body,
	"is":      Is,
}).Parse(`
package {{.PackageName}}

import(
	{{- range $pkg := .Packages }}
	"{{ $pkg | path }}"{{- end}}
)

/** Put these into resolvers.go  or similar **/
type MutationResolver struct {
	*Resolver
}

{{- range $mutationResolver := .MutationResolvers}}
func (q *MutationResolver) Mutation_{{$mutationResolver.Name}}(ctx context.Context, {{ $mutationResolver.Arguments | argsW }}, authHeader api.AuthHeader) ({{$mutationResolver.Return | pointer}}, error) { {{ body (printf "Mutation_%s" $mutationResolver.Name) $.Rewriter }}}
{{ end }}

{{- range $key, $depBody := .Rewriter.DeprecatedBodies }}
{{ if is $key "Mutation_" }}
/* {{ $depBody }} */
{{ end }}
{{ end }}
`))

var middlewareResolverTemplate = template.Must(template.New("middleware-resolver").Funcs(template.FuncMap{
	"path": Path,
	"body": MiddlewareBody,
	"is":   Is,
}).Parse(`
package {{.PackageName}}

import(
	{{- range $pkg := .Packages }}
	"{{ $pkg | path }}"{{- end}}
)

/** Put these into resolvers.go  or similar **/
type MiddlewareResolver struct {
	*Resolver
}

{{ range $middleware := .Middleware}}
func (m *MiddlewareResolver) Middleware_{{$middleware}}(md *api.MiddlewareData) error { {{ body (printf "Middleware_%s" $middleware) $.Rewriter }}}
{{ end }}

{{- range $key, $depBody := .Rewriter.DeprecatedBodies }}
{{ if is $key "Middleware_" }}
/* {{ $depBody }} */
{{ end }}
{{ end }}
`))

var webhookResolverTemplate = template.Must(template.New("webhook-resolver").Funcs(template.FuncMap{
	"path":     Path,
	"body":     MiddlewareBody,
	"typeName": TypeName,
	"is":       Is,
}).Parse(`
package {{.PackageName}}

import(
	{{- range $pkg := .Packages }}
	"{{ $pkg | path }}"{{- end}}
)

/** Put these into resolvers.go  or similar **/
type WebhookResolver struct {
	*Resolver
}

{{ range $model := .Models}}
func (w *WebhookResolver) Webhook_{{ $model.TypeName | typeName }}(ctx context.Context, event api.Event) error { {{ body (printf "Webhook_%s" ($model.TypeName | typeName)) $.Rewriter }}}
{{ end }}
`))

var executerTemplate = template.Must(template.New("executer").Funcs(template.FuncMap{
	"path":     Path,
	"typeName": TypeName,
	"ref":      Resolver_Ref,
	"untitle":  Untitle,
	"args":     Args,
	"pointer":  Pointer,
}).Parse(`
package {{.PackageName}}

import(
	{{- range $pkg := .Packages }}
	"{{ $pkg | path }}"{{- end}}
)

type Executer struct {
	fieldResolver    	{{.ResolverPackageName}}.FieldResolver
	queryResolver    	{{.ResolverPackageName}}.QueryResolver
	mutationResolver 	{{.ResolverPackageName}}.MutationResolver
	middlewareResolver 	{{.ResolverPackageName}}.MiddlewareResolver
	webhookResolver 	{{.ResolverPackageName}}.WebhookResolver
}

func NewExecuter(resolver *{{.ResolverPackageName}}.Resolver) api.ExecuterInterface {
	return Executer{fieldResolver: {{.ResolverPackageName}}.FieldResolver{Resolver: resolver}, queryResolver: {{.ResolverPackageName}}.QueryResolver{Resolver: resolver}, middlewareResolver: {{.ResolverPackageName}}.MiddlewareResolver{Resolver: resolver}, webhookResolver: {{.ResolverPackageName}}.WebhookResolver{Resolver: resolver}}
}

func (e *Executer) Middleware(md *api.MiddlewareData) error {
	var err error
	switch md.Dbody.Resolver {
		{{- range $fieldResolver := .FieldResolvers}}
		case "{{$fieldResolver.Field.TypeName | typeName }}.{{$fieldResolver.Field.Name}}":
			{
				{{- range $middleware := $fieldResolver.Middleware}}
				if err = e.middlewareResolver.Middleware_{{$middleware}}(md); err != nil {
					return err
				}
				{{- end}}
				break
			}
		{{- end }}

		{{- range $query := .Queries}}
		case "Query.{{$query.Name}}":
			{
				{{- range $middleware := $query.Middleware}}
				if err = e.middlewareResolver.Middleware_{{$middleware}}(md); err != nil {
					return err
				}
				{{- end}}
				break
			}
		{{- end }}
		{{- range $mutation := .Mutations}}
		case "Mutation.{{$mutation.Name}}":
			{
				{{- range $middleware := $mutation.Middleware}}
				if err = e.middlewareResolver.Middleware_{{$middleware}}(md); err != nil {
					return err
				}
				{{- end}}
				break
			}
		{{- end }}
	}
	return nil
}

func (e Executer) Resolve(ctx context.Context, dbody api.DBody) ([]byte, error) {
	if &dbody.Event != nil {
		var err error
		switch dbody.Event.TypeName {
			{{- range $model := .Models}} {{ if ne (len $model.LambdaOnMutate) 0 }}
			case "{{ $model.TypeName | typeName }}":
				err = e.webhookResolver.Webhook_{{ $model.TypeName | typeName }}(ctx, dbody.Event)
				return nil, err
			{{ end }} {{ end }}
		}
	} else {
		parentsBytes, err := dbody.Parents.MarshalJSON()
		if err != nil {
			return nil, err
		}

		md := &api.MiddlewareData{Ctx: ctx, Dbody: dbody}
		if err = e.Middleware(md); err != nil {
			return nil, err
		}
		ctx = md.Ctx
		dbody = md.Dbody

		response := []byte{}

		switch dbody.Resolver {
			{{- range $fieldResolver := .FieldResolvers}}
			case "{{$fieldResolver.Field.ParentTypeName }}.{{$fieldResolver.Field.Name}}":
				{
					var parents []{{$fieldResolver.Field.GoType | pointer }}
					json.Unmarshal(parentsBytes, &parents)

					// Dependent on generation loop or just direct
					/*var {{$fieldResolver.Field.Name}}s []{{$fieldResolver.Field.GoType | ref}}
					for _, parent := range parents {
						{{$fieldResolver.Field.Name}}s = fullnames.append(e.fieldResolver.{{$fieldResolver.Field.GoType | ref }}_{{$fieldResolver.Field.Name}}(ctx, parent))
					}*/
					{{$fieldResolver.Field.Name | untitle }}s, err := e.fieldResolver.{{$fieldResolver.Field.ParentTypeName }}_{{$fieldResolver.Field.Name}}(ctx, parents, dbody.AuthHeader)
					if err != nil {
						return nil, err
					}

					response, err = json.Marshal({{$fieldResolver.Field.Name | untitle }}s)
					if err != nil {
						return nil, err
					}
					break
				}
			{{- end }}

			{{- range $query := .Queries}}
			case "Query.{{$query.Name}}":
				{
					{{- range $arg := $query.Arguments }}
					var {{ $arg.Name }} {{ $arg.GoType | pointer }} 
					json.Unmarshal(dbody.Args["{{$arg.Name}}"], &{{$arg.Name}})
					{{- end }}	
					{{ $query.Return.TypeName | typeName | untitle }}, err := e.queryResolver.Query_{{$query.Name}}(ctx, {{$query.Arguments | args}}, dbody.AuthHeader)
					if err != nil {
						return nil, err
					}
		
					response, err = json.Marshal({{ $query.Return.TypeName | typeName | untitle }})
					if err != nil {
						return nil, err
					}
					break
				}
			{{- end }}
			{{- range $mutation := .Mutations}}
			case "Mutation.{{$mutation.Name}}":
				{
					{{- range $arg := $mutation.Arguments }}
					var {{ $arg.Name }} {{ $arg.GoType | pointer }} 
					json.Unmarshal(dbody.Args["{{$arg.Name}}"], &{{$arg.Name}})
					{{- end }}	
					{{ $mutation.Return.TypeName | typeName | untitle }}, err := e.mutationResolver.Mutation_{{$mutation.Name}}(ctx, {{$mutation.Arguments | args}}, dbody.AuthHeader)
					if err != nil {
						return nil, err
					}
		
					response, err = json.Marshal({{ $mutation.Return.TypeName | typeName | untitle }})
					if err != nil {
						return nil, err
					}
					break
				}
			{{- end }}
		}
		return response, nil
	}
	return nil, errors.New("No resolver found")
}
`))
