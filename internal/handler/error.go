package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/mivihan/Pull_Request_service/internal/domain"
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func respondError(w http.ResponseWriter, err error, logger *slog.Logger) {
	var domainErr *domain.DomainError
	if errors.As(err, &domainErr) {
		statusCode := mapDomainErrorToHTTP(domainErr.Code)
		respondJSON(w, statusCode, ErrorResponse{
			Error: ErrorDetail{
				Code:    string(domainErr.Code),
				Message: domainErr.Message,
			},
		})
		return
	}

	logger.Error("internal server error", "error", err)
	respondJSON(w, http.StatusInternalServerError, ErrorResponse{
		Error: ErrorDetail{
			Code:    "INTERNAL_ERROR",
			Message: "internal server error",
		},
	})
}

func mapDomainErrorToHTTP(code domain.ErrorCode) int {
	switch code {
	case domain.ErrCodeTeamExists:
		return http.StatusBadRequest
	case domain.ErrCodePRExists,
		domain.ErrCodePRMerged,
		domain.ErrCodeNotAssigned,
		domain.ErrCodeNoCandidate:
		return http.StatusConflict
	case domain.ErrCodeNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "invalid JSON format",
			},
		})
		return err
	}
	return nil
}
