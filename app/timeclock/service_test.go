package timeclock

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ariesmaulana/payroll/app/timeclock/lib"
	"github.com/ariesmaulana/payroll/app/timeclock/mock_lib"

	mocks "github.com/ariesmaulana/payroll/app/timeclock/mock_lib"
	userLib "github.com/ariesmaulana/payroll/app/user/lib"

	"github.com/ariesmaulana/payroll/common"
	"github.com/ariesmaulana/payroll/data"
	"github.com/ariesmaulana/payroll/lib/contextutil"
	"github.com/ariesmaulana/payroll/lib/test"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestServiceAddAttendancePeriod(t *testing.T) {
	t.Parallel()

	con := test.DbTestPool(t)
	timeclockStorage := NewStorage(con.Pool)
	// setup gomock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create mock user service
	userServiceMock := mocks.NewMockServiceInterface(ctrl)
	service := NewService(timeclockStorage, userServiceMock)

	// setup test users & contexts
	userId := 999
	username := "admin_user"

	adminCtx := contextutil.WithUser(con.Context, &contextutil.AuthUser{
		Id:       userId,
		Username: username,
		Role:     data.RAdmin,
	})

	employeeCtx := contextutil.WithUser(con.Context, &contextutil.AuthUser{
		Id:       userId,
		Username: username,
		Role:     data.REmployee,
	})

	noUserCtx := con.Context

	trace := &contextutil.Trace{TraceID: "test-attendance-period"}

	// Weekend (misalnya: Sabtu)
	saturday := time.Date(2025, 6, 14, 9, 0, 0, 0, time.UTC)
	// Weekday (misalnya: Senin)
	monday := time.Date(2025, 6, 16, 9, 0, 0, 0, time.UTC)

	type input struct {
		ctx context.Context
		in  *lib.AddAttendancePeriodIn
	}

	type expected struct {
		success bool
		errMsg  string
	}

	type testCase struct {
		name     string
		input    input
		expected expected
	}

	scenarios := []testCase{
		{
			name: "success add new attendance period (valid weekday)",
			input: input{
				ctx: adminCtx,
				in: &lib.AddAttendancePeriodIn{
					Trace:       trace,
					CheckInDate: monday,
					UserID:      userId,
				},
			},
			expected: expected{success: true},
		},
		{
			name: "fails when date is weekend",
			input: input{
				ctx: adminCtx,
				in: &lib.AddAttendancePeriodIn{
					Trace:       trace,
					CheckInDate: saturday,
					UserID:      userId,
				},
			},
			expected: expected{
				success: false,
				errMsg:  "Tidak bisa mengisi kehadiran saat Sabtu dan Minggu.",
			},
		},
		{
			name: "fails when user is not admin",
			input: input{
				ctx: employeeCtx,
				in: &lib.AddAttendancePeriodIn{
					Trace:       trace,
					CheckInDate: monday,
					UserID:      userId,
				},
			},
			expected: expected{
				success: false,
				errMsg:  "forbidden: Hanya admin yang bisa akses",
			},
		},
		{
			name: "fails when user is missing in context",
			input: input{
				ctx: noUserCtx,
				in: &lib.AddAttendancePeriodIn{
					Trace:       trace,
					CheckInDate: monday,
					UserID:      userId,
				},
			},
			expected: expected{
				success: false,
				errMsg:  "unauthorized",
			},
		},
		{
			name: "fails when CheckInDate is zero",
			input: input{
				ctx: adminCtx,
				in: &lib.AddAttendancePeriodIn{
					Trace:       trace,
					CheckInDate: time.Time{},
					UserID:      userId,
				},
			},
			expected: expected{
				success: false,
				errMsg:  "Checkin date wajib diisi",
			},
		},
	}

	for _, sc := range scenarios {
		sc := sc
		t.Run(sc.name, func(t *testing.T) {
			t.Parallel()
			result := service.AddAttendancePeriod(sc.input.ctx, sc.input.in)
			assert.Equal(t, sc.expected.success, result.Success)
			assert.Equal(t, sc.expected.errMsg, result.Message)
		})
	}
}

func setupUserContext(role data.UserRole) (context.Context, int, string) {
	userID := 999
	username := "test_user"
	ctx := contextutil.WithUser(context.Background(), &contextutil.AuthUser{
		Id:       userID,
		Username: username,
		Role:     role,
	})
	return ctx, userID, username
}

