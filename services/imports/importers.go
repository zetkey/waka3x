package imports

import (
	"github.com/zetkey/waka3x/models"
	"time"
)

type DataImporter interface {
	Import(*models.User, time.Time, time.Time) (<-chan *models.Heartbeat, error)
	ImportAll(*models.User) (<-chan *models.Heartbeat, error)
}
