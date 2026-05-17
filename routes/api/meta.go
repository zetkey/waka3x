package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	conf "github.com/zetkey/waka3x/config"
	"github.com/zetkey/waka3x/helpers"
	"github.com/zetkey/waka3x/models"
	routeutils "github.com/zetkey/waka3x/routes/utils"
	"github.com/zetkey/waka3x/services"
)

type MetaApiHandler struct {
	config       *conf.Config
	userSrvc     services.IUserService
	keyValueSrvc services.IKeyValueService
}

func NewMetaApiHandler(userService services.IUserService, keyValueService services.IKeyValueService) *MetaApiHandler {
	return &MetaApiHandler{
		config:       conf.Get(),
		userSrvc:     userService,
		keyValueSrvc: keyValueService,
	}
}

func (h *MetaApiHandler) RegisterRoutes(router chi.Router) {
	router.Get("/config", h.GetConfig)
	router.Get("/home", h.GetHome)
	router.Get("/imprint", h.GetImprint)
	router.Get("/setup", h.GetSetup)
	router.Post("/unsubscribe", h.PostUnsubscribe)
}

func (h *MetaApiHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	providers := make([]OidcProviderResponse, 0)
	for _, providerName := range h.config.Security.ListOidcProviders() {
		provider, err := conf.GetOidcProvider(providerName)
		if err != nil {
			continue
		}
		providers = append(providers, OidcProviderResponse{
			Name:        provider.Name,
			DisplayName: provider.DisplayName,
			LoginURL:    strings.TrimSuffix(h.config.Server.BasePath, "/") + "/oidc/" + provider.Name + "/login",
		})
	}

	routeutils.RespondJSON(w, http.StatusOK, BootstrapResponse{
		Version:                h.config.Version,
		BasePath:               h.config.Server.BasePath,
		PublicURL:              h.config.Server.GetPublicUrl(),
		DBType:                 strings.ToLower(h.config.Db.Type),
		LeaderboardEnabled:     h.config.App.LeaderboardEnabled,
		LeaderboardRequireAuth: h.config.App.LeaderboardRequireAuth,
		AllowSignup:            h.config.IsDev() || h.config.Security.AllowSignup,
		InviteCodesEnabled:     h.config.Security.InviteCodes,
		SignupCaptcha:          h.config.Security.SignupCaptcha,
		DisableLocalAuth:       h.config.Security.DisableLocalAuth,
		DisableWebAuthn:        h.config.Security.DisableWebAuthn,
		SubscriptionsEnabled:   h.config.Subscriptions.Enabled,
		SubscriptionPrice:      h.config.Subscriptions.StandardPrice,
		StripeAPIKey:           h.config.Subscriptions.StripeApiKey,
		SupportContact:         h.config.App.SupportContact,
		DataRetentionMonths:    h.config.App.DataRetentionMonths,
		DefaultWakatimeURL:     conf.WakatimeApiUrl,
		AvatarURLTemplate:      h.config.App.AvatarURLTemplate,
		OIDCProviders:          providers,
		MailEnabled:            h.config.Mail.Enabled,
		ImportEnabled:          h.config.App.ImportEnabled,
		ImportBackoffMin:       h.config.App.ImportBackoffMin,
		ImportMaxRateHours:     h.config.App.ImportMaxRate,
	})
}

func (h *MetaApiHandler) GetHome(w http.ResponseWriter, r *http.Request) {
	var (
		totalHours      int
		totalUsers      int
		currentlyOnline int
		newsbox         NewsboxResponse
	)

	if kv, err := h.keyValueSrvc.GetString(conf.KeyLatestTotalTime); err == nil && kv != nil && kv.Value != "" {
		if d, err := time.ParseDuration(kv.Value); err == nil {
			totalHours = int(d.Hours())
		}
	}
	if kv, err := h.keyValueSrvc.GetString(conf.KeyLatestTotalUsers); err == nil && kv != nil && kv.Value != "" {
		if d, err := strconv.Atoi(kv.Value); err == nil {
			totalUsers = d
		}
	}
	if kv, err := h.keyValueSrvc.GetString(conf.KeyNewsbox); err == nil && kv != nil && kv.Value != "" {
		_ = json.NewDecoder(strings.NewReader(kv.Value)).Decode(&newsbox)
	}
	if c, err := h.userSrvc.CountCurrentlyOnline(); err == nil {
		currentlyOnline = c
	}

	routeutils.RespondJSON(w, http.StatusOK, HomeStatsResponse{
		TotalHours:      totalHours,
		TotalUsers:      totalUsers,
		CurrentlyOnline: currentlyOnline,
		Newsbox:         &newsbox,
	})
}

func (h *MetaApiHandler) GetImprint(w http.ResponseWriter, r *http.Request) {
	text := "failed to load content"
	if data, err := h.keyValueSrvc.GetString(models.ImprintKey); err == nil && data != nil {
		text = data.Value
	}
	routeutils.RespondJSON(w, http.StatusOK, ImprintResponse{HTML: text})
}

func (h *MetaApiHandler) GetSetup(w http.ResponseWriter, r *http.Request) {
	response := SetupResponse{
		BaseURL:   strings.TrimSuffix(h.config.Server.GetPublicUrl(), "/") + "/api",
		PublicURL: h.config.Server.GetPublicUrl(),
	}

	if username, err := helpers.ExtractCookieAuth(r, h.config); err == nil && username != nil {
		if user, err := h.userSrvc.GetUserById(*username); err == nil && user != nil {
			response.APIKey = user.ApiKey
			response.Username = user.ID
			response.Authenticated = true
		}
	}

	routeutils.RespondJSON(w, http.StatusOK, response)
}

func (h *MetaApiHandler) PostUnsubscribe(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Token == "" {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "missing token")
		return
	}

	user, err := h.userSrvc.GetUserByUnsubscribeToken(payload.Token)
	if err != nil {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "invalid token")
		return
	}

	user.ReportsWeekly = false
	if _, err := h.userSrvc.Update(user); err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "failed to update user preferences")
		return
	}

	routeutils.RespondJSON(w, http.StatusOK, MessageResponse{Message: "successfully unsubscribed from weekly reports"})
}