func TestServiceSubmitAttendance(t *testing.T) {
	t.Parallel()
	con := test.DbTestPool(t)
	timeclockStorage := NewStorage(con.Pool)
	// setup gomock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create mock user service
	userServiceMock := mocks.NewMockServiceInterface(ctrl)
	service := NewService(timeclockStorage, userServiceMock)

	ctx, _, _ := setupUserContext(data.REmployee)
	trace := &contextutil.Trace{TraceID: "submit-attendance-test"}
	validDate := time.Date(2025, 6, 17, 9, 0, 0, 0, time.UTC)
	weekend := time.Date(2025, 6, 15, 9, 0, 0, 0, time.UTC)

	scenarios := []struct {
		name    string
		ctx     context.Context
		in      *lib.SubmitAttendanceIn
		success bool
		errMsg  string
	}{
		{
			name:    "success on weekday",
			ctx:     ctx,
			in:      &lib.SubmitAttendanceIn{Trace: trace, Period: validDate},
			success: true,
		},
		{
			name:    "fail on weekend",
			ctx:     ctx,
			in:      &lib.SubmitAttendanceIn{Trace: trace, Period: weekend},
			success: false,
			errMsg:  "Tidak bisa mengisi kehadiran saat Sabtu dan Minggu.",
		},
		{
			name:    "fail on empty period",
			ctx:     ctx,
			in:      &lib.SubmitAttendanceIn{Trace: trace},
			success: false,
			errMsg:  "Wajib pilih periode waktu checkin",
		},
		{
			name:    "unauthorized",
			ctx:     context.Background(),
			in:      &lib.SubmitAttendanceIn{Trace: trace, Period: validDate},
			success: false,
			errMsg:  "unauthorized",
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			resp := service.SubmitAttendance(sc.ctx, sc.in)
			assert.Equal(t, sc.success, resp.Success)
			assert.Equal(t, sc.errMsg, resp.Message)
		})
	}
}

func TestServiceAddOvertime(t *testing.T) {
	t.Parallel()
	con := test.DbTestPool(t)
	timeclockStorage := NewStorage(con.Pool)
	// setup gomock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create mock user service
	userServiceMock := mocks.NewMockServiceInterface(ctrl)
	service := NewService(timeclockStorage, userServiceMock)

	ctx, userId, userName := setupUserContext(data.REmployee)
	trace := &contextutil.Trace{TraceID: "add-overtime-test"}
	validPeriod := common.NewDateTime(2025, 1, 1, 18, 0, 0)

	tx, err := timeclockStorage.BeginTxWriter(ctx)
	assert.Nil(t, err)
	defer tx.Rollback(ctx)

	_, err = timeclockStorage.InsertAttendanceCheckin(ctx, userId, validPeriod, validPeriod.Add(18*time.Minute), userName)
	assert.Nil(t, err)

	err = tx.Commit(ctx)
	assert.Nil(t, err)

	scenarios := []struct {
		name    string
		ctx     context.Context
		in      *lib.AddOvertimeIn
		success bool
		errMsg  string
	}{
		{
			name:    "success",
			ctx:     ctx,
			in:      &lib.AddOvertimeIn{Trace: trace, Period: validPeriod, Hours: 2, Reason: "Project deadline"},
			success: true,
		},
		{
			name:    "fail hours over limit",
			ctx:     ctx,
			in:      &lib.AddOvertimeIn{Trace: trace, Period: validPeriod, Hours: 5, Reason: "Extra work"},
			success: false,
			errMsg:  "Jumlah jam lembur tidak boleh lebih dari 3",
		},
		{
			name:    "fail reason empty",
			ctx:     ctx,
			in:      &lib.AddOvertimeIn{Trace: trace, Period: validPeriod, Hours: 2, Reason: ""},
			success: false,
			errMsg:  "Alasan harus diisi",
		},
		{
			name:    "fail period zero",
			ctx:     ctx,
			in:      &lib.AddOvertimeIn{Trace: trace, Period: time.Time{}, Hours: 2, Reason: "Late meeting"},
			success: false,
			errMsg:  "Wajib pilih periode waktu overtime",
		},

		{
			name:    "fail period not found",
			ctx:     ctx,
			in:      &lib.AddOvertimeIn{Trace: trace, Period: common.NewDateTime(2025, 1, 2, 18, 0, 0), Hours: 2, Reason: "Late meeting"},
			success: false,
			errMsg:  "Anda belum absen di hari tersebut",
		},
		{
			name:    "fail before working hours end",
			ctx:     ctx,
			in:      &lib.AddOvertimeIn{Trace: trace, Period: validPeriod.Add(-2 * time.Hour), Hours: 2, Reason: "Early overtime"},
			success: false,
			errMsg:  "Lembur hanya bisa diajukan setelah jam kerja selesai",
		},
		{
			name:    "unauthorized",
			ctx:     context.Background(),
			in:      &lib.AddOvertimeIn{Trace: trace, Period: validPeriod, Hours: 2, Reason: "Out of scope"},
			success: false,
			errMsg:  "unauthorized",
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			resp := service.AddOvertime(sc.ctx, sc.in)
			assert.Equal(t, sc.success, resp.Success)
			assert.Equal(t, sc.errMsg, resp.Message)
		})
	}
}

