package di

import (
	"github.com/sarulabs/di/v2"
	"github.com/uptrace/bun"
	kueryDb "github.com/vukyn/kuery/bun/db"
	pkgBunHooks "github.com/vukyn/kuery/bun/hooks"
	"github.com/vukyn/kuery/log"

	"github.com/vukyn/isme/internal/constants"
)

func defineDB() *di.Def {
	def := &di.Def{
		Name:  constants.CONTAINER_NAME_DB,
		Scope: di.App,
		Build: func(ctn di.Container) (any, error) {
			cfg := GetConfig(ctn)

			// Build the dialect-aware connection config from cfg.DB. SQLite stays
			// the default; Postgres is selected via DB_DRIVER=postgres. The pragma
			// is left empty so kuery applies the shared default isme pragma.
			db, err := kueryDb.Open(kueryDb.Config{
				Driver:      kueryDb.Driver(cfg.DB.Driver),
				SQLitePath:  cfg.DB.SQLitePath,
				PostgresDSN: cfg.DB.DSN,
				Host:        cfg.DB.Host,
				Port:        cfg.DB.Port,
				User:        cfg.DB.User,
				Password:    cfg.DB.Password,
				DBName:      cfg.DB.DBName,
				SSLMode:     cfg.DB.SSLMode,
			})
			if err != nil {
				return nil, err
			}

			driver := cfg.DB.Driver
			if driver == "" {
				driver = string(kueryDb.DriverSQLite)
			}
			log.New().Infof("Database initialized with driver %q", driver)

			db.AddQueryHook(pkgBunHooks.NewQueryHook(log.New()))
			return db, nil
		},
		Close: func(obj any) error {
			db := obj.(*bun.DB)
			log.New().Debug("Database closed")
			return db.Close()
		},
	}
	return def
}

func GetDB(ctn di.Container) *bun.DB {
	return ctn.Get(constants.CONTAINER_NAME_DB).(*bun.DB)
}
