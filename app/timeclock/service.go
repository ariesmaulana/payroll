package timeclock

import (
	"context"
	"time"

	"github.com/ariesmaulana/payroll/app/timeclock/lib"
	userLib "github.com/ariesmaulana/payroll/app/user/lib"
	"github.com/ariesmaulana/payroll/common"
	"github.com/ariesmaulana/payroll/data"
	"github.com/ariesmaulana/payroll/lib/contextutil"
	log "github.com/ariesmaulana/payroll/lib/logger"
)

var _ lib.ServiceInterface = (*Service)(nil)

type Service struct {
	storage     lib.StorageInterface
	userService userLib.ServiceInterface
}

func NewService(storage lib.StorageInterface, userService userLib.ServiceInterface) *Service {
	return &Service{
		storage:     storage,
		userService: userService,
	}
}

func (s *Service) AddAttendancePeriod(ctx context.Context, in *lib.AddAttendancePeriodIn) *lib.AddAttendancePeriodOut {
	resp := lib.AddAttendancePeriodOut{}

	user, ok := contextutil.GetUser(ctx)
	if !ok {
		log.Warn(in.Trace).Msg("AddAttendancePeriod/ unauthorized access - user missing in context")
		resp.Message = "unauthorized"
		return &resp
	}

	if user.Role != data.RAdmin {
		log.Warn(in.Trace).Msg("AddAttendancePeriod/ user not admin")
		resp.Message = "forbidden: Hanya admin yang bisa akses"
		return &resp
	}

	// Validasi range tanggal
	if in.CheckInDate.IsZero() {
		log.Warn(in.Trace).
			Time("checkinDate", in.CheckInDate).
			Msg("AddAttendancePeriod/ invalid period range")
		resp.Message = "Checkin date wajib diisi"
		return &resp
	}

	if !s.isWeekDays(in.CheckInDate) {
		log.Warn(in.Trace).Msg("AddAttendancePeriod/ failed weekends")
		resp.Message = "Tidak bisa mengisi kehadiran saat Sabtu dan Minggu."
		return &resp
	}

	tx, err := s.storage.BeginTxWriter(ctx)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("AddAttendancePeriod/ Failed begin tx")
		return &resp
	}
	defer tx.Rollback(ctx)

	period := in.CheckInDate.Truncate(24 * time.Hour)
	checkin := in.CheckInDate

	payrollExists, err := s.storage.IsPayrollAlreadyRun(ctx, period)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("AddAttendancePeriod/ failed to check payroll existence")
		resp.Message = "internal error"
		return &resp
	}
	if payrollExists {
		log.Warn(in.Trace).Msg("AddAttendancePeriod/ cannot update data after payroll is processed")
		resp.Message = "Data tidak bisa diubah karena payroll sudah dijalankan"
		return &resp
	}

	// Insert ke storage
	_, err = s.storage.InsertAttendanceCheckin(ctx, in.UserID, period, checkin, user.Username)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("AddAttendancePeriod/ failed to insert attendance period")
		resp.Message = "internal error"
		return &resp
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("AddAttendancePeriod/ failed to commit")
		return &resp
	}

	resp.Success = true
	return &resp
}

