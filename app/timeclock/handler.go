package timeclock

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/ariesmaulana/payroll/app/timeclock/lib"
	"github.com/ariesmaulana/payroll/common"
	"github.com/ariesmaulana/payroll/internal/response"
	"github.com/ariesmaulana/payroll/lib/contextutil"
	log "github.com/ariesmaulana/payroll/lib/logger"
)

type Handler struct {
	service lib.ServiceInterface
}

func NewHandler(service lib.ServiceInterface) *Handler {
	return &Handler{service: service}
}

type addAttendancePeriodRequest struct {
	UserId      int    `json:"user_id"`
	CheckInDate string `json:"checkin_date"`
}

func (h *Handler) AddAttendancePeriod(w http.ResponseWriter, r *http.Request) {
	trace, ok := contextutil.GetTrace(r.Context())
	if !ok {
		log.Warn(trace).Msg("Trace not found in context")
		http.Error(w, "Trace not found", http.StatusInternalServerError)
		return
	}

	var req addAttendancePeriodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	period, err := time.Parse("2006-01-02", req.CheckInDate)
	if err != nil {
		http.Error(w, "Invalid period format, must be YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	out := h.service.AddAttendancePeriod(r.Context(), &lib.AddAttendancePeriodIn{
		Trace:       trace,
		UserID:      req.UserId,
		CheckInDate: period,
	})

	if !out.Success {
		http.Error(w, out.Message, http.StatusBadRequest)
		return
	}

	response.WriteJSON(w, http.StatusOK, trace.TraceID, true, "", nil)
}

func (h *Handler) SubmitAttendance(w http.ResponseWriter, r *http.Request) {
	trace, ok := contextutil.GetTrace(r.Context())
	if !ok {
		log.Warn(trace).Msg("Trace not found in context")
		http.Error(w, "Trace not found", http.StatusInternalServerError)
		return
	}

	out := h.service.SubmitAttendance(r.Context(), &lib.SubmitAttendanceIn{
		Trace:  trace,
		Period: common.NewDateTimeNow(),
	})

	if !out.Success {
		http.Error(w, out.Message, http.StatusBadRequest)
		return
	}

	response.WriteJSON(w, http.StatusOK, trace.TraceID, true, "", nil)
}

type addOvertimeRequest struct {
	OTDate time.Time `json:"ot_date"`
	Hours  int       `json:"hours"`
	Reason string    `json:"reason"`
}

func (h *Handler) AddOvertime(w http.ResponseWriter, r *http.Request) {
	trace, ok := contextutil.GetTrace(r.Context())
	if !ok {
		log.Warn(trace).Msg("Trace not found in context")
		http.Error(w, "Trace not found", http.StatusInternalServerError)
		return
	}

	var req addOvertimeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	out := h.service.AddOvertime(r.Context(), &lib.AddOvertimeIn{
		Trace:  trace,
		Hours:  req.Hours,
		Reason: req.Reason,
		Period: common.NewDateTimeNow(),
	})

	if !out.Success {
		http.Error(w, out.Message, http.StatusBadRequest)
		return
	}

	response.WriteJSON(w, http.StatusOK, trace.TraceID, true, "", nil)
}

func (h *Handler) CheckoutAttendance(w http.ResponseWriter, r *http.Request) {
	trace, ok := contextutil.GetTrace(r.Context())
	if !ok {
		http.Error(w, "Trace not found", http.StatusInternalServerError)
		return
	}

	out := h.service.CheckoutAttendance(r.Context(), &lib.CheckoutAttendanceIn{
		Trace:  trace,
		Period: common.NewDateToday(),
	})

	if !out.Success {
		http.Error(w, out.Message, http.StatusBadRequest)
		return
	}

	response.WriteJSON(w, http.StatusOK, trace.TraceID, true, "", nil)
}

type submitReimbursementRequest struct {
	Amount      int    `json:"amount"`
	Description string `json:"description"`
	Period      string `json:"period"` // format: YYYY-MM-DD
}

