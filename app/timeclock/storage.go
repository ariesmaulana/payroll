package timeclock

import (
	"context"
	"errors"
	"time"

	"github.com/ariesmaulana/payroll/app/timeclock/lib"
	"github.com/ariesmaulana/payroll/data"
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

func (s *Storage) InsertAttendanceCheckin(ctx context.Context, userId int, periode time.Time, checkin time.Time, createdBy string) (int, error) {
	var id int
	err := s.pool.QueryRow(ctx, `
		INSERT INTO attendances (user_id, period, checkin_time, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $4)
		RETURNING id
	`, userId, periode, checkin.Format("15:04:05"), createdBy).Scan(&id)

	return id, err
}

func (s *Storage) UpdateAttendanceCheckout(ctx context.Context, userId int, periode time.Time, checkout time.Time, updatedBy string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE attendances
		SET checkout_time = $1, updated_by = $2, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $3 AND period = $4
	`, checkout.Format("15:04:05"), updatedBy, userId, periode)
	return err
}

func (s *Storage) GetDetailAttendance(ctx context.Context, id int) (*data.Attendance, error) {
	const query = `
		SELECT id, user_id, period, checkin_time, checkout_time, created_at, updated_at, created_by, updated_by
		FROM attendances
		WHERE id = $1
	`

	row := s.pool.QueryRow(ctx, query, id)
	result := &data.Attendance{}

	err := row.Scan(
		&result.Id,
		&result.UserId,
		&result.Periode,
		&result.CheckinTime,
		&result.CheckoutTime, // <<< ini penting
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.CreatedBy,
		&result.UpdatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return result, nil
}

func (s *Storage) GetAllAttendanceByPeriod(ctx context.Context, startDate, endDate time.Time) ([]*data.Attendance, error) {
	const query = `
		SELECT id, user_id, period, checkin_time, checkout_time, created_at, updated_at, created_by, updated_by
		FROM attendances
		WHERE period BETWEEN $1 AND $2
		ORDER BY id
	`
	start := startDate.Format("2006-01-02")
	end := endDate.Format("2006-01-02")
	rows, err := s.pool.Query(ctx, query, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*data.Attendance, 0)
	for rows.Next() {
		var a data.Attendance
		err := rows.Scan(
			&a.Id,
			&a.UserId,
			&a.Periode,
			&a.CheckinTime,
			&a.CheckoutTime,
			&a.CreatedAt,
			&a.UpdatedAt,
			&a.CreatedBy,
			&a.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, &a)

	}
	return result, nil
}

func (s *Storage) InsertOvertime(ctx context.Context, userId int, period time.Time, hours int, reason, createdBy string) (int, error) {
	var id int
	err := s.pool.QueryRow(ctx, `
		INSERT INTO overtimes (user_id, period, hours, reason, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $5)
		RETURNING id
	`, userId, period, hours, reason, createdBy).Scan(&id)

	return id, err
}

func (s *Storage) GetOvertimeById(ctx context.Context, id int) (*data.Overtime, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, user_id, period, hours, reason, created_at, updated_at, created_by, updated_by
		FROM overtimes
		WHERE id = $1
	`, id)

	var ot data.Overtime
	err := row.Scan(
		&ot.Id, &ot.UserId, &ot.Period, &ot.Hours, &ot.Reason,
		&ot.CreatedAt, &ot.UpdatedAt, &ot.CreatedBy, &ot.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	return &ot, nil
}

func (s *Storage) GetOvertimeByUserId(ctx context.Context, userId int) ([]*data.Overtime, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, period, hours, reason, created_at, updated_at, created_by, updated_by
		FROM overtimes
		WHERE user_id = $1
		ORDER BY ot_date DESC
	`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*data.Overtime, 0)
	for rows.Next() {
		var ot data.Overtime
		err := rows.Scan(
			&ot.Id, &ot.UserId, &ot.Period, &ot.Hours, &ot.Reason,
			&ot.CreatedAt, &ot.UpdatedAt, &ot.CreatedBy, &ot.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, &ot)
	}
	return result, nil
}

func (s *Storage) GetDetailAttendanceByUserAndPeriod(ctx context.Context, userId int, period time.Time) (*data.Attendance, error) {
	const query = `
		SELECT id, user_id, period, checkin_time, checkout_time, created_at, updated_at, created_by, updated_by
		FROM attendances
		WHERE user_id = $1 AND period = $2
	`

	row := s.pool.QueryRow(ctx, query, userId, period)

	result := &data.Attendance{}
	err := row.Scan(
		&result.Id,
		&result.UserId,
		&result.Periode,
		&result.CheckinTime,
		&result.CheckoutTime,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.CreatedBy,
		&result.UpdatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

func (s *Storage) InsertReimbursement(ctx context.Context, userId int, period time.Time, amount int, description string, createdBy string) (int, error) {
	var id int
	err := s.pool.QueryRow(ctx, `
		INSERT INTO reimbursements (user_id, period, amount, description, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $5)
		RETURNING id
	`, userId, period, amount, description, createdBy).Scan(&id)

	return id, err
}

func (s *Storage) GetDetailReimbursement(ctx context.Context, id int) (*data.Reimbursement, error) {
	const query = `
		SELECT id, user_id, period, amount, description, created_at, updated_at, created_by, updated_by
		FROM reimbursements
		WHERE id = $1
	`

	row := s.pool.QueryRow(ctx, query, id)

	result := &data.Reimbursement{}
	err := row.Scan(
		&result.Id,
		&result.UserId,
		&result.Period,
		&result.Amount,
		&result.Description,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.CreatedBy,
		&result.UpdatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return result, nil
}

func (s *Storage) InsertPayroll(
	ctx context.Context,
	periodStart time.Time,
	periodEnd time.Time,
	totalAttendance int,
	totalOvertime int,
	totalReimbursement int,
	totalSalary int,
	createdBy string,
) (int, error) {
	const query = `
		INSERT INTO payrolls (
			period_start,
			period_end,
			total_attendance,
			total_overtime,
			total_reimbursement,
			total_salary,
			created_by,
			updated_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		RETURNING id
	`

	var id int
	err := s.pool.QueryRow(ctx, query,
		periodStart,
		periodEnd,
		totalAttendance,
		totalOvertime,
		totalReimbursement,
		totalSalary,
		createdBy,
	).Scan(&id)

	return id, err
}

func (s *Storage) IsPayrollAlreadyRun(ctx context.Context, date time.Time) (bool, error) {
	const query = `
		SELECT 1
		FROM payrolls
		WHERE period_start <= $1 AND period_end >= $1
		LIMIT 1
	`

	var exists int
	err := s.pool.QueryRow(ctx, query, date).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *Storage) InsertPayrollItem(
	ctx context.Context,
	payrollId int,
	userId int,
	attendanceCount int,
	overtimeHours int,
	reimbursementTotal int,
	totalSalary int,
	createdBy string,
) (int, error) {
	var id int
	query := `
		INSERT INTO payroll_items (
			payroll_id, user_id, attendance_count, overtime_hours,
			reimbursement_total, total_salary,
			created_by, updated_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		RETURNING id
	`
	err := s.pool.QueryRow(
		ctx, query,
		payrollId, userId, attendanceCount, overtimeHours, reimbursementTotal, totalSalary, createdBy,
	).Scan(&id)
	return id, err
}

func (s *Storage) GetPayrollItemsByPayrollID(ctx context.Context, payrollId int) ([]*data.PayrollItem, error) {
	query := `
		SELECT id, payroll_id, user_id, attendance_count, overtime_hours,
		       reimbursement_total, total_salary,
		       created_at, updated_at, created_by, updated_by
		FROM payroll_items
		WHERE payroll_id = $1
	`
	rows, err := s.pool.Query(ctx, query, payrollId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*data.PayrollItem
	for rows.Next() {
		var item data.PayrollItem
		err := rows.Scan(
			&item.Id,
			&item.PayrollId,
			&item.UserId,
			&item.AttendanceCount,
			&item.OvertimeHours,
			&item.ReimbursementTotal,
			&item.TotalSalary,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.CreatedBy,
			&item.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, nil
}

func (s *Storage) GetPayrollItemByPayrollIDAndUserID(ctx context.Context, payrollId int, userId int) (*data.PayrollItem, error) {
	query := `
		SELECT id, payroll_id, user_id, attendance_count, overtime_hours,
		       reimbursement_total, total_salary,
		       created_at, updated_at, created_by, updated_by
		FROM payroll_items
		WHERE payroll_id = $1 AND user_id = $2
	`
	row := s.pool.QueryRow(ctx, query, payrollId, userId)

	var item data.PayrollItem
	err := row.Scan(
		&item.Id,
		&item.PayrollId,
		&item.UserId,
		&item.AttendanceCount,
		&item.OvertimeHours,
		&item.ReimbursementTotal,
		&item.TotalSalary,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.CreatedBy,
		&item.UpdatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (s *Storage) GetTotalOvertimeByPeriod(ctx context.Context, startDate time.Time, endDate time.Time) (int, error) {
	const query = `
		SELECT COALESCE(SUM(hours), 0)
		FROM overtimes
		WHERE period BETWEEN $1 AND $2
	`

	var total int
	err := s.pool.QueryRow(ctx, query, startDate, endDate).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (s *Storage) GetTotalReimbursementByPeriod(ctx context.Context, startDate time.Time, endDate time.Time) (int, error) {
	const query = `
		SELECT COALESCE(SUM(amount), 0)
		FROM reimbursements
		WHERE period BETWEEN $1 AND $2
	`

	var total int
	err := s.pool.QueryRow(ctx, query, startDate, endDate).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (s *Storage) GetOvertimeHoursByPeriod(ctx context.Context, startDate time.Time, endDate time.Time) (map[int]int, error) {
	const query = `
		SELECT user_id, SUM(hours)
		FROM overtimes
		WHERE period BETWEEN $1 AND $2
		GROUP BY user_id
	`

	rows, err := s.pool.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int]int)
	for rows.Next() {
		var userId, totalHours int
		if err := rows.Scan(&userId, &totalHours); err != nil {
			return nil, err
		}
		result[userId] = totalHours
	}

	return result, nil
}

func (s *Storage) GetReimbursementTotalsByPeriod(ctx context.Context, startDate time.Time, endDate time.Time) (map[int]int, error) {
	const query = `
		SELECT user_id, SUM(amount)
		FROM reimbursements
		WHERE period BETWEEN $1 AND $2
		GROUP BY user_id
	`

	rows, err := s.pool.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int]int)
	for rows.Next() {
		var userID, total int
		if err := rows.Scan(&userID, &total); err != nil {
			return nil, err
		}
		result[userID] = total
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Storage) GetPayrollByPeriod(ctx context.Context, startDate, endDate time.Time) (*data.Payroll, error) {
	const query = `
		SELECT id, period_start, period_end, total_attendance, total_overtime,
		       total_reimbursement, total_salary, created_at, updated_at, created_by, updated_by
		FROM payrolls
		WHERE period_start = $1 AND period_end = $2
		LIMIT 1
	`
	row := s.pool.QueryRow(ctx, query, startDate, endDate)

	var p data.Payroll
	err := row.Scan(
		&p.Id,
		&p.PeriodStart,
		&p.PeriodEnd,
		&p.TotalAttendance,
		&p.TotalOvertime,
		&p.TotalReimbursement,
		&p.TotalSalary,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.CreatedBy,
		&p.UpdatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &p, nil
}

func (s *Storage) GetReimbursementsByUserAndPeriod(ctx context.Context, userId int, start, end time.Time) ([]*data.Reimbursement, error) {
	const query = `
		SELECT id, user_id, period, amount, description,
		       created_at, updated_at, created_by, updated_by
		FROM reimbursements
		WHERE user_id = $1 AND period BETWEEN $2 AND $3
	`

	rows, err := s.pool.Query(ctx, query, userId, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*data.Reimbursement
	for rows.Next() {
		var r data.Reimbursement
		err := rows.Scan(
			&r.Id,
			&r.UserId,
			&r.Period,
			&r.Amount,
			&r.Description,
			&r.CreatedAt,
			&r.UpdatedAt,
			&r.CreatedBy,
			&r.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, &r)
	}

	return result, nil
}

func (s *Storage) GetOvertimesByUserAndPeriod(ctx context.Context, userId int, start, end time.Time) ([]*data.Overtime, error) {
	const query = `
		SELECT id, user_id, period, hours, reason,
		       created_at, updated_at, created_by, updated_by
		FROM overtimes
		WHERE user_id = $1 AND period BETWEEN $2 AND $3
	`

	rows, err := s.pool.Query(ctx, query, userId, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*data.Overtime
	for rows.Next() {
		var o data.Overtime
		err := rows.Scan(
			&o.Id,
			&o.UserId,
			&o.Period,
			&o.Hours,
			&o.Reason,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.CreatedBy,
			&o.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, &o)
	}

	return result, nil
}

func (s *Storage) GetAttendancesByUserAndPeriods(ctx context.Context, userId int, start time.Time, end time.Time) ([]*data.Attendance, error) {
	const query = `
		SELECT id, user_id, period, checkin_time, checkout_time,
		       created_at, updated_at, created_by, updated_by
		FROM attendances
		WHERE user_id = $1 AND period BETWEEN $2 AND $3
	`

	rows, err := s.pool.Query(ctx, query, userId, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*data.Attendance
	for rows.Next() {
		var a data.Attendance
		err := rows.Scan(
			&a.Id,
			&a.UserId,
			&a.Periode,
			&a.CheckinTime,
			&a.CheckoutTime,
			&a.CreatedAt,
			&a.UpdatedAt,
			&a.CreatedBy,
			&a.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, &a)
	}

	return result, nil
}
