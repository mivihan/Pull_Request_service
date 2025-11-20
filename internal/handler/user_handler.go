package handler

import (
	"log/slog"
	"net/http"

	"github.com/mivihan/Pull_Request_service/internal/service"
)

type UserHandler struct {
	userService service.UserService
	logger      *slog.Logger
}

func NewUserHandler(userService service.UserService, logger *slog.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

func (h *UserHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var req SetIsActiveRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}

	if req.UserID == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "user_id is required",
			},
		})
		return
	}

	user, err := h.userService.SetIsActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		respondError(w, err, h.logger)
		return
	}

	respondJSON(w, http.StatusOK, UserResponse{
		User: mapUserToDTO(user),
	})
}

func (h *UserHandler) GetReviews(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "user_id query parameter is required",
			},
		})
		return
	}

	prs, err := h.userService.GetReviews(r.Context(), userID)
	if err != nil {
		respondError(w, err, h.logger)
		return
	}

	prDTOs := make([]PullRequestShortDTO, len(prs))
	for i, pr := range prs {
		prDTOs[i] = mapPRToShortDTO(pr)
	}

	respondJSON(w, http.StatusOK, UserReviewsResponse{
		UserID:       userID,
		PullRequests: prDTOs,
	})
}
