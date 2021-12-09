package generator

import (
	"os"
	"path"
	"text/template"

	"github.com/pkg/errors"
	c "github.com/schartey/dgraph-lambda-go/codegen/config"
	"github.com/schartey/dgraph-lambda-go/internal"
)

func GenerateConfig(filename string) error {
	if t, err := os.Open(filename); os.IsNotExist(err) {
		f, err := internal.CreateFile(filename)
		if err != nil {
			return errors.Wrap(err, "Could not create config: "+filename)
		}

		lambdaTemplate.Execute(f, struct{}{})
		f.Close()
	} else {
		t.Close()
	}
	return nil
}

func GenerateWorkspace(config *c.Config) error {
	if t, err := os.Open(config.Exec.Filename); os.IsNotExist(err) {
		f, err := internal.CreateFile(config.Exec.Filename)
		if err != nil {
			return errors.Wrap(err, "Could not create config: "+config.Exec.Filename)
		}
		template.Must(template.New("exec").Parse("package "+config.Exec.Package)).Execute(f, struct{}{})
		f.Close()
	} else {
		t.Close()
	}

	if t, err := os.Open(config.Model.Filename); os.IsNotExist(err) {
		f, err := internal.CreateFile(config.Model.Filename)
		if err != nil {
			return err
		}
		template.Must(template.New("model").Parse("package "+config.Model.Package)).Execute(f, struct{}{})
		f.Close()
	} else {
		t.Close()
	}

	if t, err := os.Open(config.Resolver.Dir); os.IsNotExist(err) {
		f, err := internal.CreateFile(config.Resolver.Dir)
		if err != nil {
			return err
		}
		f.Close()
	} else {
		t.Close()
	}

	// Generate Resolver
	if t, err := os.Open(path.Join(config.Resolver.Dir, "resolver.go")); os.IsNotExist(err) {
		f, err := internal.CreateFile(path.Join(config.Resolver.Dir, "resolver.go"))
		if err != nil {
			return err
		}

		resolverTemplate.Execute(f, struct {
			Package string
		}{
			Package: config.Resolver.Package,
		})
		f.Close()
	} else {
		t.Close()
	}

	// TODO: If lang is WASM, then this should become a server that runs a wasm instance!
	// We should probably split this one into gogen and wasm package as well
	if config.Server.Mode != c.WASM_ONLY {
		if t, err := os.Open("server.go"); os.IsNotExist(err) {
			f, err := internal.CreateFile("server.go")
			if err != nil {
				return errors.Wrap(err, "Could not create server.go")
			}

			err = serverTemplate.Execute(f, struct {
				ResolverPath     string
				ResolverPackage  string
				GeneratedPath    string
				GeneratedPackage string
				Mode             c.Mode
			}{
				ResolverPath:     path.Join(config.Root, config.Resolver.Dir),
				ResolverPackage:  config.Resolver.Package,
				GeneratedPath:    path.Join(config.Root, path.Dir(config.Exec.Filename)),
				GeneratedPackage: config.Exec.Package,
				Mode:             config.Server.Mode,
			})

			if err != nil {
				return errors.Wrap(err, "Could not execute template for server.go")
			}
			f.Close()
		} else {
			t.Close()
		}
	}
	return nil
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
  dir: lambda/resolvers
  package: resolvers
  filename_template: "{resolver}.resolver.go" # also allow "{name}.resolvers.go"

server:
  mode: server`))

var serverTemplate = template.Must(template.New("server").Parse(`package main

import (
	{{ if eq .Mode "server" }}
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"
	{{ else }}
	"fmt"
	"net/http"
	"github.com/go-chi/chi"
	{{ end }}

	"github.com/schartey/dgraph-lambda-go/api"
	"{{ .GeneratedPath }}"
	"{{ .ResolverPath }}"
)

func main() {
	{{ if eq .Mode "server" }}
	// Catch interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// WaitGroup for server shutdown
	wg := &sync.WaitGroup{}
	wg.Add(1)

	resolver := &{{ .ResolverPackage }}.Resolver{}
	executer := {{ .GeneratedPackage }}.NewExecuter(resolver)
	lambda := api.New(executer)
	srv, err := lambda.Serve(wg)
	if err != nil {
		fmt.Println(err)
	}
	
	// Interrupt signal received
	<-c
	fmt.Println("Shutdown request (Ctrl-C) caught.")
	fmt.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Println(err)
	}
	// Wait for server shutdown
	wg.Wait()
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
