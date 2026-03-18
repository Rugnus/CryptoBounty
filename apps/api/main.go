package main

import (
	"context"
	"encoding/hex"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := mustConfig()
	ctx := context.Background()

	pg, err := pgxpool.New(ctx, cfg.PostgresURL)
	must(err)
	defer pg.Close()
	must(migrate(ctx, pg))

	rdb := redis.NewClient(&redis.Options{Addr: strings.TrimPrefix(cfg.RedisURL, "redis://")})
	defer rdb.Close()

	app := fiber.New(fiber.Config{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CorsOrigin,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	s := Server{
		cfg: cfg,
		pg:  pg,
		rdb: rdb,
	}

	api := app.Group("/api")

	api.Get("/health", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"ok": true}) })

	auth := api.Group("/auth")
	auth.Get("/siwe/nonce", s.GetSIWENonce)
	auth.Post("/siwe/verify", s.PostSIWEVerify)

	api.Get("/bounties", s.GetBounties)
	api.Get("/bounties/:id", s.GetBountyByID)
	api.Get("/notifications", s.AuthRequired, s.GetNotifications)
	api.Post("/webhooks/register", s.AuthRequired, s.PostWebhookRegister)

	log.Printf("api listening on %s", cfg.Addr)
	must(app.Listen(cfg.Addr))
}

type Server struct {
	cfg Config
	pg  *pgxpool.Pool
	rdb *redis.Client
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func addrToBytesLowerHex(addr string) ([]byte, error) {
	addr = strings.TrimPrefix(strings.ToLower(addr), "0x")
	return hex.DecodeString(addr)
}

