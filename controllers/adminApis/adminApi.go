package adminapis

import (
	"encoding/base64"
	"io"

	"net/url"
	"os"

	"github.com/gofiber/fiber/v2"

	appLogger "imageConverter.pcpl2lab.ovh/utils/logger"

	httpUtils "imageConverter.pcpl2lab.ovh/controllers/utils"

	biz "imageConverter.pcpl2lab.ovh/biz"
	ic "imageConverter.pcpl2lab.ovh/imageConverter"
	models "imageConverter.pcpl2lab.ovh/models"
)

func PostNewImage(ctx *fiber.Ctx) error {
	config, err := biz.GetConfig()
	if err != nil {
		appLogger.ErrorLogger.Println(err)
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	if !ctx.Is("json") {
		appLogger.WarningLogger.Print("Invalid content type")
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	if !httpUtils.ValidateAuth(ctx, config) {
		appLogger.WarningLogger.Print("Auth error")
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	var Payload models.ImagePayload

	if err := ctx.BodyParser(&Payload); err != nil {
		appLogger.WarningLogger.Print(err.Error())
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	imageFolderPath := config.FilesPath + "/" + url.PathEscape(Payload.ID)

	if err := createFileFolder(imageFolderPath); err != nil {
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	dec, err := base64.StdEncoding.DecodeString(Payload.Image)
	if err != nil {
		appLogger.ErrorLogger.Print("Cannot read file from payload " + err.Error())
		return ctx.SendStatus(fiber.StatusNoContent)
	}

	sourceFilename := "source"
	sourcePath := imageFolderPath + "/" + sourceFilename

	if err := saveFile(sourcePath, dec); err != nil {
		return ctx.SendStatus(fiber.StatusNoContent)
	}

	queueList := createConvertCommands(config, imageFolderPath)

	ic.ConvertImage(sourcePath, queueList)
	return ctx.SendStatus(200)
}

func PostNewImageMP(ctx *fiber.Ctx) error {
	config, err := biz.GetConfig()
	if err != nil {
		appLogger.ErrorLogger.Println(err)
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	if !httpUtils.ValidateAuth(ctx, config) {
		appLogger.WarningLogger.Print("Auth error")
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	imageId := ctx.Query("imageId")

	if imageId == "" {
		appLogger.WarningLogger.Print("Invalid image id")
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	imageFolderPath := config.FilesPath + "/" + url.PathEscape(imageId)

	if err := createFileFolder(imageFolderPath); err != nil {
		appLogger.ErrorLogger.Print("Cannot create folder: " + err.Error())
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	sourceFilename := "source"
	sourcePath := imageFolderPath + "/" + sourceFilename

	file, err := ctx.FormFile("imageFile")

	if err != nil {
		appLogger.WarningLogger.Print("Invalid content type")
		return ctx.SendStatus(fiber.StatusBadRequest)

	}

	hFile, err := file.Open()
	if err != nil {
		appLogger.ErrorLogger.Print("Cannot open file: " + err.Error())
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	fbytes, err := io.ReadAll(hFile)
	if err != nil {
		appLogger.ErrorLogger.Print("Cannot read file: " + err.Error())
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	if err := saveFile(sourcePath, fbytes); err != nil {
		appLogger.ErrorLogger.Print("Cannot save file: " + err.Error())
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	queueList := createConvertCommands(config, imageFolderPath)

	ic.ConvertImage(sourcePath, queueList)
	return ctx.SendStatus(200)
}

func createFileFolder(imageFolderPath string) error {
	if _, err := os.Stat(imageFolderPath); os.IsNotExist(err) {
		errMkDir := os.Mkdir(imageFolderPath, 0755)
		if errMkDir != nil {
			appLogger.ErrorLogger.Print("Failed to create folder: " + errMkDir.Error())
			return errMkDir
		}
	}
	return nil
}

func saveFile(sourcePath string, file []byte) error {
	f, err := os.OpenFile(sourcePath, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		appLogger.ErrorLogger.Print("Cannot open file " + err.Error())
		return err
	}
	defer f.Close()

	if _, err := f.Write(file); err != nil {
		appLogger.ErrorLogger.Print("Cannot write file " + err.Error())
		return err
	}

	if err := f.Sync(); err != nil {
		appLogger.ErrorLogger.Print("Cannot sync file " + err.Error())
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
