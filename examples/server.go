package examples

import (
	"fmt"

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
	resolver := &lambda.Resolver{}
	executer := generated.NewExecuter(resolver)
	lambda := api.New(executer)
	err := lambda.Serve()
	fmt.Println(err)
}
