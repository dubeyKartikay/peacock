package app

import (
	"os"
	"strings"

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
	src, err := source.Open(options.InputPath, options.Stdin, &options.Config)
	if err != nil {
		return err
	}
	defer src.Close()

	model := tui.NewModel(src.Name(), options.Config)
	programOptions := []tea.ProgramOption{}
	if options.InputPath != "" && !options.Config.Source.FileFollow {
		programOptions = append(programOptions, tea.WithEnvironment(nonQueryEnvironment(os.Environ())))
	}
	program := tea.NewProgram(model, programOptions...)

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

func nonQueryEnvironment(base []string) []string {
	env := make([]string, 0, len(base)+2)
	for _, variable := range base {
		if strings.HasPrefix(variable, "TERM=") ||
			strings.HasPrefix(variable, "TERM_PROGRAM=") ||
			strings.HasPrefix(variable, "WT_SESSION=") ||
			strings.HasPrefix(variable, "SSH_TTY=") {
			continue
		}
		env = append(env, variable)
	}

	env = append(env,
		"TERM=xterm-256color",
		"TERM_PROGRAM=Apple_Terminal",
	)

	return env
}
