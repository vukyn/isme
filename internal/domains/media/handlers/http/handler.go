package handlers

import (
	"path/filepath"
	"strings"

	idi "github.com/vukyn/isme/internal/di"
	"github.com/vukyn/isme/internal/domains/media/models"

	pkgCtx "github.com/vukyn/kuery/ctx"
	pkgErr "github.com/vukyn/kuery/http/errors"
	pkgHttp "github.com/vukyn/kuery/http/fiber"

	"github.com/gofiber/fiber/v2"
)

// Upload proxies a multipart avatar upload (form field: file) to medioa using
// isme's server-side API key and returns { url, file_id }. The browser never
// sees the key — it only ever talks to this isme endpoint.
func Upload(c *fiber.Ctx) error {
	ctn := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
	defer ctn.Delete()

	uc, err := idi.GetMediaUsecase(ctn)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return pkgHttp.Err(c, pkgErr.InvalidRequest("file is required"))
	}

	file, err := fileHeader.Open()
	if err != nil {
		return pkgHttp.Err(c, pkgErr.InvalidRequest("failed to read uploaded file"))
	}
	defer file.Close()

	req := models.UploadRequest{
		File:        file,
		FileName:    fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
		Ext:         strings.TrimPrefix(filepath.Ext(fileHeader.Filename), "."),
		Size:        fileHeader.Size,
	}

	resp, err := uc.Upload(pkgCtx.NewContextFromFiberCtx(c), req)
	if err != nil {
		return pkgHttp.Err(c, err)
	}

	return pkgHttp.OK(c, resp)
}