// SubmitAttendance used for checkin by user it self
// it will assume:
// checkinDate = today
// createdBy = userlogin
func (s *Service) SubmitAttendance(ctx context.Context, in *lib.SubmitAttendanceIn) *lib.SubmitAttendanceOut {
	resp := lib.SubmitAttendanceOut{}

	user, ok := contextutil.GetUser(ctx)
	if !ok {
		log.Warn(in.Trace).Msg("SubmitAttendance/ unauthorized access - user missing in context")
		resp.Message = "unauthorized"
		return &resp
	}

	if in.Period.IsZero() {
		log.Warn(in.Trace).Msg("SubmitAttendance/ period missing")
		resp.Message = "Wajib pilih periode waktu checkin"
		return &resp
	}

	today := in.Period

	if !s.isWeekDays(today) {
		log.Warn(in.Trace).Msg("SubmitAttendance/ failed weekends")
		resp.Message = "Tidak bisa mengisi kehadiran saat Sabtu dan Minggu."
		return &resp
	}

	tx, err := s.storage.BeginTxWriter(ctx)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("SubmitAttendance/ Failed begin tx")
		return &resp
	}
	defer tx.Rollback(ctx)

	period := today.Truncate(24 * time.Hour)
	checkin := today

	payrollExists, err := s.storage.IsPayrollAlreadyRun(ctx, period)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("SubmitAttendance/ failed to check payroll existence")
		resp.Message = "internal error"
		return &resp
	}
	if payrollExists {
		log.Warn(in.Trace).Msg("SubmitAttendance/ cannot update data after payroll is processed")
		resp.Message = "Data tidak bisa diubah karena payroll sudah dijalankan"
		return &resp
	}

	_, err = s.storage.InsertAttendanceCheckin(ctx, user.Id, period, checkin, user.Username)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("SubmitAttendance/ failed to insert attendance period")
		resp.Message = "Terjadi kesalahan, kemungkinan anda telah tercatat di hari ini"
		return &resp
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("SubmitAttendance/ failed to commit")
		return &resp
	}

	resp.Success = true
	return &resp
}

func (s *Service) isWeekDays(date time.Time) bool {
	date = date.In(common.JakartaTZ)
	weekday := date.Weekday()
	return weekday >= time.Monday && weekday <= time.Friday
}

func (s *Service) AddOvertime(ctx context.Context, in *lib.AddOvertimeIn) *lib.AddOvertimeOut {
	resp := lib.AddOvertimeOut{}

	user, ok := contextutil.GetUser(ctx)
	if !ok {
		log.Warn(in.Trace).Msg("AddOvertime/ unauthorized")
		resp.Message = "unauthorized"
		return &resp
	}

	if in.Hours <= 0 || in.Hours > 3 {
		log.Warn(in.Trace).Msg("AddOvertime/ invalid overtimes")
		resp.Message = "Jumlah jam lembur tidak boleh lebih dari 3"
		return &resp
	}

	if in.Reason == "" {
		log.Warn(in.Trace).Msg("AddOvertime/ invalid reason")
		resp.Message = "Alasan harus diisi"
		return &resp
	}

	if in.Period.IsZero() {
		log.Warn(in.Trace).Msg("AddOvertime/ period missing")
		resp.Message = "Wajib pilih periode waktu overtime"
		return &resp
	}

	now := in.Period
	if now.Hour() < 17 && in.Period.Format("2006-01-02") == now.Format("2006-01-02") {
		log.Warn(in.Trace).Msg("AddOvertime/ invalid times")
		resp.Message = "Lembur hanya bisa diajukan setelah jam kerja selesai"
		return &resp
	}

	tx, err := s.storage.BeginTxWriter(ctx)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("AddOvertime/ Failed begin tx")
		return &resp
	}
	defer tx.Rollback(ctx)

	period := now.Truncate(24 * time.Hour)

	payrollExists, err := s.storage.IsPayrollAlreadyRun(ctx, period)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("AddOvertime/ failed to check payroll existence")
		resp.Message = "internal error"
		return &resp
	}
	if payrollExists {
		log.Warn(in.Trace).Msg("AddOvertime/ cannot update data after payroll is processed")
		resp.Message = "Data tidak bisa diubah karena payroll sudah dijalankan"
		return &resp
	}

	attn, err := s.storage.GetDetailAttendanceByUserAndPeriod(ctx, user.Id, period)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("AddOvertime/ error get attendance")
		return &resp
	}
	if attn == nil {
		log.Warn(in.Trace).Msg("AddOvertime/ attendance not found")
		resp.Message = "Anda belum absen di hari tersebut"
		return &resp
	}

	id, err := s.storage.InsertOvertime(ctx, user.Id, period, in.Hours, in.Reason, user.Username)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("AddOvertime/ Failed InsertOvertime")
		resp.Message = "internal error"
		return &resp
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("AddOvertime/ failed to commit")
		return &resp
	}

	resp.Success = true
	resp.Id = id
	return &resp
}

