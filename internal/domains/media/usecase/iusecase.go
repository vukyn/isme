package usecase

import (
	"context"

	"github.com/vukyn/isme/internal/domains/media/models"
)

type IUseCase interface {
	// Upload proxies an avatar image to medioa and returns the stored URL.
	Upload(ctx context.Context, req models.UploadRequest) (models.UploadResponse, error)
}
