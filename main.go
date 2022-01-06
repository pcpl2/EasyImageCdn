package main

import (
	"log"

	"github.com/valyala/fasthttp"

	biz "imageConverter.pcpl2lab.ovh/biz"

	aApi "imageConverter.pcpl2lab.ovh/controllers/adminApis"
)

func main() {
	log.Print("Loading config..")
	biz.InitConfiguration()
	config, err := biz.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Configuration loaded.")

	adminRequestHandler := func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Set("x-powered-by", "PHP/7.3.23")

		switch string(ctx.Path()) {
		case "/v1/newImage":
			aApi.PostNewImage(ctx)
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
