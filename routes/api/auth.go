package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/random"
	"github.com/go-chi/chi/v5"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	conf "github.com/zetkey/waka3x/config"
	"github.com/zetkey/waka3x/middlewares"
	"github.com/zetkey/waka3x/models"
	routeutils "github.com/zetkey/waka3x/routes/utils"
	"github.com/zetkey/waka3x/services"
	"github.com/zetkey/waka3x/utils"
)

type AuthApiHandler struct {
	config       *conf.Config
	userSrvc     services.IUserService
	mailSrvc     services.IMailService
	keyValueSrvc services.IKeyValueService
	webAuthnSrvc services.IWebAuthnService
}

func NewAuthApiHandler(userService services.IUserService, mailService services.IMailService, keyValueService services.IKeyValueService, webAuthnService services.IWebAuthnService) *AuthApiHandler {
	return &AuthApiHandler{
		config:       conf.Get(),
		userSrvc:     userService,
		mailSrvc:     mailService,
		keyValueSrvc: keyValueService,
		webAuthnSrvc: webAuthnService,
	}
}

func (h *AuthApiHandler) RegisterRoutes(router chi.Router) {
	router.Post("/login", h.Login)
	router.Post("/logout", h.Logout)
	router.Post("/signup", h.Signup)
	router.Post("/password/reset", h.ResetPassword)
	router.Post("/password/set", h.SetPassword)
	router.Get("/webauthn/options", h.GetWebAuthnOptions)
	router.Post("/webauthn/login", h.LoginWebAuthn)

	router.Group(func(r chi.Router) {
		r.Use(middlewares.NewAuthenticateMiddleware(h.userSrvc).Handler)
		r.Get("/users/current", h.GetCurrentUser)
		r.Put("/users/current", h.UpdateCurrentUser)
	})
}

func (h *AuthApiHandler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)

	var request UpdateCurrentUserRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	payload := request.toModel()

	if !payload.IsValid() {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "invalid user data")
		return
	}

	user.Email = payload.Email
	user.Location = payload.Location
	user.StartOfWeek = payload.StartOfWeek
	user.ReportsWeekly = payload.ReportsWeekly

	if _, err := h.userSrvc.Update(user); err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	routeutils.RespondJSON(w, http.StatusOK, newCurrentUserResponse(user))
}

func (h *AuthApiHandler) Login(w http.ResponseWriter, r *http.Request) {
	if h.config.Security.DisableLocalAuth {
		routeutils.RespondJSONError(w, http.StatusForbidden, "local authentication is disabled on this server")
		return
	}

	var login LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if login.Username == "" || login.Password == "" {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "missing credentials")
		return
	}

	user, err := h.userSrvc.GetUserById(login.Username)
	if err != nil {
		user, err = h.userSrvc.GetUserByEmail(login.Username)
		if err != nil {
			routeutils.RespondJSONError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
	}

	if !utils.ComparePassword(user.Password, login.Password, h.config.Security.PasswordSalt) {
		routeutils.RespondJSONError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Login successful
	encoded, err := h.config.Security.SecureCookie.Encode(models.AuthCookieKey, user.ID)
	if err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	user.LastLoggedInAt = models.CustomTime(time.Now())
	h.userSrvc.Update(user)

	http.SetCookie(w, h.config.CreateCookie(models.AuthCookieKey, encoded))
	routeutils.RespondJSON(w, http.StatusOK, AuthUserEnvelope{User: newCurrentUserResponse(user)})
}

func (h *AuthApiHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if user := middlewares.GetPrincipal(r); user != nil {
		h.userSrvc.FlushUserCache(user.ID)
	}
	routeutils.ClearSession(r, w)
	http.SetCookie(w, h.config.GetClearCookie(models.AuthCookieKey))
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthApiHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var signupRequest SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&signupRequest); err != nil {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	signup := signupRequest.toModel()

	if !h.config.IsDev() && !h.config.Security.AllowSignup && (!h.config.Security.InviteCodes || signup.InviteCode == "") {
		routeutils.RespondJSONError(w, http.StatusForbidden, "registration is disabled on this server")
		return
	}

	if h.config.Security.DisableLocalAuth {
		routeutils.RespondJSONError(w, http.StatusForbidden, "local authentication is disabled on this server")
		return
	}

	if err := h.consumeInvite(signup); err != nil {
		routeutils.RespondJSONError(w, http.StatusForbidden, err.Error())
		return
	}

	if !signup.IsValid() {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "invalid parameters")
		return
	}

	numUsers, _ := h.userSrvc.Count()
	user, created, err := h.userSrvc.CreateOrGet(signup, numUsers == 0)
	if err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	if !created {
		routeutils.RespondJSONError(w, http.StatusConflict, "user already exists")
		return
	}

	// Login automatically after signup
	encoded, _ := h.config.Security.SecureCookie.Encode(models.AuthCookieKey, user.ID)
	http.SetCookie(w, h.config.CreateCookie(models.AuthCookieKey, encoded))

	routeutils.RespondJSON(w, http.StatusCreated, AuthUserEnvelope{User: newCurrentUserResponse(user)})
}

