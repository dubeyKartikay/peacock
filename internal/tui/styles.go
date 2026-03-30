package tui

import (
	"charm.land/lipgloss/v2"

	appconfig "github.com/dubeyKartikay/peacock/internal/config"
	"github.com/dubeyKartikay/peacock/internal/logs"
)

const (
	barVerticalPadding   = 0
	barHorizontalPadding = 1
	generalPadding = 1
)

type styles struct {
	panel      lipgloss.Style
	status     statusStyles
	filterBar  lipgloss.Style
	timestamp  lipgloss.Style
	message    lipgloss.Style
	caller     lipgloss.Style
	context    lipgloss.Style
	raw        lipgloss.Style
	levelError lipgloss.Style
	levelWarn  lipgloss.Style
	levelInfo  lipgloss.Style
	levelDebug lipgloss.Style
	levelOther lipgloss.Style
}

type statusStyles struct {
	bar     lipgloss.Style
	live    lipgloss.Style
	paused  lipgloss.Style
	done    lipgloss.Style
	source  lipgloss.Style
	entries lipgloss.Style
	visible    lipgloss.Style
	err     lipgloss.Style
	help    lipgloss.Style
}

func defaultStyles(cfg appconfig.ThemeConfig) styles {
	borderColor := lipgloss.Color(cfg.PanelBorder)
	return styles{
		panel:      lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(borderColor),
		status:     defaultStatusStyles(cfg),
		filterBar:  lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.FilterFG)).Padding(barVerticalPadding, barHorizontalPadding),
		timestamp:  lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.TimestampFG)).Faint(cfg.TimestampFaint).PaddingRight(generalPadding),
		message:    lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.MessageFG)).PaddingRight(generalPadding),
		caller:     lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.CallerFG)).Faint(cfg.CallerFaint).PaddingRight(generalPadding),
		context:    lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.ContextFG)).Faint(cfg.ContextFaint).PaddingRight(generalPadding),
		raw:        lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.RawFG)),
		levelError: lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.LevelError)).Bold(cfg.LevelBold).PaddingRight(generalPadding),
		levelWarn:  lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.LevelWarn)).Bold(cfg.LevelBold).PaddingRight(generalPadding),
		levelInfo:  lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.LevelInfo)).Bold(cfg.LevelBold).PaddingRight(generalPadding),
		levelDebug: lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.LevelDebug)).Bold(cfg.LevelBold).PaddingRight(generalPadding),
		levelOther: lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.LevelOther)).Bold(cfg.LevelBold).PaddingRight(generalPadding),
	}
}

func defaultStatusStyles(cfg appconfig.ThemeConfig) statusStyles {
	return statusStyles{
		bar :    lipgloss.NewStyle().Padding(barVerticalPadding, barHorizontalPadding),
		live:    lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.LevelError)),
		paused:  lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.LevelWarn)).PaddingRight(generalPadding),
		done:    lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.LevelOther)).PaddingRight(generalPadding),
		source:  lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.CallerFG)).PaddingRight(generalPadding),
		entries: lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.StatusFG)).PaddingRight(generalPadding),
		help:    lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.StatusFG)).PaddingLeft(generalPadding).Faint(true),
		err:     lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.LevelError)).PaddingRight(generalPadding),
	}
}

func (s styles) renderEntry(entry logs.Entry, width int) (string,int) {
	if(!entry.Parsed) {
		return entry.Raw, lipgloss.Height(entry.Raw)
	}
	logMetadata := ""
	content := ""

	logMetadata = lipgloss.JoinHorizontal(lipgloss.Left, logMetadata, s.renderPart(entry.Timestamp))
	logMetadata = lipgloss.JoinHorizontal(lipgloss.Left, logMetadata, s.renderPart(entry.Level))

	content = lipgloss.JoinHorizontal(lipgloss.Left, content, s.renderPart(entry.Message))
	content = lipgloss.JoinHorizontal(lipgloss.Left, content, s.renderPart(entry.Caller))
	content = lipgloss.JoinHorizontal(lipgloss.Left, content, s.renderPart(entry.Context))

	view := logs.WrapHorizontalOverflow(logMetadata, content, width)
	return view, lipgloss.Height(view)
}

func (s styles) renderPart(part logs.Part) string {
	switch part.Kind {
	case logs.PartTimestamp:
		return s.timestamp.Render(part.Text)
	case logs.PartLevel:
		return s.levelStyle(part.Text).Render(part.Text)
	case logs.PartCaller:
		return s.caller.Render(part.Text)
	case logs.PartContext:
		return s.context.Render(part.Text)
	case logs.PartRaw:
		return s.raw.Render(part.Text)
	default:
		return s.message.Render(part.Text)
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
