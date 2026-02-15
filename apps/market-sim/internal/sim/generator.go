package sim

import (
	"errors"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type RandomSource interface {
	Float64() float64
}

type defaultRandom struct {
	rng *rand.Rand
}

func newDefaultRandom() defaultRandom {
	return defaultRandom{rng: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

func (d defaultRandom) Float64() float64 {
	if d.rng == nil {
		// Shouldn't happen, but keep a safe fallback.
		return rand.Float64()
	}
	return d.rng.Float64()
}

type Generator struct {
	mu            sync.Mutex
	symbols       []string
	prices        map[string]float64
	volatility    float64
	sellBias      float64
	random        RandomSource
	pausedSymbols map[string]struct{}
}

func NewGenerator(symbols []string, initialPrice float64, volatility float64, random RandomSource) *Generator {
	cleanSymbols := normalizeSymbols(symbols)
	if initialPrice <= 0 {
		initialPrice = 100
	}
	if volatility < 0 {
		volatility = 0
	}
	if random == nil {
		random = newDefaultRandom()
	}

	prices := make(map[string]float64, len(cleanSymbols))
	for _, symbol := range cleanSymbols {
		prices[symbol] = initialPrice
	}

	return &Generator{
		symbols:       cleanSymbols,
		prices:        prices,
		volatility:    volatility,
		sellBias:      0.5,
		random:        random,
		pausedSymbols: map[string]struct{}{},
	}
}

func (g *Generator) Next() []Tick {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UTC()
	ticks := make([]Tick, 0, len(g.symbols))
	for _, symbol := range g.symbols {
		if _, paused := g.pausedSymbols[symbol]; paused {
			continue
		}

		current := g.prices[symbol]
		raw := g.random.Float64()
		delta := g.priceDelta(raw)
		nextPrice := current * (1 + delta)
		if nextPrice < 0.01 {
			nextPrice = 0.01
		}
		g.prices[symbol] = nextPrice
		ticks = append(ticks, Tick{
			Symbol: symbol,
			Price:  nextPrice,
			Delta:  delta,
			TS:     now,
		})
	}
	return ticks
}

func (g *Generator) priceDelta(raw float64) float64 {
	// sellBias controls how often deltas are negative (price moves down).
	//
	// Important: we scale the magnitude so the expected drift stays ~0 even when
	// sellBias != 0.5. Otherwise sellBias > 0.5 creates a consistent downtrend
	// and the chart looks "always red".
	sellBias := g.sellBias
	if sellBias <= 0 {
		return 0
	}
	if sellBias >= 1 {
		return -g.volatility
	}

	negMax := g.volatility
	posMax := g.volatility
	if sellBias > 0.5 {
		negMax = g.volatility * (1 - sellBias) / sellBias
	} else if sellBias < 0.5 {
		posMax = g.volatility * sellBias / (1 - sellBias)
	}

	if raw < sellBias {
		magnitude := raw / sellBias
		return -magnitude * negMax
	}

	magnitude := (raw - sellBias) / (1 - sellBias)
	return magnitude * posMax
}

func (g *Generator) SetVolatility(volatility float64) error {
	if volatility < 0 {
		return errors.New("volatility must be non-negative")
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	g.volatility = volatility
	return nil
}

func (g *Generator) SetSellBias(sellBias float64) error {
	if sellBias < 0 || sellBias > 1 {
		return errors.New("sell bias must be between 0 and 1")
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	g.sellBias = sellBias
	return nil
}

func (g *Generator) PauseSymbol(symbol string) error {
	normalized := strings.TrimSpace(symbol)
	if normalized == "" {
		return errors.New("symbol is required")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	canonical, ok := g.findSymbolLocked(normalized)
	if !ok {
		return errors.New("symbol not found")
	}
	g.pausedSymbols[canonical] = struct{}{}
	return nil
}

func (g *Generator) ResumeSymbol(symbol string) error {
	normalized := strings.TrimSpace(symbol)
	if normalized == "" {
		return errors.New("symbol is required")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	canonical, ok := g.findSymbolLocked(normalized)
	if !ok {
		return errors.New("symbol not found")
	}
	delete(g.pausedSymbols, canonical)
	return nil
}

func (g *Generator) findSymbolLocked(symbol string) (string, bool) {
	for _, candidate := range g.symbols {
		if strings.EqualFold(candidate, symbol) {
			return candidate, true
		}
	}
	return "", false
}

func normalizeSymbols(symbols []string) []string {
	result := make([]string, 0, len(symbols))
	seen := map[string]struct{}{}
	for _, symbol := range symbols {
		normalized := strings.TrimSpace(symbol)
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	if len(result) == 0 {
		return []string{"BTC-USD"}
	}
	return result
}
