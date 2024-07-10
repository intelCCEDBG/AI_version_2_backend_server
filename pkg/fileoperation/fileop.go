package fileoperation

import (
	"io"
	"os"
)

func FileExists(filename string) bool {
	// To Do
	_, err := os.Stat(filename)
	return err == nil
}
func CreateFolderifNotExist(path string) {
	// To Do
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModePerm)
	}
}
func CopyFile(src string, dest string) error {
	sourceFile, err := os.Open(src) // Change "source.jpg" to your source image file
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create the destination image file
	destinationFile, err := os.Create(dest) // Change "destination.jpg" to your desired destination image file
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
func DeleteFiles(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}
	return nil
}
