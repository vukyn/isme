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

	builder.Add(defineConfig())
	builder.Add(defineDB())
	builder.Add(defineMiddleware())
	for _, def := range defineRepository() {
		builder.Add(def)
	}
	for _, def := range defineUsecase() {
		builder.Add(def)
	}
	return builder
}
