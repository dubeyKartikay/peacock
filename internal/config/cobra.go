package config

import (
	flag "github.com/spf13/pflag"
)

const (
	rootUse                    = "peacock [file]"
	rootShort                  = "Pretty JSON log viewer for stdin or tailed files"
	configFlagUsage            = "Path to a custom config file"
	followFlagName             = "follow"
	numberOfLinesFlagName      = "lines"
	numberOfLinesFlagShorthand = "n"
	numberOfLinesFlagUsage     = ""
	followFlagShorthand        = "f"
	followFlagUsage            = "Follow appended lines in file mode"
)

func RegisterFlags(flagSet *flag.FlagSet) {
	flagSet.BoolP(followFlagName, followFlagShorthand, false, followFlagUsage)
	flagSet.IntP(numberOfLinesFlagName, numberOfLinesFlagShorthand, defaultFileTailLines, numberOfLinesFlagUsage)
}

func ReadFlags(cfg *Config, flagSet *flag.FlagSet) (string, bool) {

	fileFollow, err := flagSet.GetBool(followFlagName)
	if err != nil {
		fileFollow = false
	}
	numberOfLines, err := flagSet.GetInt(numberOfLinesFlagName)
	if err != nil {
		numberOfLines = defaultFileTailLines
	}
	cfg.Source.FileFollow = fileFollow
	cfg.Source.FileTailLines = numberOfLines
	return "", false
}
