package migrations

import (
	"github.com/zetkey/waka3x/config"
	"gorm.io/gorm"
)

// see https://github.com/zetkey/waka3x/issues/817#issuecomment-3146365708

func init() {
	const name = "20250802_fix_default_coding_category"
	f := migrationFunc{
		name:       name,
		background: true,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			if err := db.Exec("update heartbeats set category = 'coding' where category = '' and type = 'file' and language != ''").Error; err != nil {
				return err
			}

			setHasRun(name, db)
			return nil
		},
	}

	registerPostMigration(f)
}
