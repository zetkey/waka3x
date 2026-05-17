package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	datastructure "github.com/duke-git/lancet/v2/datastructure/set"
	"github.com/go-chi/chi/v5"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofrs/uuid/v5"
	conf "github.com/zetkey/waka3x/config"
	"github.com/zetkey/waka3x/helpers"
	"github.com/zetkey/waka3x/middlewares"
	"github.com/zetkey/waka3x/models"
	routeutils "github.com/zetkey/waka3x/routes/utils"
	"github.com/zetkey/waka3x/services"
	"github.com/zetkey/waka3x/services/imports"
	"github.com/zetkey/waka3x/utils"
)

type SettingsApiHandler struct {
	config              *conf.Config
	userSrvc            services.IUserService
	heartbeatSrvc       services.IHeartbeatService
	durationSrvc        services.IDurationService
	summarySrvc         services.ISummaryService
	aliasSrvc           services.IAliasService
	aggregationSrvc     services.IAggregationService
	languageMappingSrvc services.ILanguageMappingService
	projectLabelSrvc    services.IProjectLabelService
	keyValueSrvc        services.IKeyValueService
	mailSrvc            services.IMailService
	apiKeySrvc          services.IApiKeyService
	webAuthnSrvc        services.IWebAuthnService
	httpClient          *http.Client
	locksMu             sync.Mutex
	aggregationLocks    map[string]bool
}

func NewSettingsApiHandler(
	userService services.IUserService,
	heartbeatService services.IHeartbeatService,
	durationService services.IDurationService,
	summaryService services.ISummaryService,
	aliasService services.IAliasService,
	aggregationService services.IAggregationService,
	languageMappingService services.ILanguageMappingService,
	projectLabelService services.IProjectLabelService,
	keyValueService services.IKeyValueService,
	mailService services.IMailService,
	apiKeyService services.IApiKeyService,
	webAuthnService services.IWebAuthnService,
) *SettingsApiHandler {
	return &SettingsApiHandler{
		config:              conf.Get(),
		userSrvc:            userService,
		heartbeatSrvc:       heartbeatService,
		durationSrvc:        durationService,
		summarySrvc:         summaryService,
		aliasSrvc:           aliasService,
		aggregationSrvc:     aggregationService,
		languageMappingSrvc: languageMappingService,
		projectLabelSrvc:    projectLabelService,
		keyValueSrvc:        keyValueService,
		mailSrvc:            mailService,
		apiKeySrvc:          apiKeyService,
		webAuthnSrvc:        webAuthnService,
		httpClient:          &http.Client{Timeout: 10 * time.Second},
		aggregationLocks:    make(map[string]bool),
	}
}

func (h *SettingsApiHandler) RegisterRoutes(router chi.Router) {
	r := chi.NewRouter()
	r.Use(middlewares.NewAuthenticateMiddleware(h.userSrvc).Handler)

	r.Get("/", h.Get)
	r.Post("/password", h.ChangePassword)
	r.Post("/username", h.ChangeUserID)
	r.Post("/api-key/reset", h.ResetAPIKey)
	r.Post("/api-keys", h.AddAPIKey)
	r.Delete("/api-keys", h.DeleteAPIKey)
	r.Post("/invite", h.GenerateInvite)
	r.Put("/unknown-projects", h.UpdateUnknownProjects)
	r.Put("/heartbeats-timeout", h.UpdateHeartbeatsTimeout)
	r.Put("/readme-stats-base-url", h.UpdateReadmeStatsBaseURL)
	r.Put("/leaderboard", h.UpdateLeaderboard)
	r.Put("/sharing", h.UpdateSharing)
	r.Post("/aliases", h.AddAlias)
	r.Delete("/aliases", h.DeleteAlias)
	r.Post("/labels", h.AddLabel)
	r.Delete("/labels", h.DeleteLabel)
	r.Post("/language-mappings", h.AddLanguageMapping)
	r.Delete("/language-mappings", h.DeleteLanguageMapping)
	r.Put("/wakatime", h.UpdateWakatime)
	r.Post("/wakatime/import", h.ImportWakatime)
	r.Post("/summaries/regenerate", h.RegenerateSummaries)
	r.Post("/data/clear", h.ClearData)
	r.Delete("/account", h.DeleteAccount)
	r.Get("/webauthn/options", h.GetWebAuthnOptions)
	r.Post("/webauthn", h.WebAuthnAdd)
	r.Delete("/webauthn", h.WebAuthnDelete)

	router.Mount("/settings", r)
}

