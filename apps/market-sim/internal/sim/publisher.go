package sim

import (
	"context"
	"sync"
	"time"
)

type TickSink interface {
	PublishTick(ctx context.Context, tick Tick) error
}

type Publisher struct {
	generator *Generator
	sink      TickSink
	interval  time.Duration

	mu      sync.Mutex
	running bool
	stopCh  chan struct{}
	doneCh  chan struct{}
}

func NewPublisher(generator *Generator, sink TickSink, interval time.Duration) *Publisher {
	if generator == nil {
		generator = NewGenerator(nil, 100, 0.005, nil)
	}
	if sink == nil {
		sink = NoopTickSink{}
	}
	if interval <= 0 {
		interval = 250 * time.Millisecond
	}

	return &Publisher{
		generator: generator,
		sink:      sink,
		interval:  interval,
	}
}

func (p *Publisher) Start() error {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return nil
	}

	stopCh := make(chan struct{})
	doneCh := make(chan struct{})
	p.stopCh = stopCh
	p.doneCh = doneCh
	p.running = true
	p.mu.Unlock()

	go p.loop(stopCh, doneCh)
	return nil
}

func (p *Publisher) Stop() error {
	p.mu.Lock()
	if !p.running {
		p.mu.Unlock()
		return nil
	}

	stopCh := p.stopCh
	doneCh := p.doneCh
	p.stopCh = nil
	p.doneCh = nil
	p.running = false
	p.mu.Unlock()

	close(stopCh)
	<-doneCh
	return nil
}

func (p *Publisher) Running() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}

func (p *Publisher) SetVolatility(volatility float64) error {
	return p.generator.SetVolatility(volatility)
}

func (p *Publisher) PauseSymbol(symbol string) error {
	return p.generator.PauseSymbol(symbol)
}

func (p *Publisher) ResumeSymbol(symbol string) error {
	return p.generator.ResumeSymbol(symbol)
}

func (p *Publisher) loop(stopCh <-chan struct{}, doneCh chan<- struct{}) {
	defer close(doneCh)
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			for _, tick := range p.generator.Next() {
				_ = p.sink.PublishTick(context.Background(), tick)
			}
		}
	}
}
