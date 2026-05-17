package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	conf "github.com/zetkey/waka3x/config"
	"github.com/zetkey/waka3x/middlewares"
	routeutils "github.com/zetkey/waka3x/routes/utils"
	"github.com/zetkey/waka3x/services"
	"github.com/zetkey/waka3x/utils"
)

type ProjectsApiHandler struct {
	config           *conf.Config
	userSrvc         services.IUserService
	heartbeatService services.IHeartbeatService
}

func NewProjectsApiHandler(userService services.IUserService, heartbeatService services.IHeartbeatService) *ProjectsApiHandler {
	return &ProjectsApiHandler{
		config:           conf.Get(),
		userSrvc:         userService,
		heartbeatService: heartbeatService,
	}
}

func (h *ProjectsApiHandler) RegisterRoutes(router chi.Router) {
	r := chi.NewRouter()
	r.Use(middlewares.NewAuthenticateMiddleware(h.userSrvc).Handler)
	r.Get("/", h.Get)

	router.Mount("/projects", r)
}

func (h *ProjectsApiHandler) Get(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	pageParams := utils.ParsePageParamsWithDefault(r, 1, 24)
	query := strings.TrimSpace(r.URL.Query().Get("q"))

	projects, err := h.heartbeatService.GetUserProjectStats(user, time.Time{}, utils.BeginOfToday(user.TZ()), query, pageParams, false)
	if err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "failed to fetch project stats")
		return
	}

	response := make([]ProjectStatResponse, len(projects))
	for i, project := range projects {
		response[i] = newProjectStatResponse(project)
	}

	routeutils.RespondJSON(w, http.StatusOK, response)
}
