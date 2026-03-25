package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"peacock/internal/app"
	appconfig "peacock/internal/config"
)

const (
	rootUse             = "peacock [file]"
	rootShort           = "Pretty JSON log viewer for stdin or tailed files"
	configFlagUsage     = "Path to a custom config file"
	followFlagName      = "follow"
	followFlagShorthand = "f"
	followFlagUsage     = "Follow appended lines in file mode"
)

func Execute(stdin *os.File) error {
	return NewRootCommand(stdin).Execute()
}

func NewRootCommand(stdin *os.File) *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:          rootUse,
		Short:        rootShort,
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := appconfig.NewViper(configPath, "")
			if err != nil {
				return err
			}
			cfg, err := appconfig.Load(v)
			if err != nil {
				return err
			}
			appconfig.ReadFlags(&cfg, cmd.Flags())
			var inputPath string
			if len(args) == 1 {
				inputPath = args[0]
			}

			runOptions := app.Options{
				Config:    cfg,
				InputPath: inputPath,
				Stdin:     stdin,
			}

			if err := app.Run(runOptions); err != nil {
				return fmt.Errorf("run peacock: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&configPath, appconfig.FlagConfig, "", configFlagUsage)
	appconfig.RegisterFlags(cmd.Flags())

	return cmd
}