type settingsResponse struct {
	User                  *CurrentUserResponse         `json:"user"`
	Aliases               []combinedAliasResponse      `json:"aliases"`
	Labels                []combinedLabelResponse      `json:"labels"`
	Projects              []string                     `json:"projects"`
	LanguageMappings      []*models.LanguageMapping    `json:"language_mappings"`
	APIKeys               []settingsAPIKeyResponse     `json:"api_keys"`
	WebAuthnCredentials   []webAuthnCredentialResponse `json:"webauthn_credentials"`
	UserFirstData         time.Time                    `json:"user_first_data"`
	SubscriptionPrice     string                       `json:"subscription_price"`
	SubscriptionsEnabled  bool                         `json:"subscriptions_enabled"`
	SupportContact        string                       `json:"support_contact"`
	DataRetentionMonths   int                          `json:"data_retention_months"`
	InviteLink            string                       `json:"invite_link,omitempty"`
	ReadmeCardCustomTitle string                       `json:"readme_card_custom_title"`
	DisableWebAuthn       bool                         `json:"disable_webauthn"`
	DefaultWakatimeURL    string                       `json:"default_wakatime_url"`
}

type combinedAliasResponse struct {
	Key    string   `json:"key"`
	Type   uint8    `json:"type"`
	Values []string `json:"values"`
}

type combinedLabelResponse struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

type settingsAPIKeyResponse struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	ReadOnly bool   `json:"read_only"`
	Main     bool   `json:"main"`
}

type webAuthnCredentialResponse struct {
	Name string `json:"name"`
}

func (h *SettingsApiHandler) Get(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	response, err := h.buildSettingsResponse(user, "")
	if err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	routeutils.RespondJSON(w, http.StatusOK, response)
}

func (h *SettingsApiHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	if user.AuthType != "local" {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "cannot reset password for non-local user")
		return
	}

	var payload struct {
		PasswordOld    string `json:"password_old"`
		PasswordNew    string `json:"password_new"`
		PasswordRepeat string `json:"password_repeat"`
	}
	if !decodeJSON(w, r, &payload) {
		return
	}

	if !models.ValidatePassword(payload.PasswordNew) || payload.PasswordNew != payload.PasswordRepeat {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "invalid parameters")
		return
	}
	if !utils.ComparePassword(user.Password, payload.PasswordOld, h.config.Security.PasswordSalt) {
		routeutils.RespondJSONError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	hash, err := utils.HashPassword(payload.PasswordNew, h.config.Security.PasswordSalt)
	if err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "failed to update password")
		return
	}
	user.Password = hash
	if _, err := h.userSrvc.Update(user); err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "failed to update password")
		return
	}
	routeutils.RespondJSON(w, http.StatusOK, MessageResponse{Message: "password was updated successfully", User: newCurrentUserResponse(user)})
}

func (h *SettingsApiHandler) ChangeUserID(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	var payload struct {
		NewUserID string `json:"new_userid"`
	}
	if !decodeJSON(w, r, &payload) {
		return
	}
	newUserID := strings.TrimSpace(payload.NewUserID)
	if !models.ValidateUsername(newUserID) || newUserID == user.ID {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "invalid username")
		return
	}
	if existing, _ := h.userSrvc.GetUserById(newUserID); existing != nil {
		routeutils.RespondJSONError(w, http.StatusConflict, "already taken")
		return
	}
	if _, err := h.userSrvc.ChangeUserId(user, newUserID); err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	http.SetCookie(w, h.config.GetClearCookie(models.AuthCookieKey))
	routeutils.RespondJSON(w, http.StatusOK, MessageResponse{Message: "username changed, please log back in"})
}

func (h *SettingsApiHandler) ResetAPIKey(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	if _, err := h.userSrvc.ResetApiKey(user); err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "failed to reset api key")
		return
	}
	routeutils.RespondJSON(w, http.StatusOK, MessageResponse{Message: "api key reset", User: newCurrentUserResponse(user)})
}

func (h *SettingsApiHandler) AddAPIKey(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name     string `json:"api_name"`
		ReadOnly bool   `json:"api_readonly"`
	}
	if !decodeJSON(w, r, &payload) {
		return
	}

	apiKey := uuid.Must(uuid.NewV4()).String()
	if _, err := h.apiKeySrvc.Create(&models.ApiKey{
		User:     middlewares.GetPrincipal(r),
		Label:    payload.Name,
		ApiKey:   apiKey,
		ReadOnly: payload.ReadOnly,
	}); err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "failed to add API key")
		return
	}
	routeutils.RespondJSON(w, http.StatusCreated, map[string]string{"message": "api key added", "api_key": apiKey})
}

