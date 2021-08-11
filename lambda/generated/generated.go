package generated

import (
	"context"

	"github.com/schartey/dgraph-lambda-go/api"
	"github.com/schartey/dgraph-lambda-go/lambda"
)

type Executer struct {
}

func NewExecuter(l *lambda.Resolver) api.ExecuterInterface {
	return nil
}

func (e *Executer) Resolve(ctx context.Context, dbody api.DBody) ([]byte, error) {
	return nil, nil
}
