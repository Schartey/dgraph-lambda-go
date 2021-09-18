package examples

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/schartey/dgraph-lambda-go/api"
	"github.com/schartey/dgraph-lambda-go/examples/lambda/generated"
	"github.com/schartey/dgraph-lambda-go/examples/lambda/resolvers"
)

func Test_Server(t *testing.T) {

	resolver := &resolvers.Resolver{}
	executer := generated.NewExecuter(resolver)
	lambda := api.New(executer)

	req := httptest.NewRequest(http.MethodPost, "/graphql-worker", nil)
	w := httptest.NewRecorder()
	lambda.Route(w, req)

	// We should get a good status code
	if want, got := http.StatusBadRequest, w.Result().StatusCode; want != got {
		t.Fatalf("expected a %d, instead got: %d", want, got)
	}

	fmt.Println("Testing example server")
}