func (h *SettingsApiHandler) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	var payload struct {
		Value string `json:"api_key_value"`
	}
	if !decodeJSON(w, r, &payload) {
		return
	}
	if payload.Value == user.ApiKey {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "main api key can only be regenerated")
		return
	}
	apiKeys, err := h.apiKeySrvc.GetByUser(user.ID)
	if err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "could not delete API key")
		return
	}
	for _, key := range apiKeys {
		if key.ApiKey == payload.Value {
			if err := h.apiKeySrvc.Delete(key); err != nil {
				routeutils.RespondJSONError(w, http.StatusInternalServerError, "could not delete API key")
				return
			}
			routeutils.RespondJSON(w, http.StatusOK, MessageResponse{Message: "API key deleted successfully"})
			return
		}
	}
	routeutils.RespondJSONError(w, http.StatusNotFound, "API key not found")
}

func (h *SettingsApiHandler) GenerateInvite(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	inviteCode := uuid.Must(uuid.NewV4()).String()[0:8]
	if err := h.keyValueSrvc.PutString(&models.KeyStringValue{
		Key:   fmt.Sprintf("%s_%s", conf.KeyInviteCode, inviteCode),
		Value: fmt.Sprintf("%s,%s", user.ID, time.Now().Format(time.RFC3339)),
	}); err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "failed to generate invite code")
		return
	}
	routeutils.RespondJSON(w, http.StatusOK, map[string]string{
		"message":     "invite generated",
		"invite_link": fmt.Sprintf("%s/signup?invite=%s", h.config.Server.GetPublicUrl(), inviteCode),
	})
}

func (h *SettingsApiHandler) UpdateUnknownProjects(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	var payload struct {
		ExcludeUnknownProjects bool `json:"exclude_unknown_projects"`
	}
	if !decodeJSON(w, r, &payload) {
		return
	}
	if h.isAggregationLocked(user.ID) {
		routeutils.RespondJSONError(w, http.StatusConflict, "summary regeneration already in progress")
		return
	}
	user.ExcludeUnknownProjects = payload.ExcludeUnknownProjects
	if _, err := h.userSrvc.Update(user); err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "failed to update setting")
		return
	}
	h.regenerateSummariesAsync(user, r)
	routeutils.RespondJSON(w, http.StatusOK, MessageResponse{Message: "regenerating summaries", User: newCurrentUserResponse(user)})
}

func (h *SettingsApiHandler) UpdateHeartbeatsTimeout(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	var payload struct {
		HeartbeatsTimeout int `json:"heartbeats_timeout"`
	}
	if !decodeJSON(w, r, &payload) {
		return
	}
	dur := time.Duration(payload.HeartbeatsTimeout) * time.Minute
	if dur < models.MinHeartbeatsTimeout || dur > models.MaxHeartbeatsTimeout {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "invalid input")
		return
	}
	user.HeartbeatsTimeoutSec = int(dur.Seconds())
	if _, err := h.userSrvc.Update(user); err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "failed to update setting")
		return
	}
	routeutils.RespondJSON(w, http.StatusOK, MessageResponse{Message: "setting updated", User: newCurrentUserResponse(user)})
}

func (h *SettingsApiHandler) UpdateReadmeStatsBaseURL(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	var payload struct {
		ReadmeStatsBaseURL string `json:"readme_stats_base_url"`
	}
	if !decodeJSON(w, r, &payload) {
		return
	}
	user.ReadmeStatsBaseUrl = payload.ReadmeStatsBaseURL
	if _, err := h.userSrvc.Update(user); err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "failed to update setting")
		return
	}
	routeutils.RespondJSON(w, http.StatusOK, MessageResponse{Message: "setting updated", User: newCurrentUserResponse(user)})
}

func (h *SettingsApiHandler) UpdateLeaderboard(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	var payload struct {
		EnableLeaderboard bool `json:"enable_leaderboard"`
	}
	if !decodeJSON(w, r, &payload) {
		return
	}
	user.PublicLeaderboard = payload.EnableLeaderboard
	if _, err := h.userSrvc.Update(user); err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "failed to update setting")
		return
	}
	routeutils.RespondJSON(w, http.StatusOK, MessageResponse{Message: "setting updated", User: newCurrentUserResponse(user)})
}

