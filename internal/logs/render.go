package logs

import (
	"fmt"
	"strings"
	"unicode/utf8"
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

func TruncateParts(parts []Part, maxWidth int) []Part {
	if maxWidth <= 0 {
		return nil
	}

	totalWidth := 0
	for _, part := range parts {
		totalWidth += utf8.RuneCountInString(part.Text)
	}
	if totalWidth <= maxWidth {
		return parts
	}

	truncated := make([]Part, 0, len(parts))
	remaining := maxWidth
	for _, part := range parts {
		width := utf8.RuneCountInString(part.Text)
		switch {
		case remaining <= 0:
			break
		case width < remaining:
			truncated = append(truncated, part)
			remaining -= width
		default:
			text := truncateString(part.Text, remaining)
			if text != "" {
				truncated = append(truncated, Part{Kind: part.Kind, Text: text, Level: part.Level})
			}
			remaining = 0
		}
		if remaining == 0 {
			break
		}
	}

	if len(truncated) == 0 {
		return []Part{{Kind: PartRaw, Text: truncateString("", maxWidth)}}
	}
	return truncated
}

func truncateString(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if utf8.RuneCountInString(text) <= maxWidth {
		return text
	}
	if maxWidth == 1 {
		return ellipsis
	}

	runes := []rune(text)
	if len(runes) >= maxWidth {
		return string(runes[:maxWidth-1]) + ellipsis
	}
	return string(runes)
}
