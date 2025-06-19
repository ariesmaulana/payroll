package user

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ariesmaulana/payroll/data"
	"github.com/ariesmaulana/payroll/lib/test"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// TestHelper provides utilities for managing test data and transactions
type TestHelper struct {
	db         *pgxpool.Pool
	ctx        context.Context
	tx         pgx.Tx
	schemaName string
	t          *testing.T
}

// Fixture represents test data that can be loaded into the database
type Fixture struct {
	Table string
	Data  map[string]interface{}
}

// NewTestHelper creates a new TestHelper instance using an existing DbTestSuite
func NewTestHelper(t *testing.T, dbSuite *test.DbTestSuite) *TestHelper {
	// Generate schema name from test name
	schemaName := strings.ToLower(strings.Replace(t.Name(), "/", "_", -1))

	return &TestHelper{
		db:         dbSuite.Pool,
		ctx:        dbSuite.Context,
		schemaName: schemaName,
		t:          t,
	}
}

// BeginTx starts a new transaction
func (h *TestHelper) BeginTx() error {
	// Set schema for this connection
	_, err := h.db.Exec(h.ctx, fmt.Sprintf("SET search_path TO %s;", h.schemaName))
	if err != nil {
		return fmt.Errorf("failed to set schema: %w", err)
	}

	tx, err := h.db.Begin(h.ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	h.tx = tx
	return nil
}

// RollbackTx rolls back the current transaction
func (h *TestHelper) CommitTx() error {
	if h.tx != nil {
		err := h.tx.Commit(h.ctx)
		h.tx = nil
		if err != nil && !strings.Contains(err.Error(), "already closed") {
			return fmt.Errorf("failed to rollback transaction: %w", err)
		}
	}
	return nil
}

func (h *TestHelper) CreateUserFixture(data *data.User) (*data.User, error) {
	if h.tx == nil {
		return nil, fmt.Errorf("no active transaction")
	}

	currentTime := time.Now()
	var id int

	err := h.tx.QueryRow(h.ctx,
		`INSERT INTO users (fullname, username, email, password_hash, role, base_salary, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at, updated_at`,
		data.Fullname,
		data.Username,
		data.Email,
		data.Password,
		data.Role,
		data.BaseSalary,
		currentTime,
		currentTime,
	).Scan(&id, &currentTime, &currentTime)

	if err != nil {
		return nil, fmt.Errorf("failed to insert user fixture: %w", err)
	}

	// Return complete user data
	data.Id = id
	data.CreatedAt = currentTime
	data.UpdatedAt = currentTime
	return data, nil
}
