package generated

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/schartey/dgraph-lambda-go/api"
	"github.com/schartey/dgraph-lambda-go/examples/lambda/model"
	"github.com/schartey/dgraph-lambda-go/examples/lambda/resolvers"
)

type Executer struct {
	api.ExecuterInterface
	fieldResolver      resolvers.FieldResolver
	queryResolver      resolvers.QueryResolver
	mutationResolver   resolvers.MutationResolver
	middlewareResolver resolvers.MiddlewareResolver
	webhookResolver    resolvers.WebhookResolver
}

func NewExecuter(resolver *resolvers.Resolver) api.ExecuterInterface {
	return Executer{fieldResolver: resolvers.FieldResolver{Resolver: resolver}, queryResolver: resolvers.QueryResolver{Resolver: resolver}, mutationResolver: resolvers.MutationResolver{Resolver: resolver}, middlewareResolver: resolvers.MiddlewareResolver{Resolver: resolver}, webhookResolver: resolvers.WebhookResolver{Resolver: resolver}}
}

func (e Executer) Resolve(ctx context.Context, request *api.Request) (response []byte, err *api.LambdaError) {
	if request.Event.Operation != "" {
		return nil, e.resolveWebhook(ctx, request)
	} else {
		parentsBytes, underlyingError := request.Parents.MarshalJSON()
		if underlyingError != nil {
			return nil, &api.LambdaError{Underlying: underlyingError, Status: http.StatusInternalServerError}
		}

		mc := &api.MiddlewareContext{Ctx: ctx, Request: request}
		if err = e.middleware(mc); err != nil {
			return nil, err
		}
		ctx = mc.Ctx
		request = mc.Request

		if strings.HasPrefix(request.Resolver, "Query.") {
			return e.resolveQuery(ctx, request)
		} else if strings.HasPrefix(request.Resolver, "Mutation.") {
			return e.resolveMutation(ctx, request)
		} else {
			return e.resolveField(ctx, request, parentsBytes)
		}
	}
}

func (e Executer) middleware(mc *api.MiddlewareContext) (err *api.LambdaError) {
	switch mc.Request.Resolver {
	case "User.active":
		{
			if err = e.middlewareResolver.Middleware_admin(mc); err != nil {
				return err
			}
			break
		}

	case "Query.getHotelByName":
		{
			if err = e.middlewareResolver.Middleware_user(mc); err != nil {
				return err
			}
			break
		}

	case "Query.getTopAuthors":
		{
			if err = e.middlewareResolver.Middleware_user(mc); err != nil {
				return err
			}
			if err = e.middlewareResolver.Middleware_admin(mc); err != nil {
				return err
			}
			break
		}

	case "Mutation.newAuthor":
		{
			if err = e.middlewareResolver.Middleware_admin(mc); err != nil {
				return err
			}
			break
		}

	}
	return nil
}

