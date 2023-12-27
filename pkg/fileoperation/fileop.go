package fileoperation

import "os"

func FileExists(filename string) bool {
	// To Do
	_, err := os.Stat(filename)
	return err == nil
}
