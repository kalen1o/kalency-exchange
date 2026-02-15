package sim

import "context"

type NoopTickSink struct{}

func (NoopTickSink) PublishTick(_ context.Context, _ Tick) error {
	return nil
}
