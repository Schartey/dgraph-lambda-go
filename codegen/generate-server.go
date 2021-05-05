package codegen

import (
	"fmt"
	"html/template"
	"os"
)

func GenerateServer() error {
	f, err := os.Create("server.go")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer f.Close()

	var data struct{}
	serverTemplate.Execute(f, data)

	return nil
}

var serverTemplate = template.Must(template.New("server").Parse(`

package main

import(
	"fmt"

	"github.com/dgraph-io/dgo/v210"
	"github.com/hasura/go-graphql-client"
	"github.com/schartey/dgraph-lambda-go/api"
	"github.com/schartey/dgraph-lambda-go/resolver"
)

func main() {
	api.RunServer(func(r *resolver.Resolver) {
		
	})
}
`))
