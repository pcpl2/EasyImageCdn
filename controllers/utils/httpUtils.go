package utils

import (
	"github.com/valyala/fasthttp"

	models "imageConverter.pcpl2lab.ovh/models"
)

func ValidateAuth(ctx *fasthttp.RequestCtx, config models.ApiConfig) bool {
	return string(ctx.Request.Header.Peek(config.APIKeyHeader)) == config.APIKey
}
