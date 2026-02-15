package sim

import "testing"

type fixedRandom struct {
	values []float64
	index  int
}

func (f *fixedRandom) Float64() float64 {
	if len(f.values) == 0 {
		return 0.5
	}
	value := f.values[f.index%len(f.values)]
	f.index++
	return value
}

func TestGeneratorNextGeneratesTickPerSymbol(t *testing.T) {
	rng := &fixedRandom{values: []float64{0.9, 0.1}}
	generator := NewGenerator([]string{"BTC-USD", "ETH-USD"}, 100, 0.01, rng)

	ticks := generator.Next()
	if len(ticks) != 2 {
		t.Fatalf("expected 2 ticks, got %d", len(ticks))
	}

	if ticks[0].Symbol != "BTC-USD" {
		t.Fatalf("expected first symbol BTC-USD, got %s", ticks[0].Symbol)
	}
	if ticks[1].Symbol != "ETH-USD" {
		t.Fatalf("expected second symbol ETH-USD, got %s", ticks[1].Symbol)
	}

	if ticks[0].Price <= 100 {
		t.Fatalf("expected first price to increase above 100, got %f", ticks[0].Price)
	}
	if ticks[1].Price >= 100 {
		t.Fatalf("expected second price to decrease below 100, got %f", ticks[1].Price)
	}
}

func TestGeneratorSetVolatilityRejectsNegative(t *testing.T) {
	generator := NewGenerator([]string{"BTC-USD"}, 100, 0.01, &fixedRandom{})
	if err := generator.SetVolatility(-0.1); err == nil {
		t.Fatal("expected negative volatility to be rejected")
	}
}

func TestGeneratorSetSellBiasRejectsOutOfRange(t *testing.T) {
	generator := NewGenerator([]string{"BTC-USD"}, 100, 0.01, &fixedRandom{})
	if err := generator.SetSellBias(-0.1); err == nil {
		t.Fatal("expected negative sell bias to be rejected")
	}
	if err := generator.SetSellBias(1.1); err == nil {
		t.Fatal("expected sell bias > 1 to be rejected")
	}
}

func TestGeneratorSellBiasSkewsToNegativeDeltas(t *testing.T) {
	rng := &fixedRandom{values: []float64{0.10, 0.20, 0.30, 0.92, 0.25, 0.95, 0.15, 0.99}}
	generator := NewGenerator([]string{"BTC-USD"}, 100, 0.01, rng)
	if err := generator.SetSellBias(0.8); err != nil {
		t.Fatalf("set sell bias failed: %v", err)
	}

	negatives := 0
	positives := 0
	for i := 0; i < 8; i++ {
		ticks := generator.Next()
		if len(ticks) != 1 {
			t.Fatalf("expected 1 tick, got %d", len(ticks))
		}
		if ticks[0].Delta < 0 {
			negatives++
		}
		if ticks[0].Delta > 0 {
			positives++
		}
	}

	if negatives <= positives {
		t.Fatalf("expected more negative deltas than positive with sell bias; negatives=%d positives=%d", negatives, positives)
	}
}

func TestGeneratorPauseResumeSymbol(t *testing.T) {
	rng := &fixedRandom{values: []float64{0.5, 0.5}}
	generator := NewGenerator([]string{"BTC-USD", "ETH-USD"}, 100, 0.01, rng)

	if err := generator.PauseSymbol("BTC-USD"); err != nil {
		t.Fatalf("pause symbol failed: %v", err)
	}

	ticks := generator.Next()
	if len(ticks) != 1 {
		t.Fatalf("expected 1 tick while BTC-USD paused, got %d", len(ticks))
	}
	if ticks[0].Symbol != "ETH-USD" {
		t.Fatalf("expected ETH-USD tick while BTC-USD paused, got %s", ticks[0].Symbol)
	}

	if err := generator.ResumeSymbol("BTC-USD"); err != nil {
		t.Fatalf("resume symbol failed: %v", err)
	}
	ticks = generator.Next()
	if len(ticks) != 2 {
		t.Fatalf("expected 2 ticks after resume, got %d", len(ticks))
	}
}
