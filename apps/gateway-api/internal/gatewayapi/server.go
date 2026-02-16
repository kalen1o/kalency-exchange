package gatewayapi

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/websocket/v2"
	"github.com/golang-jwt/jwt/v5"
	"kalency/apps/gateway-api/internal/contracts"
	"kalency/apps/gateway-api/internal/tickstream"
)

type TradingService interface {
	PlaceOrder(req contracts.PlaceOrderRequest) (contracts.OrderAck, error)
	CancelOrder(userID, orderID string) (contracts.OrderAck, error)
	OpenOrders(userID string) ([]contracts.Order, error)
	Wallet(userID string) (contracts.Wallet, error)
	ListExecutions(symbol string, limit int) ([]contracts.Execution, error)
	ListOrderBook(symbol string, depth int) (contracts.OrderBookSnapshot, error)
}

type CandleService interface {
	ListCandles(symbol, timeframe string, from, to time.Time) ([]contracts.Candle, error)
}

type TickSource interface {
	Read(ctx context.Context, lastID string, count int, block time.Duration) ([]tickstream.Tick, string, error)
}

type AdminService interface {
	StartSimulator() (map[string]any, error)
	StopSimulator() (map[string]any, error)
	SetVolatility(volatility float64) (map[string]any, error)
	PauseSymbol(symbol string) (map[string]any, error)
	ResumeSymbol(symbol string) (map[string]any, error)
	EnsureSymbol(symbol string) (map[string]any, error)
}

type Config struct {
	JWTSecret     string
	APIKeys       map[string]string
	CandleService CandleService
	AdminService  AdminService
	TickSource    TickSource
}

type tokenRequest struct {
	UserID string `json:"userId"`
}

type tokenResponse struct {
	Token string `json:"token"`
}

type authIdentity struct {
	UserID string
}

const authLocalKey = "auth.identity"

