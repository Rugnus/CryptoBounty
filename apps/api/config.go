package main

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Addr          string
	PostgresURL   string
	RedisURL      string
	JWTSecret     string
	SIWEDomain    string
	SIWEURI       string
	ChainID       int64
	NonceTTL      time.Duration
	CorsOrigin    string
	RedisStream   string
}

func mustConfig() Config {
	chainID := mustInt64Env("CHAIN_ID")
	nonceTTLSeconds := mustInt64EnvDefault("NONCE_TTL_SECONDS", 300)
	return Config{
		Addr:        envDefault("ADDR", ":8080"),
		PostgresURL: mustEnv("POSTGRES_URL"),
		RedisURL:    mustEnv("REDIS_URL"),
		JWTSecret:   mustEnv("JWT_SECRET"),
		SIWEDomain:  mustEnv("SIWE_DOMAIN"),
		SIWEURI:     mustEnv("SIWE_URI"),
		ChainID:     chainID,
		NonceTTL:    time.Duration(nonceTTLSeconds) * time.Second,
		CorsOrigin:  envDefault("CORS_ORIGIN", "http://localhost:5173"),
		RedisStream: envDefault("REDIS_STREAM", "cb_events"),
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

func mustInt64Env(k string) int64 {
	v := mustEnv(k)
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		panic("bad env " + k)
	}
	return n
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

