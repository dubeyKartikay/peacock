package config

import "fmt"

type Config struct {
	Buffer BufferConfig `mapstructure:"buffer"`
	Input  InputConfig  `mapstructure:"input"`
	Source SourceConfig `mapstructure:"source"`
	Theme  ThemeConfig  `mapstructure:"theme"`
}

type BufferConfig struct {
	MaxEntries int `mapstructure:"max_entries"`
}

type InputConfig struct {
	FilterPrompt              string `mapstructure:"filter_prompt"`
	FilterPlaceholder         string `mapstructure:"filter_placeholder"`
	FilterCharLimit           int    `mapstructure:"filter_char_limit"`
	ScannerInitialBufferBytes int    `mapstructure:"scanner_initial_buffer_bytes"`
	ScannerMaxBufferBytes     int    `mapstructure:"scanner_max_buffer_bytes"`
}

type SourceConfig struct {
	FileFollow    bool `mapstructure:"-"`
	FileTailLines int  `mapstructure:"file_tail_lines"`
	FilePoll      bool `mapstructure:"file_poll"`
	FileReopen    bool `mapstructure:"file_reopen"`
}

type ThemeConfig struct {
	PanelBorder    string `mapstructure:"panel_border"`
	StatusFG       string `mapstructure:"status_fg"`
	StatusBG       string `mapstructure:"status_bg"`
	FilterFG       string `mapstructure:"filter_fg"`
	FilterBG       string `mapstructure:"filter_bg"`
	TimestampFG    string `mapstructure:"timestamp_fg"`
	TimestampFaint bool   `mapstructure:"timestamp_faint"`
	MessageFG      string `mapstructure:"message_fg"`
	CallerFG       string `mapstructure:"caller_fg"`
	CallerFaint    bool   `mapstructure:"caller_faint"`
	ContextFG      string `mapstructure:"context_fg"`
	ContextFaint   bool   `mapstructure:"context_faint"`
	RawFG          string `mapstructure:"raw_fg"`
	LevelError     string `mapstructure:"level_error"`
	LevelWarn      string `mapstructure:"level_warn"`
	LevelInfo      string `mapstructure:"level_info"`
	LevelDebug     string `mapstructure:"level_debug"`
	LevelOther     string `mapstructure:"level_other"`
	LevelBold      bool   `mapstructure:"level_bold"`
}

func DefaultConfig() Config {
	return Config{
		Buffer: BufferConfig{
			MaxEntries: defaultMaxEntries,
		},
		Input: InputConfig{
			FilterPrompt:              defaultFilterPrompt,
			FilterPlaceholder:         defaultFilterPlaceholder,
			FilterCharLimit:           defaultFilterCharLimit,
			ScannerInitialBufferBytes: defaultScannerInitialBufferBytes,
			ScannerMaxBufferBytes:     defaultScannerMaxBufferBytes,
		},
		Source: SourceConfig{
			FileTailLines: defaultFileTailLines,
			FilePoll:      defaultFilePoll,
			FileReopen:    defaultFileReopen,
		},
		Theme: ThemeConfig{
			PanelBorder:    defaultPanelBorder,
			StatusFG:       defaultStatusFG,
			FilterFG:       defaultFilterFG,
			FilterBG:       defaultFilterBG,
			TimestampFG:    defaultTimestampFG,
			TimestampFaint: defaultTimestampFaint,
			MessageFG:      defaultMessageFG,
			CallerFG:       defaultCallerFG,
			CallerFaint:    defaultCallerFaint,
			ContextFG:      defaultContextFG,
			ContextFaint:   defaultContextFaint,
			RawFG:          defaultRawFG,
			LevelError:     defaultLevelError,
			LevelWarn:      defaultLevelWarn,
			LevelInfo:      defaultLevelInfo,
			LevelDebug:     defaultLevelDebug,
			LevelOther:     defaultLevelOther,
			LevelBold:      defaultLevelBold,
		},
	}
}

func (c Config) Validate() error {
	switch {
	case c.Buffer.MaxEntries < 1:
		return fmt.Errorf("buffer.max_entries must be greater than zero")
	case c.Input.FilterCharLimit < 1:
		return fmt.Errorf("input.filter_char_limit must be greater than zero")
	case c.Input.ScannerInitialBufferBytes < 1:
		return fmt.Errorf("input.scanner_initial_buffer_bytes must be greater than zero")
	case c.Input.ScannerMaxBufferBytes < 1:
		return fmt.Errorf("input.scanner_max_buffer_bytes must be greater than zero")
	case c.Input.ScannerInitialBufferBytes > c.Input.ScannerMaxBufferBytes:
		return fmt.Errorf("input.scanner_initial_buffer_bytes must be less than or equal to input.scanner_max_buffer_bytes")
	case c.Source.FileTailLines < 1:
		return fmt.Errorf("source.file_tail_lines must be greater than zero")
	default:
		return nil
	}
}
