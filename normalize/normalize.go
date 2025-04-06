// this package cleans up the text before training with BPE
package normalize

import (
	"golang.org/x/text/unicode/norm"

	"regexp"
	"strings"
	"unicode"
)

// Regex to match whitespace
var whitespaceRegex = regexp.MustCompile(`\s+`)

// Remove control characters and non-printable characters
func removeControlChars(text string) string {
	var b strings.Builder
	for _, r := range text {
		if unicode.IsPrint(r) || unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Normalize Unicode characters
func normalizeUnicode(text string) string {
	return norm.NFKC.String(text)
}

// Reduce whitespace to single spaces
func normalizeWhitespace(text string) string {
	return strings.TrimSpace(whitespaceRegex.ReplaceAllString(text, " "))
}

// Normalize
func Normalize(text string) string {
	// pre processing operations
	text = removeControlChars(text)
	text = normalizeUnicode(text)
	text = normalizeWhitespace(text)

	// return normalized string
	return strings.ToLower(text)
}
