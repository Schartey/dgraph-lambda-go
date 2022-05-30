package resolvers

import (
	"context"

	"github.com/miko/dgraph-lambda-go/api"
	"github.com/miko/dgraph-lambda-go/examples/lambda/model"
)

type QueryResolverInterface interface {
	Query_getApples(ctx context.Context, authHeader api.AuthHeader) ([]*model.Apple, *api.LambdaError)
	Query_getHotelByName(ctx context.Context, name string, authHeader api.AuthHeader) (*model.Hotel, *api.LambdaError)
	Query_getTopAuthors(ctx context.Context, id string, authHeader api.AuthHeader) ([]*model.Author, *api.LambdaError)
}

type QueryResolver struct {
	*Resolver
}

func (q *QueryResolver) Query_getApples(ctx context.Context, authHeader api.AuthHeader) ([]*model.Apple, *api.LambdaError) {
	return nil, nil
}

func (q *QueryResolver) Query_getHotelByName(ctx context.Context, name string, authHeader api.AuthHeader) (*model.Hotel, *api.LambdaError) {
	return nil, nil
}

func (q *QueryResolver) Query_getTopAuthors(ctx context.Context, id string, authHeader api.AuthHeader) ([]*model.Author, *api.LambdaError) {
	return nil, nil
}
