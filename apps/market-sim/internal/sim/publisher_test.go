package sim

import (
	"context"
	"sync"
	"testing"
	"time"
)

type recordingSink struct {
	mu     sync.Mutex
	ticks  []Tick
	notify chan struct{}
}

func newRecordingSink() *recordingSink {
	return &recordingSink{notify: make(chan struct{}, 32)}
}

func (s *recordingSink) PublishTick(_ context.Context, tick Tick) error {
	s.mu.Lock()
	s.ticks = append(s.ticks, tick)
	s.mu.Unlock()
	select {
	case s.notify <- struct{}{}:
	default:
	}
	return nil
}

func (s *recordingSink) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.ticks)
}

func (s *recordingSink) WaitForAtLeast(n int, timeout time.Duration) bool {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	for {
		if s.Count() >= n {
			return true
		}
		select {
		case <-s.notify:
		case <-deadline.C:
			return false
		}
	}
}

func TestPublisherStartAndStop(t *testing.T) {
	rng := &fixedRandom{values: []float64{0.8}}
	generator := NewGenerator([]string{"BTC-USD"}, 100, 0.01, rng)
	sink := newRecordingSink()
	publisher := NewPublisher(generator, sink, 10*time.Millisecond)

	if err := publisher.Start(); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	if !sink.WaitForAtLeast(1, 200*time.Millisecond) {
		t.Fatal("expected at least one published tick")
	}

	if err := publisher.Stop(); err != nil {
		t.Fatalf("stop failed: %v", err)
	}

	countAfterStop := sink.Count()
	time.Sleep(40 * time.Millisecond)
	if sink.Count() != countAfterStop {
		t.Fatal("expected no additional ticks after stop")
	}
}
