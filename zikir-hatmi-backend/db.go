package main

import (
	"context"
	"errors"
	"log"
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

	// Wait for the database to become reachable. Postgres inside a
	// freshly-started compose/Portainer stack can take several seconds to
	// accept connections even after the container is up. Rather than failing
	// fast (which would crash the backend and send it into a restart loop),
	// we retry with a bounded backoff so the backend can wait for the db to
	// finish its initialisation.
	maxWait := 120 * time.Second
	deadline := time.Now().Add(maxWait)
	backoff := 500 * time.Millisecond
	const maxBackoff = 5 * time.Second

	var pool *pgxpool.Pool
	for {
		pool, err = pgxpool.NewWithConfig(ctx, config)
		if err == nil {
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err = pool.Ping(pingCtx)
			cancel()
			if err == nil {
				break
			}
			pool.Close()
		}

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if time.Now().After(deadline) {
			return nil, err
		}

		log.Printf("database not ready yet (%v), retrying in %s", err, backoff)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}

		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
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
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		-- migration: add title column if missing
		DO $$ BEGIN
			ALTER TABLE hatims ADD COLUMN IF NOT EXISTS title TEXT NOT NULL DEFAULT '';
		EXCEPTION WHEN others THEN NULL;
		END $$;
	`)
	return err
}
