package config

import (
	"fmt"
	"testing"

	"github.com/schartey/dgraph-lambda-go/codegen/parser"
)

var moduleName = "test"
var autobindValues = []string{"github.com/schartey/dgraph-lambda-go/examples/models"}
var models = []string{"User", "Author", "Apple", "Figure", "Hotel"}
var fieldResolvers = []string{"User.reputation", "User.rank", "User.active", "Post.additionalInfo", "Figure.size"}

var queryResolvers = []string{"getApples", "getTopAuthors", "getHotelByName"}
var ignoreQueries = []string{"ignoredQuery"}

var mutationResolvers = []string{"newAuthor"}
var ignoreMutations = []string{"ignoredMutation"}

var middlewareResolvers = []string{"user", "admin"}

func Test_Config(t *testing.T) {

	config, err := LoadConfig("test", "../../test_resources/lamdba.yaml")
	if err != nil {
		fmt.Println(err.Error())
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

	for _, m := range models {
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
	for _, q := range ignoreQueries {
		if containsQueryResolver(q, config.ParsedTree.ResolverTree.Queries) {
			fmt.Println("Parsed non lambda query: " + q)
			t.FailNow()
		}
	}

	for _, m := range mutationResolvers {
		if !containsMutationResolver(m, config.ParsedTree.ResolverTree.Mutations) {
			fmt.Println("Missing mutation-resolver after parsing: " + m)
			t.FailNow()
		}
	}
	for _, m := range ignoreMutations {
		if containsMutationResolver(m, config.ParsedTree.ResolverTree.Mutations) {
			fmt.Println("Parsed non lambda mutation: " + m)
			t.FailNow()
		}
	}

	for _, m := range middlewareResolvers {
		if !containsMiddlewareResolver(m, config.ParsedTree.Middleware) {
			fmt.Println("Missing middleware resolver: " + m)
			t.FailNow()
		}
	}
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
		if fmt.Sprintf("%s.%s", fieldResolver.Field.ParentTypeName, fieldResolver.Field.Name) == fieldResolverName {
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
