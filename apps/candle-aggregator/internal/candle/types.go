package candle

import "time"

type Tick struct {
	Symbol string
	Price  float64
	Volume float64
	TS     time.Time
}

type Candle struct {
	Symbol      string
	Timeframe   string
	BucketStart time.Time
	Open        float64
	High        float64
	Low         float64
	Close       float64
	Volume      float64
}
