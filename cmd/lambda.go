package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dgraph-io/dgo/v210"
	"github.com/dgraph-io/dgo/v210/protos/api"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"gitlab.com/trendsnap/trendgraph/dgraph-lambda-go/request"
	"gitlab.com/trendsnap/trendgraph/dgraph-lambda-go/resolver"
	"google.golang.org/grpc"
)

func main() {

	dgraphUrl := os.Getenv("DGRAPH_URL")
	fmt.Println(dgraphUrl)

	/* Setup DQL-Client */
	conn, err := grpc.Dial(dgraphUrl, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	dql := dgo.NewDgraphClient(api.NewDgraphClient(conn))

	/* Setup Resolver */
	res := resolver.NewResolver(dql)
	err = res.LoadPlugins("plugins")
	if err != nil {
		panic(err)
	}

	/* Setup Router */
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/graphql-worker", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request")
		decoder := json.NewDecoder(r.Body)

		var dbody request.DBody
		err := decoder.Decode(&dbody)
		if err != nil {
			fmt.Println(err.Error())
		}

		err = res.Resolve(r.Context(), &dbody)
		if err != nil {
			fmt.Println(err.Error())
		}
	})
	log.Println("listening on 8686")
	log.Fatal(http.ListenAndServe(":8686", r))
}
