package user

import (
	"context"
	"time"

	"github.com/ariesmaulana/payroll/app/user/lib"
	"github.com/ariesmaulana/payroll/common"
	"github.com/ariesmaulana/payroll/data"
	"github.com/ariesmaulana/payroll/lib/database"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var _ lib.StorageInterface = (*Storage)(nil)

type Storage struct {
	pool *pgxpool.Pool
}

func NewStorage(pool *pgxpool.Pool) *Storage {
	return &Storage{pool: pool}
}

func (s *Storage) BeginTxReader(ctx context.Context) (pgx.Tx, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// BeginTxWriter starts a read-write transaction and returns a pointer to pgx.Tx
func (s *Storage) BeginTxWriter(ctx context.Context) (pgx.Tx, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadWrite})
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (s *Storage) InsertUser(ctx context.Context, fullname string, username string, email string, password string, baseSalary int, joinDate time.Time) (int, error) {
	var id int
	currentTime := time.Now()
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (fullname, username, email, password_hash, base_salary, join_date, created_at, updated_at) 
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
         RETURNING id`,
		fullname, username, email, password, baseSalary, joinDate, currentTime, currentTime).Scan(&id)
	return id, err
}

func (s *Storage) GetUserByUsername(ctx context.Context, username string) (*data.User, database.ErrType, error) {
	user := &data.User{}
	err := s.pool.QueryRow(ctx,
		`SELECT  id, fullname, username, email, password_hash, role, base_salary, join_date, created_at, updated_at
         FROM users WHERE username = $1`,
		username).Scan(&user.Id, &user.Fullname, &user.Username, &user.Email, &user.Password,
		&user.Role, &user.BaseSalary, &user.JoinDate, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		// Return ErrNotFound error type when no rows are found
		if err == pgx.ErrNoRows {
			return nil, database.ErrNotFound, err
		} else {
			return nil, database.ErrUnset, err
		}
	}

	// Convert timezone to Asia/Jakarta
	user.JoinDate = common.TruncateToJakartaDate(user.JoinDate)
	user.CreatedAt = common.TruncateToJakartaDate(user.CreatedAt)
	user.UpdatedAt = common.TruncateToJakartaDate(user.UpdatedAt)
	return user, database.ErrUnset, nil
}

func (s *Storage) GetAllUserBaseSalary(ctx context.Context) (map[int]int, database.ErrType, error) {
	query := `SELECT id, base_salary FROM users`
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, database.ErrUnset, err
	}
	defer rows.Close()

	result := make(map[int]int)
	for rows.Next() {
		var id int
		var salary int
		if err := rows.Scan(&id, &salary); err != nil {
			return nil, database.ErrUnset, err
		}
		result[id] = salary
	}

	return result, database.ErrUnset, nil
}
