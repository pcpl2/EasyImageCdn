package imageconverter

import (
	"bytes"
	"image"
	"os"
	"strconv"

	_ "image/jpeg"
	png "image/png"

	"golang.org/x/image/draw"

	"github.com/chai2010/webp"
	"github.com/strukturag/libheif/go/heif"

	models "easy-image-cdn.pcpl2lab.ovh/models"
	appLogger "easy-image-cdn.pcpl2lab.ovh/utils/logger"
)

func ConvertImage(imagePath string, command []models.ConvertCommand) {
	orginalImage, err := openFile(imagePath)
	if err != nil {
		appLogger.ErrorLogger.Println(err)
		return
	}

	for _, cmd := range command {
		if cmd.ConvertRes {
			org := *orginalImage
			resized := image.NewRGBA(image.Rect(0, 0, cmd.TargetRes.Width, cmd.TargetRes.Height))
			draw.Draw(resized, resized.Bounds(), image.White, image.Point{}, draw.Src)
			draw.ApproxBiLinear.Scale(resized, resized.Bounds(), org, org.Bounds(), draw.Src, nil)
			_ = saveFile(cmd.Path, strconv.Itoa(cmd.TargetRes.Width)+"x"+strconv.Itoa(cmd.TargetRes.Height), resized, cmd.WebP, cmd.Heic)
		} else {
			_ = saveFile(cmd.Path, "source", *orginalImage, cmd.WebP, cmd.Heic)
		}
	}
}

func saveFile(filePath string, fileName string, image image.Image, toWebp bool, toHeic bool) error {
	path := filePath + fileName
	b := new(bytes.Buffer)
	if toWebp {
		if err := webp.Encode(b, image, &webp.Options{Quality: 80}); err != nil {
			appLogger.ErrorLogger.Println(err)
			return err
		}
		path = path + ".webp"
	} else if toHeic {
		ctx, err := heif.EncodeFromImage(image, heif.CompressionHEVC, 80, heif.LosslessModeEnabled, heif.LoggingLevelFull)
		if err != nil {
			appLogger.ErrorLogger.Printf("failed to heif encode image: %v\n", err)
			return err
		}
		path = path + ".heif"
		if err := ctx.WriteToFile(path); err != nil {
			return err
		} else {
			return nil
		}
	} else {
		if err := png.Encode(b, image); err != nil {
			appLogger.ErrorLogger.Println(err)
			return err
		}
	}
	newFile, err := os.Create(path)
	if err != nil {
		return err
	}

	defer newFile.Close()

	_, err = b.WriteTo(newFile)

	return err
}

func openFile(filePath string) (*image.Image, error) {
	file, err := os.Open(filePath)

	if err != nil {
		appLogger.ErrorLogger.Println("Cannot open file: " + err.Error())
		return nil, err
	}
	defer file.Close()

	decodedImage, _, err := image.Decode(file)

	if err != nil && decodedImage == nil {
		appLogger.ErrorLogger.Println("Cannot read file: " + err.Error())
		return nil, nil
	}

	return &decodedImage, nil
}
