package tui

import (
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	appconfig "github.com/dubeyKartikay/peacock/internal/config"
	"github.com/dubeyKartikay/peacock/internal/logs"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func TestEntryRingEvictsOldest(t *testing.T) {
	ring := newEntryRing(3)
	ring.AppendBatch([]logs.Entry{
		testEntry("first"),
		testEntry("second"),
		testEntry("third"),
		testEntry("fourth"),
	})

	got := ringMessages(ring.Newest(0))
	want := []string{"second", "third", "fourth"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("unexpected ring order: got %v want %v", got, want)
	}
}

func TestPauseBuffersNewEntriesUntilResume(t *testing.T) {
	m := newSizedModel(appconfig.DefaultConfig())
	m = updateModel(t, m, EntryMsg{Entry: logs.ParseLine(`{"level":"info","time":"2026-03-25T12:00:00Z","message":"first"}`)})

	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeySpace})
	if !m.paused {
		t.Fatal("expected model to pause after space")
	}
	pausedView := stripANSI(m.viewport.View())

	m = updateModel(t, m, EntryMsg{Entry: logs.ParseLine(`{"level":"error","time":"2026-03-25T12:00:01Z","message":"second"}`)})
	if got := m.queuedEntries.Len(); got != 1 {
		t.Fatalf("expected 1 queued entry, got %d", got)
	}
	if got := stripANSI(m.viewport.View()); got != pausedView {
		t.Fatalf("expected paused viewport to stay frozen\nwant: %q\n got: %q", pausedView, got)
	}

	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeySpace})
	if m.paused {
		t.Fatal("expected model to resume after second space")
	}
	if got := m.queuedEntries.Len(); got != 0 {
		t.Fatalf("expected queued entries to flush on resume, got %d", got)
	}
	if got := stripANSI(m.viewport.View()); !strings.Contains(got, "second") {
		t.Fatalf("expected resumed viewport to include buffered entry, got %q", got)
	}
}

func TestResumeFlushPreservesQueuedOrderWithinCapacity(t *testing.T) {
	cfg := appconfig.DefaultConfig()
	cfg.Buffer.MaxEntries = 3

	m := newSizedModel(cfg)
	for _, message := range []string{"first", "second", "third", "fourth"} {
		m = updateModel(t, m, EntryMsg{Entry: logs.ParseLine(`{"level":"info","time":"2026-03-25T12:00:00Z","message":"` + message + `"}`)})
		if message == "first" {
			m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeySpace})
		}
	}

	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeySpace})

	got := ringMessages(m.inBufferEntries.Newest(0))
	want := []string{"second", "third", "fourth"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("unexpected resumed order: got %v want %v", got, want)
	}
}

func TestFilterUsesLiteralSubstringMatching(t *testing.T) {
	m := newSizedModel(appconfig.DefaultConfig())
	m = updateModel(t, m, EntryMsg{Entry: logs.ParseLine(`{"level":"info","time":"2026-03-25T12:00:00Z","message":"health check ok"}`)})
	m = updateModel(t, m, EntryMsg{Entry: logs.ParseLine(`{"level":"error","time":"2026-03-25T12:00:01Z","message":"database timeout"}`)})

	m = updateModel(t, m, tea.KeyPressMsg{Code: '/', Text: "/"})
	for _, key := range []tea.KeyPressMsg{
		{Code: 't', Text: "t"},
		{Code: 'i', Text: "i"},
		{Code: 'm', Text: "m"},
		{Code: 'e', Text: "e"},
		{Code: 'o', Text: "o"},
		{Code: 'u', Text: "u"},
		{Code: 't', Text: "t"},
	} {
		m = updateModel(t, m, key)
	}
	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})

	if got, want := len(m.filters), 1; got != want {
		t.Fatalf("expected %d active filter, got %d", want, got)
	}
	if got := m.filters[0]; got != "timeout" {
		t.Fatalf("expected timeout filter, got %q", got)
	}

	filtered := m.filteredEntries(0)
	if got, want := len(filtered), 1; got != want {
		t.Fatalf("expected %d filtered entry, got %d", want, got)
	}
	if got := filtered[0].Message.Text; got != "database timeout" {
		t.Fatalf("unexpected filtered message %q", got)
	}
}

func TestRenderHighlights(t *testing.T) {
	s := defaultStyles(appconfig.DefaultConfig().Theme)
	s.highlight = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("11"))

	const text = "database timeout while reading"
	if got := s.renderHighlights(text, Filters{"missing"}); got != text {
		t.Fatalf("expected no-match path to return original string, got %q", got)
	}

	got := s.renderHighlights(text, Filters{"timeout"})
	if stripped := stripANSI(got); stripped != text {
		t.Fatalf("highlight changed text: got %q want %q", stripped, text)
	}
	if !strings.Contains(got, "\x1b[") {
		t.Fatalf("expected highlighted output to contain ANSI styling, got %q", got)
	}
}

func newSizedModel(cfg appconfig.Config) model {
	m := NewModel("stdin", cfg).(model)
	m.width = 80
	m.height = 20
	m.syncViewport(true, false)
	return m
}

func updateModel(t *testing.T, m model, msg tea.Msg) model {
	t.Helper()
	updated, _ := m.Update(msg)
	result, ok := updated.(model)
	if !ok {
		t.Fatal("expected concrete model result")
	}
	return result
}

func ringMessages(entries []*logs.Entry) []string {
	messages := make([]string, 0, len(entries))
	for _, entry := range entries {
		messages = append(messages, entry.Message.Text)
	}
	return messages
}

func testEntry(message string) logs.Entry {
	return logs.Entry{
		Message: logs.Part{Text: message},
		Search:  message,
	}
}

func stripANSI(text string) string {
	return ansiPattern.ReplaceAllString(text, "")
}
