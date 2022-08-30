package imageconverter

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"os"
	"strconv"

	_ "image/jpeg"
	png "image/png"

	"github.com/chai2010/webp"
	"golang.org/x/image/draw"
	models "imageConverter.pcpl2lab.ovh/models"
)

func ConvertImage(imagePath string, command []models.ConvertCommand) {
	orginalImage, err := openFile(imagePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		//TODO log error
		return
	}

	for _, cmd := range command {
		if cmd.ConvertRes {
			org := *orginalImage
			resized := image.NewNRGBA(image.Rect(0, 0, cmd.TargetRes.Width, cmd.TargetRes.Height))
			draw.Draw(resized, resized.Bounds(), image.White, image.Point{}, draw.Src)
			draw.ApproxBiLinear.Scale(resized, resized.Bounds(), org, org.Bounds(), draw.Src, nil)
			saveFile(cmd.Path, strconv.Itoa(cmd.TargetRes.Width)+"x"+strconv.Itoa(cmd.TargetRes.Height), resized, cmd.WebP)
		} else {
			saveFile(cmd.Path, "source", *orginalImage, cmd.WebP)
		}
	}
}

func saveFile(filePath string, fileName string, image image.Image, toWebp bool) error {
	path := filePath + fileName
	b := new(bytes.Buffer)
	if toWebp {
		if err := webp.Encode(b, image, &webp.Options{Lossless: true}); err != nil {
			log.Println(err)
		}
		path = path + ".webp"
	} else {
		png.Encode(b, image)
	}
	newFile, err := os.Create(path)
	if err != nil {
		return err
	}

	defer newFile.Close()

	b.WriteTo(newFile)

	return nil
}

func openFile(filePath string) (*image.Image, error) {
	file, err := os.Open(filePath)

	if err != nil {
		println("Cannot open file: " + err.Error())
		return nil, err
	}
	defer file.Close()

	image, _, err := image.Decode(file)

	if err != nil && image == nil {
		println("Cannot read image: " + err.Error())
		return nil, nil
	}

	return &image, nil
}
