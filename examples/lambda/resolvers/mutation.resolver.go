package resolvers

import (
	"context"

	"github.com/schartey/dgraph-lambda-go/api"
)

type MutationResolverInterface interface {
	Mutation_newAuthor(ctx context.Context, name string, authHeader api.AuthHeader) (string, *api.LambdaError)
}

type MutationResolver struct {
	*Resolver
}

func (q *MutationResolver) Mutation_newAuthor(ctx context.Context, name string, authHeader api.AuthHeader) (string, *api.LambdaError) {
	return "", nil
}
