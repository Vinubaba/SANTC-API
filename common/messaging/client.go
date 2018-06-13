package messaging

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
)

type Client struct {
	googlePubSubClient *pubsub.Client
	options            []option.ClientOption
	topic              *pubsub.Topic
	subscription       *pubsub.Subscription
	projectID          string
}

type ClientOptions struct {
	ProjectID      string
	Topic          string
	Subscription   string
	CredentialPath string
}

type SubscribeCallbackFunc func(ctx context.Context, msg Message)

func New(config ClientOptions) (*Client, error) {
	var err error
	client := &Client{}
	client.googlePubSubClient, err = pubsub.NewClient(context.Background(), config.ProjectID, option.WithCredentialsFile(config.CredentialPath))
	if err != nil {
		return nil, err
	}
	if config.Topic != "" {
		client.topic = client.googlePubSubClient.Topic(config.Topic)
	}
	if config.Subscription != "" {
		client.subscription = client.googlePubSubClient.Subscription(config.Subscription)
	}
	return client, nil
}

func (s *Client) GetPubSubClient() *pubsub.Client {
	return s.googlePubSubClient
}

func (s *Client) Subscribe(ctx context.Context, callback SubscribeCallbackFunc) error {
	err := s.subscription.Receive(ctx, s.provideReceiveHandler(callback))
	if err != nil {
		return errors.Wrapf(err, "failed to pull messages from Google Pub/Sub subscription %s", s.subscription)
	}

	return nil
}

func (s *Client) Topics(ctx context.Context) *pubsub.TopicIterator {
	return s.googlePubSubClient.Topics(ctx)
}

func (s *Client) provideReceiveHandler(callback SubscribeCallbackFunc) func(ctx context.Context, pubSubMsg *pubsub.Message) {
	return func(ctx context.Context, pubSubMsg *pubsub.Message) {
		msg := s.newMessageFromPubSubMessage(pubSubMsg)

		callback(ctx, msg)
	}
}

func (s *Client) newMessageFromPubSubMessage(pubSubMsg *pubsub.Message) (msg Message) {
	msg.ID = pubSubMsg.ID
	msg.Data = pubSubMsg.Data
	msg.Attributes = pubSubMsg.Attributes
	msg.PublishTime = pubSubMsg.PublishTime

	msg.RegisterAck(func() error {
		pubSubMsg.Ack()
		return nil
	})
	msg.RegisterNack(func() error {
		pubSubMsg.Nack()
		return nil
	})
	if msg.Attributes == nil {
		msg.Attributes = make(map[string]string)
	}
	return
}

func (s *Client) addToContextFromMessage(ctx context.Context, key string, msg Message) context.Context {
	return context.WithValue(ctx, key, msg.Attributes[key])
}

func (s *Client) Publish(ctx context.Context, message Message) (err error) {
	msg := s.newPubSubMessageFromMessage(ctx, message)

	if _, err = s.topic.Publish(ctx, msg).Get(ctx); err != nil {
		err = errors.Wrapf(err, fmt.Sprintf("failed to publish in Google Pub/Sub topic %s", s.topic))
	}

	return err
}

func (s *Client) newPubSubMessageFromMessage(ctx context.Context, message Message) *pubsub.Message {
	msg := &pubsub.Message{
		ID:         message.ID,
		Data:       message.Data,
		Attributes: message.Attributes,
	}
	if msg.Attributes == nil {
		msg.Attributes = make(map[string]string)
	}

	return msg
}
