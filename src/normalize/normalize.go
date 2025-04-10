// this package cleans up the text before training with BPE
package normalize

import (
	"golang.org/x/text/unicode/norm"

	"regexp"
	"strings"
	"unicode"
)

// Regex to match whitespace
var pdRegex = regexp.MustCompile(`\s+`)

// check for emoji
func isEmoji(r rune) bool {
	// Rough filter for common emoji ranges (can be expanded)
	return (r >= 0x1F600 && r <= 0x1F64F) || // Emoticons
		(r >= 0x1F300 && r <= 0x1F5FF) || // Misc Symbols and Pictographs
		(r >= 0x1F680 && r <= 0x1F6FF) || // Transport & Map
		(r >= 0x2600 && r <= 0x26FF) || // Misc symbols
		(r >= 0x2700 && r <= 0x27BF) || // Dingbats
		(r >= 0x1F900 && r <= 0x1F9FF) || // Supplemental Symbols and Pictographs
		(r >= 0x1FA70 && r <= 0x1FAFF) || // Symbols and Pictographs Extended-A
		unicode.Is(unicode.Variation_Selector, r) // Skin tone modifiers etc.
}

// Remove some characters that we do not want to parse (control chars, emojis, etc.)
func removeChars(sText string) string {
	var dBuilder strings.Builder
	for _, r := range sText {
		if (unicode.IsPrint(r) || unicode.IsSpace(r)) && !isEmoji(r) {
			dBuilder.WriteRune(r)
		}
	}
	return dBuilder.String()
}

// Normalize Unicode characters
func normalizeUnicode(sText string) string {
	return norm.NFKC.String(sText)
}

// Reduce whitespace to single spaces
func normalizeWhitespace(sText string) string {
	return strings.TrimSpace(pdRegex.ReplaceAllString(sText, " "))
}

// Normalize
func Normalize(sText string) string {
	// pre processing operations
	sText = removeChars(sText)
	sText = normalizeUnicode(sText)
	sText = normalizeWhitespace(sText)

	// return normalized string
	return strings.ToLower(sText)
}
