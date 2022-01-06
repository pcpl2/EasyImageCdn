package utils

import (
	"errors"
	"fmt"
	"os"

	"github.com/valyala/fasthttp"

	models "imageConverter.pcpl2lab.ovh/models"
)

func ValidateAuth(ctx *fasthttp.RequestCtx, config models.ApiConfig) bool {
	return string(ctx.Request.Header.Peek(config.APIKeyHeader)) == config.APIKey
}

func SendFileHTTP(ctx *fasthttp.RequestCtx, config models.ApiConfig, id string, fileName string) {
	filePath := fmt.Sprintf("%s/%s/%s", config.FilesPath, id, fileName)
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		ctx.Error("", fasthttp.StatusInternalServerError)
	}

	ctx.SendFile(filePath)
}
