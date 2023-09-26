package helpers

import (
	"os"
	"regexp"
)

func CreateDirIfNotExists(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, DIR_DEFAULT_PERMISSIONS)
		if err != nil {
			panic("cannot create dir " + path + ": " + err.Error())
		}
	}
}

func ParseScheduledDelete(filename string) (string, string, bool) {
	re := regexp.MustCompile(DATE_TIME_REGEX)
	matches := re.FindStringSubmatch(filename)

	if len(matches) != 4 {
		return "", "", false
	}

	if len(matches[3]) == 0 {
		return "", "", false
	}

	return matches[2], matches[3][1:], true
}
