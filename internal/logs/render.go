package logs

import (
	"fmt"
	"strings"
	"charm.land/lipgloss/v2"
	"github.com/muesli/reflow/wordwrap"
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

func FormatEntry(entry Entry) []Part {
	if !entry.Parsed {
		return []Part{{Kind: PartRaw, Text: entry.Raw}}
	}

	parts := make([]Part, 0, partCapacity)
	if entry.Timestamp != "" {
		parts = append(parts, Part{Kind: PartTimestamp, Text: entry.Timestamp + timestampSuffix})
	}

	if entry.Level != ""{
		parts = append(parts, Part{Kind: PartLevel, Text: fmt.Sprintf(levelTextFormat, entry.Level), Level: entry.Level})
	}

	message := entry.Message
	if message == "" && len(entry.Context) == 0 && entry.Caller == "" {
		message = entry.Raw
	}
	if message != "" {
		parts = append(parts, Part{Kind: PartMessage, Text: message})
	}

	if entry.Caller != "" {
		parts = append(parts, Part{Kind: PartCaller, Text: callerPrefix + entry.Caller})
	}

	if len(entry.Context) > 0 {
		parts = append(parts, Part{Kind: PartContext, Text: contextPrefix + renderContext(entry.Context)})
	}

	return parts
}

func RenderPlain(entry Entry) string {
	parts := FormatEntry(entry)
	var builder strings.Builder
	for _, part := range parts {
		builder.WriteString(part.Text)
	}
	return builder.String()
}

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
	text = wordwrap.String(text, maxWidth)
	return text
}
