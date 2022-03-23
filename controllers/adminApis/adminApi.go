package adminapis

import (
	"encoding/base64"
	"fmt"
	"strconv"

	"log"
	"net/url"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/h2non/bimg"

	httpUtils "imageConverter.pcpl2lab.ovh/controllers/utils"

	biz "imageConverter.pcpl2lab.ovh/biz"
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
	log.Print("Hello")

	imageFolderPath := config.FilesPath + "/" + url.PathEscape(Payload.ID)

	if _, err := os.Stat(imageFolderPath); os.IsNotExist(err) {
		errMkDir := os.Mkdir(imageFolderPath, 0755)
		if errMkDir != nil {
			ctx.SendStatus(fiber.StatusNoContent)
			log.Print("*ERROR* Failed to create folder " + errMkDir.Error())
			return
		}
	}

	dec, err := base64.StdEncoding.DecodeString(Payload.Image)
	if err != nil {
		panic(err)
	}

	sourceFilename := "source"
	sourcePath := imageFolderPath + "/" + sourceFilename

	f, err := os.OpenFile(sourcePath, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		ctx.SendStatus(fiber.StatusNoContent)
		log.Print("*ERROR* Cannot open file " + err.Error())
		return
	}
	defer f.Close()

	if _, err := f.Write(dec); err != nil {
		panic(err)
	}
	if err := f.Sync(); err != nil {
		panic(err)
	}

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

	convertImage(sourcePath, queueList)
}

func convertImage(imagePath string, command []models.ConvertCommand) {
	buffer, err := bimg.Read(imagePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	for _, element := range command {
		imagePrt := bimg.NewImage(buffer)
		imageName := "source"
		imageExtension := ""

		if element.ConvertRes {
			_, err := imagePrt.ForceResize(element.TargetRes.Width, element.TargetRes.Height)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			imageName = strconv.Itoa(element.TargetRes.Width) + "x" + strconv.Itoa(element.TargetRes.Height)
		}

		if element.WebP {
			_, err := imagePrt.Convert(bimg.WEBP)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			imageExtension = ".webp"
		}

		bimg.Write(element.Path+imageName+imageExtension, imagePrt.Image())
	}
}
