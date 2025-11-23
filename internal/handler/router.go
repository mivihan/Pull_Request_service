package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/mivihan/Pull_Request_service/internal/middleware"
	"github.com/mivihan/Pull_Request_service/internal/service"
)

func NewRouter(
	teamService service.TeamService,
	userService service.UserService,
	prService service.PRService,
	logger *slog.Logger,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logging(logger))
	r.Use(middleware.Recovery(logger))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	teamHandler := NewTeamHandler(teamService, logger)
	userHandler := NewUserHandler(userService, logger)
	prHandler := NewPRHandler(prService, logger)
	statsHandler := NewStatsHandler(prService, logger)

	r.Post("/team/add", teamHandler.CreateTeam)
	r.Get("/team/get", teamHandler.GetTeam)
	r.Post("/team/deactivateUsers", teamHandler.DeactivateUsers)

	r.Post("/users/setIsActive", userHandler.SetIsActive)
	r.Get("/users/getReview", userHandler.GetReviews)

	r.Post("/pullRequest/create", prHandler.CreatePR)
	r.Post("/pullRequest/merge", prHandler.MergePR)
	r.Post("/pullRequest/reassign", prHandler.ReassignReviewer)

	r.Get("/stats/reviewers", statsHandler.GetReviewerStats)
	r.Get("/stats/pullRequests", statsHandler.GetPRStats)

	return r
}
