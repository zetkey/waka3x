package middlewares

import (
	"net/http"

	"github.com/zetkey/waka3x/models"
	routeutils "github.com/zetkey/waka3x/routes/utils"
)

func SetPrincipal(r *http.Request, user *models.User) {
	routeutils.SetPrincipal(r, user)
}

func GetPrincipal(r *http.Request) *models.User {
	return routeutils.GetPrincipal(r)
}
