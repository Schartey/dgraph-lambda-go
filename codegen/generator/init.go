package generator

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/schartey/dgraph-lambda-go/codegen/config"
)

func GenerateConfig() {
	if t, err := os.Open("lambda.yaml"); os.IsNotExist(err) {
		f, err := createFile("lambda.yaml")
		if err != nil {
			panic(err)
		}

		lambdaTemplate.Execute(f, struct{}{})
		f.Close()
	} else {
		t.Close()
	}
}

func Init(config *config.Config) error {

	f, err := createFile(config.Exec.Filename)
	if err != nil {
		panic(err)
	}
	template.Must(template.New("exec").Parse("package "+config.Exec.Package)).Execute(f, struct{}{})
	f.Close()

	f, err = createFile(config.Model.Filename)
	if err != nil {
		panic(err)
	}
	template.Must(template.New("model").Parse("package "+config.Model.Package)).Execute(f, struct{}{})
	f.Close()
	f, err = createFile(config.Resolver.Dir)
	if err != nil {
		panic(err)
	}
	f.Close()

	// Generate Resolver
	f, err = createFile(path.Join(config.Resolver.Dir, "resolver.go"))
	if err != nil {
		panic(err)
	}

	resolverTemplate.Execute(f, struct {
		Package string
	}{
		Package: config.Resolver.Package,
	})
	f.Close()

	f, err = os.Create("server.go")
	if err != nil {
		fmt.Println(err.Error())
	}

	serverTemplate.Execute(f, struct {
		ResolverPath     string
		ResolverPackage  string
		GeneratedPath    string
		GeneratedPackage string
		Standalone       bool
	}{
		ResolverPath:     path.Join(config.Root, config.Resolver.Dir),
		ResolverPackage:  config.Resolver.Package,
		GeneratedPath:    path.Join(config.Root, path.Dir(config.Exec.Filename)),
		GeneratedPackage: config.Exec.Package,
		Standalone:       config.Server.Standalone,
	})
	f.Close()

	return nil
}

func createFile(p string) (*os.File, error) {
	path := p
	file := ""
	if strings.Contains(filepath.Base(p), ".") {
		path = filepath.Dir(p)
		file = filepath.Base(p)
	}

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		print(err.Error())
		return nil, err
	}
	if file == "" {
		return nil, nil
	}
	return os.Create(p)
}

var resolverTemplate = template.Must(template.New("resolver").Parse(`package {{ .Package }}

// Add objects to your desire
type Resolver struct {
}`))

var lambdaTemplate = template.Must(template.New("lambda").Parse(`schema:
  - ./*.graphql

exec:
  filename: lambda/generated/generated.go
  package: generated

model:
  filename: lambda/model/models_gen.go
  package: model

autobind:
  # - "github.com/schartey/dgraph-lambda-go/examples/models"

resolver:
  layout: follow-schema
  dir: lambda/resolvers
  package: resolvers
  filename_template: "{resolver}.resolver.go" # also allow "{name}.resolvers.go"

server:
  standalone: true`))

var serverTemplate = template.Must(template.New("server").Parse(`package main

import (
	"fmt"

	{{ if not .Standalone }}
	"net/http"
	"github.com/go-chi/chi"
	{{ end }}

	"github.com/schartey/dgraph-lambda-go/api"
	"{{ .GeneratedPath }}"
	"{{ .ResolverPath }}"
)

func main() {
	{{ if .Standalone }}
	resolver := &{{ .ResolverPackage }}.Resolver{}
	executer := {{ .GeneratedPackage }}.NewExecuter(resolver)
	lambda := api.New(executer)
	err := lambda.Serve()
	fmt.Println(err)
	{{ else }}
	r := chi.NewRouter()

	resolver := &{{ .ResolverPackage }}.Resolver{}
	executer := {{ .GeneratedPackage }}.NewExecuter(resolver)
	lambda := api.New(executer)

	r.Post("/graphql-worker", lambda.Route)

	fmt.Println("Lambda listening on 8686")
	fmt.Println(http.ListenAndServe(":8686", r))

	{{ end }}
}
`))
