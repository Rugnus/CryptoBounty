package main

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	RPCURL                string
	ChainID               int64
	Confirmations         uint64
	PollInterval          time.Duration
	EscrowAddress         string
	PostgresURL           string
	RedisURL              string
	RedisStream           string
	BackfillBlocksOnStart uint64
}

func mustConfig() Config {
	chainID := mustInt64Env("CHAIN_ID")
	confirmations := mustUint64EnvDefault("CONFIRMATIONS", 2)
	pollMs := mustInt64EnvDefault("POLL_INTERVAL_MS", 2000)
	backfill := mustUint64EnvDefault("BACKFILL_BLOCKS", 500)
	return Config{
		RPCURL:                mustEnv("RPC_URL"),
		ChainID:               chainID,
		Confirmations:         confirmations,
		PollInterval:          time.Duration(pollMs) * time.Millisecond,
		EscrowAddress:         mustEnv("ESCROW_ADDRESS"),
		PostgresURL:           mustEnv("POSTGRES_URL"),
		RedisURL:              mustEnv("REDIS_URL"),
		RedisStream:           envDefault("REDIS_STREAM", "cb_events"),
		BackfillBlocksOnStart: backfill,
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

func mustUint64EnvDefault(k string, def uint64) uint64 {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		panic("bad env " + k)
	}
	return n
}

