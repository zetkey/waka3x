package mail

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/datetime"
	conf "github.com/zetkey/waka3x/config"
	"github.com/zetkey/waka3x/helpers"
	"github.com/zetkey/waka3x/models"
	"github.com/zetkey/waka3x/services"
	"github.com/zetkey/waka3x/services/mail/templates"
	"github.com/zetkey/waka3x/utils"
)

const (
	tplNamePasswordReset               = "reset_password"
	tplNameImportNotification          = "import_finished"
	tplNameWakatimeFailureNotification = "wakatime_connection_failure"
	tplNameReport                      = "report"
	tplNameSubscriptionNotification    = "subscription_expiring"
	subjectPasswordReset               = "Waka3x - Password Reset"
	subjectImportNotification          = "Waka3x - Data Import Finished"
	subjectWakatimeFailureNotification = "Waka3x - WakaTime Connection Failure"
	subjectReport                      = "Waka3x - Report from %s"
	subjectSubscriptionNotification    = "Waka3x - Subscription expiring / expired"
)

type SendingService interface {
	Send(*models.Mail) error
}

type MailService struct {
	config         *conf.Config
	sendingService SendingService
	templates      utils.TemplateMap
}

func NewMailService() (services.IMailService, error) {
	config := conf.Get()

	var sendingService SendingService
	sendingService = &NoopSendingService{}

	if config.Mail.Enabled {
		if config.Mail.Provider == conf.MailProviderSmtp {
			sendingService = NewSMTPSendingService(config.Mail.Smtp)
		}
	}

	// Use local file system when in 'dev' environment, go embed file system otherwise
	templateFS := conf.ChooseFS("services/mail/templates", templates.TemplateFiles)
	loadedTemplates, err := utils.LoadTemplates(templateFS, defaultTemplateFuncs())
	if err != nil {
		return nil, fmt.Errorf("failed to load email templates: %w", err)
	}

	return &MailService{sendingService: sendingService, config: config, templates: loadedTemplates}, nil
}

func (m *MailService) SendPasswordReset(recipient *models.User, resetLink string) error {
	tpl, err := m.getPasswordResetTemplate(PasswordResetTplData{ResetLink: resetLink})
	if err != nil {
		return err
	}
	mail := &models.Mail{
		From:    models.MailAddress(m.config.Mail.Sender),
		To:      models.MailAddresses([]models.MailAddress{models.MailAddress(recipient.Email)}),
		Subject: subjectPasswordReset,
	}
	mail.WithHTML(tpl.String())
	return m.sendingService.Send(mail)
}

func (m *MailService) SendWakatimeFailureNotification(recipient *models.User, numFailures int) error {
	tpl, err := m.getWakatimeFailureNotificationTemplate(WakatimeFailureNotificationNotificationTplData{
		PublicUrl:   m.config.Server.PublicUrl,
		NumFailures: numFailures,
	})
	if err != nil {
		return err
	}
	mail := &models.Mail{
		From:    models.MailAddress(m.config.Mail.Sender),
		To:      models.MailAddresses([]models.MailAddress{models.MailAddress(recipient.Email)}),
		Subject: subjectWakatimeFailureNotification,
	}
	mail.WithHTML(tpl.String())
	return m.sendingService.Send(mail)
}

func (m *MailService) SendImportNotification(recipient *models.User, duration time.Duration, numHeartbeats int) error {
	tpl, err := m.getImportNotificationTemplate(ImportNotificationTplData{
		PublicUrl:     m.config.Server.PublicUrl,
		Duration:      fmt.Sprintf("%.0f seconds", duration.Seconds()),
		NumHeartbeats: numHeartbeats,
	})
	if err != nil {
		return err
	}
	mail := &models.Mail{
		From:    models.MailAddress(m.config.Mail.Sender),
		To:      models.MailAddresses([]models.MailAddress{models.MailAddress(recipient.Email)}),
		Subject: subjectImportNotification,
	}
	mail.WithHTML(tpl.String())
	return m.sendingService.Send(mail)
}

