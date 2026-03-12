package events

import "context"

// NoopEventBus silently discards all events. Useful in tests
// and when NATS is not configured.
type NoopEventBus struct{}

func NewNoopEventBus() *NoopEventBus { return &NoopEventBus{} }

func (b *NoopEventBus) Publish(_ context.Context, _ Event) error { return nil }
func (b *NoopEventBus) Close() error                              { return nil }
