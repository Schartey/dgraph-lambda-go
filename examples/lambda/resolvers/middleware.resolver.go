
package resolvers

import(
	"github.com/schartey/dgraph-lambda-go/api"
)

/** Put these into resolvers.go  or similar **/
type MiddlewareResolver struct {
	*Resolver
}


func (m *MiddlewareResolver) Middleware_admin(md *api.MiddlewareData) *api.LambdaError {      
	return nil
}

func (m *MiddlewareResolver) Middleware_user(md *api.MiddlewareData) *api.LambdaError {      
	return nil
}

