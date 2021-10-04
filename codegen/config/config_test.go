package config

import (
	"fmt"
	"path"
	"path/filepath"
	"testing"

	"github.com/schartey/dgraph-lambda-go/codegen/parser"
	"github.com/schartey/dgraph-lambda-go/internal"
	"github.com/stretchr/testify/assert"
)

var autobindValues = []string{"github.com/schartey/dgraph-lambda-go/examples/models"}
var models = []string{"User", "Author", "Apple", "Figure", "Hotel"}
var fieldResolvers = []string{"User.reputation", "User.rank", "User.active", "Post.additionalInfo", "Figure.size"}

var queryResolvers = []string{"getApples", "getTopAuthors", "getHotelByName"}

var mutationResolvers = []string{"newAuthor"}

var middlewareResolvers = []string{"user", "admin"}

func Test_LoadConfig(t *testing.T) {
	config, err := LoadConfigFile("github.com/schartey/dgraph-lambda-go", "../../lambda.yaml")
	assert.NoError(t, err)
	err = config.LoadConfig("../../lambda.yaml")
	assert.NoError(t, err)

	assert.Contains(t, filepath.ToSlash(config.SchemaFilename[0]), path.Join("dgraph-lambda-go", "examples", "test.graphql"))
	assert.Equal(t, "examples/lambda/generated/generated.go", config.Exec.Filename)
	assert.Equal(t, "generated", config.Exec.Package)
	assert.Equal(t, "examples/lambda/model/models_gen.go", config.Model.Filename)
	assert.Equal(t, "model", config.Model.Package)
	assert.Equal(t, "github.com/schartey/dgraph-lambda-go/examples/models", config.AutoBind[0])
	assert.Equal(t, "follow-schema", config.Resolver.Layout)
	assert.Equal(t, "examples/lambda/resolvers", config.Resolver.Dir)
	assert.Equal(t, "resolvers", config.Resolver.Package)
	assert.Equal(t, "{resolver}.resolver.go", config.Resolver.FilenameTemplate)
	assert.Equal(t, true, config.Server.Standalone)
	assert.Equal(t, "github.com/schartey/dgraph-lambda-go", config.Root)
	assert.NotNil(t, config.DefaultModelPackage)
	assert.Equal(t, "model", config.DefaultModelPackage.Name)
	assert.Equal(t, "github.com/schartey/dgraph-lambda-go/examples/lambda/model", config.DefaultModelPackage.PkgPath)
	assert.Equal(t, 3, len(config.Sources))
	assert.Contains(t, config.Sources[2].Name, "dgraph-lambda-go/examples/test.graphql")
}

func Test_LoadConfig_Fail(t *testing.T) {
	// Non existent file
	_, err := LoadConfigFile("github.com/schartey/dgraph-lambda-go", "./lambda.yaml")
	assert.Error(t, err)

	// Invalid file type
	_, err = LoadConfigFile("github.com/schartey/dgraph-lambda-go", "./config.go")
	assert.Error(t, err)

	for i := 1; i < 6; i++ {
		// Invalid file type
		_, err = LoadConfigFile("github.com/schartey/dgraph-lambda-go", fmt.Sprintf("../../test_resources/faulty%d.yaml", i))
		assert.Error(t, err)
	}
}

func Test_loadSchema(t *testing.T) {
	config, err := LoadConfigFile("github.com/schartey/dgraph-lambda-go", "../../lambda.yaml")
	assert.NoError(t, err)
	err = config.LoadConfig("../../lambda.yaml")
	assert.NoError(t, err)

	err = config.loadSchema()
	assert.NoError(t, err)
	assert.NotNil(t, config.Schema)
}

func Test_Config(t *testing.T) {
	moduleName, err := internal.GetModuleName()
	if err != nil {
		t.FailNow()
	}

	config, err := LoadConfigFile(moduleName, "../../lambda.yaml")
	if err != nil {
		fmt.Println(err.Error())
		t.FailNow()
	}
	err = config.LoadConfig("../../lambda.yaml")
	if err != nil {
		t.FailNow()
	}

	for _, value := range autobindValues {
		if !contains(value, config.AutoBind) {
			fmt.Println("Autobind Value missing: " + value)
			t.FailNow()
		}
	}
	// Check all values parsed from lambda.yaml

	if err := config.LoadSchema(); err != nil {
		fmt.Println(err.Error())
		t.FailNow()
	}

	/*	for _, m := range models {
			if !containsModel(m, config.ParsedTree.ModelTree.Models) {
				fmt.Println("Missing model after parsing: " + m)
				t.FailNow()
			}
		}

		for _, f := range fieldResolvers {
			if !containsFieldResolver(f, config.ParsedTree.ResolverTree.FieldResolvers) {
				fmt.Println("Missing field-resolver after parsing: " + f)
				t.FailNow()
			}
		}

		for _, q := range queryResolvers {
			if !containsQueryResolver(q, config.ParsedTree.ResolverTree.Queries) {
				fmt.Println("Missing query-resolver after parsing: " + q)
				t.FailNow()
			}
		}

		for _, m := range mutationResolvers {
			if !containsMutationResolver(m, config.ParsedTree.ResolverTree.Mutations) {
				fmt.Println("Missing mutation-resolver after parsing: " + m)
				t.FailNow()
			}
		}

		for _, m := range middlewareResolvers {
			if !containsMiddlewareResolver(m, config.ParsedTree.Middleware) {
				fmt.Println("Missing middleware resolver: " + m)
				t.FailNow()
			}
		}*/
}

func contains(s string, arr []string) bool {
	for _, v := range arr {
		if s == v {
			return true
		}
	}
	return false
}

func containsModel(modelName string, models map[string]*parser.Model) bool {
	for _, model := range models {
		if model.Name == modelName {
			return true
		}
	}
	return false
}

func containsFieldResolver(fieldResolverName string, fieldResolvers map[string]*parser.FieldResolver) bool {
	for _, fieldResolver := range fieldResolvers {
		if fmt.Sprintf("%s.%s", fieldResolver.Parent.Name, fieldResolver.Field.Name) == fieldResolverName {
			return true
		}
	}
	return false
}

func containsQueryResolver(queryResolverName string, queries map[string]*parser.Query) bool {
	for _, query := range queries {
		if query.Name == queryResolverName {
			return true
		}
	}
	return false
}

func containsMutationResolver(mutationResolverName string, mutations map[string]*parser.Mutation) bool {
	for _, mutation := range mutations {
		if mutation.Name == mutationResolverName {
			return true
		}
	}
	return false
}

func containsMiddlewareResolver(middlewareResolverName string, middleware map[string]string) bool {
	for _, m := range middleware {
		if m == middlewareResolverName {
			return true
		}
	}
	return false
}
