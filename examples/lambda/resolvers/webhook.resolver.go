package resolvers

import (
	"context"

	"github.com/schartey/dgraph-lambda-go/api"
)

type WebhookResolverInterface interface {
	Webhook_Hotel(ctx context.Context, event *api.Event) *api.LambdaError
	Webhook_User(ctx context.Context, event *api.Event) *api.LambdaError
}

type WebhookResolver struct {
	*Resolver
}

func (w *WebhookResolver) Webhook_Hotel(ctx context.Context, event *api.Event) *api.LambdaError {
	return nil
}

func (w *WebhookResolver) Webhook_User(ctx context.Context, event *api.Event) *api.LambdaError {
	return nil
}
