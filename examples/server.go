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

		// Global Middleware
		r.Use(func(hf resolver.HandlerFunc) resolver.HandlerFunc {
			return func(c context.Context, b []byte, parents []byte, ah resolver.AuthHeader) ([]byte, error) {
				// For example authentication.
				// Add user to context
				return hf(c, b, parents, ah)
			}
		})

		// Middleware on specific resolver
		r.UseOnResolver("Mutation.createUser", func(hf resolver.HandlerFunc) resolver.HandlerFunc {
			return func(c context.Context, b []byte, parents []byte, ah resolver.AuthHeader) ([]byte, error) {
				// For example authentication.
				// Add user to context
				return b, nil
			}
		})

		// Query/Mutation Resolver
		r.ResolveFunc("Mutation.createUser", func(ctx context.Context, input []byte, parents []byte, ah resolver.AuthHeader) ([]byte, error) {
			var createUserInput CreateUserInput
			json.Unmarshal(input, &createUserInput)

			// Do Something

			resp := `
			{
				"id": "0x1"	
			}`
			return ([]byte)(resp), nil
		})

		// Field Resolver
		r.ResolveFunc("User.complexProperty", func(ctx context.Context, input []byte, parents []byte, ah resolver.AuthHeader) ([]byte, error) {
			fmt.Println(string(parents))

			resp := `
			[ "complexPropertyValue" ]`
			return ([]byte)(resp), nil
		})
	})

	if err != nil {
		fmt.Println(err.Error())
	}
}
