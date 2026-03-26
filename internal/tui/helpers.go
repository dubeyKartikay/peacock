package tui

import "unicode/utf8"

const ellipsis = "…"

func truncateText(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if utf8.RuneCountInString(text) <= maxWidth {
		return text
	}
	if maxWidth == minViewportDimension {
		return ellipsis
	}

	runes := []rune(text)
	return string(runes[:maxWidth-1]) + ellipsis
}
