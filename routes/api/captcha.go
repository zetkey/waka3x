package api

import (
	"net/http"

	"github.com/dchest/captcha"
	"github.com/go-chi/chi/v5"
	conf "github.com/zetkey/waka3x/config"
	routeutils "github.com/zetkey/waka3x/routes/utils"
)

type CaptchaHandler struct {
	config *conf.Config
}

func NewCaptchaHandler() *CaptchaHandler {
	return &CaptchaHandler{
		config: conf.Get(),
	}
}

func (h *CaptchaHandler) RegisterRoutes(router chi.Router) {
	router.Get("/captcha/new", h.New)
	router.Get("/captcha/{id}.png", captcha.Server(captcha.StdWidth, captcha.StdHeight).ServeHTTP)
}

func (h *CaptchaHandler) New(w http.ResponseWriter, r *http.Request) {
	id := captcha.New()
	routeutils.RespondJSON(w, http.StatusOK, CaptchaResponse{
		ID:       id,
		ImageURL: h.config.Server.BasePath + "/api/captcha/" + id + ".png",
	})
}
