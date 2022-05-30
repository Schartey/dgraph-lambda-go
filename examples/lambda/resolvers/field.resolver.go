package resolvers

import (
	"context"

	"github.com/miko/dgraph-lambda-go/api"
	"github.com/miko/dgraph-lambda-go/examples/lambda/model"
)

type FieldResolverInterface interface {
	User_active(ctx context.Context, parents []*model.User, authHeader api.AuthHeader) ([]bool, *api.LambdaError)
	Post_additionalInfo(ctx context.Context, parents []*model.Post, authHeader api.AuthHeader) ([]string, *api.LambdaError)
	User_rank(ctx context.Context, parents []*model.User, authHeader api.AuthHeader) ([]int64, *api.LambdaError)
	User_reputation(ctx context.Context, parents []*model.User, authHeader api.AuthHeader) ([]int64, *api.LambdaError)
	Figure_size(ctx context.Context, parents []*model.Figure, authHeader api.AuthHeader) ([]int64, *api.LambdaError)
}

type FieldResolver struct {
	*Resolver
}

func (f *FieldResolver) User_active(ctx context.Context, parents []*model.User, authHeader api.AuthHeader) ([]bool, *api.LambdaError) {
	return nil, nil
}

func (f *FieldResolver) Post_additionalInfo(ctx context.Context, parents []*model.Post, authHeader api.AuthHeader) ([]string, *api.LambdaError) {
	return nil, nil
}

func (f *FieldResolver) User_rank(ctx context.Context, parents []*model.User, authHeader api.AuthHeader) ([]int64, *api.LambdaError) {
	return nil, nil
}

func (f *FieldResolver) User_reputation(ctx context.Context, parents []*model.User, authHeader api.AuthHeader) ([]int64, *api.LambdaError) {
	return nil, nil
}

func (f *FieldResolver) Figure_size(ctx context.Context, parents []*model.Figure, authHeader api.AuthHeader) ([]int64, *api.LambdaError) {
	return nil, nil
}