func (h *SettingsApiHandler) UpdateSharing(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	var payload struct {
		MaxDays            int  `json:"max_days"`
		ShareProjects      bool `json:"share_projects"`
		ShareLanguages     bool `json:"share_languages"`
		ShareEditors       bool `json:"share_editors"`
		ShareOSs           bool `json:"share_oss"`
		ShareMachines      bool `json:"share_machines"`
		ShareLabels        bool `json:"share_labels"`
		ShareActivityChart bool `json:"share_activity_chart"`
	}
	if !decodeJSON(w, r, &payload) {
		return
	}
	user.ShareDataMaxDays = payload.MaxDays
	user.ShareProjects = payload.ShareProjects
	user.ShareLanguages = payload.ShareLanguages
	user.ShareEditors = payload.ShareEditors
	user.ShareOSs = payload.ShareOSs
	user.ShareMachines = payload.ShareMachines
	user.ShareLabels = payload.ShareLabels
	user.ShareActivityChart = payload.ShareActivityChart
	if _, err := h.userSrvc.Update(user); err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "failed to update sharing")
		return
	}
	routeutils.RespondJSON(w, http.StatusOK, MessageResponse{Message: "sharing updated", User: newCurrentUserResponse(user)})
}

func (h *SettingsApiHandler) AddAlias(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	var payload struct {
		Type  uint8  `json:"type"`
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if !decodeJSON(w, r, &payload) {
		return
	}
	alias := &models.Alias{UserID: user.ID, Type: payload.Type, Key: payload.Key, Value: payload.Value}
	if _, err := h.aliasSrvc.Create(alias); err != nil {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "invalid input")
		return
	}
	routeutils.RespondJSON(w, http.StatusCreated, MessageResponse{Message: "alias added successfully"})
}

func (h *SettingsApiHandler) DeleteAlias(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	var payload struct {
		Type uint8  `json:"type"`
		Key  string `json:"key"`
	}
	if !decodeJSON(w, r, &payload) {
		return
	}
	aliases, err := h.aliasSrvc.GetByUserAndKeyAndType(user.ID, payload.Key, payload.Type)
	if err != nil || len(aliases) == 0 {
		routeutils.RespondJSONError(w, http.StatusNotFound, "aliases not found")
		return
	}
	if err := h.aliasSrvc.DeleteMulti(aliases); err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "could not delete aliases")
		return
	}
	routeutils.RespondJSON(w, http.StatusOK, MessageResponse{Message: "aliases deleted successfully"})
}

func (h *SettingsApiHandler) AddLabel(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	var payload struct {
		Keys  []string `json:"key"`
		Value string   `json:"value"`
	}
	if !decodeJSON(w, r, &payload) {
		return
	}
	for _, key := range payload.Keys {
		label := &models.ProjectLabel{UserID: user.ID, ProjectKey: key, Label: payload.Value}
		if !label.IsValid() {
			routeutils.RespondJSONError(w, http.StatusBadRequest, "invalid input")
			return
		}
		if _, err := h.projectLabelSrvc.Create(label); err != nil {
			routeutils.RespondJSONError(w, http.StatusBadRequest, "invalid input")
			return
		}
	}
	routeutils.RespondJSON(w, http.StatusCreated, MessageResponse{Message: "label added successfully"})
}

func (h *SettingsApiHandler) DeleteLabel(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	var payload struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if !decodeJSON(w, r, &payload) {
		return
	}
	labels, err := h.projectLabelSrvc.GetByUser(user.ID)
	if err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "could not delete label")
		return
	}
	for _, label := range labels {
		if label.Label == payload.Key && label.ProjectKey == payload.Value {
			if err := h.projectLabelSrvc.Delete(label); err != nil {
				routeutils.RespondJSONError(w, http.StatusInternalServerError, "could not delete label")
				return
			}
			routeutils.RespondJSON(w, http.StatusOK, MessageResponse{Message: "label deleted successfully"})
			return
		}
	}
	routeutils.RespondJSONError(w, http.StatusNotFound, "label not found")
}

func (h *SettingsApiHandler) AddLanguageMapping(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	var payload struct {
		Extension string `json:"extension"`
		Language  string `json:"language"`
	}
	if !decodeJSON(w, r, &payload) {
		return
	}
	extension := strings.TrimPrefix(payload.Extension, ".")
	mapping := &models.LanguageMapping{UserID: user.ID, Extension: extension, Language: payload.Language}
	if _, err := h.languageMappingSrvc.Create(mapping); err != nil {
		routeutils.RespondJSONError(w, http.StatusConflict, "mapping already exists")
		return
	}
	routeutils.RespondJSON(w, http.StatusCreated, MessageResponse{Message: "mapping added successfully"})
}

