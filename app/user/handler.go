package user

import (
	"encoding/json"
	"net/http"

	"github.com/ariesmaulana/payroll/app/user/lib"
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

type loginRequest struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	trace, ok := contextutil.GetTrace(r.Context())
	if !ok {
		log.Warn(trace).Msg("Trace not found in context")
		http.Error(w, "Trace not found", http.StatusInternalServerError)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	out := h.service.Login(r.Context(), &lib.LoginIn{
		Trace:    trace,
		UserName: req.UserName,
		Password: req.Password,
	})
	if !out.Success {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	response.WriteJSON(w, http.StatusOK, trace.TraceID, true, "", out.Token)

}
