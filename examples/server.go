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

type UserData struct {
	Id              string `json:"id"`
	Username        string `json:"username"`
	ComplexProperty string `json:"complexProperty"`
}

func Run() {
	err := api.RunServer(func(r *resolver.Resolver, gql *graphql.Client, dql *dgo.Dgraph) {

		// Global Middleware
		r.Use(func(hf resolver.HandlerFunc) resolver.HandlerFunc {
			return func(c context.Context, b []byte, parents []byte, ah resolver.AuthHeader) (interface{}, error) {
				// For example authentication.
				// Add user to context
				return hf(c, b, parents, ah)
			}
		})

		// Middleware on specific resolver
		r.UseOnResolver("Mutation.createUser", func(hf resolver.HandlerFunc) resolver.HandlerFunc {
			return func(c context.Context, b []byte, parents []byte, ah resolver.AuthHeader) (interface{}, error) {
				// For example authentication.
				// Add user to context
				return b, nil
			}
		})

		// Query/Mutation Resolver
		r.ResolveFunc("Mutation.createUser", func(ctx context.Context, input []byte, parents []byte, ah resolver.AuthHeader) (interface{}, error) {
			var createUserInput CreateUserInput
			json.Unmarshal(input, &createUserInput)

			// Do Something
			user := UserData{
				Id:       "0x1",
				Username: createUserInput.Username,
			}
			return user, nil
		})

		// Field Resolver
		r.ResolveFunc("UserData.complexProperty", func(ctx context.Context, input []byte, parents []byte, ah resolver.AuthHeader) (interface{}, error) {
			var userParents []UserData
			json.Unmarshal(parents, &userParents)

			var complexProperties []string
			for _, userParent := range userParents {
				complexProperties = append(complexProperties, fmt.Sprintf("VeryComplex - %s", userParent.Id))
			}

			return complexProperties, nil
		})

		// Webhook
		r.WebhookFunc("UserData", func(ctx context.Context, event resolver.Event) error {
			fmt.Println(event.Operation)

			return nil
		})
	})

	if err != nil {
		fmt.Println(err.Error())
	}
}
