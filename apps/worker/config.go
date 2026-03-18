package main

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	PostgresURL string
	RedisURL    string
	RedisStream string
	Group       string
	Consumer    string
	Block       time.Duration

	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string
	SMTPFrom string
}

func mustConfig() Config {
	return Config{
		PostgresURL: mustEnv("POSTGRES_URL"),
		RedisURL:    mustEnv("REDIS_URL"),
		RedisStream: envDefault("REDIS_STREAM", "cb_events"),
		Group:       envDefault("REDIS_GROUP", "cb_worker"),
		Consumer:    envDefault("REDIS_CONSUMER", "c1"),
		Block:       time.Duration(mustInt64EnvDefault("REDIS_BLOCK_MS", 2000)) * time.Millisecond,

		SMTPHost: os.Getenv("SMTP_HOST"),
		SMTPPort: int(mustInt64EnvDefault("SMTP_PORT", 0)),
		SMTPUser: os.Getenv("SMTP_USER"),
		SMTPPass: os.Getenv("SMTP_PASS"),
		SMTPFrom: os.Getenv("SMTP_FROM"),
	}
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic("missing env " + k)
	}
	return v
}

func envDefault(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func mustInt64EnvDefault(k string, def int64) int64 {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		panic("bad env " + k)
	}
	return n
}

