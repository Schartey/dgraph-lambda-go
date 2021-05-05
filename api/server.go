package api

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
	"github.com/schartey/dgraph-lambda-go/resolver"
	"google.golang.org/grpc"
)

func RunServer(setupResolver func(r *resolver.Resolver)) error {
	dgraphUrl := os.Getenv("DGRAPH_URL")

	/* Setup DQL-Client */
	conn, err := grpc.Dial(dgraphUrl, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer conn.Close()

	dql := dgo.NewDgraphClient(api.NewDgraphClient(conn))

	/* Setup Resolver */
	res := resolver.NewResolver(dql)
	setupResolver(res)

	/* Setup Router */
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/graphql-worker", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request")
		decoder := json.NewDecoder(r.Body)

		var dbody resolver.DBody
		err := decoder.Decode(&dbody)
		if err != nil {
			fmt.Println(err.Error())
		}

		response, err := res.Resolve(r.Context(), &dbody)
		if err != nil {
			fmt.Println(err.Error())
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
		}
		w.Write(response)
	})
	fmt.Println("Lambda listening on 8686")
	fmt.Println(http.ListenAndServe(":8686", r))

	return nil
}
