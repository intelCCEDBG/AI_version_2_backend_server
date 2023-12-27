package cropping

import (
	"image"
	"image/draw"
	"image/png"
	"io"
	"os"
	"recorder/pkg/logger"
)

func Crop_image(inputImagePath string, coordinates []float64, outputImagePath string) error {
	file, err := os.Open(inputImagePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Decode the input image file
	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	// Extract bounding box coordinates
	xMin, yMin, xMax, yMax := int(coordinates[0]), int(coordinates[1]), int(coordinates[2]), int(coordinates[3])

	// Calculate the width and height of the cropped region
	width := xMax - xMin
	height := yMax - yMin

	// Create a new image with the cropped size
	croppedImg := image.NewRGBA(image.Rect(0, 0, width, height))

	// Define the crop rectangle in the source image
	cropRect := image.Rect(xMin, yMin, xMax, yMax)

	// Perform cropping by copying the specified region
	draw.Draw(croppedImg, croppedImg.Bounds(), img, cropRect.Min, draw.Src)

	// Create an output file to save the cropped image as a PNG file
	outputFile, err := os.Create(outputImagePath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	// Encode and save the cropped image to a PNG file
	err = png.Encode(outputFile, croppedImg)
	if err != nil {
		return err
	}

	return nil
}
func Switch_picture_if_exist(image_with_path string) {
	_, err := os.Stat(image_with_path)
	if err == nil {
		err = copyFile(image_with_path, image_with_path[:len(image_with_path)-4]+"_old.png")
		if err != nil {
			logger.Error(err.Error())
		}
	}
}

func copyFile(src, dst string) error {
	// Open the source file for reading
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create the destination file for writing
	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	// Copy the content from source to destination
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}
