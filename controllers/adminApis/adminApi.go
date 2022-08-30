package adminapis

import (
	"encoding/base64"

	"log"
	"net/url"
	"os"

	"github.com/gofiber/fiber/v2"

	httpUtils "imageConverter.pcpl2lab.ovh/controllers/utils"

	biz "imageConverter.pcpl2lab.ovh/biz"
	ic "imageConverter.pcpl2lab.ovh/imageConverter"
	models "imageConverter.pcpl2lab.ovh/models"
)

func PostNewImage(ctx *fiber.Ctx) {
	config, err := biz.GetConfig()
	if err != nil {
		ctx.SendStatus(fiber.StatusUnauthorized)
		log.Fatal(err)
	}

	if !ctx.Is("json") {
		ctx.SendStatus(fiber.StatusBadRequest)
		log.Print("*ERROR* invalid content type")
		return
	}

	if !httpUtils.ValidateAuth(ctx, config) {
		ctx.SendStatus(fiber.StatusUnauthorized)
		log.Print("*ERROR* Auth error")
		return
	}

	var Payload models.ImagePayload

	if err := ctx.BodyParser(&Payload); err != nil {
		log.Print("*ERROR* " + err.Error())
		ctx.SendStatus(fiber.StatusBadRequest)
		return
	}

	imageFolderPath := config.FilesPath + "/" + url.PathEscape(Payload.ID)

	if err := createFileFolder(config, imageFolderPath, ctx); err != nil {
		return
	}

	dec, err := base64.StdEncoding.DecodeString(Payload.Image)
	if err != nil {
		ctx.SendStatus(fiber.StatusNoContent)
		log.Print("*ERROR* Cannot read file from payload " + err.Error())
		return
	}

	sourceFilename := "source"
	sourcePath := imageFolderPath + "/" + sourceFilename

	if err := saveFile(config, sourcePath, dec, ctx); err != nil {
		return
	}

	queueList := createConvertCommands(config, imageFolderPath)

	ic.ConvertImage(sourcePath, queueList)
}

func createFileFolder(config models.ApiConfig, imageFolderPath string, ctx *fiber.Ctx) error {
	if _, err := os.Stat(imageFolderPath); os.IsNotExist(err) {
		errMkDir := os.Mkdir(imageFolderPath, 0755)
		if errMkDir != nil {
			ctx.SendStatus(fiber.StatusNoContent)
			log.Print("*ERROR* Failed to create folder " + errMkDir.Error())
			return errMkDir
		}
	}
	return nil
}

func saveFile(config models.ApiConfig, sourcePath string, file []byte, ctx *fiber.Ctx) error {
	f, err := os.OpenFile(sourcePath, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		ctx.SendStatus(fiber.StatusNoContent)
		log.Print("*ERROR* Cannot open file " + err.Error())
		return err
	}
	defer f.Close()

	if _, err := f.Write(file); err != nil {
		ctx.SendStatus(fiber.StatusNoContent)
		log.Print("*ERROR* Cannot write file " + err.Error())
		return err
	}

	if err := f.Sync(); err != nil {
		ctx.SendStatus(fiber.StatusNoContent)
		log.Print("*ERROR* Cannot sync file " + err.Error())
		return err
	}

	return nil
}

func createConvertCommands(config models.ApiConfig, imageFolderPath string) []models.ConvertCommand {
	queueList := []models.ConvertCommand{}
	queueList = append(queueList, models.ConvertCommand{
		Path:       imageFolderPath + "/",
		WebP:       true,
		ConvertRes: false,
	})

	for _, element := range config.Resolutions {
		queueList = append(queueList, models.ConvertCommand{
			Path:       imageFolderPath + "/",
			WebP:       true,
			ConvertRes: true,
			TargetRes:  element,
		})

		queueList = append(queueList, models.ConvertCommand{
			Path:       imageFolderPath + "/",
			WebP:       false,
			ConvertRes: true,
			TargetRes:  element,
		})
	}
	return queueList
}
