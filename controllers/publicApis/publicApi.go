package publicapis

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"

	biz "imageConverter.pcpl2lab.ovh/biz"
)

func GetImage(ctx *fiber.Ctx, id string, fileName string) {
	config, err := biz.GetConfig()
	if err != nil {
		ctx.SendStatus(fiber.StatusInternalServerError)
		log.Fatal(err)
	}
	acceptHeader := string(ctx.Get("Accept"))

	fileNameWithEx := fileName

	if strings.Contains(acceptHeader, "image/webp") {
		log.Printf("Send webp file.")
		fileNameWithEx = fmt.Sprintf("%s.webp", fileName)
	}
	filePath := fmt.Sprintf("%s/%s/%s", config.FilesPath, id, fileNameWithEx)
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		ctx.SendStatus(fiber.StatusNotFound)
		return
	}
	ctx.SendFile(filePath, true)

	//	httpUtils.SendFileHTTP(ctx, config, id, fileNameWithEx)
}