func (s *Service) CheckoutAttendance(ctx context.Context, in *lib.CheckoutAttendanceIn) *lib.CheckoutAttendanceOut {
	resp := lib.CheckoutAttendanceOut{}

	user, ok := contextutil.GetUser(ctx)
	if !ok {
		log.Warn(in.Trace).Msg("CheckoutAttendance/ unauthorized access - user missing in context")
		resp.Message = "unauthorized"
		return &resp
	}

	if in.Period.IsZero() {
		log.Warn(in.Trace).Msg("CheckoutAttendance/ period missing")
		resp.Message = "Wajib pilih periode waktu overtime"
		return &resp
	}

	today := in.Period
	time := common.NewDateTimeNow()

	tx, err := s.storage.BeginTxWriter(ctx)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("CheckoutAttendance/ Failed begin tx")
		return &resp
	}
	defer tx.Rollback(ctx)

	err = s.storage.UpdateAttendanceCheckout(ctx, user.Id, today, time, user.Username)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("CheckoutAttendance/ failed to update checkout")
		resp.Message = "Anda belum check-in atau sudah checkout"
		return &resp
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("CheckoutAttendance/ failed to commit")
		resp.Message = "commit error"
		return &resp
	}

	resp.Success = true
	return &resp
}

func (s *Service) SubmitReimbursement(ctx context.Context, in *lib.SubmitReimbursementIn) *lib.SubmitReimbursementOut {
	resp := lib.SubmitReimbursementOut{}

	user, ok := contextutil.GetUser(ctx)
	if !ok {
		log.Warn(in.Trace).Msg("SubmitReimbursement/ unauthorized access - user missing in context")
		resp.Message = "unauthorized"
		return &resp
	}

	if in.Period.IsZero() {
		log.Warn(in.Trace).Msg("SubmitReimbursement/ period is missing")
		resp.Message = "Periode wajib diisi"
		return &resp
	}

	if in.Amount <= 0 {
		log.Warn(in.Trace).Msg("SubmitReimbursement/ invalid amount")
		resp.Message = "Jumlah reimbursement harus lebih dari 0"
		return &resp
	}

	if in.Description == "" {
		log.Warn(in.Trace).Msg("SubmitReimbursement/ description is empty")
		resp.Message = "Deskripsi reimbursement wajib diisi"
		return &resp
	}

	tx, err := s.storage.BeginTxWriter(ctx)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("SubmitReimbursement/ failed to begin tx")
		return &resp
	}
	defer tx.Rollback(ctx)

	payrollExists, err := s.storage.IsPayrollAlreadyRun(ctx, in.Period)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("SubmitReimbursement/ failed to check payroll existence")
		resp.Message = "internal error"
		return &resp
	}
	if payrollExists {
		log.Warn(in.Trace).Msg("SubmitReimbursement/ cannot update data after payroll is processed")
		resp.Message = "Data tidak bisa diubah karena payroll sudah dijalankan"
		return &resp
	}

	id, err := s.storage.InsertReimbursement(ctx, user.Id, in.Period, in.Amount, in.Description, user.Username)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("SubmitReimbursement/ insert error")
		return &resp
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("SubmitReimbursement/ commit error")
		return &resp
	}

	resp.Success = true
	resp.Id = id
	return &resp
}

