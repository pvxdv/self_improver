package response

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// Response
// @Description Response is the standard API response structure.
// @Description Used for success and error responses.
type Response struct {
	Status string `json:"status" example:"Error"`
	Error  string `json:"error,omitempty" example:"failed to..."`
}

const (
	StatusOK    = "OK"
	StatusError = "Error"
)

func OK() Response {
	return Response{
		Status: StatusOK,
	}
}

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}

func RespondWithJSON(w http.ResponseWriter, statusCode int, data interface{}, logger *zap.SugaredLogger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Errorf("failed to encode response: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func RespondWithError(w http.ResponseWriter, statusCode int, message string, logger *zap.SugaredLogger) {
	RespondWithJSON(w, statusCode, Error(message), logger)
}
