package database

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, url string, log *slog.Logger) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	tryToConn := 3
	var lastErr error
	for i := 1; i <= tryToConn; i++ {
		err := pool.Ping(ctx)
		if err == nil {
			log.Info("success conection to DB", "try:", i)
			return pool, nil
		}
		lastErr = err
		log.Warn("failed to connection to DB, retrying...", "error", lastErr)
		delay := 1 * i

		select {
		case <-time.After(time.Second * time.Duration(delay)):

		case <-ctx.Done():
			log.Error("context canceled during BD retry.")
			return nil, ctx.Err()
		}

	}
	pool.Close()
	return nil, lastErr
}