func (h *SettingsApiHandler) DeleteLanguageMapping(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	var payload struct {
		MappingID uint `json:"mapping_id"`
	}
	if !decodeJSON(w, r, &payload) {
		return
	}
	mapping, err := h.languageMappingSrvc.GetById(payload.MappingID)
	if err != nil || mapping == nil {
		routeutils.RespondJSONError(w, http.StatusNotFound, "mapping not found")
		return
	}
	if mapping.UserID != user.ID {
		routeutils.RespondJSONError(w, http.StatusForbidden, "not allowed to delete mapping")
		return
	}
	if err := h.languageMappingSrvc.Delete(mapping); err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "could not delete mapping")
		return
	}
	routeutils.RespondJSON(w, http.StatusOK, MessageResponse{Message: "mapping deleted successfully"})
}

func (h *SettingsApiHandler) UpdateWakatime(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	var payload struct {
		APIURL string `json:"api_url"`
		APIKey string `json:"api_key"`
	}
	if !decodeJSON(w, r, &payload) {
		return
	}
	if payload.APIURL == conf.WakatimeApiUrl || payload.APIKey == "" {
		payload.APIURL = ""
	}
	if payload.APIURL != "" && routeutils.ValidateWakatimeUrl(payload.APIURL) != nil {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "invalid WakaTime API URL")
		return
	}
	if user.WakatimeApiKey == "" && payload.APIKey != "" && !h.validateWakatimeKey(payload.APIKey, payload.APIURL) {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "failed to connect to WakaTime, API key or endpoint URL invalid?")
		return
	}
	if _, err := h.userSrvc.SetWakatimeApiCredentials(user, payload.APIKey, payload.APIURL); err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "failed to update WakaTime credentials")
		return
	}
	routeutils.RespondJSON(w, http.StatusOK, MessageResponse{Message: "WakaTime API key updated", User: newCurrentUserResponse(user)})
}

func (h *SettingsApiHandler) ImportWakatime(w http.ResponseWriter, r *http.Request) {
	if !h.config.App.ImportEnabled {
		routeutils.RespondJSONError(w, http.StatusForbidden, "imports are disabled on this server")
		return
	}
	user := middlewares.GetPrincipal(r)
	if user.WakatimeApiKey == "" {
		routeutils.RespondJSONError(w, http.StatusForbidden, "not connected to WakaTime")
		return
	}

	var payload struct {
		UseLegacyImporter bool `json:"use_legacy_importer"`
	}
	if !decodeJSON(w, r, &payload) {
		return
	}

	kvKeyLastImport := fmt.Sprintf("%s_%s", conf.KeyLastImport, user.ID)
	kvKeyLastImportSuccess := fmt.Sprintf("%s_%s", conf.KeyLastImportSuccess, user.ID)

	importer := imports.NewWakatimeImporter(user.WakatimeApiKey, payload.UseLegacyImporter)
	if err := importer.Validate(user); err != nil {
		routeutils.RespondJSONError(w, http.StatusForbidden, fmt.Sprintf("failed to import: %v", err))
		return
	}

	if !h.config.IsDev() {
		lastImport, _ := time.Parse(time.RFC822, h.keyValueSrvc.MustGetString(kvKeyLastImport).Value)
		if time.Since(lastImport) < time.Duration(h.config.App.ImportBackoffMin)*time.Minute {
			routeutils.RespondJSONError(w, http.StatusTooManyRequests, fmt.Sprintf("too many data imports, please wait %d minutes between import requests", h.config.App.ImportBackoffMin))
			return
		}

		lastImportSuccess, _ := time.Parse(time.RFC822, h.keyValueSrvc.MustGetString(kvKeyLastImportSuccess).Value)
		if time.Since(lastImportSuccess) < time.Duration(h.config.App.ImportMaxRate)*time.Hour {
			routeutils.RespondJSONError(w, http.StatusTooManyRequests, fmt.Sprintf("too many data imports, last import ran less than %d hours ago", h.config.App.ImportMaxRate))
			return
		}
	}

	go h.importWakatime(user, importer, kvKeyLastImportSuccess, r)

	_ = h.keyValueSrvc.PutString(&models.KeyStringValue{
		Key:   kvKeyLastImport,
		Value: time.Now().Format(time.RFC822),
	})

	routeutils.RespondJSON(w, http.StatusAccepted, MessageResponse{Message: "import started, this will take several minutes"})
}

func (h *SettingsApiHandler) RegenerateSummaries(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	if h.isAggregationLocked(user.ID) {
		routeutils.RespondJSONError(w, http.StatusConflict, "summary regeneration already in progress")
		return
	}
	h.regenerateSummariesAsync(user, r)
	routeutils.RespondJSON(w, http.StatusAccepted, MessageResponse{Message: "summaries are being regenerated"})
}

