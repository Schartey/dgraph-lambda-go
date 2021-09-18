
package generated

import(
	"github.com/schartey/dgraph-lambda-go/api"
	"context"
	"errors"
	"net/http"
	"encoding/json"
	"github.com/schartey/dgraph-lambda-go/examples/lambda/model"
	"github.com/schartey/dgraph-lambda-go/examples/lambda/resolvers"
)

type Executer struct {
	fieldResolver    	resolvers.FieldResolver
	queryResolver    	resolvers.QueryResolver
	mutationResolver 	resolvers.MutationResolver
	middlewareResolver 	resolvers.MiddlewareResolver
	webhookResolver 	resolvers.WebhookResolver
}

func NewExecuter(resolver *resolvers.Resolver) api.ExecuterInterface {
	return Executer{fieldResolver: resolvers.FieldResolver{Resolver: resolver}, queryResolver: resolvers.QueryResolver{Resolver: resolver}, mutationResolver: resolvers.MutationResolver{Resolver: resolver}, middlewareResolver: resolvers.MiddlewareResolver{Resolver: resolver}, webhookResolver: resolvers.WebhookResolver{Resolver: resolver}}
}

func (e *Executer) Middleware(md *api.MiddlewareData) (err *api.LambdaError) {
	switch md.Dbody.Resolver {
		case "User.active":
			{
				if err = e.middlewareResolver.Middleware_admin(md); err != nil {
					return err
				}
				break
			}
		
		case "Query.getHotelByName":
			{
				if err = e.middlewareResolver.Middleware_user(md); err != nil {
					return err
				}
				break
			}
		
		case "Query.getTopAuthors":
			{
				if err = e.middlewareResolver.Middleware_user(md); err != nil {
					return err
				}
				if err = e.middlewareResolver.Middleware_admin(md); err != nil {
					return err
				}
				break
			}
		
		case "Mutation.newAuthor":
			{
				if err = e.middlewareResolver.Middleware_admin(md); err != nil {
					return err
				}
				break
			}
		
	}
	return nil
}

