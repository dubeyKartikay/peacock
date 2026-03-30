package tui

import (
	"github.com/muesli/reflow/truncate"
	"charm.land/lipgloss/v2"
)

const ellipsis = "…"

func truncateText(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if lipgloss.Width(text) <= maxWidth {
		return text
	}
	if maxWidth == minViewportDimension {
		return ellipsis
	}
	return truncate.StringWithTail(text, uint(maxWidth), ellipsis)
}
func (f * Filters) pop () string {
	if len(*f) > 0  {
		last := (*f)[len(*f)-1]
		*f = (*f)[:len(*f)-1]
		return last
	}
	return ""
}
