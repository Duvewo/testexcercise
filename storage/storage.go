package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDB(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, connString)
}
