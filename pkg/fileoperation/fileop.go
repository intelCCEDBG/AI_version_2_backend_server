package fileoperation

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
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

func ZipFiles(zipfile string, files ...string) error {
	// create the zip file
	zipFile, err := os.Create(zipfile)
	if err != nil {
		return err
	}
	defer zipFile.Close()
	// create a new zip archive
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()
	// add files to the archive
	for _, file := range files {
		if err = addFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	return nil
}

func addFileToZip(zipWriter *zip.Writer, filePath string) error {
	// open the file
	fileToZip, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer fileToZip.Close()
	// get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}
	// create a file header
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	// change to deflate to gain better compression
	header.Method = zip.Deflate
	header.Name = filepath.Base(filePath)
	// create a writer for the file in the archive
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	// write the file to the archive
	_, err = io.Copy(writer, fileToZip)
	if err != nil {
		return err
	}
	return nil
}
