package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// EventBus defines the publish/subscribe contract.
type EventBus interface {
	Publish(ctx context.Context, event Event) error
	Close() error
}

// NATSEventBus implements EventBus using NATS JetStream.
type NATSEventBus struct {
	conn   *nats.Conn
	js     jetstream.JetStream
	stream jetstream.Stream
}

// NewNATSEventBus connects to NATS, ensures the JetStream stream exists,
// and returns a ready-to-use event bus.
func NewNATSEventBus(url, streamName string, subjects []string) (*NATSEventBus, error) {
	nc, err := nats.Connect(url,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(10),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("jetstream new: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      streamName,
		Subjects:  subjects,
		Retention: jetstream.LimitsPolicy,
		MaxAge:    7 * 24 * time.Hour,
		Storage:   jetstream.FileStorage,
	})
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("jetstream create stream: %w", err)
	}

	slog.Info("NATS JetStream connected", "stream", streamName, "subjects", subjects)

	return &NATSEventBus{conn: nc, js: js, stream: stream}, nil
}

// Publish serializes the event to JSON and publishes to the event subject.
func (b *NATSEventBus) Publish(ctx context.Context, event Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("events: marshal: %w", err)
	}

	_, err = b.js.Publish(ctx, event.Subject, payload)
	if err != nil {
		return fmt.Errorf("events: publish %s: %w", event.Subject, err)
	}
	return nil
}

// Close drains and closes the NATS connection.
func (b *NATSEventBus) Close() error {
	b.conn.Close()
	return nil
}
