package generator

import (
	"strings"
)

// -----------------------------------------------------------------------------

func capitalizeFirstLetter(s string) string {
	runes := []rune(s)
	return strings.ToUpper(string(runes[0])) + string(runes[1:])
}
