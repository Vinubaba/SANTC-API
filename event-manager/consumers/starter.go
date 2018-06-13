package consumers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Vinubaba/SANTC-API/common/log"
	"github.com/Vinubaba/SANTC-API/common/messaging"
	"github.com/Vinubaba/SANTC-API/event-manager/shared"

	"github.com/pkg/errors"
)

type Consumer struct {
	Config        *shared.AppConfig `inject:""`
	Logger        *log.Logger       `inject:""`
	PubSubClient  *messaging.Client `inject:""`
	EventHandlers []interface {
		CanHandle(event Event) bool
		Handle(ctx context.Context, event Event) error
		Name() string
	}
}

func (c *Consumer) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			c.Logger.Info(ctx, "starting consumer")
			if err := c.subscribe(ctx); err != nil {
				c.Logger.Warn(ctx, "Consumer stopped")
				time.Sleep(time.Second)
			}
		}
	}
}

func (c *Consumer) subscribe(ctx context.Context) error {
	callback := func(ctx context.Context, msg messaging.Message) {
		msg.Ack()

		event := Event{}
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			c.Logger.Err(ctx, "failed to unmarshal the message data", "err", err, "messageId", msg.ID)
			return
		}

		for _, eventHandler := range c.EventHandlers {
			if eventHandler.CanHandle(event) {
				if err := eventHandler.Handle(ctx, event); err != nil {
					c.Logger.Err(ctx, "failed to handle message", "err", err, "messageId", msg.ID, "handler", eventHandler.Name())
				} else {
					c.Logger.Info(ctx, "message successfully handled !", "messageId", msg.ID, "handler", eventHandler.Name())
				}
				return
			}
		}
		c.Logger.Warn(ctx, "no handlers were able to consume this message", "messageId", msg.ID)
		return
	}
	if err := c.PubSubClient.Subscribe(ctx, callback); err != nil {
		return errors.Wrap(err, "failed to subscribe to the messaging system")
	}
	return nil
}
