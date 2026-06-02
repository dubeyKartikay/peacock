package tui

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	appconfig "github.com/dubeyKartikay/peacock/internal/config"
	"github.com/dubeyKartikay/peacock/internal/logs"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)
var errTestSource = errors.New("read failed")

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

	m = applyFilter(t, m, "timeout")

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

func TestMultipleFiltersUseAndSemanticsAndBackspaceRemovesLastFilter(t *testing.T) {
	m := newSizedModel(appconfig.DefaultConfig())
	for _, line := range []string{
		`{"level":"info","message":"database ok","service":"api"}`,
		`{"level":"error","message":"database timeout","service":"api"}`,
		`{"level":"error","message":"cache timeout","service":"worker"}`,
	} {
		m = updateModel(t, m, EntryMsg{Entry: logs.ParseLine(line)})
	}

	m = applyFilter(t, m, "timeout")
	m = applyFilter(t, m, "database")

	filtered := m.filteredEntries(0)
	if got, want := len(filtered), 1; got != want {
		t.Fatalf("expected %d entry matching both filters, got %d", want, got)
	}
	if got := filtered[0].Message.Text; got != "database timeout" {
		t.Fatalf("unexpected filtered message %q", got)
	}

	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeyBackspace})

	filtered = m.filteredEntries(0)
	if got, want := len(filtered), 2; got != want {
		t.Fatalf("expected %d entries after removing last filter, got %d", want, got)
	}
}

func TestEscapeCancelsFilterEntryWithoutAddingFilter(t *testing.T) {
	m := newSizedModel(appconfig.DefaultConfig())
	m = updateModel(t, m, tea.KeyPressMsg{Code: '/', Text: "/"})
	m = typeText(t, m, "timeout")
	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeyEsc})

	if m.filterActive {
		t.Fatal("expected filter mode to close after escape")
	}
	if len(m.filters) != 0 {
		t.Fatalf("expected no filters after cancel, got %v", m.filters)
	}
}

func TestResumeFlushesQueuedEntriesThroughActiveFilters(t *testing.T) {
	m := newSizedModel(appconfig.DefaultConfig())
	m = applyFilter(t, m, "timeout")
	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeySpace})

	m = updateModel(t, m, EntryMsg{Entry: logs.ParseLine(`{"level":"info","message":"health check ok"}`)})
	m = updateModel(t, m, EntryMsg{Entry: logs.ParseLine(`{"level":"error","message":"database timeout"}`)})
	if got := m.queuedEntries.Len(); got != 2 {
		t.Fatalf("expected 2 queued entries while paused, got %d", got)
	}

	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeySpace})

	if m.paused {
		t.Fatal("expected model to resume")
	}
	if got := m.queuedEntries.Len(); got != 0 {
		t.Fatalf("expected queued entries to flush, got %d", got)
	}
	filtered := m.filteredEntries(0)
	if got, want := len(filtered), 1; got != want {
		t.Fatalf("expected %d visible filtered entry after resume, got %d", want, got)
	}
	if got := filtered[0].Message.Text; got != "database timeout" {
		t.Fatalf("unexpected filtered message %q", got)
	}
}

func TestSourceDoneQuitsOnlyInNonFollowMode(t *testing.T) {
	cfg := appconfig.DefaultConfig()
	cfg.Source.FileFollow = false
	nonFollow := newSizedModel(cfg)

	nonFollow, cmd := updateModelWithCmd(t, nonFollow, SourceDoneMsg{})
	if !nonFollow.sourceDone {
		t.Fatal("expected non-follow source to be marked done")
	}
	if cmd == nil {
		t.Fatal("expected non-follow source done to return quit command")
	}

	cfg.Source.FileFollow = true
	follow := newSizedModel(cfg)
	follow, cmd = updateModelWithCmd(t, follow, SourceDoneMsg{})
	if !follow.sourceDone {
		t.Fatal("expected follow source to be marked done")
	}
	if cmd != nil {
		t.Fatal("expected follow source done to keep UI alive")
	}
}

func TestStatusRendersStateSourceErrorsAndFilters(t *testing.T) {
	m := newSizedModel(appconfig.DefaultConfig())
	m.width = 160
	m = applyFilter(t, m, "timeout")
	m = updateModel(t, m, tea.KeyPressMsg{Code: tea.KeySpace})
	m = updateModel(t, m, SourceErrMsg{Err: errTestSource})

	status := stripANSI(m.renderStatus())
	for _, fragment := range []string{"PAUSED", "stdin", "timeout", "read failed"} {
		if !strings.Contains(status, fragment) {
			t.Fatalf("expected status to contain %q, got %q", fragment, status)
		}
	}
}

func newSizedModel(cfg appconfig.Config) model {
	m := NewModel("stdin", cfg).(model)
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

func updateModelWithCmd(t *testing.T, m model, msg tea.Msg) (model, tea.Cmd) {
	t.Helper()

	updated, cmd := m.Update(msg)
	result, ok := updated.(model)
	if !ok {
		t.Fatal("expected concrete model result")
	}
	return result, cmd
}

func applyFilter(t *testing.T, m model, query string) model {
	t.Helper()

	m = updateModel(t, m, tea.KeyPressMsg{Code: '/', Text: "/"})
	m = typeText(t, m, query)
	return updateModel(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
}

func typeText(t *testing.T, m model, text string) model {
	t.Helper()

	for _, char := range text {
		m = updateModel(t, m, tea.KeyPressMsg{Code: rune(char), Text: string(char)})
	}
	return m
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
