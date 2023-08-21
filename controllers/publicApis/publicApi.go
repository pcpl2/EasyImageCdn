package publicapis

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"

	biz "easy-image-cdn.pcpl2lab.ovh/biz"
	appLogger "easy-image-cdn.pcpl2lab.ovh/utils/logger"
)

func GetImage(ctx *fiber.Ctx, id string, fileName string) error {
	config, err := biz.GetConfig()
	if err != nil {
		appLogger.ErrorLogger.Print(err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	acceptHeader := ctx.Get("Accept")

	fileNameWithEx := fileName

	if strings.Contains(acceptHeader, "image/webp") {
		fileNameWithEx = fmt.Sprintf("%s.webp", fileName)
	}
	filePath := fmt.Sprintf("%s/%s/%s", config.FilesPath, id, fileNameWithEx)
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		return ctx.SendStatus(fiber.StatusNotFound)
	}
	return ctx.SendFile(filePath, true)
}