func (h *AuthApiHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetPrincipal(r)
	routeutils.RespondJSON(w, http.StatusOK, newCurrentUserResponse(user))
}

func (h *AuthApiHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if !h.config.Mail.Enabled {
		routeutils.RespondJSONError(w, http.StatusNotImplemented, "mailing is disabled on this server")
		return
	}

	var request ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if user, err := h.userSrvc.GetUserByEmail(request.Email); user != nil && err == nil {
		if user.AuthType != "local" {
			routeutils.RespondJSONError(w, http.StatusBadRequest, "password reset is only available for local users")
			return
		}

		u, err := h.userSrvc.GenerateResetToken(user)
		if err != nil {
			routeutils.RespondJSONError(w, http.StatusInternalServerError, "failed to generate reset token")
			return
		}

		go func(user *models.User, r *http.Request) {
			link := fmt.Sprintf("%s/set-password?token=%s", h.config.Server.GetPublicUrl(), user.ResetToken)
			if err := h.mailSrvc.SendPasswordReset(user, link); err != nil {
				conf.Log().Request(r).Error("failed to send password reset mail", "userID", user.ID, "error", err)
			}
		}(u, r)
	}

	routeutils.RespondJSON(w, http.StatusAccepted, MessageResponse{Message: "an e-mail was sent if the address is registered"})
}

func (h *AuthApiHandler) SetPassword(w http.ResponseWriter, r *http.Request) {
	var request SetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	payload := models.SetPasswordRequest{
		Password:       request.Password,
		PasswordRepeat: request.PasswordRepeat,
		Token:          request.Token,
	}

	user, err := h.userSrvc.GetUserByResetToken(payload.Token)
	if err != nil {
		routeutils.RespondJSONError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	if !payload.IsValid() {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "invalid parameters")
		return
	}

	hash, err := utils.HashPassword(payload.Password, h.config.Security.PasswordSalt)
	if err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "failed to set new password")
		return
	}

	user.Password = hash
	user.ResetToken = ""
	if _, err := h.userSrvc.Update(user); err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "failed to save new password")
		return
	}

	routeutils.RespondJSON(w, http.StatusOK, MessageResponse{Message: "password updated successfully"})
}

func (h *AuthApiHandler) GetWebAuthnOptions(w http.ResponseWriter, r *http.Request) {
	if h.config.Security.DisableWebAuthn {
		routeutils.RespondJSONError(w, http.StatusForbidden, "webauthn is disabled on this server")
		return
	}

	options, sessionData, err := conf.WebAuthn.BeginDiscoverableLogin()
	if err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "failed to begin login")
		return
	}
	if routeutils.SetWebAuthnSession(sessionData, r, w) != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "failed to set session")
		return
	}
	routeutils.RespondJSON(w, http.StatusOK, options)
}

func (h *AuthApiHandler) LoginWebAuthn(w http.ResponseWriter, r *http.Request) {
	if h.config.Security.DisableWebAuthn {
		routeutils.RespondJSONError(w, http.StatusForbidden, "webauthn authentication is disabled on this server")
		return
	}

	var request WebAuthnLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.AssertionJSON == "" {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "missing assertion")
		return
	}

	sessionData, err := routeutils.GetWebAuthnSession(r)
	if err != nil {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "session expired")
		return
	}

	par, err := protocol.ParseCredentialRequestResponseBytes([]byte(request.AssertionJSON))
	if err != nil {
		routeutils.RespondJSONError(w, http.StatusBadRequest, "invalid assertion format")
		return
	}

	userHandler := func(rawID, userHandle []byte) (webauthn.User, error) {
		userHandleStr := string(userHandle)
		user, err := h.userSrvc.GetUserByWebAuthnID(userHandleStr)
		if err != nil || user == nil {
			return nil, fmt.Errorf("no user found for webauthn id: %s", userHandleStr)
		}
		if err := h.webAuthnSrvc.LoadCredentialIntoUser(user); err != nil {
			return nil, err
		}
		return user, nil
	}

	userInterface, credential, err := conf.WebAuthn.ValidatePasskeyLogin(userHandler, *sessionData, par)
	if err != nil {
		routeutils.RespondJSONError(w, http.StatusUnauthorized, "authentication failed")
		return
	}

	user := userInterface.(*models.User)
	if user.AuthType != "local" {
		routeutils.RespondJSONError(w, http.StatusUnauthorized, "non-local user cannot be authenticated with webauthn")
		return
	}
	if err := h.webAuthnSrvc.UpdateCredential(credential); err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "failed to update credential")
		return
	}

	h.finishUserLogin(user, w)
	routeutils.RespondJSON(w, http.StatusOK, AuthUserEnvelope{User: newCurrentUserResponse(user)})
}

func (h *AuthApiHandler) GetOidcLogin(w http.ResponseWriter, r *http.Request) {
	provider := h.getOidcProvider(w, r)
	if provider == nil {
		return
	}
	state := routeutils.SetNewOidcState(r, w)
	http.Redirect(w, r, provider.OAuth2.AuthCodeURL(state), http.StatusFound)
}