func (h *SettingsApiHandler) ClearData(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	go func(user *models.User, r *http.Request) {
		if err := h.summarySrvc.DeleteByUser(user.ID); err != nil {
			conf.Log().Request(r).Error("failed to clear summaries", "error", err)
		}
		if err := h.durationSrvc.DeleteByUser(user); err != nil {
			conf.Log().Request(r).Error("failed to clear durations", "error", err)
		}
		if err := h.heartbeatSrvc.DeleteByUser(user); err != nil {
			conf.Log().Request(r).Error("failed to clear heartbeats", "error", err)
		}
		user.HasData = false
		if _, err := h.userSrvc.Update(user); err != nil {
			conf.Log().Request(r).Error("failed to update user after clear data", "error", err)
		}
	}(user, r)
	routeutils.RespondJSON(w, http.StatusAccepted, MessageResponse{Message: "deletion in progress"})
}

func (h *SettingsApiHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	go func(user *models.User, r *http.Request) {
		if err := h.userSrvc.Delete(user); err != nil {
			conf.Log().Request(r).Error("failed to delete user", "userID", user.ID, "error", err)
		}
	}(user, r)
	routeutils.ClearSession(r, w)
	http.SetCookie(w, h.config.GetClearCookie(models.AuthCookieKey))
	routeutils.RespondJSON(w, http.StatusAccepted, MessageResponse{Message: "account deletion queued"})
}

func (h *SettingsApiHandler) GetWebAuthnOptions(w http.ResponseWriter, r *http.Request) {
	if h.config.Security.DisableWebAuthn {
		routeutils.RespondJSONError(w, http.StatusForbidden, "webauthn is disabled on this server")
		return
	}
	user := middlewares.GetPrincipal(r)
	if user.AuthType != "local" {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "webauthn is only available for local users")
		return
	}
	if err := h.webAuthnSrvc.LoadCredentialIntoUser(user); err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "error while loading webauthn credentials")
		return
	}
	webAuthnOptions, session, err := conf.WebAuthn.BeginRegistration(user, webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
		RequireResidentKey: protocol.ResidentKeyRequired(),
	}))
	if err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "error while getting webauthn registration options")
		return
	}
	if err = routeutils.SetWebAuthnSession(session, r, w); err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "error while setting webauthn session")
		return
	}
	routeutils.RespondJSON(w, http.StatusOK, webAuthnOptions)
}

func (h *SettingsApiHandler) WebAuthnAdd(w http.ResponseWriter, r *http.Request) {
	if h.config.Security.DisableWebAuthn {
		routeutils.RespondJSONError(w, http.StatusForbidden, "webauthn is disabled on this server")
		return
	}
	user := middlewares.GetPrincipal(r)
	if user.AuthType != "local" {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "cannot add webauthn authenticator for non-local user")
		return
	}
	var payload struct {
		AuthenticatorName string `json:"authenticator_name"`
		CredentialJSON    string `json:"credential_json"`
	}
	if !decodeJSON(w, r, &payload) {
		return
	}
	if strings.TrimSpace(payload.AuthenticatorName) == "" {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "authenticator name must not be empty")
		return
	}
	if err := h.webAuthnSrvc.LoadCredentialIntoUser(user); err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "could not load webauthn credentials")
		return
	}
	for _, c := range user.Credentials {
		if c.Name == payload.AuthenticatorName {
			routeutils.RespondJSONError(w, http.StatusBadRequest, "authenticator name already in use")
			return
		}
	}
	sessionData, err := routeutils.GetWebAuthnSession(r)
	if err != nil {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "could not get webauthn session data")
		return
	}
	pcc, err := protocol.ParseCredentialCreationResponseBytes([]byte(payload.CredentialJSON))
	if err != nil {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "could not parse credential creation response")
		return
	}
	credential, err := conf.WebAuthn.CreateCredential(user, *sessionData, pcc)
	if err != nil {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "could not create webauthn credential")
		return
	}
	if _, err := h.webAuthnSrvc.CreateCredential(credential, user, payload.AuthenticatorName); err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "could not store webauthn credential")
		return
	}
	routeutils.RespondJSON(w, http.StatusCreated, MessageResponse{Message: "webauthn authenticator added successfully"})
}

func (h *SettingsApiHandler) WebAuthnDelete(w http.ResponseWriter, r *http.Request) {
	if h.config.Security.DisableWebAuthn {
		routeutils.RespondJSONError(w, http.StatusForbidden, "webauthn is disabled on this server")
		return
	}
	user := middlewares.GetPrincipal(r)
	var payload struct {
		CredentialName string `json:"credential_name"`
	}
	if !decodeJSON(w, r, &payload) {
		return
	}
	credential, err := h.webAuthnSrvc.GetCredentialByUserAndName(user, payload.CredentialName)
	if err != nil || credential == nil {
		routeutils.RespondJSONError(w, http.StatusNotFound, "webauthn credential not found")
		return
	}
	if err := h.webAuthnSrvc.DeleteCredential(credential); err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "could not delete webauthn credential")
		return
	}
	routeutils.RespondJSON(w, http.StatusOK, MessageResponse{Message: "webauthn authenticator deleted successfully"})
}

