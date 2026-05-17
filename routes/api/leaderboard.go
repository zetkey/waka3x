package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/go-chi/chi/v5"
	conf "github.com/zetkey/waka3x/config"
	"github.com/zetkey/waka3x/middlewares"
	"github.com/zetkey/waka3x/models"
	routeutils "github.com/zetkey/waka3x/routes/utils"
	"github.com/zetkey/waka3x/services"
	"github.com/zetkey/waka3x/utils"
)

type LeaderboardApiHandler struct {
	config             *conf.Config
	userSrvc           services.IUserService
	leaderboardService services.ILeaderboardService
}

func NewLeaderboardApiHandler(userService services.IUserService, leaderboardService services.ILeaderboardService) *LeaderboardApiHandler {
	return &LeaderboardApiHandler{
		config:             conf.Get(),
		userSrvc:           userService,
		leaderboardService: leaderboardService,
	}
}

var allowedLeaderboardAggregations = map[string]uint8{
	"language": models.SummaryLanguage,
}

func (h *LeaderboardApiHandler) RegisterRoutes(router chi.Router) {
	r := chi.NewRouter()
	if h.config.App.LeaderboardRequireAuth {
		r.Use(middlewares.NewAuthenticateMiddleware(h.userSrvc).Handler)
	} else {
		r.Use(middlewares.NewAuthenticateMiddleware(h.userSrvc).WithOptionalFor("/").Handler)
	}
	r.Get("/", h.Get)

	router.Mount("/leaderboard", r)
}

func (h *LeaderboardApiHandler) Get(w http.ResponseWriter, r *http.Request) {
	if h.leaderboardService == nil {
		routeutils.RespondJSONError(w, http.StatusNotFound, "leaderboard is disabled")
		return
	}

	byParam := strings.ToLower(r.URL.Query().Get("by"))
	keyParam := strings.ToLower(r.URL.Query().Get("key"))
	pageParams := utils.ParsePageParamsWithDefault(r, 1, 100)
	user := middlewares.GetPrincipal(r)

	var leaderboard models.Leaderboard
	var err error
	var userLanguages map[string][]string
	var topKeys []string

	if byParam == "" {
		leaderboard, err = h.leaderboardService.GetByInterval(h.leaderboardService.GetDefaultScope(), pageParams, true)
		if err == nil && user != nil && !leaderboard.HasUser(user.ID) {
			if count, err := h.leaderboardService.CountUsers(true); err == nil && count > int64(pageParams.PageSize) {
				if ownLeaderboard, err := h.leaderboardService.GetByIntervalAndUser(h.leaderboardService.GetDefaultScope(), user.ID, true); err == nil && len(ownLeaderboard) > 0 {
					leaderboard = append(leaderboard, ownLeaderboard[0])
				}
			}
		}
	} else {
		by, ok := allowedLeaderboardAggregations[byParam]
		if !ok {
			routeutils.RespondJSONError(w, http.StatusBadRequest, fmt.Sprintf("unsupported aggregation '%s'", byParam))
			return
		}

		leaderboard, err = h.leaderboardService.GetAggregatedByInterval(h.leaderboardService.GetDefaultScope(), &by, pageParams, true)
		if err == nil && user != nil {
			if count, err := h.leaderboardService.CountUsers(true); err == nil && count > int64(pageParams.PageSize) {
				if ownLeaderboard, err := h.leaderboardService.GetAggregatedByIntervalAndUser(h.leaderboardService.GetDefaultScope(), user.ID, &by, true); err == nil {
					leaderboard.AddMany(ownLeaderboard)
				}
			}
		}

		userLeaderboards := slice.GroupWith[*models.LeaderboardItemRanked, string](leaderboard, func(item *models.LeaderboardItemRanked) string {
			return item.UserID
		})
		userLanguages = map[string][]string{}
		for userID, items := range userLeaderboards {
			userLanguages[userID] = models.Leaderboard(items).TopKeysByUser(models.SummaryLanguage, userID)
		}

		topKeys = leaderboard.TopKeys(by)
		if len(topKeys) > 0 {
			if keyParam == "" {
				keyParam = topKeys[0]
			}
			leaderboard = leaderboard.TopByKey(by, keyParam)
		}
	}

	if err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "failed to fetch leaderboard")
		return
	}

	leaderboard.FilterEmpty()

	response := make([]LeaderboardItemResponse, len(leaderboard))
	for i, item := range leaderboard {
		response[i] = newLeaderboardItemResponse(item)
	}

	routeutils.RespondJSON(w, http.StatusOK, LeaderboardResponse{
		Items:         response,
		By:            byParam,
		Key:           keyParam,
		TopKeys:       topKeys,
		UserLanguages: userLanguages,
		IntervalLabel: h.leaderboardService.GetDefaultScope().GetHumanReadable(),
		LastUpdated:   leaderboard.LastUpdate(),
	})
}
