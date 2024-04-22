package ssim

import (
	"image"
	"os"
)

func Ssim_cal(image1 string, image2 string) (float64, error) {
	img1File, err := os.Open(image1)
	if err != nil {
		return 0.0, err
	}
	defer img1File.Close()
	img1RGB, _, err := image.Decode(img1File)
	if err != nil {
		return 0.0, err
	}
	img2File, err := os.Open(image2)
	if err != nil {
		return 0.0, err
	}
	defer img2File.Close()
	img2RGB, _, err := image.Decode(img2File)
	if err != nil {
		return 0.0, err
	}
	img1 := RGBtoGRAY(img1RGB)
	img2 := RGBtoGRAY(img2RGB)
	ssimValue := Ssim(img1, img2)
	return ssimValue, nil
}

func RGBtoGRAY(originalImg image.Image) *image.Gray {
	grayImg := image.NewGray(originalImg.Bounds())
	for y := originalImg.Bounds().Min.Y; y < originalImg.Bounds().Max.Y; y++ {
		for x := originalImg.Bounds().Min.X; x < originalImg.Bounds().Max.X; x++ {
			grayImg.Set(x, y, originalImg.At(x, y))
		}
	}
	return grayImg
}
