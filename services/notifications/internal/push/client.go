package push

import (
	"context"
	"log/slog"
	"time"

	"github.com/mashmool0/inama/services/notifications/internal/model"
)

type Client interface {
	SendNotification(ctx context.Context, notification model.Notification) error
}

type NopClient struct{}

func (NopClient) SendNotification(context.Context, model.Notification) error {
	return nil
}

type Dispatcher interface {
	Dispatch(ctx context.Context, notification model.Notification)
}

type AsyncDispatcher struct {
	Client  Client
	Logger  *slog.Logger
	Timeout time.Duration
}

func (d AsyncDispatcher) Dispatch(ctx context.Context, notification model.Notification) {
	go func() {
		sendCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), d.timeout())
		defer cancel()

		if err := d.Client.SendNotification(sendCtx, notification); err != nil && d.Logger != nil {
			d.Logger.Error("push delivery failed", "notification_id", notification.ID, "recipient_id", notification.RecipientID, "error", err)
		}
	}()
}

func (d AsyncDispatcher) timeout() time.Duration {
	if d.Timeout <= 0 {
		return 5 * time.Second
	}
	return d.Timeout
}