func NewServer(cfg Config, trading TradingService) *fiber.App {
	app := fiber.New()
	secret := cfg.JWTSecret
	if secret == "" {
		secret = "dev-secret"
	}
	candleService := cfg.CandleService
	adminService := cfg.AdminService
	tickSource := cfg.TickSource

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Authorization,Content-Type,X-API-Key",
		AllowMethods: "GET,POST,DELETE,OPTIONS",
	}))

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	app.Post("/v1/auth/token", func(c *fiber.Ctx) error {
		var req tokenRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
		}
		req.UserID = strings.TrimSpace(req.UserID)
		if req.UserID == "" {
			return fiber.NewError(fiber.StatusBadRequest, "userId is required")
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": req.UserID,
			"exp": time.Now().Add(24 * time.Hour).Unix(),
		})
		signed, err := token.SignedString([]byte(secret))
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to sign token")
		}
		return c.JSON(tokenResponse{Token: signed})
	})

	protected := app.Group("/v1", requireAuth(secret, cfg.APIKeys))

	protected.Post("/orders", func(c *fiber.Ctx) error {
		identity := c.Locals(authLocalKey).(authIdentity)

		var req contracts.PlaceOrderRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
		}

		req.UserID = strings.TrimSpace(req.UserID)
		if req.UserID != "" && req.UserID != identity.UserID {
			return fiber.NewError(fiber.StatusForbidden, "userId does not match authenticated identity")
		}
		req.UserID = identity.UserID

		ack, err := trading.PlaceOrder(req)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return c.Status(fiber.StatusCreated).JSON(ack)
	})

	protected.Delete("/orders/:orderId", func(c *fiber.Ctx) error {
		identity := c.Locals(authLocalKey).(authIdentity)
		orderID := strings.TrimSpace(c.Params("orderId"))
		if orderID == "" {
			return fiber.NewError(fiber.StatusBadRequest, "orderId is required")
		}

		ack, err := trading.CancelOrder(identity.UserID, orderID)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return c.JSON(ack)
	})

	protected.Get("/orders/open", func(c *fiber.Ctx) error {
		identity := c.Locals(authLocalKey).(authIdentity)
		orders, err := trading.OpenOrders(identity.UserID)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return c.JSON(orders)
	})

	protected.Get("/wallet", func(c *fiber.Ctx) error {
		identity := c.Locals(authLocalKey).(authIdentity)
		wallet, err := trading.Wallet(identity.UserID)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return c.JSON(wallet)
	})

	protected.Post("/admin/sim/start", func(c *fiber.Ctx) error {
		if adminService == nil {
			return fiber.NewError(fiber.StatusServiceUnavailable, "admin service unavailable")
		}
		out, err := adminService.StartSimulator()
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return c.JSON(out)
	})

	protected.Post("/admin/sim/stop", func(c *fiber.Ctx) error {
		if adminService == nil {
			return fiber.NewError(fiber.StatusServiceUnavailable, "admin service unavailable")
		}
		out, err := adminService.StopSimulator()
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return c.JSON(out)
	})

	protected.Post("/admin/sim/volatility-profile", func(c *fiber.Ctx) error {
		if adminService == nil {
			return fiber.NewError(fiber.StatusServiceUnavailable, "admin service unavailable")
		}
		var req struct {
			Volatility float64 `json:"volatility"`
		}
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
		}
		out, err := adminService.SetVolatility(req.Volatility)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return c.JSON(out)
	})

	protected.Post("/admin/symbols/:symbol/pause", func(c *fiber.Ctx) error {
		if adminService == nil {
			return fiber.NewError(fiber.StatusServiceUnavailable, "admin service unavailable")
		}
		symbol := strings.TrimSpace(c.Params("symbol"))
		if symbol == "" {
			return fiber.NewError(fiber.StatusBadRequest, "symbol is required")
		}
		out, err := adminService.PauseSymbol(symbol)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return c.JSON(out)
	})

	protected.Post("/admin/symbols/:symbol/resume", func(c *fiber.Ctx) error {
		if adminService == nil {
			return fiber.NewError(fiber.StatusServiceUnavailable, "admin service unavailable")
		}
		symbol := strings.TrimSpace(c.Params("symbol"))
		if symbol == "" {
			return fiber.NewError(fiber.StatusBadRequest, "symbol is required")
		}
		out, err := adminService.ResumeSymbol(symbol)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return c.JSON(out)
	})

	protected.Post("/admin/symbols/:symbol/ensure", func(c *fiber.Ctx) error {
		if adminService == nil {
			return fiber.NewError(fiber.StatusServiceUnavailable, "admin service unavailable")
		}
		symbol := strings.TrimSpace(c.Params("symbol"))
		if symbol == "" {
			return fiber.NewError(fiber.StatusBadRequest, "symbol is required")
		}
		out, err := adminService.EnsureSymbol(symbol)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return c.JSON(out)
	})

	app.Get("/v1/markets/:symbol/trades", func(c *fiber.Ctx) error {
		symbol := strings.TrimSpace(c.Params("symbol"))
		if symbol == "" {
			return fiber.NewError(fiber.StatusBadRequest, "symbol is required")
		}
		limit := c.QueryInt("limit", 100)
		if limit <= 0 {
			return fiber.NewError(fiber.StatusBadRequest, "limit must be positive")
		}

		trades, err := trading.ListExecutions(symbol, limit)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return c.JSON(trades)
	})

	app.Get("/v1/markets/:symbol/book", func(c *fiber.Ctx) error {
		symbol := strings.TrimSpace(c.Params("symbol"))
		if symbol == "" {
			return fiber.NewError(fiber.StatusBadRequest, "symbol is required")
		}
		depth := c.QueryInt("depth", 20)
		if depth <= 0 {
			return fiber.NewError(fiber.StatusBadRequest, "depth must be positive")
		}

		snapshot, err := trading.ListOrderBook(symbol, depth)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return c.JSON(snapshot)
	})

	app.Get("/v1/markets/:symbol/candles", func(c *fiber.Ctx) error {
		if candleService == nil {
			return fiber.NewError(fiber.StatusServiceUnavailable, "candle service unavailable")
		}

		symbol := strings.TrimSpace(c.Params("symbol"))
		if symbol == "" {
			return fiber.NewError(fiber.StatusBadRequest, "symbol is required")
		}

		timeframe := normalizeTimeframe(c.Query("tf", "1m"))
		if !isSupportedTimeframe(timeframe) {
			return fiber.NewError(fiber.StatusBadRequest, "unsupported timeframe")
		}

		from, err := parseOptionalTime(c.Query("from"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid from timestamp")
		}
		to, err := parseOptionalTime(c.Query("to"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid to timestamp")
		}

		candles, err := candleService.ListCandles(symbol, timeframe, from, to)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return c.JSON(candles)
	})

	app.Get("/ws/trades/:symbol", websocket.New(func(conn *websocket.Conn) {
		symbol := strings.TrimSpace(conn.Params("symbol"))
		if symbol == "" {
			_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"symbol is required"}`))
			_ = conn.Close()
			return
		}

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		sentTradeIDs := map[string]struct{}{}
		for range ticker.C {
			trades, err := trading.ListExecutions(symbol, 20)
			if err != nil {
				_ = conn.WriteJSON(map[string]any{"type": "error", "message": err.Error()})
				continue
			}
			for _, trade := range trades {
				if _, seen := sentTradeIDs[trade.TradeID]; seen {
					continue
				}
				if err := conn.WriteJSON(map[string]any{"type": "trade", "data": trade}); err != nil {
					_ = conn.Close()
					return
				}
				sentTradeIDs[trade.TradeID] = struct{}{}
			}
		}
	}))

	app.Get("/ws/ticks/:symbol", websocket.New(func(conn *websocket.Conn) {
		if tickSource == nil {
			_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"tick stream not enabled"}`))
			_ = conn.Close()
			return
		}

		symbol := strings.TrimSpace(conn.Params("symbol"))
		if symbol == "" {
			_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"symbol is required"}`))
			_ = conn.Close()
			return
		}

		lastID := "$"
		for {
			ticks, nextID, err := tickSource.Read(context.Background(), lastID, 250, 2*time.Second)
			if err != nil {
				_ = conn.WriteJSON(map[string]any{"type": "error", "message": err.Error()})
				lastID = nextID
				continue
			}
			lastID = nextID

			for _, tick := range ticks {
				if !strings.EqualFold(strings.TrimSpace(tick.Symbol), symbol) {
					continue
				}
				if err := conn.WriteJSON(map[string]any{"type": "tick", "data": tick}); err != nil {
					_ = conn.Close()
					return
				}
			}
		}
	}))

	return app
}

func normalizeTimeframe(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func isSupportedTimeframe(value string) bool {
	switch value {
	case "1s", "5s", "1m", "5m", "1h":
		return true
	default:
		return false
	}
}

func parseOptionalTime(raw string) (time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		return time.Time{}, err
	}
	return parsed.UTC(), nil
}

func requireAuth(jwtSecret string, apiKeys map[string]string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		identity, err := authenticate(c, jwtSecret, apiKeys)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}
		c.Locals(authLocalKey, identity)
		return c.Next()
	}
}

func authenticate(c *fiber.Ctx, jwtSecret string, apiKeys map[string]string) (authIdentity, error) {
	apiKey := strings.TrimSpace(c.Get("X-API-Key"))
	if apiKey != "" {
		if userID, ok := apiKeys[apiKey]; ok && strings.TrimSpace(userID) != "" {
			return authIdentity{UserID: strings.TrimSpace(userID)}, nil
		}
		return authIdentity{}, errors.New("invalid API key")
	}

	authorization := strings.TrimSpace(c.Get("Authorization"))
	if !strings.HasPrefix(strings.ToLower(authorization), "bearer ") {
		return authIdentity{}, errors.New("missing authentication")
	}
	rawToken := strings.TrimSpace(authorization[len("Bearer "):])
	if rawToken == "" {
		return authIdentity{}, errors.New("missing bearer token")
	}

	parsed, err := jwt.Parse(rawToken, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || !parsed.Valid {
		return authIdentity{}, errors.New("invalid token")
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return authIdentity{}, errors.New("invalid token claims")
	}
	subjectRaw, ok := claims["sub"].(string)
	if !ok {
		return authIdentity{}, errors.New("token subject missing")
	}
	subject := strings.TrimSpace(subjectRaw)
	if subject == "" {
		return authIdentity{}, errors.New("token subject missing")
	}
	return authIdentity{UserID: subject}, nil
}
