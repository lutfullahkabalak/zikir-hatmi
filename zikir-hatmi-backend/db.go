package main

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func initDB(ctx context.Context) (*pgxpool.Pool, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}

	if err := ensureSchema(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func ensureSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS hatims (
			id UUID PRIMARY KEY,
			share_code TEXT UNIQUE NOT NULL,
			title TEXT NOT NULL DEFAULT '',
			count INT NOT NULL DEFAULT 0,
			target INT NOT NULL DEFAULT 50,
			password_hash TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS hatim_tokens (
			token TEXT PRIMARY KEY,
			hatim_id UUID NOT NULL REFERENCES hatims(id) ON DELETE CASCADE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '30 days')
		);

		-- migration: add title column if missing
		DO $$ BEGIN
			ALTER TABLE hatims ADD COLUMN IF NOT EXISTS title TEXT NOT NULL DEFAULT '';
		EXCEPTION WHEN others THEN NULL;
		END $$;

		-- migration: add expires_at column to hatim_tokens if missing
		DO $$ BEGIN
			ALTER TABLE hatim_tokens ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '30 days');
		EXCEPTION WHEN others THEN NULL;
		END $$;
	`)
	return err
}
