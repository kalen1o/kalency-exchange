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
	"kalency/apps/candle-aggregator/internal/candle"
	"kalency/apps/candle-aggregator/internal/httpapi"
	"kalency/apps/candle-aggregator/internal/store"
)

func main() {
	port := getEnv("PORT", "8083")
	redisAddr := strings.TrimSpace(os.Getenv("REDIS_ADDR"))
	streamKey := getEnv("CANDLE_TICK_STREAM", "kalency:v1:stream:ticks")
	prefix := getEnv("CANDLE_KEY_PREFIX", "v1")
	startID := getEnv("CANDLE_START_ID", "$")
	batchSize := getEnvInt("CANDLE_BATCH_SIZE", 100)
	blockMS := getEnvInt("CANDLE_BLOCK_MS", 250)
	ttlHours := getEnvInt("CANDLE_TTL_HOURS", 720)

	if redisAddr == "" {
		log.Fatal("REDIS_ADDR is required")
	}

	client := redis.NewClient(&redis.Options{Addr: redisAddr})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		log.Fatalf("redis ping failed: %v", err)
	}

	ttl := time.Duration(ttlHours) * time.Hour
	tickSource := store.NewRedisTickStreamSource(client, streamKey)
	candleStore := store.NewRedisCandleStore(client, prefix)
	svc := candle.NewService(candleStore, candle.Config{TTL: ttl})

	go runAggregator(context.Background(), tickSource, svc, startID, batchSize, time.Duration(blockMS)*time.Millisecond)

	server := httpapi.NewServer()
	addr := ":" + port
	log.Printf("candle-aggregator listening on %s (redis=%s stream=%s)", addr, redisAddr, streamKey)
	if err := http.ListenAndServe(addr, server); err != nil {
		_ = client.Close()
		log.Fatal(err)
	}
}

type tickSource interface {
	Read(ctx context.Context, lastID string, count int, block time.Duration) ([]candle.Tick, string, error)
}

func runAggregator(ctx context.Context, source tickSource, svc *candle.Service, lastID string, batchSize int, block time.Duration) {
	currentID := lastID
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		ticks, nextID, err := source.Read(ctx, currentID, batchSize, block)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("tick read failed: %v", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		currentID = nextID
		for _, tick := range ticks {
			if err := svc.ProcessTick(ctx, tick); err != nil {
				log.Printf("process tick failed: %v", err)
			}
		}
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
