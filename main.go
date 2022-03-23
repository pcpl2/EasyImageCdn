package main

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/logger"

	biz "imageConverter.pcpl2lab.ovh/biz"

	aApi "imageConverter.pcpl2lab.ovh/controllers/adminApis"
	pApi "imageConverter.pcpl2lab.ovh/controllers/publicApis"
	utils "imageConverter.pcpl2lab.ovh/controllers/utils"
)

func main() {
	log.Print("Loading config..")
	biz.InitConfiguration()
	config, err := biz.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	if config.APIKey == "" || config.APIKey == "00000000-0000-0000-0000-000000000000" {
		log.Fatalln("*ERROR* The application will not start without setting the value API_KEY")
	}

	log.Print("Configuration loaded.")

	logger := logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	})

	adminApp := fiber.New(fiber.Config{
		BodyLimit:    config.MaxFileSize * 1024 * 1024,
		ServerHeader: "",
	})

	adminApp.Use(logger)

	adminApp.Post("/v1/newImage", func(c *fiber.Ctx) error {
		aApi.PostNewImage(c)
		return nil
	})

	if len(config.AdminHTTPAddr) > 0 {
		log.Printf("Starting HTTP server on %q", config.AdminHTTPAddr)
		go func() {
			if err := adminApp.Listen(config.AdminHTTPAddr); err != nil {
				log.Fatalf("error in ListenAndServe: %s", err)
			}
		}()
	}

	publicApp := fiber.New(fiber.Config{
		ServerHeader: "",
	})

	publicApp.Use(logger)
	publicApp.Use(etag.New())
	publicApp.Use(compress.New(compress.Config{
		Level: compress.LevelBestCompression, // 1
	}))

	publicApp.Get("/*", func(c *fiber.Ctx) error {
		spath := utils.DeleteEmpty(strings.Split(string(c.Path()), "/"))
		if len(spath) < 2 {
			c.SendStatus(fiber.StatusNotFound)
			return nil
		}

		pApi.GetImage(c, spath[0], spath[1])
		return nil
	})

	if len(config.PublicHttpAddr) > 0 {
		log.Printf("Starting HTTP server on %q", config.PublicHttpAddr)
		go func() {
			if err := publicApp.Listen(config.PublicHttpAddr); err != nil {
				log.Fatalf("error in ListenAndServe: %s", err)
			}
		}()
	}

	select {}
}
