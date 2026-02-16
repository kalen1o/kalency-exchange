package sim

import "time"

// Tick is a synthetic market price update for a symbol.
type Tick struct {
	Symbol string    `json:"symbol"`
	Price  float64   `json:"price"`
	Delta  float64   `json:"delta"`
	Volume float64   `json:"volume"`
	TS     time.Time `json:"ts"`
}
