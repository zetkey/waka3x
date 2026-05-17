package routes

import (
	"github.com/zetkey/waka3x/config"
)

func defaultErrorRedirectTarget() string {
	return config.Get().Server.BasePath + "/"
}
