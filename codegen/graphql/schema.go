package graphql

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

func SchemaLoaderFromExternal(host string) (*ast.Schema, error) {
	return nil, errors.New("Not implemented")
}

func SchemaLoaderFromFile(path string) (*ast.Schema, error) {
	schemaFile, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open schema")
	}
	schemaFile.Close()

	schemaInput, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read schema")
	}

	var sources []*ast.Source

	sources = append(sources, &ast.Source{Input: SchemaInputs + DirectiveDefs})
	sources = append(sources, &ast.Source{Name: schemaFile.Name(), Input: string(schemaInput)})

	schema := gqlparser.MustLoadSchema(sources...)

	return schema, nil
}