func (m *MailService) SendReport(recipient *models.User, report *models.Report) error {
	tpl, err := m.getReportTemplate(ReportTplData{report})
	if err != nil {
		return err
	}
	mail := &models.Mail{
		From:            models.MailAddress(m.config.Mail.Sender),
		To:              models.MailAddresses([]models.MailAddress{models.MailAddress(recipient.Email)}),
		Subject:         fmt.Sprintf(subjectReport, helpers.FormatDateHuman(time.Now().In(recipient.TZ()))),
		LinkUnsubscribe: recipient.UnsubscribeLink(),
	}
	mail.WithHTML(tpl.String())
	return m.sendingService.Send(mail)
}

func (m *MailService) SendSubscriptionNotification(recipient *models.User, hasExpired bool) error {
	tpl, err := m.getSubscriptionNotificationTemplate(SubscriptionNotificationTplData{
		PublicUrl:           m.config.Server.PublicUrl,
		DataRetentionMonths: m.config.App.DataRetentionMonths,
		HasExpired:          hasExpired,
	})
	if err != nil {
		return err
	}
	mail := &models.Mail{
		From:    models.MailAddress(m.config.Mail.Sender),
		To:      models.MailAddresses([]models.MailAddress{models.MailAddress(recipient.Email)}),
		Subject: subjectSubscriptionNotification,
	}
	mail.WithHTML(tpl.String())
	return m.sendingService.Send(mail)
}

func (m *MailService) getPasswordResetTemplate(data PasswordResetTplData) (*bytes.Buffer, error) {
	var rendered bytes.Buffer
	if err := m.templates[m.fmtName(tplNamePasswordReset)].Execute(&rendered, data); err != nil {
		return nil, err
	}
	return &rendered, nil
}

func (m *MailService) getWakatimeFailureNotificationTemplate(data WakatimeFailureNotificationNotificationTplData) (*bytes.Buffer, error) {
	var rendered bytes.Buffer
	if err := m.templates[m.fmtName(tplNameWakatimeFailureNotification)].Execute(&rendered, data); err != nil {
		return nil, err
	}
	return &rendered, nil
}

func (m *MailService) getImportNotificationTemplate(data ImportNotificationTplData) (*bytes.Buffer, error) {
	var rendered bytes.Buffer
	if err := m.templates[m.fmtName(tplNameImportNotification)].Execute(&rendered, data); err != nil {
		return nil, err
	}
	return &rendered, nil
}

func (m *MailService) getReportTemplate(data ReportTplData) (*bytes.Buffer, error) {
	var rendered bytes.Buffer
	if err := m.templates[m.fmtName(tplNameReport)].Execute(&rendered, data); err != nil {
		return nil, err
	}
	return &rendered, nil
}

func (m *MailService) getSubscriptionNotificationTemplate(data SubscriptionNotificationTplData) (*bytes.Buffer, error) {
	var rendered bytes.Buffer
	if err := m.templates[m.fmtName(tplNameSubscriptionNotification)].Execute(&rendered, data); err != nil {
		return nil, err
	}
	return &rendered, nil
}

func (m *MailService) fmtName(name string) string {
	return fmt.Sprintf("%s.tpl.html", name)
}

func defaultTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"json":           utils.Json,
		"date":           helpers.FormatDateHuman,
		"datetime":       helpers.FormatDateTimeHuman,
		"datetimetz":     helpers.FormatDateTimeHumanTZ,
		"simpledate":     helpers.FormatDate,
		"simpledatetime": helpers.FormatDateTime,
		"duration":       helpers.FmtWakatimeDuration,
		"floordate":      datetime.BeginOfDay,
		"ceildate":       utils.CeilDate,
		"title":          strings.Title,
		"join":           strings.Join,
		"lower":          strings.ToLower,
		"htmlSafe": func(html string) template.HTML {
			return template.HTML(html)
		},
		"urlSafe": func(s string) template.URL {
			return template.URL(s)
		},
		"cssSafe": func(s string) template.CSS {
			return template.CSS(s)
		},
	}
}
