package app

import (
	"context"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"

	appconfig "github.com/dubeyKartikay/peacock/internal/config"
	"github.com/dubeyKartikay/peacock/internal/source"
	tui "github.com/dubeyKartikay/peacock/internal/tui"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	readSource := readBatch(ctx, src.Events(), options.Config.Input.BatchMaxEntries)
	model := tui.NewModel(src.Name(), options.Config, readSource)
	programOptions := []tea.ProgramOption{}
	if options.InputPath != "" && !options.Config.Source.FileFollow {
		programOptions = append(programOptions, tea.WithEnvironment(nonQueryEnvironment(os.Environ())))
	}
	program := tea.NewProgram(model, programOptions...)

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
