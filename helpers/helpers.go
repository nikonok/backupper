package helpers

import "os"

func CreateDirIfNotExists(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, DIR_DEFAULT_PERMISSIONS)
		if err != nil {
			panic("cannot create dir " + path + ": " + err.Error())
		}
	}
}