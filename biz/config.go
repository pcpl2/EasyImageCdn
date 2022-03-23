package biz

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	models "imageConverter.pcpl2lab.ovh/models"
)

var config models.ApiConfig
var loaded = false

func InitConfiguration() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error with loading .env file: " + err.Error())
	}

	maxFilesize, _ := strconv.Atoi(os.Getenv("MAX_FILE_SIZE"))
	cacheTime, _ := strconv.Atoi(os.Getenv("CACHE_TIME"))

	//TODO validate all configuration
	//TODO Default configuration if enviroment is empty
	//if os.Getenv("PUBLIC_HTTP_ADDR") == "" {
	//	log.Print("Not declared public http addr")
	//}

	config = models.ApiConfig{
		AdminHTTPAddr:  os.Getenv("ADMIN_HTTP_ADDR"),
		PublicHttpAddr: os.Getenv("PUBLIC_HTTP_ADDR"),
		APIKey:         os.Getenv("API_KEY"),
		APIKeyHeader:   os.Getenv("API_KEY_HEADER"),
		FilesPath:      os.Getenv("FILES_PATH"),
		MaxFileSize:    maxFilesize,
		CacheTime:      cacheTime,
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