func (h *Handler) SubmitReimbursement(w http.ResponseWriter, r *http.Request) {
	trace, ok := contextutil.GetTrace(r.Context())
	if !ok {
		log.Warn(trace).Msg("Trace not found in context")
		http.Error(w, "Trace not found", http.StatusInternalServerError)
		return
	}

	var req submitReimbursementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	period, err := time.Parse("2006-01-02", req.Period)
	if err != nil {
		http.Error(w, "Invalid period format, must be YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	out := h.service.SubmitReimbursement(r.Context(), &lib.SubmitReimbursementIn{
		Trace:       trace,
		Period:      period,
		Amount:      req.Amount,
		Description: req.Description,
	})

	if !out.Success {
		http.Error(w, out.Message, http.StatusBadRequest)
		return
	}

	response.WriteJSON(w, http.StatusOK, trace.TraceID, true, "", nil)
}

type runPayrollRequest struct {
	Start string `json:"start"` // format: YYYY-MM-DD
	End   string `json:"end"`   // format: YYYY-MM-DD
}

func (h *Handler) RunPayroll(w http.ResponseWriter, r *http.Request) {
	trace, ok := contextutil.GetTrace(r.Context())
	if !ok {
		http.Error(w, "Trace not found", http.StatusInternalServerError)
		return
	}

	var req runPayrollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	start, err := time.Parse("2006-01-02", req.Start)
	if err != nil {
		http.Error(w, "Invalid start date format", http.StatusBadRequest)
		return
	}
	end, err := time.Parse("2006-01-02", req.End)
	if err != nil {
		http.Error(w, "Invalid end date format", http.StatusBadRequest)
		return
	}

	out := h.service.RunPayroll(r.Context(), &lib.RunPayrollIn{
		Trace:       trace,
		PeriodStart: start,
		PeriodEnd:   end,
	})

	if !out.Success {
		http.Error(w, out.Message, http.StatusBadRequest)
		return
	}

	response.WriteJSON(w, http.StatusOK, trace.TraceID, true, "", out)
}

func (h *Handler) GenerateSelfPaySlip(w http.ResponseWriter, r *http.Request) {
	trace, ok := contextutil.GetTrace(r.Context())
	if !ok {
		http.Error(w, "Trace not found", http.StatusInternalServerError)
		return
	}

	monthStr := r.URL.Query().Get("month")
	yearStr := r.URL.Query().Get("year")
	if monthStr == "" || yearStr == "" {
		http.Error(w, "Query param 'month' dan 'year' wajib diisi", http.StatusBadRequest)
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		http.Error(w, "Param 'month' harus angka", http.StatusBadRequest)
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		http.Error(w, "Param 'year' harus angka", http.StatusBadRequest)
		return
	}

	out := h.service.GenerateSelfPaySlip(r.Context(), &lib.GenerateSelfPaySlipIn{
		Trace: trace,
		Month: month,
		Year:  year,
	})

	if !out.Success {
		http.Error(w, out.Message, http.StatusBadRequest)
		return
	}

	response.WriteJSON(w, http.StatusOK, trace.TraceID, true, "", out)
}

func (h *Handler) GenerateAllPaySlips(w http.ResponseWriter, r *http.Request) {
	trace, ok := contextutil.GetTrace(r.Context())
	if !ok {
		http.Error(w, "Trace not found", http.StatusInternalServerError)
		return
	}

	monthStr := r.URL.Query().Get("month")
	yearStr := r.URL.Query().Get("year")
	if monthStr == "" || yearStr == "" {
		http.Error(w, "Query param 'month' dan 'year' wajib diisi", http.StatusBadRequest)
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		http.Error(w, "Param 'month' harus angka", http.StatusBadRequest)
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		http.Error(w, "Param 'year' harus angka", http.StatusBadRequest)
		return
	}

	out := h.service.GenerateAllPaySlips(r.Context(), &lib.GenerateAllPaySlipsIn{
		Trace: trace,
		Month: month,
		Year:  year,
	})

	if !out.Success {
		http.Error(w, out.Message, http.StatusBadRequest)
		return
	}

	response.WriteJSON(w, http.StatusOK, trace.TraceID, true, "", out)
}