func TestServiceCheckoutAttendance(t *testing.T) {
	t.Parallel()
	con := test.DbTestPool(t)
	timeclockStorage := NewStorage(con.Pool)
	// setup gomock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create mock user service
	userServiceMock := mocks.NewMockServiceInterface(ctrl)
	service := NewService(timeclockStorage, userServiceMock)

	ctx, _, _ := setupUserContext(data.REmployee)
	trace := &contextutil.Trace{TraceID: "checkout-attendance-test"}
	validDate := time.Date(2025, 6, 17, 18, 0, 0, 0, time.UTC)

	scenarios := []struct {
		name    string
		ctx     context.Context
		in      *lib.CheckoutAttendanceIn
		success bool
		errMsg  string
	}{
		{
			name:    "success checkout",
			ctx:     ctx,
			in:      &lib.CheckoutAttendanceIn{Trace: trace, Period: validDate},
			success: true,
		},
		{
			name:    "fail period zero",
			ctx:     ctx,
			in:      &lib.CheckoutAttendanceIn{Trace: trace},
			success: false,
			errMsg:  "Wajib pilih periode waktu overtime",
		},
		{
			name:    "unauthorized",
			ctx:     context.Background(),
			in:      &lib.CheckoutAttendanceIn{Trace: trace, Period: validDate},
			success: false,
			errMsg:  "unauthorized",
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			resp := service.CheckoutAttendance(sc.ctx, sc.in)
			assert.Equal(t, sc.success, resp.Success)
			assert.Equal(t, sc.errMsg, resp.Message)
		})
	}
}

func TestServiceSubmitReimbursement(t *testing.T) {
	t.Parallel()
	con := test.DbTestPool(t)
	timeclockStorage := NewStorage(con.Pool)
	// setup gomock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create mock user service
	userServiceMock := mocks.NewMockServiceInterface(ctrl)
	service := NewService(timeclockStorage, userServiceMock)

	ctx, _, _ := setupUserContext(data.REmployee)
	trace := &contextutil.Trace{TraceID: "submit-reimbursement-test"}
	validDate := time.Date(2025, 6, 17, 0, 0, 0, 0, time.UTC)

	scenarios := []struct {
		name    string
		ctx     context.Context
		in      *lib.SubmitReimbursementIn
		success bool
		errMsg  string
	}{
		{
			name:    "success submit",
			ctx:     ctx,
			in:      &lib.SubmitReimbursementIn{Trace: trace, Period: validDate, Amount: 100000, Description: "Makan siang kantor"},
			success: true,
		},
		{
			name:    "fail period zero",
			ctx:     ctx,
			in:      &lib.SubmitReimbursementIn{Trace: trace, Amount: 100000, Description: "Transport"},
			success: false,
			errMsg:  "Periode wajib diisi",
		},
		{
			name:    "fail amount zero",
			ctx:     ctx,
			in:      &lib.SubmitReimbursementIn{Trace: trace, Period: validDate, Amount: 0, Description: "Transport"},
			success: false,
			errMsg:  "Jumlah reimbursement harus lebih dari 0",
		},
		{
			name:    "fail empty description",
			ctx:     ctx,
			in:      &lib.SubmitReimbursementIn{Trace: trace, Period: validDate, Amount: 100000},
			success: false,
			errMsg:  "Deskripsi reimbursement wajib diisi",
		},
		{
			name:    "unauthorized",
			ctx:     context.Background(),
			in:      &lib.SubmitReimbursementIn{Trace: trace, Period: validDate, Amount: 100000, Description: "Makan siang kantor"},
			success: false,
			errMsg:  "unauthorized",
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			resp := service.SubmitReimbursement(sc.ctx, sc.in)
			assert.Equal(t, sc.success, resp.Success)
			assert.Equal(t, sc.errMsg, resp.Message)
		})
	}
}

