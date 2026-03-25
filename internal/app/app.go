package app

import (
	"os"

	tea "charm.land/bubbletea/v2"

	appconfig "peacock/internal/config"
	"peacock/internal/logs"
	"peacock/internal/source"
	tui "peacock/internal/tui"
)

type Options struct {
	Config    appconfig.Config
	InputPath string
	Stdin     *os.File
}

func Run(options Options) error {
	src, err := source.Open(options.InputPath, options.Stdin, options.Config)
	if err != nil {
		return err
	}
	defer src.Close()

	model := tui.NewModel(src.Name(), options.Config)
	program := tea.NewProgram(model)

	go func() {
		for event := range src.Events() {
			switch {
			case event.Line != nil:
				program.Send(tui.EntryMsg{Entry: logs.ParseLine(*event.Line)})
			case event.Err != nil:
				program.Send(tui.SourceErrMsg{Err: event.Err})
			case event.Done:
				program.Send(tui.SourceDoneMsg{})
			}
		}
	}()

	if _, err := program.Run(); err != nil {
		return err
	}

	return nil
}
