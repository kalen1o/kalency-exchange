package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"kalency/apps/market-sim/internal/httpapi"
	"kalency/apps/market-sim/internal/sim"
	"kalency/apps/market-sim/internal/store"
)

func main() {
	port := getEnv("PORT", "8082")
	redisAddr := strings.TrimSpace(os.Getenv("REDIS_ADDR"))
	streamKey := getEnv("SIM_STREAM_KEY", "kalency:v1:stream:ticks")
	symbols := parseSymbols(getEnv("SIM_SYMBOLS", "BTC-USD,ETH-USD"))
	initialPrice := getEnvFloat("SIM_INITIAL_PRICE", 100)
	volatility := getEnvFloat("SIM_VOLATILITY", 0.005)
	sellBias := getEnvUnitFloat("SIM_SELL_BIAS", 0.65)
	intervalMS := getEnvInt("SIM_INTERVAL_MS", 250)
	startOnBoot := getEnvBool("SIM_START_ON_BOOT", true)

	generator := sim.NewGenerator(symbols, initialPrice, volatility, nil)
	if err := generator.SetSellBias(sellBias); err != nil {
		log.Printf("invalid SIM_SELL_BIAS value; using default 0.65: %v", err)
		_ = generator.SetSellBias(0.65)
	}
	sink, closeSink := newSink(redisAddr, streamKey)
	defer closeSink()

	publisher := sim.NewPublisher(generator, sink, time.Duration(intervalMS)*time.Millisecond)
	if startOnBoot {
		if err := publisher.Start(); err != nil {
			log.Printf("failed to auto-start simulator: %v", err)
		}
	}

	server := httpapi.NewServer(publisher)
	addr := ":" + port
	log.Printf(
		"market-sim listening on %s (symbols=%v interval_ms=%d sell_bias=%.2f redis=%q stream=%q running=%t)",
		addr,
		symbols,
		intervalMS,
		sellBias,
		redisAddr,
		streamKey,
		publisher.Running(),
	)
	if err := http.ListenAndServe(addr, server); err != nil {
		log.Fatal(err)
	}
}

func newSink(redisAddr, streamKey string) (sim.TickSink, func()) {
	if redisAddr == "" {
		log.Printf("redis sink disabled (REDIS_ADDR not set)")
		return sim.NoopTickSink{}, func() {}
	}

	client := redis.NewClient(&redis.Options{Addr: redisAddr})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("redis sink disabled (ping failed): %v", err)
		_ = client.Close()
		return sim.NoopTickSink{}, func() {}
	}

	log.Printf("redis sink enabled at %s", redisAddr)
	return store.NewRedisTickStreamSink(client, streamKey), func() {
		_ = client.Close()
	}
}

func getEnv(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(name string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func getEnvFloat(name string, fallback float64) float64 {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvUnitFloat(name string, fallback float64) float64 {
	value := getEnvFloat(name, fallback)
	if value < 0 || value > 1 {
		return fallback
	}
	return value
}

func getEnvBool(name string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return parsed
}

func parseSymbols(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		symbol := strings.TrimSpace(part)
		if symbol == "" {
			continue
		}
		result = append(result, symbol)
	}
	if len(result) == 0 {
		return []string{"BTC-USD"}
	}
	return result
}
