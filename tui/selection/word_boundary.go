package selection

import (
	"unicode"
)

// IsWordChar returns true if the rune is a word character.
func IsWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

// findWordBoundaries finds the word containing the character at position col.
func findWordBoundaries(line string, col int) (start, end int) {
	runes := []rune(line)
	if len(runes) == 0 {
		return 0, 0
	}

	if col < 0 {
		col = 0
	}
	if col >= len(runes) {
		col = len(runes) - 1
	}

	if !IsWordChar(runes[col]) {
		return col, col
	}

	start = col
	for start > 0 && IsWordChar(runes[start-1]) {
		start--
	}

	end = col
	for end < len(runes)-1 && IsWordChar(runes[end+1]) {
		end++
	}

	return start, end + 1
}
