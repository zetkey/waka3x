package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/securecookie"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/zetkey/waka3x/config"
	"github.com/zetkey/waka3x/mocks"
	"github.com/zetkey/waka3x/models"
	"github.com/zetkey/waka3x/utils"
)

func TestAuthApiHandler_LoginUpgradesPasswordHashCreatedWithoutSalt(t *testing.T) {
	const (
		username = "alice"
		password = "correct-password"
		newSalt  = "new-production-password-salt"
	)

	legacyHash, err := utils.HashPassword(password, "")
	require.NoError(t, err)

	cfg := config.Empty()
	cfg.Security.PasswordSalt = newSalt
	cfg.Security.CookieMaxAgeSec = 172800
	cfg.Security.InsecureCookies = true
	cfg.Security.SecureCookie = securecookie.New(
		bytes.Repeat([]byte("h"), 64),
		bytes.Repeat([]byte("b"), 32),
	)
	config.Set(cfg)

	user := &models.User{ID: username, Password: legacyHash, AuthType: "local"}
	userService := new(mocks.UserServiceMock)
	userService.On("GetUserById", username).Return(user, nil)
	userService.On("Update", mock.MatchedBy(func(updated *models.User) bool {
		return updated.ID == username &&
			updated.Password != legacyHash &&
			utils.ComparePassword(updated.Password, password, newSalt)
	})).Return(user, nil)

	handler := NewAuthApiHandler(userService, nil, nil, nil)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/login",
		strings.NewReader(`{"username":"alice","password":"correct-password"}`),
	)
	rec := httptest.NewRecorder()

	handler.Login(rec, req)
	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.NotEmpty(t, res.Cookies())
	userService.AssertExpectations(t)
}
