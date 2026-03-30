package logs

import (
	"strings"
	"charm.land/lipgloss/v2"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
)

const (
	callerPrefix     = " caller="
	contextPrefix    = " "
	contextSeparator = " "
	ellipsis         = "…"
	levelTextFormat  = "[%s] "
	partCapacity     = 5
	timestampSuffix  = " "
)

func renderContext(fields []Field) string {
	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		parts = append(parts, field.Key+"="+field.Value)
	}
	return strings.Join(parts, contextSeparator)
}

func WrapHorizontalOverflow(logMetadata string, content string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	totalWidth := lipgloss.Width(logMetadata) + lipgloss.Width(content)

	if totalWidth <= maxWidth {
		return lipgloss.JoinHorizontal(lipgloss.Left, logMetadata, content)
	}
	if (lipgloss.Width(logMetadata)) <= maxWidth {
		wrappedContent := wrapString(content, maxWidth-lipgloss.Width(logMetadata))
		return lipgloss.JoinHorizontal(lipgloss.Left, logMetadata, wrappedContent)
	}
	tuncatedLine := wrapString(lipgloss.JoinHorizontal(lipgloss.Left, logMetadata, content), maxWidth)
	return tuncatedLine

}

func wrapString(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if lipgloss.Width(text) <= maxWidth {
		return text
	}
	text = wrap.String(wordwrap.String(text, maxWidth), maxWidth)
	return text
}