func (h *SettingsApiHandler) buildSettingsResponse(user *models.User, inviteLink string) (*settingsResponse, error) {
	mappings, _ := h.languageMappingSrvc.GetByUser(user.ID)
	aliases, err := h.aliasSrvc.GetByUser(user.ID)
	if err != nil {
		return nil, fmt.Errorf("error while loading aliases")
	}
	labels, err := h.projectLabelSrvc.GetByUserGroupedInverted(user.ID)
	if err != nil {
		return nil, fmt.Errorf("error while loading labels")
	}
	projects, err := routeutils.GetEffectiveProjectsList(user, h.heartbeatSrvc, h.aliasSrvc)
	if err != nil {
		return nil, fmt.Errorf("error while loading projects")
	}
	firstData, _ := h.heartbeatSrvc.GetFirstByUser(user)
	apiKeys, err := h.apiKeySrvc.GetByUser(user.ID)
	if err != nil {
		return nil, fmt.Errorf("error while loading api keys")
	}
	_ = h.webAuthnSrvc.LoadCredentialIntoUser(user)

	readmeCardTitle := "Wakapi.dev Stats"
	if err, maxRange := helpers.ResolveMaximumRange(user.ShareDataMaxDays); err == nil {
		readmeCardTitle += fmt.Sprintf(" (%v)", maxRange.GetHumanReadable())
	}

	return &settingsResponse{
		User:                  newCurrentUserResponse(user),
		Aliases:               combineAliases(aliases),
		Labels:                combineLabels(labels),
		Projects:              projects,
		LanguageMappings:      mappings,
		APIKeys:               combineAPIKeys(user, apiKeys),
		WebAuthnCredentials:   combineWebAuthnCredentials(user.Credentials),
		UserFirstData:         firstData,
		SubscriptionPrice:     h.config.Subscriptions.StandardPrice,
		SubscriptionsEnabled:  h.config.Subscriptions.Enabled,
		SupportContact:        h.config.App.SupportContact,
		DataRetentionMonths:   h.config.App.DataRetentionMonths,
		InviteLink:            inviteLink,
		ReadmeCardCustomTitle: readmeCardTitle,
		DisableWebAuthn:       h.config.Security.DisableWebAuthn,
		DefaultWakatimeURL:    conf.WakatimeApiUrl,
	}, nil
}

func combineAliases(aliases []*models.Alias) []combinedAliasResponse {
	grouped := map[string]*combinedAliasResponse{}
	for _, alias := range aliases {
		key := fmt.Sprintf("%s_%d", alias.Key, alias.Type)
		if _, ok := grouped[key]; !ok {
			grouped[key] = &combinedAliasResponse{Key: alias.Key, Type: alias.Type}
		}
		grouped[key].Values = append(grouped[key].Values, alias.Value)
	}
	result := make([]combinedAliasResponse, 0, len(grouped))
	for _, alias := range grouped {
		sort.Strings(alias.Values)
		result = append(result, *alias)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Type == result[j].Type {
			return result[i].Key < result[j].Key
		}
		return result[i].Type < result[j].Type
	})
	return result
}

func combineLabels(labels map[string][]*models.ProjectLabel) []combinedLabelResponse {
	result := make([]combinedLabelResponse, 0, len(labels))
	for _, group := range labels {
		if len(group) == 0 {
			continue
		}
		item := combinedLabelResponse{Key: group[0].Label}
		for _, label := range group {
			item.Values = append(item.Values, label.ProjectKey)
		}
		sort.Strings(item.Values)
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Key < result[j].Key })
	return result
}

func combineAPIKeys(user *models.User, apiKeys []*models.ApiKey) []settingsAPIKeyResponse {
	result := []settingsAPIKeyResponse{{
		Name:     "Main API Key",
		Value:    user.ApiKey,
		ReadOnly: false,
		Main:     true,
	}}
	for _, key := range apiKeys {
		result = append(result, settingsAPIKeyResponse{Name: key.Label, Value: key.ApiKey, ReadOnly: key.ReadOnly})
	}
	return result
}

