
package resolvers

import(
	"github.com/schartey/dgraph-lambda-go/api"
	"context"
	"github.com/schartey/dgraph-lambda-go/examples/lambda/model"
)

/** Put these into resolvers.go  or similar **/
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

