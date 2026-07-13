package push

import (
	"context"

	"github.com/mashmool0/inama/services/notifications/internal/model"
)

type Client interface {
	SendNotification(ctx context.Context, notification model.Notification) error
}

type NopClient struct{}

func (NopClient) SendNotification(context.Context, model.Notification) error {
	return nil
}