func TestServiceGenerateSelfPaySlip(t *testing.T) {
	t.Parallel()
	con := test.DbTestPool(t)
	timeclockStorage := NewStorage(con.Pool)
	// setup gomock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create mock user service
	userServiceMock := mocks.NewMockServiceInterface(ctrl)
	service := NewService(timeclockStorage, userServiceMock)

	// Setup user dan payroll data
	ctx, userId, _ := setupUserContext(data.REmployee)
	trace := &contextutil.Trace{TraceID: "self-payslip-test"}

	periodStart := common.NewDate(2025, 1, 1)
	periodEnd := common.NewDate(2025, 1, 31)

	tx, err := timeclockStorage.BeginTxWriter(ctx)
	assert.Nil(t, err)
	defer tx.Rollback(ctx)

	payrollID, err := timeclockStorage.InsertPayroll(ctx, periodStart, periodEnd, 10, 5, 100000, 1000000, "admin")
	assert.Nil(t, err)

	_, err = timeclockStorage.InsertPayrollItem(ctx, payrollID, userId, 10, 5, 100000, 1000000, "admin")
	assert.Nil(t, err)

	scenarios := []struct {
		name    string
		ctx     context.Context
		in      *lib.GenerateSelfPaySlipIn
		success bool
		errMsg  string
	}{
		{
			name:    "success",
			ctx:     ctx,
			in:      &lib.GenerateSelfPaySlipIn{Trace: trace, Month: 1, Year: 2025},
			success: true,
		},
		{
			name:    "unauthorized",
			ctx:     context.Background(),
			in:      &lib.GenerateSelfPaySlipIn{Trace: trace, Month: 1, Year: 2025},
			success: false,
			errMsg:  "unauthorized",
		},
		{
			name:    "invalid input",
			ctx:     ctx,
			in:      &lib.GenerateSelfPaySlipIn{Trace: trace, Month: 0, Year: 2025},
			success: false,
			errMsg:  "Bulan atau tahun tidak valid",
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			resp := service.GenerateSelfPaySlip(sc.ctx, sc.in)
			assert.Equal(t, sc.success, resp.Success)
			assert.Equal(t, sc.errMsg, resp.Message)
		})
	}
}

func TestServiceGenerateAllPaySlips(t *testing.T) {
	t.Parallel()
	con := test.DbTestPool(t)
	timeclockStorage := NewStorage(con.Pool)
	// setup gomock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create mock user service
	userServiceMock := mocks.NewMockServiceInterface(ctrl)
	service := NewService(timeclockStorage, userServiceMock)

	ctx, _, _ := setupUserContext(data.RAdmin)
	trace := &contextutil.Trace{TraceID: "admin-payslip-test"}

	periodStart := common.NewDate(2025, 1, 1)
	periodEnd := common.NewDate(2025, 1, 31)

	tx, err := timeclockStorage.BeginTxWriter(ctx)
	assert.Nil(t, err)
	defer tx.Rollback(ctx)

	payrollID, err := timeclockStorage.InsertPayroll(ctx, periodStart, periodEnd, 10, 5, 100000, 1000000, "admin")
	assert.Nil(t, err)

	_ = []int{1, 2, 3}
	for i := 1; i <= 3; i++ {
		_, err := timeclockStorage.InsertPayrollItem(ctx, payrollID, i, 10, 2, 50000, 500000, "admin")
		assert.Nil(t, err)
	}

	scenarios := []struct {
		name    string
		ctx     context.Context
		in      *lib.GenerateAllPaySlipsIn
		success bool
		errMsg  string
	}{
		{
			name:    "success",
			ctx:     ctx,
			in:      &lib.GenerateAllPaySlipsIn{Trace: trace, Month: 1, Year: 2025},
			success: true,
		},
		{
			name:    "unauthorized",
			ctx:     context.Background(),
			in:      &lib.GenerateAllPaySlipsIn{Trace: trace, Month: 1, Year: 2025},
			success: false,
			errMsg:  "unauthorized",
		},
		{
			name:    "invalid input",
			ctx:     ctx,
			in:      &lib.GenerateAllPaySlipsIn{Trace: trace, Month: 0, Year: 2025},
			success: false,
			errMsg:  "Bulan atau tahun tidak valid",
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			resp := service.GenerateAllPaySlips(sc.ctx, sc.in)
			assert.Equal(t, sc.success, resp.Success)
			assert.Equal(t, sc.errMsg, resp.Message)
		})
	}
}

