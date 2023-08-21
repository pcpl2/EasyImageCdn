package imageconverter

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"

	models "easy-image-cdn.pcpl2lab.ovh/models"
)

func TestOpenImage(t *testing.T) {
	_, err := openFile("../testImages/350x150.png")
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestConvertImageTo64X64(t *testing.T) {
	ConvertImage("../testImages/350x150.png", []models.ConvertCommand{
		{
			Path:       "../testImages" + "/",
			WebP:       false,
			ConvertRes: true,
			TargetRes:  models.ResElement{Width: 64, Height: 64},
		},
	})

	checksum, err := calcuateCheckSumForFile("../testImages/64x64")
	if err != nil {
		t.Fatalf(err.Error())
	}

	if *checksum != "75f2b66fc23eae1022f73979c76c6f4c3bc1f46f920b3a8cbfab48f68221e991" {
		t.Failed()
	}

	cleanAfterTest("../testImages/64x64")
}

func TestConvertImageTo64X64WebP(t *testing.T) {
	ConvertImage("../testImages/350x150.png", []models.ConvertCommand{
		{
			Path:       "../testImages" + "/",
			WebP:       true,
			ConvertRes: true,
			TargetRes:  models.ResElement{Width: 64, Height: 64},
		},
	})

	checksum, err := calcuateCheckSumForFile("../testImages/64x64.webp")
	if err != nil {
		t.Fatalf(err.Error())
	}

	if *checksum != "2cb03ec0c8d424fad03a355a930efe7548f1f6b1f729cccabbf0ddc676025eeb" {
		t.Failed()
	}

	cleanAfterTest("../testImages/64x64.webp")
}

func calcuateCheckSumForFile(filePath string) (*string, error) {
	s, err := os.ReadFile(filePath)
	hasher := sha256.New()
	hasher.Write(s)
	if err != nil {
		return nil, err
	}

	hash := (hex.EncodeToString(hasher.Sum(nil)))

	return &hash, nil
}

func cleanAfterTest(filePath string) {
	os.Remove(filePath)
}
