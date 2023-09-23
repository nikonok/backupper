package helpers

import (
	"regexp"
)

func ParseScheduledDelete(filename string) (string, string, bool) {
	re := regexp.MustCompile(DATE_TIME_REGEX)
	matches := re.FindStringSubmatch(filename)

	if len(matches) != 4 {
		return "", "", false
	}

	return matches[2], matches[3][1:], true
}
