package handler

import (
	"log/slog"
	"net/http"

	"github.com/mivihan/Pull_Request_service/internal/service"
)

type TeamHandler struct {
	teamService service.TeamService
	logger      *slog.Logger
}

func NewTeamHandler(teamService service.TeamService, logger *slog.Logger) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
		logger:      logger,
	}
}

func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var req CreateTeamRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}

	if req.TeamName == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "team_name is required",
			},
		})
		return
	}

	members := make([]service.TeamMemberInput, len(req.Members))
	for i, m := range req.Members {
		members[i] = service.TeamMemberInput{
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		}
	}

	team, err := h.teamService.CreateTeam(r.Context(), req.TeamName, members)
	if err != nil {
		respondError(w, err, h.logger)
		return
	}

	respondJSON(w, http.StatusCreated, TeamResponse{
		Team: mapTeamWithMembersToDTO(team),
	})
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "team_name query parameter is required",
			},
		})
		return
	}

	team, err := h.teamService.GetTeam(r.Context(), teamName)
	if err != nil {
		respondError(w, err, h.logger)
		return
	}

	respondJSON(w, http.StatusOK, mapTeamWithMembersToDTO(team))
}

func (h *TeamHandler) DeactivateUsers(w http.ResponseWriter, r *http.Request) {
	var req DeactivateUsersRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}

	if req.TeamName == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "team_name is required",
			},
		})
		return
	}

	if req.UserIDs == nil {
		req.UserIDs = []string{}
	}

	result, err := h.teamService.DeactivateTeamUsers(r.Context(), req.TeamName, req.UserIDs)
	if err != nil {
		respondError(w, err, h.logger)
		return
	}

	respondJSON(w, http.StatusOK, DeactivateUsersResponse{
		TeamName:        result.TeamName,
		DeactivatedCount: result.DeactivatedCount,
		AffectedPRCount:  result.AffectedPRCount,
	})
}