func TestServiceRunPayroll(t *testing.T) {
	t.Parallel()

	con := test.DbTestPool(t)
	timeclockStorage := NewStorage(con.Pool)

	// Setup gomock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userServiceMock := mock_lib.NewMockServiceInterface(ctrl)

	service := NewService(timeclockStorage, userServiceMock)

	// Setup context dan user
	ctx, _, userName := setupUserContext(data.RAdmin)
	trace := &contextutil.Trace{TraceID: "run-payroll-test"}

	start := common.NewDate(2025, 1, 1)
	end := common.NewDate(2025, 1, 31)

	// Setup seed data
	tx, err := timeclockStorage.BeginTxWriter(ctx)
	assert.Nil(t, err)

	// Attendance untuk user 1 & 2
	for i := 0; i < 10; i++ {
		date := start.AddDate(0, 0, i)
		_, err := timeclockStorage.InsertAttendanceCheckin(ctx, 1, date, date.Add(9*time.Hour), userName)
		assert.Nil(t, err)
		_, err = timeclockStorage.InsertAttendanceCheckin(ctx, 2, date, date.Add(9*time.Hour), userName)
		assert.Nil(t, err)
	}

	// Overtime
	_, err = timeclockStorage.InsertOvertime(ctx, 1, start, 3, "test overtime", userName)
	assert.Nil(t, err)
	_, err = timeclockStorage.InsertOvertime(ctx, 2, start, 2, "test overtime", userName)
	assert.Nil(t, err)

	// Reimbursement
	_, err = timeclockStorage.InsertReimbursement(ctx, 1, start, 100000, "test", userName)
	assert.Nil(t, err)
	_, err = timeclockStorage.InsertReimbursement(ctx, 2, start, 50000, "test", userName)
	assert.Nil(t, err)

	err = tx.Commit(ctx)
	assert.Nil(t, err)

	type expected struct {
		success bool
		message string
	}

	type testCase struct {
		name     string
		ctx      context.Context
		in       *lib.RunPayrollIn
		mock     func()
		expected expected
	}

	scenarios := []testCase{
		{
			name: "success run payroll",
			ctx:  ctx,
			in: &lib.RunPayrollIn{
				Trace:       trace,
				PeriodStart: start,
				PeriodEnd:   end,
			},
			mock: func() {
				fmt.Println(">>> MOCK SET")
				userServiceMock.EXPECT().
					UserSalary(gomock.Any(), gomock.AssignableToTypeOf(&userLib.UserSalaryIn{})).
					Return(&userLib.UserSalaryOut{
						Success: true,
						Result: map[int]int{
							1: 3000000,
							2: 2000000,
						},
					}).Times(1)
			},
			expected: expected{
				success: true,
				message: "",
			},
		},
		{
			name: "fail unauthorized",
			ctx:  context.Background(),
			in: &lib.RunPayrollIn{
				Trace:       trace,
				PeriodStart: start,
				PeriodEnd:   end,
			},
			mock: func() {},
			expected: expected{
				success: false,
				message: "unauthorized",
			},
		},
		{
			name: "fail invalid period",
			ctx:  ctx,
			in: &lib.RunPayrollIn{
				Trace:       trace,
				PeriodStart: end,
				PeriodEnd:   start,
			},
			mock: func() {},
			expected: expected{
				success: false,
				message: "Periode tidak valid",
			},
		},
		{
			name: "fail user salary not found",
			ctx:  ctx,
			in: &lib.RunPayrollIn{
				Trace:       trace,
				PeriodStart: start,
				PeriodEnd:   end,
			},
			mock: func() {
				userServiceMock.EXPECT().
					UserSalary(gomock.Any(), gomock.AssignableToTypeOf(&userLib.UserSalaryIn{})).
					Return(&userLib.UserSalaryOut{
						Success: false,
						Result:  nil,
					}).Times(1)
			},
			expected: expected{
				success: false,
				message: "tidak ditemukan employee",
			},
		},
	}

	for _, sc := range scenarios {
		sc := sc
		t.Run(sc.name, func(t *testing.T) {
			sc.mock()
			out := service.RunPayroll(sc.ctx, sc.in)
			assert.Equal(t, sc.expected.success, out.Success)
			assert.Equal(t, sc.expected.message, out.Message)
		})
	}
}
