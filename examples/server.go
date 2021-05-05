package examples

import (
	"fmt"

	"github.com/schartey/dgraph-lambda-go/api"
	"github.com/schartey/dgraph-lambda-go/resolver"
)

func Run() {
	err := api.RunServer(func(r *resolver.Resolver) {
	})

	if err != nil {
		fmt.Println(err.Error())
	}
}
