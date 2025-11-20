package handler

import (
	"log/slog"
	"net/http"

	"github.com/mivihan/Pull_Request_service/internal/service"
)

type PRHandler struct {
	prService service.PRService
	logger    *slog.Logger
}

func NewPRHandler(prService service.PRService, logger *slog.Logger) *PRHandler {
	return &PRHandler{
		prService: prService,
		logger:    logger,
	}
}

func (h *PRHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req CreatePRRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}

	if req.PullRequestID == "" || req.PullRequestName == "" || req.AuthorID == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "pull_request_id, pull_request_name, and author_id are required",
			},
		})
		return
	}

	pr, err := h.prService.CreatePR(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		respondError(w, err, h.logger)
		return
	}

	respondJSON(w, http.StatusCreated, PRResponse{
		PR: mapPRToDTO(pr),
	})
}

func (h *PRHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	var req MergePRRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}

	if req.PullRequestID == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "pull_request_id is required",
			},
		})
		return
	}

	pr, err := h.prService.MergePR(r.Context(), req.PullRequestID)
	if err != nil {
		respondError(w, err, h.logger)
		return
	}

	respondJSON(w, http.StatusOK, PRResponse{
		PR: mapPRToDTO(pr),
	})
}

func (h *PRHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	var req ReassignReviewerRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}

	if req.PullRequestID == "" || req.OldUserID == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "pull_request_id and old_user_id are required",
			},
		})
		return
	}

	pr, replacedBy, err := h.prService.ReassignReviewer(r.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		respondError(w, err, h.logger)
		return
	}

	respondJSON(w, http.StatusOK, ReassignResponse{
		PR:         mapPRToDTO(pr),
		ReplacedBy: replacedBy,
	})
}
