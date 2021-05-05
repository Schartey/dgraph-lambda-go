package examples

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/dgo/v210"
	"github.com/hasura/go-graphql-client"
	"github.com/schartey/dgraph-lambda-go/api"
	"github.com/schartey/dgraph-lambda-go/resolver"
)

type CreateUserInput struct {
	Username string `json:"username"`
}

func Run() {
	err := api.RunServer(func(r *resolver.Resolver, gql *graphql.Client, dql *dgo.Dgraph) {

		r.Use(func(hf resolver.HandlerFunc) resolver.HandlerFunc {
			return func(c context.Context, b []byte, ah resolver.AuthHeader) ([]byte, error) {
				// For example authentication.
				// Add user to context
				return b, nil
			}
		})

		r.UseOnResolver("Mutation.createUser", func(hf resolver.HandlerFunc) resolver.HandlerFunc {
			return func(c context.Context, b []byte, ah resolver.AuthHeader) ([]byte, error) {
				// For example authentication.
				// Add user to context
				return b, nil
			}
		})

		r.ResolveFunc("Mutation.createUser", func(ctx context.Context, input []byte, ah resolver.AuthHeader) ([]byte, error) {
			var createUserInput CreateUserInput
			json.Unmarshal(input, &createUserInput)

			// Do Something

			resp := `
			{
				"id": "0x1"	
			}`
			return ([]byte)(resp), nil
		})
	})

	if err != nil {
		fmt.Println(err.Error())
	}
}
