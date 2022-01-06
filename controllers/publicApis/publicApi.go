package publicapis

import (
	"fmt"
	"log"
	"strings"

	"github.com/valyala/fasthttp"
	httpUtils "imageConverter.pcpl2lab.ovh/controllers/utils"

	biz "imageConverter.pcpl2lab.ovh/biz"
)

func GetImage(ctx *fasthttp.RequestCtx, id string, fileName string) {
	config, err := biz.GetConfig()
	if err != nil {
		ctx.Error("", fasthttp.StatusInternalServerError)
		log.Fatal(err)
	}
	acceptHeader := string(ctx.Request.Header.Peek("Accept"))

	fileNameWithEx := fileName

	if strings.Contains(acceptHeader, "image/webp") {
		log.Printf("Send webp file.")
		fileNameWithEx = fmt.Sprintf("%s.webp", fileName)
	}

	httpUtils.SendFileHTTP(ctx, config, id, fileNameWithEx)
}
