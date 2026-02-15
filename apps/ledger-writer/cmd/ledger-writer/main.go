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
	"kalency/apps/ledger-writer/internal/httpapi"
	"kalency/apps/ledger-writer/internal/ledger"
	"kalency/apps/ledger-writer/internal/store"
)

func main() {
	port := getEnv("PORT", "8084")
	redisAddr := strings.TrimSpace(os.Getenv("REDIS_ADDR"))
	streamKey := getEnv("LEDGER_STREAM_KEY", "kalency:v1:stream:executions")
	startID := getEnv("LEDGER_START_ID", "$")
	batchSize := getEnvInt("LEDGER_BATCH_SIZE", 100)
	blockMS := getEnvInt("LEDGER_BLOCK_MS", 250)
	postgresDSN := strings.TrimSpace(os.Getenv("POSTGRES_DSN"))

	if redisAddr == "" {
		log.Fatal("REDIS_ADDR is required")
	}

	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		_ = redisClient.Close()
		log.Fatalf("redis ping failed: %v", err)
	}

	var (
		sink      ledger.ExecutionSink
		closeSink func()
	)
	if postgresDSN == "" {
		log.Printf("postgres sink disabled (POSTGRES_DSN not set)")
		sink = store.LogSink{}
		closeSink = func() {}
	} else {
		pgSink, err := store.NewPostgresSink(context.Background(), postgresDSN)
		if err != nil {
			_ = redisClient.Close()
			log.Fatalf("postgres connect failed: %v", err)
		}
		sink = pgSink
		closeSink = pgSink.Close
	}
	defer closeSink()

	source := store.NewRedisExecutionStreamSource(redisClient, streamKey)
	svc := ledger.NewService(sink)
	go runLedgerWriter(context.Background(), source, svc, startID, batchSize, time.Duration(blockMS)*time.Millisecond)

	server := httpapi.NewServer()
	addr := ":" + port
	log.Printf("ledger-writer listening on %s (redis=%s stream=%s)", addr, redisAddr, streamKey)
	if err := http.ListenAndServe(addr, server); err != nil {
		_ = redisClient.Close()
		log.Fatal(err)
	}
}

type executionSource interface {
	Read(ctx context.Context, lastID string, count int, block time.Duration) ([]ledger.ExecutionEvent, string, error)
}

func runLedgerWriter(ctx context.Context, source executionSource, svc *ledger.Service, lastID string, batchSize int, block time.Duration) {
	currentID := lastID
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		events, nextID, err := source.Read(ctx, currentID, batchSize, block)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("execution read failed: %v", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		currentID = nextID
		for _, event := range events {
			if err := svc.Handle(ctx, event); err != nil {
				log.Printf("write execution failed: %v", err)
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
