package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"kalency/apps/matching-engine/internal/httpapi"
	"kalency/apps/matching-engine/internal/matching"
	"kalency/apps/matching-engine/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	engine, tradeSource := newRuntime()
	server := httpapi.NewServer(engine, tradeSource)

	addr := ":" + port
	log.Printf("matching-engine listening on %s", addr)
	if err := http.ListenAndServe(addr, server); err != nil {
		log.Fatal(err)
	}
}

func newRuntime() (*matching.Engine, httpapi.TradeSource) {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		engine := matching.NewEngine()
		return engine, engine
	}

	client := redis.NewClient(&redis.Options{Addr: redisAddr})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("redis integration disabled (ping failed): %v", err)
		_ = client.Close()
		engine := matching.NewEngine()
		return engine, engine
	}

	log.Printf("redis integration enabled at %s", redisAddr)
	openOrderStore := store.NewRedisOpenOrdersStore(client, "kalency:v1")
	streamSink := store.NewRedisExecutionStreamSink(client, "kalency:v1:stream:executions")
	streamReader := store.NewRedisExecutionStreamReader(client, "kalency:v1:stream:executions")

	engine := matching.NewEngineWithStoreAndSink(openOrderStore, streamSink)
	return engine, streamReader
}
