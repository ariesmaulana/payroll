package lib

import (
	"context"
	"time"

	"github.com/ariesmaulana/payroll/data"
	"github.com/jackc/pgx/v4"

	"github.com/ariesmaulana/payroll/lib/database"
)

type StorageInterface interface {
	BeginTxReader(ctx context.Context) (pgx.Tx, error)
	BeginTxWriter(ctx context.Context) (pgx.Tx, error)

	InsertUser(ctx context.Context, fullname string, username string, email string, password string, baseSalary int, joinDate time.Time) (int, error)
	GetUserByUsername(ctx context.Context, username string) (*data.User, database.ErrType, error)

	GetAllUserBaseSalary(ctx context.Context) (map[int]int, database.ErrType, error)
}
