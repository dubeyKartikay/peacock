package logs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	boolFalseValue    = false
	levelFatal        = "fatal"
	levelError        = "error"
	levelWarn         = "warn"
	levelWarning      = "warning"
	levelInfo         = "info"
	levelDebug        = "debug"
	lineJoinSeparator = " "
	nullValue         = "NULL"
	searchQuoteChars  = " \t\n\r\""
)

var (
	levelKeys     = []string{"level"}
	timestampKeys = []string{"time", "timestamp"}
	messageKeys   = []string{"message", "msg"}
	callerKeys    = []string{"caller", "file"}
)

func ParseLine(line string) Entry {
	entry := Entry{Raw: line, Search: line}

	decoder := json.NewDecoder(strings.NewReader(line))

	var payload map[string]any
	if err := decoder.Decode(&payload); err != nil {
		return entry
	}

	entry.Parsed = true
	entry.Level = Part{
		Kind: PartLevel,
		Text: extractString(payload, levelKeys...),
	}
	entry.Timestamp = Part{
		Kind: PartTimestamp,
		Text: extractString(payload, timestampKeys...),
	}

	entry.Message = Part{
		Kind: PartMessage,
		Text: extractString(payload, messageKeys...),
	}
	entry.Caller = Part{
		Kind: PartCaller,
		Text: extractString(payload, callerKeys...),
	}
	entry.Context = Part{
		Kind: PartContext,
		Text: renderContext(extractContext(payload)),
	}
	entry.Search = buildSearchText(entry)

	return entry
}

func extractString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := payload[key]
		if !ok {
			continue
		}
		delete(payload, key)
		switch v := value.(type) {
		case string:
			return v
		case json.Number:
			return v.String()
		default:
			return fmt.Sprint(v)
		}
	}
	return ""
}

func extractContext(payload map[string]any) []Field {
	if len(payload) == 0 {
		return nil
	}

	keys := make([]string, 0, len(payload))
	for key := range payload {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	fields := make([]Field, 0, len(keys))
	for _, key := range keys {
		fields = append(fields, Field{Key: key, Value: stringifyValue(payload[key])})
	}
	return fields
}

func stringifyValue(value any) string {
	switch v := value.(type) {
	case nil:
		return nullValue
	case string:
		if strings.ContainsAny(v, searchQuoteChars) {
			return strconv.Quote(v)
		}
		return v
	case json.Number:
		return v.String()
	case bool:
		return strconv.FormatBool(v)
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprint(v)
		}
		return string(bytes.TrimSpace(data))
	}
}

func normalizeLevel(level string) string {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case levelFatal:
		return levelFatal
	case levelError:
		return levelError
	case levelWarn, levelWarning:
		return levelWarn
	case levelInfo:
		return levelInfo
	case levelDebug:
		return levelDebug
	default:
		return strings.ToUpper(strings.TrimSpace(level))
	}
}

func buildSearchText(entry Entry) string {
	if !entry.Parsed {
		return entry.Raw
	}
	parts := []string{entry.Timestamp.Text, entry.Level.Text, entry.Message.Text, entry.Caller.Text, entry.Context.Text}
	return strings.Join(parts, lineJoinSeparator)
}
