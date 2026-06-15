package di

import (
	"github.com/vukyn/kuery/log"

	"github.com/sarulabs/di/v2"
)

func NewBuilder() *di.EnhancedBuilder {
	builder, err := di.NewEnhancedBuilder()
	if err != nil {
		log.New().Fatal("Failed to create builder", err)
	}

	defs := []*di.Def{
		defineConfig(),
		defineMedioaClient(),
		defineDB(),
		defineScheduler(),
		defineScheduleProvider(),
		defineCache(),
		defineMiddleware(),
	}
	defs = append(defs, defineRepository()...)
	defs = append(defs, defineUsecase()...)
	for _, def := range defs {
		if err := builder.Add(def); err != nil {
			log.New().Fatal("Failed to add definition to builder", err)
		}
	}
	return builder
}
