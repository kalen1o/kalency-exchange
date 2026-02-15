package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"kalency/apps/gateway-api/internal/candleclient"
	"kalency/apps/gateway-api/internal/chartclient"
	"kalency/apps/gateway-api/internal/gatewayapi"
	"kalency/apps/gateway-api/internal/marketsimclient"
	"kalency/apps/gateway-api/internal/matchingclient"
	"kalency/apps/gateway-api/internal/tickstream"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	matchingEngineURL := os.Getenv("MATCHING_ENGINE_URL")
	if matchingEngineURL == "" {
		matchingEngineURL = "http://localhost:8081"
	}
	candleRedisAddr := strings.TrimSpace(os.Getenv("CANDLE_REDIS_ADDR"))
	if candleRedisAddr == "" {
		candleRedisAddr = strings.TrimSpace(os.Getenv("REDIS_ADDR"))
	}
	candleKeyPrefix := strings.TrimSpace(os.Getenv("CANDLE_KEY_PREFIX"))
	if candleKeyPrefix == "" {
		candleKeyPrefix = "v1"
	}
	tickStreamKey := strings.TrimSpace(os.Getenv("TICK_STREAM_KEY"))
	if tickStreamKey == "" {
		tickStreamKey = "kalency:v1:stream:ticks"
	}
	chartRenderGatewayURL := strings.TrimSpace(os.Getenv("CHART_RENDER_GATEWAY_URL"))
	marketSimURL := strings.TrimSpace(os.Getenv("MARKET_SIM_URL"))

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret"
	}

	tradingClient := matchingclient.NewHTTPClient(matchingEngineURL)
	candleService, tickSource, closeCandleService := newRedisIntegrations(candleRedisAddr, candleKeyPrefix, tickStreamKey)
	defer closeCandleService()

	var chartService gatewayapi.ChartService
	if chartRenderGatewayURL != "" {
		chartService = chartclient.NewHTTPClient(chartRenderGatewayURL)
	}

	var adminService gatewayapi.AdminService
	if marketSimURL != "" {
		adminService = marketsimclient.NewHTTPClient(marketSimURL)
	}

	apiServer := gatewayapi.NewServer(gatewayapi.Config{
		JWTSecret:     jwtSecret,
		APIKeys:       parseAPIKeys(os.Getenv("API_KEYS")),
		CandleService: candleService,
		ChartService:  chartService,
		AdminService:  adminService,
		TickSource:    tickSource,
	}, tradingClient)

	addr := ":" + port
	log.Printf(
		"gateway-api listening on %s (matching-engine=%s candles-enabled=%t chart-gateway=%q market-sim=%q)",
		addr,
		matchingEngineURL,
		candleService != nil,
		chartRenderGatewayURL,
		marketSimURL,
	)
	if err := apiServer.Listen(addr); err != nil {
		log.Fatal(err)
	}
}

func newRedisIntegrations(redisAddr, keyPrefix, tickStreamKey string) (gatewayapi.CandleService, gatewayapi.TickSource, func()) {
	if strings.TrimSpace(redisAddr) == "" {
		return nil, nil, func() {}
	}

	client := redis.NewClient(&redis.Options{Addr: redisAddr})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("candle integration disabled (redis ping failed): %v", err)
		_ = client.Close()
		return nil, nil, func() {}
	}

	log.Printf("redis integration enabled (redis=%s prefix=%s tickStream=%s)", redisAddr, keyPrefix, tickStreamKey)
	return candleclient.NewRedisClient(client, keyPrefix), tickstream.NewRedisTickStreamSource(client, tickStreamKey), func() {
		_ = client.Close()
	}
}

func parseAPIKeys(raw string) map[string]string {
	result := map[string]string{}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return result
	}

	pairs := strings.Split(raw, ",")
	for _, pair := range pairs {
		parts := strings.Split(strings.TrimSpace(pair), ":")
		if len(parts) != 2 {
			continue
		}
		apiKey := strings.TrimSpace(parts[0])
		userID := strings.TrimSpace(parts[1])
		if apiKey == "" || userID == "" {
			continue
		}
		result[apiKey] = userID
	}
	return result
}
