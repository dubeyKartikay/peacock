package cli

import (
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/spf13/cobra"

	"github.com/dubeyKartikay/peacock/internal/app"
	appconfig "github.com/dubeyKartikay/peacock/internal/config"
)

// Version is set at build time via -ldflags.
var Version = "dev"

const (
	rootUse             = "peacock [file]"
	rootShort           = "Pretty JSON log viewer for stdin or tailed files"
	configFlagUsage     = "Path to a custom config file"
	followFlagName      = "follow"
	followFlagShorthand = "f"
	followFlagUsage     = "Follow appended lines in file mode"
	cpuProfileFlagName  = "cpuprofile"
	cpuProfileFlagUsage = "Write CPU profile to this file"
)

func Execute(stdin *os.File) error {
	return NewRootCommand(stdin).Execute()
}

func NewRootCommand(stdin *os.File) *cobra.Command {
	var configPath string
	var cpuProfilePath string

	cmd := &cobra.Command{
		Use:          rootUse,
		Short:        rootShort,
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cpuProfilePath != "" {
				f, err := os.Create(cpuProfilePath)
				if err != nil {
					return fmt.Errorf("create cpu profile: %w", err)
				}
				defer f.Close()
				if err := pprof.StartCPUProfile(f); err != nil {
					return fmt.Errorf("start cpu profile: %w", err)
				}
				defer pprof.StopCPUProfile()
			}

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
	cmd.Flags().StringVar(&cpuProfilePath, cpuProfileFlagName, "", cpuProfileFlagUsage)
	appconfig.RegisterFlags(cmd.Flags())

	cmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(Version)
		},
	})

	return cmd
}
