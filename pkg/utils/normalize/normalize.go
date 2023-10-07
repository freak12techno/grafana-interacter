package normalize

import (
	"regexp"
	"strings"
)

func NormalizeString(input string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9]+")
	return strings.ToLower(reg.ReplaceAllString(input, ""))
}
