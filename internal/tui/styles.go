package tui

import (
	"strings"

	"charm.land/lipgloss/v2"

	appconfig "github.com/dubeyKartikay/peacock/internal/config"
	"github.com/dubeyKartikay/peacock/internal/logs"
	ac "github.com/petar-dambovaliev/aho-corasick"
)

const (
	barVerticalPadding   = 0
	barHorizontalPadding = 1
	generalPadding       = 1
)

type styles struct {
	aho_corasick ac.AhoCorasickBuilder
	panel        lipgloss.Style
	status       statusStyles
	filterBar    lipgloss.Style
	timestamp    lipgloss.Style
	message      lipgloss.Style
	caller       lipgloss.Style
	context      lipgloss.Style
	raw          lipgloss.Style
	levelError   lipgloss.Style
	levelWarn    lipgloss.Style
	levelInfo    lipgloss.Style
	levelDebug   lipgloss.Style
	levelOther   lipgloss.Style
	highlight    lipgloss.Style
}

type statusStyles struct {
	bar     lipgloss.Style
	live    lipgloss.Style
	paused  lipgloss.Style
	done    lipgloss.Style
	source  lipgloss.Style
	filter  lipgloss.Style
	entries lipgloss.Style
	visible lipgloss.Style
	err     lipgloss.Style
	help    lipgloss.Style
}

func defaultStyles(cfg appconfig.ThemeConfig) styles {
	borderColor := lipgloss.Color(cfg.PanelBorder)
	builder := ac.NewAhoCorasickBuilder(ac.Opts{
		AsciiCaseInsensitive: false,
		MatchOnlyWholeWords:  false,
		MatchKind:            ac.LeftMostLongestMatch,
		DFA:                  true,
	})
	return styles{
		aho_corasick: builder,
		panel:        lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(borderColor),
		status:       defaultStatusStyles(cfg),
		filterBar:    lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.FilterFG)).Padding(barVerticalPadding, barHorizontalPadding),
		timestamp:    lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.TimestampFG)).Faint(cfg.TimestampFaint).PaddingRight(generalPadding),
		message:      lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.MessageFG)).PaddingRight(generalPadding),
		caller:       lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.CallerFG)).Faint(cfg.CallerFaint).PaddingRight(generalPadding),
		context:      lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.ContextFG)).Faint(cfg.ContextFaint).PaddingRight(generalPadding),
		raw:          lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.RawFG)),
		highlight:    lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.HighlightFG)).Background(lipgloss.Color(cfg.HighlightBG)).Bold(true),
		levelError:   lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.LevelError)).Bold(cfg.LevelBold).PaddingRight(generalPadding),
		levelWarn:    lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.LevelWarn)).Bold(cfg.LevelBold).PaddingRight(generalPadding),
		levelInfo:    lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.LevelInfo)).Bold(cfg.LevelBold).PaddingRight(generalPadding),
		levelDebug:   lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.LevelDebug)).Bold(cfg.LevelBold).PaddingRight(generalPadding),
		levelOther:   lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.LevelOther)).Bold(cfg.LevelBold).PaddingRight(generalPadding),
	}
}

func defaultStatusStyles(cfg appconfig.ThemeConfig) statusStyles {
	return statusStyles{
		bar:     lipgloss.NewStyle().Padding(barVerticalPadding, barHorizontalPadding),
		live:    lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.LevelError)),
		paused:  lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.LevelWarn)).PaddingRight(generalPadding),
		done:    lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.LevelOther)).PaddingRight(generalPadding),
		source:  lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.CallerFG)).PaddingRight(generalPadding),
		filter:  lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.FilterFG)).Background(lipgloss.Color(cfg.FilterBG)).PaddingLeft(generalPadding).PaddingRight(generalPadding).MarginLeft(generalPadding),
		entries: lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.StatusFG)).PaddingRight(generalPadding),
		help:    lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.StatusFG)).PaddingLeft(generalPadding).Faint(true),
		err:     lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.LevelError)).PaddingRight(generalPadding),
	}
}

func (s styles) renderEntry(entry *logs.Entry, filters Filters, width int, invalidateRenderCache bool) string {
	if cached, ok := entry.GetCachedRender(width, invalidateRenderCache); ok {
		return cached
	}
	if !entry.Parsed {
		entry.CacheRender(width, entry.Raw, lipgloss.Height(entry.Raw))
		return entry.Raw
	}
	logMetadata := ""
	content := ""

	logMetadata = lipgloss.JoinHorizontal(lipgloss.Left, logMetadata, s.renderPart(entry.Timestamp, filters))
	logMetadata = lipgloss.JoinHorizontal(lipgloss.Left, logMetadata, s.renderPart(entry.Level, filters))

	content = lipgloss.JoinHorizontal(lipgloss.Left, content, s.renderPart(entry.Message, filters))
	content = lipgloss.JoinHorizontal(lipgloss.Left, content, s.renderPart(entry.Caller, filters))
	content = lipgloss.JoinHorizontal(lipgloss.Left, content, s.renderPart(entry.Context, filters))

	view := logs.WrapHorizontalOverflow(logMetadata, content, width)
	entry.CacheRender(width, view, lipgloss.Height(view))
	return view
}

func (s styles) renderHighlights(str string, filter Filters) string {
	matcher := s.aho_corasick.Build(filter)
	matches := matcher.FindAll(str)
	if len(matches) == 0 {
		return str
	}

	var highlighted strings.Builder
	highlighted.Grow(len(str))
	cursor := 0
	for _, match := range matches {
		start, end := match.Start(), match.End()
		highlighted.WriteString(str[cursor:start])
		highlighted.WriteString(s.highlight.Render(str[start:end]))
		cursor = end
	}
	highlighted.WriteString(str[cursor:])
	return highlighted.String()
}

func (s styles) renderPart(part logs.Part, filters Filters) string {
	text := s.renderHighlights(part.Text, filters)

	switch part.Kind {
	case logs.PartTimestamp:
		return s.timestamp.Render(text)
	case logs.PartLevel:
		return s.levelStyle(part.Text).Render(text)
	case logs.PartCaller:
		return s.caller.Render(text)
	case logs.PartContext:
		return s.context.Render(text)
	case logs.PartRaw:
		return s.raw.Render(text)
	default:
		return s.message.Render(text)
	}
}

func (s styles) levelStyle(level string) lipgloss.Style {
	switch level {
	case "fatal", "error":
		return s.levelError
	case "warn":
		return s.levelWarn
	case "info":
		return s.levelInfo
	case "debug":
		return s.levelDebug
	default:
		return s.levelOther
	}
}
