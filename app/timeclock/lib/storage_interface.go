package lib

import (
	"context"
	"time"

	"github.com/ariesmaulana/payroll/data"
	"github.com/jackc/pgx/v4"
)

type StorageInterface interface {
	BeginTxReader(ctx context.Context) (pgx.Tx, error)
	BeginTxWriter(ctx context.Context) (pgx.Tx, error)

	InsertAttendanceCheckin(ctx context.Context, userId int, period time.Time, checkin time.Time, createdBy string) (int, error)
	UpdateAttendanceCheckout(ctx context.Context, userId int, period time.Time, checkout time.Time, updatedBy string) error

	GetDetailAttendance(ctx context.Context, id int) (*data.Attendance, error)
	GetAllAttendanceByPeriod(ctx context.Context, startDate, endDate time.Time) ([]*data.Attendance, error)
	GetDetailAttendanceByUserAndPeriod(ctx context.Context, userId int, period time.Time) (*data.Attendance, error)
	// Get list of attendance records for a user in the given period range
	GetAttendancesByUserAndPeriods(ctx context.Context, userId int, start time.Time, end time.Time) ([]*data.Attendance, error)

	InsertOvertime(ctx context.Context, userId int, period time.Time, hours int, reason, createdBy string) (int, error)
	GetOvertimeById(ctx context.Context, id int) (*data.Overtime, error)
	GetOvertimeByUserId(ctx context.Context, userId int) ([]*data.Overtime, error)

	//GetTotalOvertimeByPeriod Get total overtime hours (accumulated) in the given period for all users
	GetTotalOvertimeByPeriod(ctx context.Context, startDate time.Time, endDate time.Time) (int, error)

	//GetOvertimeHoursByPeriod will return map[userId]totalHours on this period
	GetOvertimeHoursByPeriod(ctx context.Context, startDate time.Time, endDate time.Time) (map[int]int, error)

	// Get list of overtime entries for a user in the given period range
	GetOvertimesByUserAndPeriod(ctx context.Context, userId int, start, end time.Time) ([]*data.Overtime, error)

	InsertReimbursement(ctx context.Context, userId int, period time.Time, amount int, description string, createdBy string) (int, error)
	GetDetailReimbursement(ctx context.Context, id int) (*data.Reimbursement, error)
	// Get total reimbursement amount (accumulated) in the given period for all users
	GetTotalReimbursementByPeriod(ctx context.Context, startDate time.Time, endDate time.Time) (int, error)

	// Get total reimbursement amount per user in the given period.
	// key = user_id, value = total amount reimbursed
	GetReimbursementTotalsByPeriod(ctx context.Context, startDate time.Time, endDate time.Time) (map[int]int, error)

	//GetReimbursementsByUserAndPeriod Get list of reimbursement entries for a user in the given period range
	GetReimbursementsByUserAndPeriod(ctx context.Context, userId int, start, end time.Time) ([]*data.Reimbursement, error)

	// InsertPayroll inserts a new payroll record for a specific period.
	//
	// totalAttendance: total number of attendance records within the period.
	// For example, if 10 users attended 5 days each, the total is 50.
	//
	// totalOvertime: total hours of overtime worked within the period.
	// For example, if 3 users each worked 2 hours, the total is 6.
	//
	// totalReimbursement: total reimbursement amount (in currency) submitted within the period.
	// For example, Rp100.000 + Rp150.000 = Rp250.000.
	//
	// totalSalary: final calculated total salary payout for all employees in this period.
	//
	// This function returns the generated payroll ID or an error if insert fails.
	InsertPayroll(ctx context.Context, periodStart time.Time, periodEnd time.Time, totalAttendance int, totalOvertime int,
		totalReimbursement int, totalSalary int, createdBy string) (int, error)
	IsPayrollAlreadyRun(ctx context.Context, date time.Time) (bool, error)

	// GetPayrollByPeriod retrieves payroll metadata for a given period (start to end).
	// It returns nil if no payroll is found.
	GetPayrollByPeriod(ctx context.Context, startDate time.Time, endDate time.Time) (*data.Payroll, error)

	// PayrollItem is the detail salary breakdown per user in a payroll period
	InsertPayrollItem(ctx context.Context, payrollId int, userId int, attendanceCount int, overtimeHours int, reimbursementTotal int, totalSalary int, createdBy string) (int, error)

	// GetPayrollItemsByPayrollID returns all payroll items for a specific payroll batch
	GetPayrollItemsByPayrollID(ctx context.Context, payrollId int) ([]*data.PayrollItem, error)

	// GetPayrollItemByPayrollIDAndUserID returns one user's payroll item in a specific payroll
	GetPayrollItemByPayrollIDAndUserID(ctx context.Context, payrollId int, userId int) (*data.PayrollItem, error)
}
