package di

import (
	"github.com/vukyn/isme/internal/constants"

	"github.com/sarulabs/di/v2"
	pkgCache "github.com/vukyn/kuery/cache"
)

func defineCache() *di.Def {
	def := &di.Def{
		Name:  constants.CONTAINER_NAME_CACHE,
		Scope: di.App,
		Build: func(ctn di.Container) (any, error) {
			return pkgCache.NewCache[string, string](), nil
		},
		Close: func(obj any) error {
			cache := obj.(*pkgCache.Cache[string, string])
			cache.Close()
			return nil
		},
	}
	return def
}

func GetCache(ctn di.Container) *pkgCache.Cache[string, string] {
	return ctn.Get(constants.CONTAINER_NAME_CACHE).(*pkgCache.Cache[string, string])
}
