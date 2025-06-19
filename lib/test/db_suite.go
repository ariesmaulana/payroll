package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type DbTestSuite struct {
	Pool    *pgxpool.Pool
	Context context.Context
}

func DbTestPool(t *testing.T) *DbTestSuite {
	t.Helper()
	ctx := context.Background()

	//Normalize t.Name()
	schemaName := strings.ToLower(strings.Replace(t.Name(), "/", "_", -1))

	connectionString := "postgresql://postgres:localdb123@localhost/test_db_1?sslmode=disable"

	poolConfig, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	// Set up BeforeAcquire before creating the pool
	// this part make sure for every connection always use the correct schema
	poolConfig.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
		// Set search_path for every new connection
		_, err := conn.Exec(ctx, fmt.Sprintf("SET search_path TO %s;", schemaName))
		if err != nil {
			t.Fatalf("set schema failed. err: %v", err)
		}
		return true
	}

	pool, err := pgxpool.ConnectConfig(ctx, poolConfig)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	// Make sure every test using dbtest suite always close the pool & drop the schema
	t.Cleanup(func() {
		_, err = pool.Exec(ctx, "DROP SCHEMA "+schemaName+" CASCADE")
		if err != nil {
			t.Fatalf("db cleanup failed. err: %v", err)
		}
		pool.Close()
	})

	// create the schema
	_, err = pool.Exec(ctx, "CREATE SCHEMA "+schemaName)
	if err != nil {
		t.Fatalf("schema creation failed. err: %v", err)
	}

	// use schema
	query := fmt.Sprintf("SET search_path TO %s;", schemaName)
	_, err = pool.Exec(ctx, query)
	if err != nil {
		t.Fatalf("error while switching to schema. err: %v", err)
	}

	// populate the table
	schemaDir := "../../schema" // Adjust path if needed
	files, err := filepath.Glob(filepath.Join(schemaDir, "*.sql"))
	if err != nil {
		t.Fatalf("failed to read schema directory: %v", err)
	}

	if len(files) == 0 {
		t.Fatalf("no SQL files found in schema directory")
	}

	for _, schemaPath := range files {
		file, err := os.ReadFile(schemaPath)
		if err != nil {
			t.Fatalf("error reading file %s: %v", schemaPath, err)
		}

		_, err = pool.Exec(ctx, string(file))
		if err != nil {
			fmt.Printf("error executing %s: %v\n", schemaPath, err)
			t.Fatal("Failed to execute file")
		} else {
			fmt.Printf("executed %s successfully\n", schemaPath)
		}
	}

	return &DbTestSuite{
		Pool:    pool,
		Context: ctx,
	}
}
