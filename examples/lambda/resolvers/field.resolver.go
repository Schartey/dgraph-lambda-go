
package resolvers

import(
	"github.com/schartey/dgraph-lambda-go/api"
	"context"
	"github.com/schartey/dgraph-lambda-go/examples/lambda/model"
)

/** Put these into resolvers.go  or similar **/
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
