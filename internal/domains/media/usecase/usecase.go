package usecase

import (
	"context"

	"github.com/vukyn/isme/internal/config"
	"github.com/vukyn/isme/internal/domains/media/exceptions"
	"github.com/vukyn/isme/internal/domains/media/models"

	pkgBase "github.com/vukyn/kuery/http/base"
	pkgErr "github.com/vukyn/kuery/http/errors"
	"github.com/vukyn/kuery/medioa"
)

// pathAvatars is the virtual destination folder in medioa for avatar images.
const pathAvatars = "avatars"

type usecase struct {
	cfg *config.Config
	// medioaClient may be nil when MEDIOA_API_KEY is unset — Upload guards on it.
	medioaClient *medioa.Client
}

func NewUsecase(cfg *config.Config, medioaClient *medioa.Client) IUseCase {
	return &usecase{cfg: cfg, medioaClient: medioaClient}
}

func (u *usecase) Upload(ctx context.Context, req models.UploadRequest) (models.UploadResponse, error) {
	if err := req.Validate(); err != nil {
		return models.UploadResponse{}, pkgErr.InvalidRequest(err.Error())
	}

	// A nil client means MEDIOA_API_KEY was not configured at boot. This is an
	// operator problem, not a caller one — surface it as an upstream failure.
	if u.medioaClient == nil {
		return models.UploadResponse{}, pkgErr.Forward(pkgBase.Response{
			Code:    502,
			Message: "media service is not configured",
		})
	}

	// Avatars are small images — single-shot upload under the avatars path.
	result, err := u.medioaClient.Upload(ctx, medioa.UploadInput{
		File:        req.File,
		FileName:    req.FileName,
		ContentType: req.ContentType,
		Ext:         req.Ext,
		Path:        pathAvatars,
	})
	if err != nil {
		return models.UploadResponse{}, exceptions.MapMediaError(err)
	}

	return models.UploadResponse{
		URL:      result.URL,
		FileID:   result.FileID,
		FileName: result.FileName,
		FileSize: result.FileSize,
	}, nil
}
