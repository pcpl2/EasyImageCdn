package publicapis

import (
	"log"

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

	httpUtils.SendFileHTTP(ctx, config, id, fileName)
}
