package firebase

import (
	"context"
	"fmt"

	pubsub "cloud.google.com/go/pubsub/v2"
	"github.com/rs/zerolog/log"
)

type PubSubMetrics interface {
	PublishSuccess()
	PublishFailure()
}

type PubSubTopicPublisherClient struct {
	metrics   PubSubMetrics
	publisher *pubsub.Publisher
}

func NewPubSubTopicPublisherClient(publisher *pubsub.Publisher, metrics PubSubMetrics) *PubSubTopicPublisherClient {
	return &PubSubTopicPublisherClient{
		metrics:   metrics,
		publisher: publisher,
	}
}

func (c *PubSubTopicPublisherClient) PublishMessage(ctx context.Context, msg []byte) error {
	result := c.publisher.Publish(ctx, &pubsub.Message{Data: msg})
	id, err := result.Get(ctx)
	if err != nil {
		log.Error().
			Err(err).
			Str("pubsub_message_id", id).
			Msg("pubsub publish failed")
		c.metrics.PublishFailure()
		return fmt.Errorf("webhook client publish topic: %w", err)
	}
	log.Info().
		Str("pubsub_message_id", id).
		Msg("pubsub publish result")
	c.metrics.PublishSuccess()
	return nil
}
