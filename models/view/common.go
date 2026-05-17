package view

import (
	conf "github.com/zetkey/waka3x/config"
	"github.com/zetkey/waka3x/models"
)

type BasicViewModel interface {
	SetError(string)
	SetSuccess(string)
}

type Messages struct {
	Success string
	Error   string
}

type SharedViewModel struct {
	Messages
	LeaderboardEnabled bool
	InvitesEnabled     bool
}

type SharedLoggedInViewModel struct {
	SharedViewModel
	User *models.User
}

func NewSharedViewModel(c *conf.Config, messages *Messages) SharedViewModel {
	vm := SharedViewModel{
		LeaderboardEnabled: c.App.LeaderboardEnabled,
		InvitesEnabled:     c.Security.InviteCodes,
	}
	if messages != nil {
		vm.Messages = *messages
	}
	return vm
}

func (m *Messages) SetError(message string) {
	m.Error = message
}

func (m *Messages) SetSuccess(message string) {
	m.Success = message
}

func (m SharedLoggedInViewModel) ApiKey() string {
	if m.User != nil {
		return m.User.ApiKey
	}
	return ""
}
