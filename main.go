package main

import (
	"log"
	"strings"

	"github.com/valyala/fasthttp"

	biz "imageConverter.pcpl2lab.ovh/biz"

	aApi "imageConverter.pcpl2lab.ovh/controllers/adminApis"
	pApi "imageConverter.pcpl2lab.ovh/controllers/publicApis"
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

	adminRequestHandler := func(ctx *fasthttp.RequestCtx) {
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

	publicRequestHandler := func(ctx *fasthttp.RequestCtx) {
		spath := strings.Split(string(ctx.Path()), "/")
		if len(spath) < 3 {
			ctx.Error("", fasthttp.StatusNotFound)
			return
		}

		if spath[1] == "" {
			ctx.Error("", fasthttp.StatusNotFound)
			return
		}

		if spath[2] == "" {
			ctx.Error("", fasthttp.StatusNotFound)
			return
		}
		pApi.GetImage(ctx, spath[1], spath[2])
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
