package response

import (
	"encoding/json"
	"net/http"
)

type ApiResponse struct {
	Success bool        `json:"success"`
	Msg     string      `json:"msg"`
	Trace   string      `json:"trace"`
	Data    interface{} `json:"data"`
}

func WriteJSON(w http.ResponseWriter, statusCode int, trace string, success bool, msg string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	resp := ApiResponse{
		Success: success,
		Msg:     msg,
		Trace:   trace,
		Data:    data,
	}
	json.NewEncoder(w).Encode(resp)
}
