package examples

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/schartey/dgraph-lambda-go/api"
	"github.com/schartey/dgraph-lambda-go/lambda"
	"github.com/schartey/dgraph-lambda-go/lambda/generated"
)

func RunWithServer() {
	resolver := &lambda.Resolver{}
	executer := generated.NewExecuter(resolver)
	lambda := api.New(executer)
	err := lambda.Serve()
	fmt.Println(err)
}

func RunWithRoute() {
	r := chi.NewRouter()

	resolver := &lambda.Resolver{}
	executer := generated.NewExecuter(resolver)
	lambda := api.New(executer)

	r.Post("/graphql-worker", lambda.Route)

	fmt.Println("Lambda listening on 8686")
	fmt.Println(http.ListenAndServe(":8686", r))
}
