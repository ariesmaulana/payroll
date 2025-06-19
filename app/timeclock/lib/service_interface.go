package lib

import (
	"context"
	"time"

	"github.com/ariesmaulana/payroll/data"
	"github.com/ariesmaulana/payroll/lib/contextutil"
)

type ServiceInterface interface {
	AddAttendancePeriod(ctx context.Context, in *AddAttendancePeriodIn) *AddAttendancePeriodOut
	SubmitAttendance(ctx context.Context, in *SubmitAttendanceIn) *SubmitAttendanceOut
	AddOvertime(ctx context.Context, in *AddOvertimeIn) *AddOvertimeOut
	CheckoutAttendance(ctx context.Context, in *CheckoutAttendanceIn) *CheckoutAttendanceOut

	SubmitReimbursement(ctx context.Context, in *SubmitReimbursementIn) *SubmitReimbursementOut

	RunPayroll(ctx context.Context, in *RunPayrollIn) *RunPayrollOut

	GenerateSelfPaySlip(ctx context.Context, in *GenerateSelfPaySlipIn) *GenerateSelfPaySlipOut
	GenerateAllPaySlips(ctx context.Context, in *GenerateAllPaySlipsIn) *GenerateAllPaySlipsOut
}

type AddAttendancePeriodIn struct {
	Trace       *contextutil.Trace
	UserID      int
	CheckInDate time.Time
}

type AddAttendancePeriodOut struct {
	Success bool
	Message string
}

type SubmitAttendanceIn struct {
	Trace  *contextutil.Trace
	Period time.Time
}

type SubmitAttendanceOut struct {
	Success bool
	Message string
}

type AddOvertimeIn struct {
	Trace  *contextutil.Trace
	Period time.Time
	Hours  int
	Reason string
}

type AddOvertimeOut struct {
	Success bool
	Message string

	// Id this is id overtime, we need return this for testing purpose
	Id int
}

type CheckoutAttendanceIn struct {
	Trace  *contextutil.Trace
	Period time.Time
}

type CheckoutAttendanceOut struct {
	Success bool
	Message string
}

type SubmitReimbursementIn struct {
	Trace       *contextutil.Trace
	Amount      int
	Description string
	Period      time.Time
}

type SubmitReimbursementOut struct {
	Success bool
	Message string

	// Id this is id reimbursment, we need return this for testing purpose
	Id int
}

type RunPayrollIn struct {
	Trace       *contextutil.Trace
	PeriodStart time.Time
	PeriodEnd   time.Time
}

type RunPayrollOut struct {
	Success   bool
	Message   string
	PayrollId int // for testing purpose
}

type GenerateSelfPaySlipIn struct {
	Trace *contextutil.Trace
	Month int
	Year  int
}

type GenerateSelfPaySlipOut struct {
	Success bool
	Message string

	TotalSalary       int
	ListReimbursement []*data.Reimbursement
	ListOvertimes     []*data.Overtime
	ListAttendAnce    []*data.Attendance
}

type GenerateAllPaySlipsIn struct {
	Trace *contextutil.Trace
	Month int
	Year  int
}

type GenerateAllPaySlipsOut struct {
	Success          bool
	Message          string
	TotalSalaryAll   int
	ListUserPayslips []*data.UserPayslip
}
