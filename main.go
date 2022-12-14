package main

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/logger"

	biz "easy-image-cdn.pcpl2lab.ovh/biz"
	appLogger "easy-image-cdn.pcpl2lab.ovh/utils/logger"

	aApi "easy-image-cdn.pcpl2lab.ovh/controllers/adminApis"
	pApi "easy-image-cdn.pcpl2lab.ovh/controllers/publicApis"
	utils "easy-image-cdn.pcpl2lab.ovh/controllers/utils"
)

func main() {
	appLogger.StartLogger()
	appLogger.InfoLogger.Print("Loading config..")
	biz.InitConfiguration()
	config, err := biz.GetConfig()
	if err != nil {
		appLogger.ErrorLogger.Fatal(err)
	}

	if config.APIKey == "" || config.APIKey == "00000000-0000-0000-0000-000000000000" {
		appLogger.ErrorLogger.Fatalln("*ERROR* The application will not start without setting the value API_KEY")
	}

	appLogger.InfoLogger.Print("Configuration loaded.")

	fiberLogger := logger.New(logger.Config{
		Format:     "INFO: ${time} [${ip}]:${port} ${status}${latency} - ${method} ${path} ${ua}\n",
		TimeFormat: "2006/01/02 15:04:05",
		Output:     appLogger.LoggerWritter,
	})

	adminApp := fiber.New(fiber.Config{
		BodyLimit:             config.MaxFileSize * 1024 * 1024,
		DisableStartupMessage: true,
		ServerHeader:          "",
	})

	adminApp.Use(fiberLogger)

	adminApp.Post("/v1/newImage", func(c *fiber.Ctx) error {
		return aApi.PostNewImage(c)
	})

	adminApp.Post("/v1/newImageMp", func(c *fiber.Ctx) error {
		return aApi.PostNewImageMP(c)
	})

	appLogger.InfoLogger.Printf("Starting HTTP server on 0.0.0.0:9324")
	go func() {
		if err := adminApp.Listen("0.0.0.0:9324"); err != nil {
			appLogger.ErrorLogger.Fatalf("error in adminApp.Listen: %s", err)
		}
	}()

	publicApp := fiber.New(fiber.Config{
		ServerHeader:          "",
		DisableStartupMessage: true,
	})

	publicApp.Use(fiberLogger)
	publicApp.Use(etag.New())
	publicApp.Use(compress.New(compress.Config{
		Level: compress.LevelBestCompression,
	}))

	publicApp.Use(cache.New(cache.Config{
		Next: func(c *fiber.Ctx) bool {
			return c.Query("refresh") == "true"
		},
		Expiration:   time.Duration(config.CacheTime) * time.Minute,
		CacheControl: true,
	}))

	publicApp.Get("/*", func(c *fiber.Ctx) error {
		spath := utils.DeleteEmpty(strings.Split(c.Path(), "/"))
		fileName := "source"
		if len(spath) < 1 {
			return c.SendStatus(fiber.StatusNotFound)
		} else if len(spath) == 2 {
			fileName = spath[1]
		}

		return pApi.GetImage(c, spath[0], fileName)
	})

	appLogger.InfoLogger.Printf("Starting HTTP server on 0.0.0.0:9555")
	go func() {
		if err := publicApp.Listen("0.0.0.0:9555"); err != nil {
			appLogger.ErrorLogger.Fatalf("error in publicApp.Listen: %s", err)
		}
	}()

	select {}
}