func (e Executer) resolveField(ctx context.Context, request *api.Request, parentsBytes []byte) (response []byte, err *api.LambdaError) {
	switch request.Resolver {
	case "User.active":
		{
			var parents []*model.User
			json.Unmarshal(parentsBytes, &parents)

			result, err := e.fieldResolver.User_active(ctx, parents, request.AuthHeader)
			if err != nil {
				return nil, err
			}

			var underlyingError error
			response, underlyingError = json.Marshal(result)
			if underlyingError != nil {
				return nil, &api.LambdaError{Underlying: underlyingError, Status: http.StatusInternalServerError}
			} else {
				return response, nil
			}
			break
		}
	case "Post.additionalInfo":
		{
			var parents []*model.Post
			json.Unmarshal(parentsBytes, &parents)

			result, err := e.fieldResolver.Post_additionalInfo(ctx, parents, request.AuthHeader)
			if err != nil {
				return nil, err
			}

			var underlyingError error
			response, underlyingError = json.Marshal(result)
			if underlyingError != nil {
				return nil, &api.LambdaError{Underlying: underlyingError, Status: http.StatusInternalServerError}
			} else {
				return response, nil
			}
			break
		}
	case "User.rank":
		{
			var parents []*model.User
			json.Unmarshal(parentsBytes, &parents)

			result, err := e.fieldResolver.User_rank(ctx, parents, request.AuthHeader)
			if err != nil {
				return nil, err
			}

			var underlyingError error
			response, underlyingError = json.Marshal(result)
			if underlyingError != nil {
				return nil, &api.LambdaError{Underlying: underlyingError, Status: http.StatusInternalServerError}
			} else {
				return response, nil
			}
			break
		}
	case "User.reputation":
		{
			var parents []*model.User
			json.Unmarshal(parentsBytes, &parents)

			result, err := e.fieldResolver.User_reputation(ctx, parents, request.AuthHeader)
			if err != nil {
				return nil, err
			}

			var underlyingError error
			response, underlyingError = json.Marshal(result)
			if underlyingError != nil {
				return nil, &api.LambdaError{Underlying: underlyingError, Status: http.StatusInternalServerError}
			} else {
				return response, nil
			}
			break
		}
	case "Figure.size":
		{
			var parents []*model.Figure
			json.Unmarshal(parentsBytes, &parents)

			result, err := e.fieldResolver.Figure_size(ctx, parents, request.AuthHeader)
			if err != nil {
				return nil, err
			}

			var underlyingError error
			response, underlyingError = json.Marshal(result)
			if underlyingError != nil {
				return nil, &api.LambdaError{Underlying: underlyingError, Status: http.StatusInternalServerError}
			} else {
				return response, nil
			}
			break
		}
	}

	return nil, &api.LambdaError{Underlying: errors.New("could not find query resolver"), Status: http.StatusNotFound}
}

func (e Executer) resolveQuery(ctx context.Context, request *api.Request) (response []byte, err *api.LambdaError) {
	switch request.Resolver {
	case "Query.getApples":
		{
			result, err := e.queryResolver.Query_getApples(ctx, request.AuthHeader)
			if err != nil {
				return nil, err
			}

			var underlyingError error
			response, underlyingError = json.Marshal(result)
			if underlyingError != nil {
				return nil, &api.LambdaError{Underlying: underlyingError, Status: http.StatusInternalServerError}
			} else {
				return response, nil
			}
			break
		}
	case "Query.getHotelByName":
		{
			var name string
			json.Unmarshal(request.Args["name"], &name)
			result, err := e.queryResolver.Query_getHotelByName(ctx, name, request.AuthHeader)
			if err != nil {
				return nil, err
			}

			var underlyingError error
			response, underlyingError = json.Marshal(result)
			if underlyingError != nil {
				return nil, &api.LambdaError{Underlying: underlyingError, Status: http.StatusInternalServerError}
			} else {
				return response, nil
			}
			break
		}
	case "Query.getTopAuthors":
		{
			var id string
			json.Unmarshal(request.Args["id"], &id)
			result, err := e.queryResolver.Query_getTopAuthors(ctx, id, request.AuthHeader)
			if err != nil {
				return nil, err
			}

			var underlyingError error
			response, underlyingError = json.Marshal(result)
			if underlyingError != nil {
				return nil, &api.LambdaError{Underlying: underlyingError, Status: http.StatusInternalServerError}
			} else {
				return response, nil
			}
			break
		}
	}

	return nil, &api.LambdaError{Underlying: errors.New("could not find query resolver"), Status: http.StatusNotFound}
}

func (e Executer) resolveMutation(ctx context.Context, request *api.Request) (response []byte, err *api.LambdaError) {
	switch request.Resolver {
	case "Mutation.newAuthor":
		{
			var name string
			json.Unmarshal(request.Args["name"], &name)
			result, err := e.mutationResolver.Mutation_newAuthor(ctx, name, request.AuthHeader)
			if err != nil {
				return nil, err
			}

			var underlyingError error
			response, underlyingError = json.Marshal(result)
			if underlyingError != nil {
				return nil, &api.LambdaError{Underlying: underlyingError, Status: http.StatusInternalServerError}
			} else {
				return response, nil
			}
			break
		}
	}

	return nil, &api.LambdaError{Underlying: errors.New("could not find query resolver"), Status: http.StatusNotFound}
}

func (e Executer) resolveWebhook(ctx context.Context, request *api.Request) (err *api.LambdaError) {

	return &api.LambdaError{Underlying: errors.New("could not find webhook resolver"), Status: http.StatusNotFound}
}
