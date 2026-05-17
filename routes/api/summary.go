package api

import (
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/go-chi/chi/v5"
	"github.com/zetkey/waka3x/helpers"
	"github.com/zetkey/waka3x/models"
	"github.com/zetkey/waka3x/models/view"
	routeutils "github.com/zetkey/waka3x/routes/utils"
	"net/http"

	conf "github.com/zetkey/waka3x/config"
	"github.com/zetkey/waka3x/middlewares"
	"github.com/zetkey/waka3x/services"
	"github.com/zetkey/waka3x/utils"
)

const (
	dailyStatsMinRangeDays = 3
	dailyStatsMaxRangeDays = 31
)

type SummaryApiHandler struct {
	config         *conf.Config
	userSrvc       services.IUserService
	summarySrvc    services.ISummaryService
	heartbeatsSrvc services.IHeartbeatService
	durationSrvc   services.IDurationService
	aliasSrvc      services.IAliasService
}

func NewSummaryApiHandler(userService services.IUserService, summaryService services.ISummaryService, heartbeatsService services.IHeartbeatService, durationService services.IDurationService, aliasService services.IAliasService) *SummaryApiHandler {
	return &SummaryApiHandler{
		summarySrvc:    summaryService,
		userSrvc:       userService,
		heartbeatsSrvc: heartbeatsService,
		durationSrvc:   durationService,
		aliasSrvc:      aliasService,
		config:         conf.Get(),
	}
}

func (h *SummaryApiHandler) RegisterRoutes(router chi.Router) {
	r := chi.NewRouter()
	r.Use(middlewares.NewAuthenticateMiddleware(h.userSrvc).Handler)
	r.Get("/", h.Get)
	r.Get("/details", h.GetDetails)

	router.Mount("/summary", r)
}

// @Summary Retrieve a summary
// @ID get-summary
// @Tags summary
// @Produce json
// @Param interval query string false "Interval identifier" Enums(today, yesterday, week, month, year, 7_days, last_7_days, 30_days, last_30_days, 6_months, last_6_months, 12_months, last_12_months, last_year, any, all_time)
// @Param from query string false "Start date (e.g. '2021-02-07')"
// @Param to query string false "End date (e.g. '2021-02-08')"
// @Param recompute query bool false "Whether to recompute the summary from raw heartbeat or use cache"
// @Param project query string false "Project to filter by"
// @Param language query string false "Language to filter by"
// @Param editor query string false "Editor to filter by"
// @Param operating_system query string false "OS to filter by"
// @Param machine query string false "Machine to filter by"
// @Param label query string false "Project label to filter by"
// @Security ApiKeyAuth
// @Success 200 {object} models.Summary
// @Router /summary [get]
func (h *SummaryApiHandler) Get(w http.ResponseWriter, r *http.Request) {
	summary, err, status := routeutils.LoadUserSummary(h.summarySrvc, r)
	if err != nil {
		routeutils.RespondJSONError(w, status, err.Error())
		return
	}

	helpers.RespondJSON(w, r, http.StatusOK, summary)
}

