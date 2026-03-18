package main

import (
	"context"
	_ "embed"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed db.sql
var schemaSQL string

func migrate(ctx context.Context, pg *pgxpool.Pool) error {
	// naive splitter is OK for our simple schema file
	stmts := strings.Split(schemaSQL, ";")
	for _, s := range stmts {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, err := pg.Exec(ctx, s); err != nil {
			return err
		}
	}
	return nil
}

