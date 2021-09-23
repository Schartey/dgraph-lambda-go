package examples

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/schartey/dgraph-lambda-go/api"
	"github.com/schartey/dgraph-lambda-go/examples/lambda/generated"
	"github.com/schartey/dgraph-lambda-go/examples/lambda/resolvers"
)

func RunWithServer() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	resolver := &resolvers.Resolver{}
	executer := generated.NewExecuter(resolver)
	lambda := api.New(executer)
	srv, err := lambda.Serve(wg)
	if err != nil {
		fmt.Println(err)
	}
	<-c
	fmt.Println("Shutdown request (Ctrl-C) caught.")
	fmt.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		fmt.Println(err)
	}
	wg.Wait()
}

func RunWithRoute() {
	r := chi.NewRouter()

	resolver := &resolvers.Resolver{}
	executer := generated.NewExecuter(resolver)
	lambda := api.New(executer)

	r.Post("/graphql-worker", lambda.Route)

	fmt.Println("Lambda listening on 8686")
	fmt.Println(http.ListenAndServe(":8686", r))
}