func (h *SummaryApiHandler) GetDetails(w http.ResponseWriter, r *http.Request) {
	summaryParams, err := helpers.ParseSummaryParams(r)
	if err != nil {
		routeutils.RespondJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	summary, err, status := routeutils.LoadUserSummaryByParams(h.summarySrvc, summaryParams)
	if err != nil {
		routeutils.RespondJSONError(w, status, err.Error())
		return
	}

	summaryWithoutFilter, err, status := routeutils.LoadUserSummaryWithoutFilter(h.summarySrvc, summaryParams)
	if err != nil {
		routeutils.RespondJSONError(w, status, err.Error())
		return
	}

	user := middlewares.GetPrincipal(r)
	firstData, err := h.heartbeatsSrvc.GetFirstByUser(user)
	if err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "failed to load first heartbeat")
		return
	}

	var timeline []*view.TimelineViewModel
	if rangeDays := summaryParams.RangeDays(); rangeDays >= dailyStatsMinRangeDays && rangeDays <= dailyStatsMaxRangeDays {
		if dailyStatsSummaries, err := h.fetchSplitSummaries(summaryParams); err == nil {
			timeline = view.NewTimelineViewModel(dailyStatsSummaries)
		}
	}

	var hourlyBreakdown view.HourlyBreakdownsViewModel
	hourlyBreakdownFrom := summaryParams.From
	if summaryParams.RangeDays() > 1 {
		hourlyBreakdownFrom = summaryParams.To.Add(-24 * time.Hour)
	}
	if durations, err := h.durationSrvc.Get(hourlyBreakdownFrom, summaryParams.To, summaryParams.User, summaryParams.Filters, nil, false); err == nil && len(durations) <= 200 {
		hourlyBreakdown = view.NewHourlyBreakdownViewModel(view.NewHourlyBreakdownItems(durations, func(t uint8, k string) string {
			s, _ := h.aliasSrvc.GetAliasOrDefault(user.ID, t, k)
			return s
		}))
	}

	routeutils.RespondJSON(w, http.StatusOK, SummaryDetailsResponse{
		Summary:             summary,
		AvailableFilters:    newAvailableFiltersResponse(summaryWithoutFilter),
		EditorColors:        routeutils.FilterColors(h.config.App.GetEditorColors(), summary.Editors),
		LanguageColors:      routeutils.FilterColors(h.config.App.GetLanguageColors(), summary.Languages),
		OSColors:            routeutils.FilterColors(h.config.App.GetOSColors(), summary.OperatingSystems),
		Timeline:            newTimelineResponse(timeline),
		HourlyBreakdown:     newHourlyBreakdownResponse(hourlyBreakdown),
		HourlyBreakdownFrom: hourlyBreakdownFrom,
		UserFirstData:       firstData,
		DataRetentionMonths: h.config.App.DataRetentionMonths,
		UserDataExpiring: h.config.Subscriptions.Enabled &&
			h.config.App.DataRetentionMonths > 0 &&
			!firstData.IsZero() &&
			!user.HasActiveSubscription() &&
			time.Now().AddDate(0, -h.config.App.DataRetentionMonths, 0).After(firstData),
		ProjectDetails: summaryParams.IsProjectDetails(),
		Project:        summaryParams.GetProjectFilter(),
	})
}

func (h *SummaryApiHandler) fetchSplitSummaries(params *models.SummaryParams) ([]*models.Summary, error) {
	summaries := make([]*models.Summary, 0)
	intervals := utils.SplitRangeByDays(params.From, params.To)
	for _, interval := range intervals {
		curSummary, err := h.summarySrvc.Aliased(interval[0], interval[1], params.User, h.summarySrvc.Retrieve, params.Filters, nil, false)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, curSummary)
	}
	return summaries, nil
}

func newAvailableFiltersResponse(summary *models.Summary) AvailableFiltersResponse {
	return AvailableFiltersResponse{
		Projects:   slice.Map(summary.Projects, func(_ int, item *models.SummaryItem) string { return item.Key }),
		Languages:  slice.Map(summary.Languages, func(_ int, item *models.SummaryItem) string { return item.Key }),
		Machines:   slice.Map(summary.Machines, func(_ int, item *models.SummaryItem) string { return item.Key }),
		Labels:     slice.Map(summary.Labels, func(_ int, item *models.SummaryItem) string { return item.Key }),
		Categories: slice.Map(summary.Categories, func(_ int, item *models.SummaryItem) string { return item.Key }),
	}
}

func newTimelineResponse(timeline []*view.TimelineViewModel) []TimelineResponse {
	response := make([]TimelineResponse, 0, len(timeline))
	for _, day := range timeline {
		projects := make([]TimelineItemResponse, 0, len(day.Projects))
		for _, project := range day.Projects {
			projects = append(projects, TimelineItemResponse{
				Name:     project.Name,
				Duration: int64(project.Duration / time.Second),
			})
		}
		response = append(response, TimelineResponse{
			Date:     day.Date,
			Projects: projects,
		})
	}
	return response
}

func newHourlyBreakdownResponse(breakdown view.HourlyBreakdownsViewModel) []HourlyBreakdownProjectResponse {
	response := make([]HourlyBreakdownProjectResponse, 0, len(breakdown))
	for _, project := range breakdown {
		items := make([]HourlyBreakdownItemResponse, 0, len(project.Items))
		for _, item := range project.Items {
			items = append(items, HourlyBreakdownItemResponse{
				FromTime: item.FromTime,
				Duration: int64(item.Duration / time.Second),
				Entity:   item.Entity,
			})
		}
		response = append(response, HourlyBreakdownProjectResponse{
			Project: project.Project,
			Items:   items,
		})
	}
	return response
}
