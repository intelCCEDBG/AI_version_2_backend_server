package fileoperation

import "os"

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