func combineWebAuthnCredentials(credentials []*models.WebAuthnCredential) []webAuthnCredentialResponse {
	result := make([]webAuthnCredentialResponse, 0, len(credentials))
	for _, credential := range credentials {
		result = append(result, webAuthnCredentialResponse{Name: credential.Name})
	}
	return result
}

func (h *SettingsApiHandler) validateWakatimeKey(apiKey string, baseURL string) bool {
	if baseURL == "" {
		baseURL = conf.WakatimeApiUrl
	}

	request, err := http.NewRequest(http.MethodGet, baseURL+conf.WakatimeApiUserUrl, nil)
	if err != nil {
		return false
	}
	request.Header = http.Header{
		"Accept": []string{"application/json"},
		"Authorization": []string{
			fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(apiKey))),
		},
	}

	if _, err = utils.RaiseForStatus(h.httpClient.Do(request)); err != nil {
		return false
	}
	return true
}

func (h *SettingsApiHandler) importWakatime(user *models.User, importer *imports.WakatimeImporter, kvKeyLastImportSuccess string, r *http.Request) {
	start := time.Now()
	countBefore, _ := h.heartbeatSrvc.CountByUser(user)

	var (
		stream      <-chan *models.Heartbeat
		importError error
	)
	if latest, err := h.heartbeatSrvc.GetLatestByOriginAndUser(imports.OriginWakatime, user); latest == nil || err != nil {
		stream, importError = importer.ImportAll(user)
	} else {
		stream, importError = importer.Import(user, latest.Time.T(), time.Now())
	}
	if importError != nil {
		conf.Log().Request(r).Error("wakatime import for user failed", "userID", user.ID, "error", importError)
		return
	}

	_ = h.keyValueSrvc.PutString(&models.KeyStringValue{
		Key:   kvKeyLastImportSuccess,
		Value: time.Now().Format(time.RFC822),
	})

	count := 0
	batch := make([]*models.Heartbeat, 0, h.config.App.ImportBatchSize)
	insert := func(items []*models.Heartbeat) {
		if err := h.heartbeatSrvc.InsertBatch(items); err != nil {
			slog.Warn("failed to insert imported heartbeat, already existing?", "error", err)
		}
	}

	for heartbeat := range stream {
		count++
		batch = append(batch, heartbeat)
		if len(batch) == h.config.App.ImportBatchSize {
			insert(batch)
			batch = make([]*models.Heartbeat, 0, h.config.App.ImportBatchSize)
		}
	}
	if len(batch) > 0 {
		insert(batch)
	}

	countAfter, _ := h.heartbeatSrvc.CountByUser(user)
	slog.Info("downloaded heartbeats for user", "count", count, "userID", user.ID, "importedCount", countAfter-countBefore)

	if err := h.regenerateSummaries(user); err != nil {
		conf.Log().Request(r).Error("failed to regenerate summaries after wakatime import", "userID", user.ID, "error", err)
	}

	if !user.HasData {
		user.HasData = true
		if _, err := h.userSrvc.Update(user); err != nil {
			conf.Log().Request(r).Error("failed to set has_data flag after import", "userID", user.ID, "error", err)
		}
	}

	if user.Email != "" && h.mailSrvc != nil {
		if err := h.mailSrvc.SendImportNotification(user, time.Since(start), int(countAfter-countBefore)); err != nil {
			conf.Log().Request(r).Error("failed to send import notification mail", "userID", user.ID, "error", err)
		}
	}
}

func (h *SettingsApiHandler) regenerateSummaries(user *models.User) error {
	if err := h.summarySrvc.DeleteByUser(user.ID); err != nil {
		conf.Log().Error("failed to clear summaries", "error", err)
		return err
	}
	if err := h.aggregationSrvc.AggregateSummaries(datastructure.New(user.ID)); err != nil {
		conf.Log().Error("failed to regenerate summaries", "error", err)
		return err
	}
	return nil
}

func (h *SettingsApiHandler) regenerateSummariesAsync(user *models.User, r *http.Request) {
	go func(user *models.User, r *http.Request) {
		h.toggleAggregationLock(user.ID, true)
		defer h.toggleAggregationLock(user.ID, false)
		if err := h.regenerateSummaries(user); err != nil {
			conf.Log().Request(r).Error("failed to regenerate summaries", "userID", user.ID, "error", err)
		}
	}(user, r)
}

func (h *SettingsApiHandler) toggleAggregationLock(userID string, locked bool) {
	h.locksMu.Lock()
	defer h.locksMu.Unlock()
	h.aggregationLocks[userID] = locked
}

func (h *SettingsApiHandler) isAggregationLocked(userID string) bool {
	h.locksMu.Lock()
	defer h.locksMu.Unlock()
	return h.aggregationLocks[userID]
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}
