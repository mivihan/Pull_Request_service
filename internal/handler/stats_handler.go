package handler

import (
	"log/slog"
	"net/http"
	"sort"

	"github.com/mivihan/Pull_Request_service/internal/service"
)

type StatsHandler struct {
	prService service.PRService
	logger    *slog.Logger
}

func NewStatsHandler(prService service.PRService, logger *slog.Logger) *StatsHandler {
	return &StatsHandler{
		prService: prService,
		logger:    logger,
	}
}

func (h *StatsHandler) GetReviewerStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.prService.GetReviewerStats(r.Context())
	if err != nil {
		respondError(w, err, h.logger)
		return
	}

	reviewers := make([]ReviewerStatDTO, 0, len(stats))
	for userID, count := range stats {
		reviewers = append(reviewers, ReviewerStatDTO{
			UserID:           userID,
			AssignmentsCount: count,
		})
	}

	sort.Slice(reviewers, func(i, j int) bool {
		return reviewers[i].AssignmentsCount > reviewers[j].AssignmentsCount
	})

	respondJSON(w, http.StatusOK, ReviewerStatsResponse{
		Reviewers: reviewers,
	})
}

func (h *StatsHandler) GetPRStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.prService.GetPRStats(r.Context())
	if err != nil {
		respondError(w, err, h.logger)
		return
	}

	response := PRStatsResponse{
		Open:   stats["OPEN"],
		Merged: stats["MERGED"],
	}

	respondJSON(w, http.StatusOK, response)
}