func (e Executer) Resolve(ctx context.Context, dbody api.DBody) (response []byte, err *api.LambdaError) {
	if dbody.Event.Operation != "" {
		switch dbody.Event.TypeName {       
			case "Hotel":
				err = e.webhookResolver.Webhook_Hotel(ctx, dbody.Event)
				return nil, err
			        
			case "User":
				err = e.webhookResolver.Webhook_User(ctx, dbody.Event)
				return nil, err
			 
		}
	} else {
		parentsBytes, underlyingError := dbody.Parents.MarshalJSON()
		if underlyingError != nil {
			return nil, &api.LambdaError{Underlying: underlyingError, Status: http.StatusInternalServerError}	
		}
		

		md := &api.MiddlewareData{Ctx: ctx, Dbody: dbody}
		if err = e.Middleware(md); err != nil {
			return nil, err
		}
		ctx = md.Ctx
		dbody = md.Dbody

		response := []byte{}

		switch dbody.Resolver {
			case "User.active":
				{
					var parents []*model.User
					json.Unmarshal(parentsBytes, &parents)

					// Dependent on generation loop or just direct
					/*var actives []bool
					for _, parent := range parents {
						actives = fullnames.append(e.fieldResolver.bool_active(ctx, parent))
					}*/
					result, err := e.fieldResolver.User_active(ctx, parents, dbody.AuthHeader)
					if err != nil {
						return nil, err
					}

					var underlyingError error
					response, underlyingError = json.Marshal(result)
					if underlyingError != nil {
						return nil, &api.LambdaError{Underlying: underlyingError, Status: http.StatusInternalServerError}
					}
					break
				}
			case "Post.additionalInfo":
				{
					var parents []*model.Post
					json.Unmarshal(parentsBytes, &parents)

					// Dependent on generation loop or just direct
					/*var additionalInfos []string
					for _, parent := range parents {
						additionalInfos = fullnames.append(e.fieldResolver.string_additionalInfo(ctx, parent))
					}*/
					result, err := e.fieldResolver.Post_additionalInfo(ctx, parents, dbody.AuthHeader)
					if err != nil {
						return nil, err
					}

					var underlyingError error
					response, underlyingError = json.Marshal(result)
					if underlyingError != nil {
						return nil, &api.LambdaError{Underlying: underlyingError, Status: http.StatusInternalServerError}
					}
					break
				}
			case "User.rank":
				{
					var parents []*model.User
					json.Unmarshal(parentsBytes, &parents)

					// Dependent on generation loop or just direct
					/*var ranks []int64
					for _, parent := range parents {
						ranks = fullnames.append(e.fieldResolver.int64_rank(ctx, parent))
					}*/
					result, err := e.fieldResolver.User_rank(ctx, parents, dbody.AuthHeader)
					if err != nil {
						return nil, err
					}

					var underlyingError error
					response, underlyingError = json.Marshal(result)
					if underlyingError != nil {
						return nil, &api.LambdaError{Underlying: underlyingError, Status: http.StatusInternalServerError}
					}
					break
				}
			case "User.reputation":
				{
					var parents []*model.User
					json.Unmarshal(parentsBytes, &parents)

					// Dependent on generation loop or just direct
					/*var reputations []int64
					for _, parent := range parents {
						reputations = fullnames.append(e.fieldResolver.int64_reputation(ctx, parent))
					}*/
					result, err := e.fieldResolver.User_reputation(ctx, parents, dbody.AuthHeader)
					if err != nil {
						return nil, err
					}

					var underlyingError error
					response, underlyingError = json.Marshal(result)
					if underlyingError != nil {
						return nil, &api.LambdaError{Underlying: underlyingError, Status: http.StatusInternalServerError}
					}
					break
				}
			case "Figure.size":
				{
					var parents []*model.Figure
					json.Unmarshal(parentsBytes, &parents)

					// Dependent on generation loop or just direct
					/*var sizes []int64
					for _, parent := range parents {
						sizes = fullnames.append(e.fieldResolver.int64_size(ctx, parent))
					}*/
					result, err := e.fieldResolver.Figure_size(ctx, parents, dbody.AuthHeader)
					if err != nil {
						return nil, err
					}

					var underlyingError error
					response, underlyingError = json.Marshal(result)
					if underlyingError != nil {
						return nil, &api.LambdaError{Underlying: underlyingError, Status: http.StatusInternalServerError}
					}
					break
				}
			case "Query.getApples":
				{	
					result, err := e.queryResolver.Query_getApples(ctx, dbody.AuthHeader)
					if err != nil {
						return nil, err
					}

					var underlyingError error
					response, underlyingError = json.Marshal(result)
					if underlyingError != nil {
						return nil, &api.LambdaError{Underlying: underlyingError, Status: http.StatusInternalServerError}
					}
					break
				}
			case "Query.getHotelByName":
				{
					var name string 
					json.Unmarshal(dbody.Args["name"], &name)	
					result, err := e.queryResolver.Query_getHotelByName(ctx, name, dbody.AuthHeader)
					if err != nil {
						return nil, err
					}

					var underlyingError error
					response, underlyingError = json.Marshal(result)
					if underlyingError != nil {
						return nil, &api.LambdaError{Underlying: underlyingError, Status: http.StatusInternalServerError}
					}
					break
				}
			case "Query.getTopAuthors":
				{
					var id string 
					json.Unmarshal(dbody.Args["id"], &id)	
					result, err := e.queryResolver.Query_getTopAuthors(ctx, id, dbody.AuthHeader)
					if err != nil {
						return nil, err
					}

					var underlyingError error
					response, underlyingError = json.Marshal(result)
					if underlyingError != nil {
						return nil, &api.LambdaError{Underlying: underlyingError, Status: http.StatusInternalServerError}
					}
					break
				}
			case "Mutation.newAuthor":
				{
					var name string 
					json.Unmarshal(dbody.Args["name"], &name)	
					result, err := e.mutationResolver.Mutation_newAuthor(ctx, name, dbody.AuthHeader)
					if err != nil {
						return nil, err
					}

					var underlyingError error
					response, underlyingError = json.Marshal(result)
					if underlyingError != nil {
						return nil, &api.LambdaError{Underlying: underlyingError, Status: http.StatusInternalServerError}
					}
					break
				}
		}
		return response, nil
	}
	return nil, &api.LambdaError{Underlying: errors.New("No resolver found"), Status: http.StatusNotFound}
}
