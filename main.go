package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/h2non/bimg"

	"github.com/valyala/fasthttp"
)

type apiConfig struct {
	AdminHTTPAddr  string `json:"adminHttpAddr"`
	PublicHttpAddr string `json:"publicHttpAddr"`
	APIKey         string `json:"apiKey"`
	APIKeyHeader   string `json:"apiKeyHeader"`
	FilesPath      string `json:"filesPath"`
	ConvertToRes   string `json:"convertToRes"`
	MaxFileSize    int    `json:"maxFileSize"`
}

type imagePayload struct {
	ID    string `json:"id"`
	Image string `json:"image"`
}

type resElement struct {
	Width  int
	Height int
}

type convertCommand struct {
	Path       string
	WebP       bool
	ConvertRes bool
	TargetRes  resElement
}

var config apiConfig
var resolutions []resElement

func loadConfig() error {
	if os.Getenv("IN_DOCKER") == "1" {
		println("I'm in Docker :)")

		maxFilesize, _ := strconv.Atoi(os.Getenv("MAX_FILE_SIZE"))

		config = apiConfig{
			AdminHTTPAddr:  os.Getenv("ADMIN_HTTP_ADDR"),
			PublicHttpAddr: os.Getenv("PUBLIC_HTTP_ADDR"),
			APIKey:         os.Getenv("API_KEY"),
			APIKeyHeader:   os.Getenv("API_KEY_HEADER"),
			FilesPath:      os.Getenv("FILES_PATH"),
			ConvertToRes:   os.Getenv("CONVERT_TO_RES"),
			MaxFileSize:    maxFilesize,
		}
	} else {
		configFile, err := os.Open("config.json")
		if err != nil {
			return err
		}
		defer configFile.Close()

		byteValue, err := ioutil.ReadAll(configFile)
		if err != nil {
			return err
		}

		err = json.Unmarshal(byteValue, &config)
		if err != nil {
			return err
		}
	}

	return nil
}

func loadResolutions() {
	resolutions = []resElement{}
	resList := strings.Split(config.ConvertToRes, ",")
	for _, element := range resList {
		res := strings.Split(element, "x")
		width, _ := strconv.Atoi(res[0])
		height, _ := strconv.Atoi(res[1])
		resolutions = append(resolutions, resElement{
			Width:  width,
			Height: height,
		})
	}
}

func validateAuth(ctx *fasthttp.RequestCtx) bool {
	return string(ctx.Request.Header.Peek(config.APIKeyHeader)) == config.APIKey
}

func postNewImage(ctx *fasthttp.RequestCtx) {
	if !ctx.IsPost() {
		ctx.Error("", fasthttp.StatusNoContent)
		log.Print("*ERROR* invalid method")
		return
	} else if !strings.Contains(string(ctx.Request.Header.ContentType()), "application/json") {
		ctx.Error("", fasthttp.StatusNoContent)
		log.Print("*ERROR* invalid content type")
		return
	}

	if !validateAuth(ctx) {
		ctx.Error("", fasthttp.StatusNoContent)
		log.Print("*ERROR* Auth error")
		return
	}

	var Payload imagePayload
	err := json.Unmarshal(ctx.Request.Body(), &Payload)

	if err != nil {
		ctx.Error("", fasthttp.StatusNoContent)
		log.Print("*ERROR* Failed to parse payload " + err.Error())
		return
	}

	imageFolderPath := config.FilesPath + "/" + url.PathEscape(Payload.ID)

	if _, err := os.Stat(imageFolderPath); os.IsNotExist(err) {
		errMkDir := os.Mkdir(imageFolderPath, 0755)
		if errMkDir != nil {
			ctx.Error("", fasthttp.StatusNoContent)
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
		ctx.Error("", fasthttp.StatusNoContent)
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

	queueList := []convertCommand{}

	queueList = append(queueList, convertCommand{
		Path:       imageFolderPath + "/",
		WebP:       true,
		ConvertRes: false,
	})

	for _, element := range resolutions {
		queueList = append(queueList, convertCommand{
			Path:       imageFolderPath + "/",
			WebP:       true,
			ConvertRes: true,
			TargetRes:  element,
		})

		queueList = append(queueList, convertCommand{
			Path:       imageFolderPath + "/",
			WebP:       false,
			ConvertRes: true,
			TargetRes:  element,
		})
	}

	convertImage(sourcePath, queueList)
}

func convertImage(imagePath string, command []convertCommand) {
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

func main() {
	log.Print("Loading config..")
	err := loadConfig()
	if err != nil {
		log.Print("*ERROR* Failed to load config " + err.Error())
		return
	}
	loadResolutions()

	log.Print("Configuration loaded.")

	adminRequestHandler := func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Set("x-powered-by", "PHP/7.3.23")

		switch string(ctx.Path()) {
		case "/v1/newImage":
			postNewImage(ctx)
		}
	}

	adminServer := &fasthttp.Server{
		Handler:               adminRequestHandler,
		NoDefaultServerHeader: true,
		MaxRequestBodySize:    config.MaxFileSize * 1024 * 1024,
	}

	if len(config.AdminHTTPAddr) > 0 {
		log.Printf("Starting HTTP server on %q", config.AdminHTTPAddr)
		go func() {
			if err := adminServer.ListenAndServe(config.AdminHTTPAddr); err != nil {
				log.Fatalf("error in ListenAndServe: %s", err)
			}
		}()
	}

	fs := &fasthttp.FS{
		Root:               config.FilesPath,
		GenerateIndexPages: false,
		Compress:           true,
		AcceptByteRange:    false,
	}

	fsHandler := fs.NewRequestHandler()

	publicRequestHandler := func(ctx *fasthttp.RequestCtx) {
		fsHandler(ctx)
	}

	publicServer := &fasthttp.Server{
		Handler:               publicRequestHandler,
		NoDefaultServerHeader: true,
	}

	if len(config.PublicHttpAddr) > 0 {
		log.Printf("Starting HTTP server on %q", config.PublicHttpAddr)
		go func() {
			if err := publicServer.ListenAndServe(config.PublicHttpAddr); err != nil {
				log.Fatalf("error in ListenAndServe: %s", err)
			}
		}()
	}

	select {}
}
