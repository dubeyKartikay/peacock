package tui

import (
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	appconfig "github.com/dubeyKartikay/peacock/internal/config"
	"github.com/dubeyKartikay/peacock/internal/logs"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func TestPauseBuffersNewEntriesUntilResume(t *testing.T) {
	m := newSizedModel()
	m = updateModel(t, m, EntryMsg{Entry: logs.ParseLine(`{"level":"info","time":"2026-03-25T12:00:00Z","message":"first"}`)})

	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeySpace})
	if !m.paused {
		t.Fatal("expected model to pause after space")
	}
	pausedView := stripANSI(m.viewport.View())

	m = updateModel(t, m, EntryMsg{Entry: logs.ParseLine(`{"level":"error","time":"2026-03-25T12:00:01Z","message":"second"}`)})
	if got := len(m.queuedEntries); got != 1 {
		t.Fatalf("expected 1 queued entry, got %d", got)
	}
	if got := stripANSI(m.viewport.View()); got != pausedView {
		t.Fatalf("expected paused viewport to stay frozen\nwant: %q\n got: %q", pausedView, got)
	}

	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeySpace})
	if m.paused {
		t.Fatal("expected model to resume after second space")
	}
	if got := len(m.queuedEntries); got != 0 {
		t.Fatalf("expected queued entries to flush, got %d", got)
	}
	if got := len(m.inBufferEntries); got != 2 {
		t.Fatalf("expected resumed buffer to contain 2 entries, got %d", got)
	}
	if got := stripANSI(m.viewport.View()); !strings.Contains(got, "second") {
		t.Fatalf("expected resumed viewport to include buffered entry, got %q", got)
	}
}

func TestSlashFilterCommitsLiteralSubstringFilterOnEnter(t *testing.T) {
	m := newSizedModel()
	m = updateModel(t, m, EntryMsg{Entry: logs.ParseLine(`{"level":"info","time":"2026-03-25T12:00:00Z","message":"health check ok"}`)})
	m = updateModel(t, m, EntryMsg{Entry: logs.ParseLine(`{"level":"error","time":"2026-03-25T12:00:01Z","message":"database timeout"}`)})

	m = updateModel(t, m, tea.KeyPressMsg{Text: "/", Code: '/'})
	for _, ch := range "timeout" {
		m = updateModel(t, m, tea.KeyPressMsg{Text: string(ch), Code: ch})
	}

	if !m.filterActive {
		t.Fatal("expected filter mode to be active while typing")
	}
	if got := m.filterInput.Value(); got != "timeout" {
		t.Fatalf("expected filter input timeout, got %q", got)
	}
	if got := len(m.filters); got != 0 {
		t.Fatalf("expected filter not to commit before enter, got %d committed filters", got)
	}

	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.filterActive {
		t.Fatal("expected enter to exit filter mode")
	}
	if got, want := len(m.filters), 1; got != want {
		t.Fatalf("expected %d committed filter, got %d", want, got)
	}
	if m.filters[0] != "timeout" {
		t.Fatalf("expected committed filter timeout, got %q", m.filters[0])
	}

	filtered := m.filteredEntryIndexes()
	if got, want := len(filtered), 1; got != want {
		t.Fatalf("expected %d filtered entry, got %d", want, got)
	}
	if filtered[0].Message.Text != "database timeout" {
		t.Fatalf("unexpected filtered message %q", filtered[0].Message.Text)
	}
	if got := stripANSI(m.viewport.View()); strings.Contains(got, "health check ok") || !strings.Contains(got, "database timeout") {
		t.Fatalf("expected viewport to show only filtered entry, got %q", got)
	}

	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeyBackspace})
	if got := len(m.filters); got != 0 {
		t.Fatalf("expected backspace to remove last filter, got %d", got)
	}
	if got := len(m.filteredEntryIndexes()); got != 2 {
		t.Fatalf("expected both entries after removing filter, got %d", got)
	}
}

func TestCommittedFilterWithNoMatchesShowsNoVisibleEntries(t *testing.T) {
	m := newSizedModel()
	m = updateModel(t, m, EntryMsg{Entry: logs.ParseLine(`{"level":"info","time":"2026-03-25T12:00:00Z","message":"health check ok"}`)})

	m.filters = Filters{"zzz"}
	m.syncViewport(true)

	if filtered := m.filteredEntryIndexes(); len(filtered) != 0 {
		t.Fatalf("expected no matching entries, got %#v", filtered)
	}
	if got := m.visibleEntryCount(); got != 0 {
		t.Fatalf("expected zero visible entries, got %d", got)
	}
	if got := stripANSI(m.viewport.View()); strings.Contains(got, "health check ok") {
		t.Fatalf("expected viewport to hide non-matching entry, got %q", got)
	}
}

func newSizedModel() model {
	m := NewModel("stdin", appconfig.DefaultConfig()).(model)
	m.width = 80
	m.height = 20
	m.syncViewport(true)
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

func stripANSI(text string) string {
	return ansiPattern.ReplaceAllString(text, "")
}