func (h *AuthApiHandler) GetOidcCallback(w http.ResponseWriter, r *http.Request) {
	provider := h.getOidcProvider(w, r)
	if provider == nil {
		return
	}

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	routeutils.ClearOidcIdTokenPayload(r, w)

	if savedState := routeutils.GetOidcState(r); state == "" || savedState != state {
		http.Redirect(w, r, fmt.Sprintf("%s/login?error=invalid_oidc_state", h.config.Server.BasePath), http.StatusFound)
		return
	}
	routeutils.ClearOidcState(r, w)

	authToken, err := provider.OAuth2.Exchange(conf.GetOidcContext(r.Context()), code)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s/login?error=oidc_exchange_failed", h.config.Server.BasePath), http.StatusFound)
		return
	}

	rawIDToken, ok := authToken.Extra("id_token").(string)
	if !ok {
		http.Redirect(w, r, fmt.Sprintf("%s/login?error=oidc_token_missing", h.config.Server.BasePath), http.StatusFound)
		return
	}

	idTokenPayload, err := routeutils.DecodeOidcIdToken(rawIDToken, provider, conf.GetOidcContext(r.Context()))
	if err != nil || idTokenPayload == nil {
		http.Redirect(w, r, fmt.Sprintf("%s/login?error=oidc_token_invalid", h.config.Server.BasePath), http.StatusFound)
		return
	}

	user, err := h.userSrvc.GetUserByOidc(provider.Name, idTokenPayload.Subject)
	if err != nil {
		if !h.config.IsDev() && !h.config.Security.OidcAllowSignup {
			http.Redirect(w, r, fmt.Sprintf("%s/login?error=registration_disabled", h.config.Server.BasePath), http.StatusFound)
			return
		}

		signup := models.SignupFromOidcIdToken(idTokenPayload)
		if !signup.IsValid() {
			http.Redirect(w, r, fmt.Sprintf("%s/login?error=invalid_oidc_signup", h.config.Server.BasePath), http.StatusFound)
			return
		}
		if newUsername := h.coalesceExistingUser(signup.Username); newUsername != signup.Username {
			signup.Username = newUsername
		}

		newUser, created, err := h.userSrvc.CreateOrGet(signup, false)
		if err != nil || !created {
			http.Redirect(w, r, fmt.Sprintf("%s/login?error=oidc_signup_failed", h.config.Server.BasePath), http.StatusFound)
			return
		}
		user = newUser
	}

	routeutils.SetOidcIdTokenPayload(idTokenPayload, r, w)
	h.finishUserLogin(user, w)
	http.Redirect(w, r, fmt.Sprintf("%s/dashboard", h.config.Server.BasePath), http.StatusFound)
}

func (h *AuthApiHandler) consumeInvite(signup *models.Signup) error {
	if signup.InviteCode == "" {
		return nil
	}

	inviteCodeKey := fmt.Sprintf("%s_%s", conf.KeyInviteCode, signup.InviteCode)
	kv, _ := h.keyValueSrvc.GetString(inviteCodeKey)
	if kv == nil || kv.Value == "" {
		return fmt.Errorf("invite code invalid or expired")
	}

	parts := strings.Split(kv.Value, ",")
	if len(parts) != 2 {
		return fmt.Errorf("invite code invalid or expired")
	}

	invitedDate, _ := time.Parse(time.RFC3339, parts[1])
	if time.Since(invitedDate) > 24*time.Hour {
		return fmt.Errorf("invite code invalid or expired")
	}

	signup.InvitedBy = parts[0]
	if err := h.keyValueSrvc.DeleteString(inviteCodeKey); err != nil {
		conf.Log().Error("failed to revoke invite code", "inviteCodeKey", inviteCodeKey, "error", err)
	}
	return nil
}

func (h *AuthApiHandler) finishUserLogin(user *models.User, w http.ResponseWriter) bool {
	encoded, err := h.config.Security.SecureCookie.Encode(models.AuthCookieKey, user.ID)
	if err != nil {
		routeutils.RespondJSONError(w, http.StatusInternalServerError, "failed to create session")
		return false
	}

	user.LastLoggedInAt = models.CustomTime(time.Now())
	h.userSrvc.Update(user)
	http.SetCookie(w, h.config.CreateCookie(models.AuthCookieKey, encoded))
	return true
}

func (h *AuthApiHandler) getOidcProvider(w http.ResponseWriter, r *http.Request) *conf.OidcProvider {
	providerName := chi.URLParam(r, "provider")
	provider, err := conf.GetOidcProvider(providerName)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("%s/login?error=oidc_provider_not_found", h.config.Server.BasePath), http.StatusFound)
		return nil
	}
	return provider
}

func (h *AuthApiHandler) coalesceExistingUser(username string) string {
	if u, _ := h.userSrvc.GetUserById(username); u != nil {
		return fmt.Sprintf("%s-%s", username, strings.ToLower(random.RandString(6)))
	}
	return username
}
