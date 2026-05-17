package mail

import "github.com/zetkey/waka3x/models"

type PasswordResetTplData struct {
	ResetLink string
}

type ImportNotificationTplData struct {
	PublicUrl     string
	Duration      string
	NumHeartbeats int
}

type WakatimeFailureNotificationNotificationTplData struct {
	PublicUrl   string
	NumFailures int
}

type ReportTplData struct {
	Report *models.Report
}

type SubscriptionNotificationTplData struct {
	PublicUrl           string
	HasExpired          bool
	DataRetentionMonths int
}
