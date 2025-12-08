package di

import (
	"isme/cache"
	"isme/internal/constants"

	"github.com/sarulabs/di/v2"
)

func defineCache() *di.Def {
	def := &di.Def{
		Name:  constants.CONTAINER_NAME_CACHE,
		Scope: di.App,
		Build: func(ctn di.Container) (any, error) {
			return cache.NewCache(), nil
		},
		Close: func(obj any) error {
			cache := obj.(*cache.Cache)
			cache.Close()
			return nil
		},
	}
	return def
}

func GetCache(ctn di.Container) *cache.Cache {
	return ctn.Get(constants.CONTAINER_NAME_CACHE).(*cache.Cache)
}
