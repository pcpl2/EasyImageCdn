package utils

import (
	"github.com/gofiber/fiber/v2"

	models "easy-image-cdn.pcpl2lab.ovh/models"
)

func ValidateAuth(ctx *fiber.Ctx, config models.ApiConfig) bool {
	return ctx.Get(config.APIKeyHeader) == config.APIKey
}
