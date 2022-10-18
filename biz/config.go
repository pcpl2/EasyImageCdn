package biz

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	appLogger "imageConverter.pcpl2lab.ovh/utils/logger"
	models "imageConverter.pcpl2lab.ovh/models"
)

var config models.ApiConfig
var loaded = false

func InitConfiguration() {
	if os.Getenv("IN_DOCKER") != "1" {
		err := godotenv.Load(".env")
		if err != nil {
			appLogger.ErrorLogger.Println("Error with loading .env file: " + err.Error())
		}
	}

	maxFilesize, _ := strconv.Atoi(os.Getenv("MAX_FILE_SIZE"))
	cacheTime, _ := strconv.Atoi(os.Getenv("CACHE_TIME"))

	//TODO validate all configuration

	config = models.ApiConfig{
		APIKey:       os.Getenv("API_KEY"),
		APIKeyHeader: os.Getenv("API_KEY_HEADER"),
		FilesPath:    "/var/lib/images",
		MaxFileSize:  maxFilesize,
		CacheTime:    cacheTime,
	}

	loadResolutions()
}

func loadResolutions() {
	config.Resolutions = []models.ResElement{}
	resList := strings.Split(os.Getenv("CONVERT_TO_RES"), ",")
	for _, element := range resList {
		res := strings.Split(element, "x")
		width, _ := strconv.Atoi(res[0])
		height, _ := strconv.Atoi(res[1])
		config.Resolutions = append(config.Resolutions, models.ResElement{
			Width:  width,
			Height: height,
		})
	}

	loaded = true
}

func GetConfig() (models.ApiConfig, error) {
	if !loaded {
		return config, errors.New("configuration not initialized")
	} else {
		return config, nil
	}
}
