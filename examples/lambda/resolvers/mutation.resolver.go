
package resolvers

import(
	"github.com/schartey/dgraph-lambda-go/api"
	"context"
)

/** Put these into resolvers.go  or similar **/
type MutationResolver struct {
	*Resolver
}
func (q *MutationResolver) Mutation_newAuthor(ctx context.Context, name string, authHeader api.AuthHeader) (string, *api.LambdaError) {    
	return "", nil
}

