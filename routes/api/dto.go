package api

import (
	"time"

	conf "github.com/zetkey/waka3x/config"
	"github.com/zetkey/waka3x/models"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SignupRequest struct {
	Username       string `json:"username"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	PasswordRepeat string `json:"password_repeat"`
	Location       string `json:"location"`
	CaptchaID      string `json:"captcha_id"`
	Captcha        string `json:"captcha"`
	InviteCode     string `json:"invite_code"`
}

func (r SignupRequest) toModel() *models.Signup {
	passwordRepeat := r.PasswordRepeat
	if passwordRepeat == "" {
		passwordRepeat = r.Password
	}

	return &models.Signup{
		Username:       r.Username,
		Email:          r.Email,
		Password:       r.Password,
		PasswordRepeat: passwordRepeat,
		Location:       r.Location,
		CaptchaId:      r.CaptchaID,
		Captcha:        r.Captcha,
		InviteCode:     r.InviteCode,
	}
}

type UpdateCurrentUserRequest struct {
	Email             string `json:"email"`
	Location          string `json:"location"`
	StartOfWeek       int    `json:"start_of_week"`
	ReportsWeekly     bool   `json:"reports_weekly"`
	PublicLeaderboard bool   `json:"public_leaderboard"`
}

func (r UpdateCurrentUserRequest) toModel() models.UserDataUpdate {
	return models.UserDataUpdate{
		Email:             r.Email,
		Location:          r.Location,
		StartOfWeek:       r.StartOfWeek,
		ReportsWeekly:     r.ReportsWeekly,
		PublicLeaderboard: r.PublicLeaderboard,
	}
}

type CurrentUserResponse struct {
	ID                     string    `json:"id"`
	APIKey                 string    `json:"api_key"`
	Email                  string    `json:"email"`
	Location               string    `json:"location"`
	StartOfWeek            int       `json:"start_of_week"`
	CreatedAt              time.Time `json:"created_at"`
	LastLoggedInAt         time.Time `json:"last_logged_in_at"`
	AuthType               string    `json:"auth_type"`
	HasData                bool      `json:"has_data"`
	ReportsWeekly          bool      `json:"reports_weekly"`
	PublicLeaderboard      bool      `json:"public_leaderboard"`
	ExcludeUnknownProjects bool      `json:"exclude_unknown_projects"`
	HeartbeatsTimeoutMin   int       `json:"heartbeats_timeout_min"`
	ShareDataMaxDays       int       `json:"share_data_max_days"`
	ShareEditors           bool      `json:"share_editors"`
	ShareLanguages         bool      `json:"share_languages"`
	ShareProjects          bool      `json:"share_projects"`
	ShareOSs               bool      `json:"share_oss"`
	ShareMachines          bool      `json:"share_machines"`
	ShareLabels            bool      `json:"share_labels"`
	ShareActivityChart     bool      `json:"share_activity_chart"`
	HasActiveSubscription  bool      `json:"has_active_subscription"`
	AvatarURL              string    `json:"avatar_url"`
	WakatimeConnected      bool      `json:"wakatime_connected"`
	WakatimeAPIURL         string    `json:"wakatime_api_url"`
	ReadmeStatsBaseURL     string    `json:"readme_stats_base_url"`
}

func newCurrentUserResponse(user *models.User) *CurrentUserResponse {
	if user == nil {
		return nil
	}

	cfg := conf.Get()
	return &CurrentUserResponse{
		ID:                     user.ID,
		APIKey:                 user.ApiKey,
		Email:                  user.Email,
		Location:               user.Location,
		StartOfWeek:            user.StartOfWeek,
		CreatedAt:              user.CreatedAt.T(),
		LastLoggedInAt:         user.LastLoggedInAt.T(),
		AuthType:               user.AuthType,
		HasData:                user.HasData,
		ReportsWeekly:          user.ReportsWeekly,
		PublicLeaderboard:      user.PublicLeaderboard,
		ExcludeUnknownProjects: user.ExcludeUnknownProjects,
		HeartbeatsTimeoutMin:   user.HeartbeatsTimeoutMin(),
		ShareDataMaxDays:       user.ShareDataMaxDays,
		ShareEditors:           user.ShareEditors,
		ShareLanguages:         user.ShareLanguages,
		ShareProjects:          user.ShareProjects,
		ShareOSs:               user.ShareOSs,
		ShareMachines:          user.ShareMachines,
		ShareLabels:            user.ShareLabels,
		ShareActivityChart:     user.ShareActivityChart,
		HasActiveSubscription:  user.HasActiveSubscription(),
		AvatarURL:              user.AvatarURL(cfg.App.AvatarURLTemplate),
		WakatimeConnected:      user.WakatimeApiKey != "",
		WakatimeAPIURL:         user.WakaTimeURL(""),
		ReadmeStatsBaseURL:     user.ReadmeStatsBaseUrl,
	}
}

type AuthUserEnvelope struct {
	User *CurrentUserResponse `json:"user"`
}

type ProjectStatResponse struct {
	Project         string    `json:"project"`
	TotalHeartbeats int64     `json:"total_heartbeats"`
	TopLanguage     string    `json:"top_language"`
	FirstHeartbeat  time.Time `json:"first_heartbeat"`
	LastHeartbeat   time.Time `json:"last_heartbeat"`
}

func newProjectStatResponse(project *models.ProjectStats) ProjectStatResponse {
	return ProjectStatResponse{
		Project:         project.Project,
		TotalHeartbeats: project.Count,
		TopLanguage:     project.TopLanguage,
		FirstHeartbeat:  project.First.T(),
		LastHeartbeat:   project.Last.T(),
	}
}

type PublicUserResponse struct {
	ID                    string `json:"id"`
	AvatarURL             string `json:"avatar_url"`
	HasActiveSubscription bool   `json:"has_active_subscription"`
}

func newPublicUserResponse(user *models.User) *PublicUserResponse {
	if user == nil {
		return nil
	}

	return &PublicUserResponse{
		ID:                    user.ID,
		AvatarURL:             user.AvatarURL(conf.Get().App.AvatarURLTemplate),
		HasActiveSubscription: user.HasActiveSubscription(),
	}
}

type LeaderboardItemResponse struct {
	Rank         uint                `json:"rank"`
	UserID       string              `json:"user_id"`
	Interval     string              `json:"interval"`
	AggregatedBy *uint8              `json:"aggregated_by,omitempty"`
	Key          *string             `json:"key,omitempty"`
	Total        int64               `json:"total"`
	UpdatedAt    time.Time           `json:"updated_at"`
	User         *PublicUserResponse `json:"user,omitempty"`
}

func newLeaderboardItemResponse(item *models.LeaderboardItemRanked) LeaderboardItemResponse {
	return LeaderboardItemResponse{
		Rank:         item.Rank,
		UserID:       item.UserID,
		Interval:     item.Interval,
		AggregatedBy: item.By,
		Key:          item.Key,
		Total:        int64(item.Total / time.Second),
		UpdatedAt:    item.CreatedAt.T(),
		User:         newPublicUserResponse(item.User),
	}
}

type LeaderboardResponse struct {
	Items         []LeaderboardItemResponse `json:"items"`
	By            string                    `json:"by"`
	Key           string                    `json:"key"`
	TopKeys       []string                  `json:"top_keys"`
	UserLanguages map[string][]string       `json:"user_languages"`
	IntervalLabel string                    `json:"interval_label"`
	LastUpdated   time.Time                 `json:"last_updated"`
}

type AvailableFiltersResponse struct {
	Projects   []string `json:"projects"`
	Languages  []string `json:"languages"`
	Machines   []string `json:"machines"`
	Labels     []string `json:"labels"`
	Categories []string `json:"categories"`
}

type TimelineItemResponse struct {
	Name     string `json:"name"`
	Duration int64  `json:"duration"`
}

type TimelineResponse struct {
	Date     time.Time              `json:"date"`
	Projects []TimelineItemResponse `json:"projects"`
}

type HourlyBreakdownItemResponse struct {
	FromTime time.Time `json:"from_time"`
	Duration int64     `json:"duration"`
	Entity   string    `json:"entity"`
}

type HourlyBreakdownProjectResponse struct {
	Project string                        `json:"project"`
	Items   []HourlyBreakdownItemResponse `json:"items"`
}

type HourlyActivityResponse struct {
	Hour     int   `json:"hour"`
	Duration int64 `json:"duration"`
}

type SummaryDetailsResponse struct {
	Summary             *models.Summary                  `json:"summary"`
	AvailableFilters    AvailableFiltersResponse         `json:"available_filters"`
	EditorColors        map[string]string                `json:"editor_colors"`
	LanguageColors      map[string]string                `json:"language_colors"`
	OSColors            map[string]string                `json:"os_colors"`
	AICodingRatio       float64                          `json:"ai_coding_ratio"`
	Timeline            []TimelineResponse               `json:"timeline"`
	HourlyBreakdown     []HourlyBreakdownProjectResponse `json:"hourly_breakdown"`
	HourlyBreakdownFrom time.Time                        `json:"hourly_breakdown_from"`
	HourlyActivity      []HourlyActivityResponse         `json:"hourly_activity"`
	UserFirstData       time.Time                        `json:"user_first_data"`
	DataRetentionMonths int                              `json:"data_retention_months"`
	UserDataExpiring    bool                             `json:"user_data_expiring"`
	ProjectDetails      bool                             `json:"project_details"`
	Project             string                           `json:"project"`
}

type OidcProviderResponse struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	LoginURL    string `json:"login_url"`
}

type BootstrapResponse struct {
	Version                string                 `json:"version"`
	BasePath               string                 `json:"base_path"`
	PublicURL              string                 `json:"public_url"`
	DBType                 string                 `json:"db_type"`
	LeaderboardEnabled     bool                   `json:"leaderboard_enabled"`
	LeaderboardRequireAuth bool                   `json:"leaderboard_require_auth"`
	AllowSignup            bool                   `json:"allow_signup"`
	InviteCodesEnabled     bool                   `json:"invite_codes_enabled"`
	SignupCaptcha          bool                   `json:"signup_captcha"`
	DisableLocalAuth       bool                   `json:"disable_local_auth"`
	DisableWebAuthn        bool                   `json:"disable_webauthn"`
	SubscriptionsEnabled   bool                   `json:"subscriptions_enabled"`
	SubscriptionPrice      string                 `json:"subscription_price"`
	StripeAPIKey           string                 `json:"stripe_api_key"`
	SupportContact         string                 `json:"support_contact"`
	DataRetentionMonths    int                    `json:"data_retention_months"`
	DefaultWakatimeURL     string                 `json:"default_wakatime_url"`
	AvatarURLTemplate      string                 `json:"avatar_url_template"`
	OIDCProviders          []OidcProviderResponse `json:"oidc_providers"`
	MailEnabled            bool                   `json:"mail_enabled"`
	ImportEnabled          bool                   `json:"import_enabled"`
	ImportBackoffMin       int                    `json:"import_backoff_min"`
	ImportMaxRateHours     int                    `json:"import_max_rate_hours"`
}

type SetupResponse struct {
	APIKey        string `json:"api_key"`
	BaseURL       string `json:"base_url"`
	PublicURL     string `json:"public_url"`
	Username      string `json:"username,omitempty"`
	Authenticated bool   `json:"authenticated"`
}

type CaptchaResponse struct {
	ID       string `json:"id"`
	ImageURL string `json:"image_url"`
}

type HomeStatsResponse struct {
	TotalHours      int         `json:"total_hours"`
	TotalUsers      int         `json:"total_users"`
	CurrentlyOnline int         `json:"currently_online"`
	Newsbox         interface{} `json:"newsbox,omitempty"`
}

type ImprintResponse struct {
	HTML string `json:"html"`
}

type MessageResponse struct {
	Message string               `json:"message"`
	User    *CurrentUserResponse `json:"user,omitempty"`
}

type ResetPasswordRequest struct {
	Email string `json:"email"`
}

type SetPasswordRequest struct {
	Password       string `json:"password"`
	PasswordRepeat string `json:"password_repeat"`
	Token          string `json:"token"`
}

type WebAuthnLoginRequest struct {
	AssertionJSON string `json:"assertion_json"`
}