func (s *Service) RunPayroll(ctx context.Context, in *lib.RunPayrollIn) *lib.RunPayrollOut {
	resp := lib.RunPayrollOut{}

	user, ok := contextutil.GetUser(ctx)
	if !ok {
		log.Warn(in.Trace).Msg("RunPayroll/ unauthorized")
		resp.Message = "unauthorized"
		return &resp
	}

	// Cek periode valid
	if in.PeriodStart.IsZero() || in.PeriodEnd.IsZero() || in.PeriodEnd.Before(in.PeriodStart) {
		log.Warn(in.Trace).Msg("RunPayroll/ invalid period")
		resp.Message = "Periode tidak valid"
		return &resp
	}

	userSalaries := s.userService.UserSalary(ctx, &userLib.UserSalaryIn{
		Trace: in.Trace,
	})

	if !userSalaries.Success {
		log.Warn(in.Trace).Msg("RunPayroll/ invalid user salaries")
		resp.Message = "tidak ditemukan employee"
		return &resp
	}

	tx, err := s.storage.BeginTxWriter(ctx)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("RunPayroll/ begin tx failed")
		resp.Message = "internal error"
		return &resp
	}
	defer tx.Rollback(ctx)

	salaries := userSalaries.Result
	// ambil semua user yang punya attendance
	attendances, err := s.storage.GetAllAttendanceByPeriod(ctx, in.PeriodStart, in.PeriodEnd)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("RunPayroll/ fetch attendance summary failed")
		resp.Message = "internal error"
		return &resp
	}
	totalAttendance := len(attendances)

	// key is userId and value total attendance in this period
	userAttendanceMap := make(map[int]int)
	for _, att := range attendances {
		userAttendanceMap[att.UserId]++
	}

	baseSalariesPerUser := calculateProratedSalary(salaries, userAttendanceMap, in.PeriodStart, in.PeriodEnd)
	totalBaseSalariesThisMonth := sumValuesMap(baseSalariesPerUser)

	// total overtime (jam)
	totalOvertime, err := s.storage.GetTotalOvertimeByPeriod(ctx, in.PeriodStart, in.PeriodEnd)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("RunPayroll/ error total overtime")
		resp.Message = "internal error"
		return &resp
	}

	usersOvertime, err := s.storage.GetOvertimeHoursByPeriod(ctx, in.PeriodStart, in.PeriodEnd)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("RunPayroll/ error total overtime hours")
		resp.Message = "internal error"
		return &resp
	}
	baseSalaryOverTimes := calculateOvertimeSalary(salaries, userAttendanceMap, usersOvertime, in.PeriodStart, in.PeriodEnd)
	totalOverTimeSalary := sumValuesMap(baseSalaryOverTimes)

	// total reimbursement (rupiah)
	totalReimbursement, err := s.storage.GetTotalReimbursementByPeriod(ctx, in.PeriodStart, in.PeriodEnd)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("RunPayroll/ error total reimbursement")
		resp.Message = "internal error"
		return &resp
	}
	totalReimbursementPerUser, err := s.storage.GetReimbursementTotalsByPeriod(ctx, in.PeriodStart, in.PeriodEnd)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("RunPayroll/ error total reimbursement per user")
		resp.Message = "internal error"
		return &resp
	}
	totalSalaryThisPeriod := totalBaseSalariesThisMonth + totalOverTimeSalary + totalReimbursement

	payrollId, err := s.storage.InsertPayroll(ctx, in.PeriodStart, in.PeriodEnd, totalAttendance, totalOvertime, totalReimbursement, totalSalaryThisPeriod, user.Username)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("RunPayroll/ insert payroll failed")
		resp.Message = "internal error"
		return &resp
	}

	for userId := range userAttendanceMap {
		attendanceCount := userAttendanceMap[userId]
		overtimeHours := usersOvertime[userId]

		userSalary := baseSalariesPerUser[userId]
		userOvertimeSalary := baseSalaryOverTimes[userId]
		userReimbursement := totalReimbursementPerUser[userId]
		userTotalSalary := userSalary + userOvertimeSalary + userReimbursement

		_, err = s.storage.InsertPayrollItem(ctx, payrollId, userId, attendanceCount, overtimeHours, userReimbursement, userTotalSalary, user.Username)
		if err != nil {
			log.Error(in.Trace).Err(err).Msgf("RunPayroll/ insert payroll item user_id=%d failed", userId)
			resp.Message = "internal error"
			return &resp
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("RunPayroll/ commit failed")
		resp.Message = "internal error"
		return &resp
	}

	resp.Success = true
	resp.PayrollId = payrollId
	return &resp
}

