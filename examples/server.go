package examples

import (
	"fmt"

	"github.com/dgraph-io/dgo/v210"
	"github.com/hasura/go-graphql-client"
	"github.com/schartey/dgraph-lambda-go/api"
	"github.com/schartey/dgraph-lambda-go/resolver"
)

func Run() {
	err := api.RunServer(func(r *resolver.Resolver, gql *graphql.Client, dql *dgo.Dgraph) {
	})

	if err != nil {
		fmt.Println(err.Error())
	}
}
