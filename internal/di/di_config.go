package di

import (
	"isme/internal/config"
	"isme/internal/constants"

	"fmt"

	"github.com/sarulabs/di/v2"
	"github.com/vukyn/kuery/log"
)

func defineConfig() *di.Def {
	def := &di.Def{
		Name:  constants.CONTAINER_NAME_CONFIG,
		Scope: di.App,
		Build: func(ctn di.Container) (any, error) {
			cfg, err := config.LoadConfig(".env")
			if err != nil {
				return nil, err
			}
			fmt.Println(">>>>  Config loaded")
			return cfg, nil
		},
		Close: func(obj any) error {
			log.New().Debug("Config destroyed")
			return nil
		},
	}
	return def
}

func GetConfig(ctn di.Container) *config.Config {
	return ctn.Get(constants.CONTAINER_NAME_CONFIG).(*config.Config)
}