func (s *Service) GenerateSelfPaySlip(ctx context.Context, in *lib.GenerateSelfPaySlipIn) *lib.GenerateSelfPaySlipOut {
	resp := &lib.GenerateSelfPaySlipOut{}

	user, ok := contextutil.GetUser(ctx)
	if !ok {
		log.Warn(in.Trace).Msg("GenerateSelfPaySlip/ unauthorized")
		resp.Message = "unauthorized"
		return resp
	}

	if in.Month <= 0 || in.Month > 12 || in.Year <= 0 {
		log.Warn(in.Trace).Msg("GenerateSelfPaySlip/ invalid input")
		resp.Message = "Bulan atau tahun tidak valid"
		return resp
	}

	location := common.JakartaTZ
	periodStart := time.Date(in.Year, time.Month(in.Month), 1, 0, 0, 0, 0, location)
	periodEnd := periodStart.AddDate(0, 1, -1)

	payroll, err := s.storage.GetPayrollByPeriod(ctx, periodStart, periodEnd)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("GenerateSelfPaySlip/ get payroll failed")
		resp.Message = "internal error"
		return resp
	}
	if payroll == nil {
		log.Warn(in.Trace).Msg("GenerateSelfPaySlip/ payroll not found")
		resp.Message = "Payroll belum tersedia untuk periode ini"
		return resp
	}

	item, err := s.storage.GetPayrollItemByPayrollIDAndUserID(ctx, payroll.Id, user.Id)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("GenerateSelfPaySlip/ get payroll item failed")
		resp.Message = "internal error"
		return resp
	}
	if item == nil {
		log.Warn(in.Trace).Msg("GenerateSelfPaySlip/ payroll item not found")
		resp.Message = "Data payslip tidak tersedia"
		return resp
	}

	reimbursements, err := s.storage.GetReimbursementsByUserAndPeriod(ctx, user.Id, periodStart, periodEnd)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("GenerateSelfPaySlip/ get reimbursements failed")
		resp.Message = "internal error"
		return resp
	}

	overtimes, err := s.storage.GetOvertimesByUserAndPeriod(ctx, user.Id, periodStart, periodEnd)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("GenerateSelfPaySlip/ get overtimes failed")
		resp.Message = "internal error"
		return resp
	}

	attendances, err := s.storage.GetAttendancesByUserAndPeriods(ctx, user.Id, periodStart, periodEnd)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("GenerateSelfPaySlip/ get attendances failed")
		resp.Message = "internal error"
		return resp
	}

	resp.Success = true
	resp.TotalSalary = item.TotalSalary
	resp.ListReimbursement = reimbursements
	resp.ListOvertimes = overtimes
	resp.ListAttendAnce = attendances
	return resp
}

func (s *Service) GenerateAllPaySlips(ctx context.Context, in *lib.GenerateAllPaySlipsIn) *lib.GenerateAllPaySlipsOut {
	resp := &lib.GenerateAllPaySlipsOut{}

	user, ok := contextutil.GetUser(ctx)
	if !ok || user.Role != data.RAdmin {
		log.Warn(in.Trace).Msg("GenerateAllPaySlips/ unauthorized")
		resp.Message = "unauthorized"
		return resp
	}

	if in.Month <= 0 || in.Month > 12 || in.Year <= 0 {
		log.Warn(in.Trace).Msg("GenerateAllPaySlips/ invalid input")
		resp.Message = "Bulan atau tahun tidak valid"
		return resp
	}

	location := common.JakartaTZ
	periodStart := time.Date(in.Year, time.Month(in.Month), 1, 0, 0, 0, 0, location)
	periodEnd := periodStart.AddDate(0, 1, -1)

	payroll, err := s.storage.GetPayrollByPeriod(ctx, periodStart, periodEnd)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("GenerateAllPaySlips/ get payroll failed")
		resp.Message = "internal error"
		return resp
	}
	if payroll == nil {
		log.Warn(in.Trace).Msg("GenerateAllPaySlips/ payroll not found")
		resp.Message = "Payroll belum tersedia untuk periode ini"
		return resp
	}

	items, err := s.storage.GetPayrollItemsByPayrollID(ctx, payroll.Id)
	if err != nil {
		log.Error(in.Trace).Err(err).Msg("GenerateAllPaySlips/ get payroll items failed")
		resp.Message = "internal error"
		return resp
	}

	var totalSalaryAll int
	var payslips []*data.UserPayslip
	for _, item := range items {
		payslip := &data.UserPayslip{
			UserID:           item.UserId,
			TotalSalary:      item.TotalSalary,
			AttendanceCount:  item.AttendanceCount,
			OvertimeHours:    item.OvertimeHours,
			ReimbursementSum: item.ReimbursementTotal,
		}
		payslips = append(payslips, payslip)
		totalSalaryAll += item.TotalSalary
	}

	resp.Success = true
	resp.TotalSalaryAll = totalSalaryAll
	resp.ListUserPayslips = payslips
	return resp
}
