package data

import (
	"time"

	"github.com/jackc/pgtype"
)

type Attendance struct {
	Id           int
	UserId       int
	Periode      time.Time
	CheckinTime  time.Time
	CheckoutTime pgtype.Timestamp
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CreatedBy    string
	UpdatedBy    string
}

type Overtime struct {
	Id        int
	UserId    int
	Period    time.Time
	Hours     int
	Reason    string
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy string
	UpdatedBy string
}

type Reimbursement struct {
	Id          int
	UserId      int
	Period      time.Time
	Amount      int
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   string
	UpdatedBy   string
}

// Payroll is the main record that marks payroll has been processed for a specific period
type Payroll struct {
	Id                 int       // unique ID
	PeriodStart        time.Time // start date of the payroll period
	PeriodEnd          time.Time // end date of the payroll period
	TotalAttendance    int       // number of employees who had attendance in this period
	TotalOvertime      int       // total overtime hours from all employees
	TotalReimbursement int       // total reimbursement nominal from all employees
	TotalSalary        int       // total salary paid for all employees (including base, overtime, reimbursement)
	CreatedAt          time.Time
	UpdatedAt          time.Time
	CreatedBy          string
	UpdatedBy          string
}

// PayrollItem stores detailed payroll breakdown per employee in a specific payroll batch
type PayrollItem struct {
	Id                 int // unique ID
	PayrollId          int // foreign key to payroll.id
	UserId             int // user/employee this payroll item belongs to
	AttendanceCount    int // total days present during the payroll period
	OvertimeHours      int // total hours of overtime in the payroll period
	ReimbursementTotal int // total amount of approved reimbursements
	TotalSalary        int // final take-home pay for this user (base + overtime + reimbursement)
	CreatedAt          time.Time
	UpdatedAt          time.Time
	CreatedBy          string
	UpdatedBy          string
}

type UserPayslip struct {
	UserID           int
	TotalSalary      int
	AttendanceCount  int
	OvertimeHours    int
	ReimbursementSum int
}
