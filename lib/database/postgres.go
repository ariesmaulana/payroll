package database

import (
	"context"
	"fmt"

	"github.com/ariesmaulana/payroll/config"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

func NewPostgresPool(cfg *config.Config) (*pgxpool.Pool, error) {
	connectionString := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?pool_max_conns=10",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	poolConfig, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse pool config: %v", err)
	}

	// make sure we select the correct schemae
	poolConfig.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
		// Set search_path for every new connection
		_, err := conn.Exec(ctx, "SET search_path TO public;")
		if err != nil {
			// We can't fail here since BeforeAcquire doesn't accept error returns
			// Instead, we'll log the error and return false to reject the connection
			fmt.Printf("Failed to set search_path: %v\n", err)
			return false
		}
		return true
	}

	pool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %v", err)
	}

	return pool, nil
}
