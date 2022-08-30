package biz

import (
	"fmt"
	"os"
	"strconv"

	"github.com/h2non/bimg"
	models "imageConverter.pcpl2lab.ovh/models"
)

func ConvertImage(imagePath string, command []models.ConvertCommand) {
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
