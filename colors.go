package logger

import (
	"os"
)

// ColorConfig contains configuration options to write
// colored text to the console.
type colorConfig struct {
	enableColors bool
}

// Default ANSI color code definitions.
// The variable contains a function that will be padded by the
// matching color. You can also specify replace values after the string
// using printf.
var (
	colPurple      = color("\033[1;35m", "\033[0m")
	colPurpleLight = color("\033[0;35m", "\033[0m")
	colRed         = color("\033[1;31m", "\033[0m")
	colYellow      = color("\033[1;33m", "\033[0m")
	colBlue        = color("\033[1;34m", "\033[0m")
	colBlueLight   = color("\033[0;34m", "\033[0m")
	colCyan        = color("\033[1;36m", "\033[0m")
	colGreen       = color("\033[0;32m", "\033[0m")
)

// Color returns a function that pads the string with the given color code
func color(code, termination string) func(str string) string {
	return func(str string) string {
		return code + str + termination
	}
}

// NewColorConfig prepares and creates a new color config.
// This function could panic because of low level system access
func newColorConfig(enable bool) (conf *colorConfig) {
	conf = &colorConfig{}

	// Validate if ANSI codes are supported by the terminal
	if enable {
		if _, exist := os.LookupEnv("TERMINAL_DISABLE_COLORS"); exist {
			return
		} else if _, exist := os.LookupEnv("TERMINAL_ENABLE_COLORS"); exist {
			conf.enableColors = true
			return
		}

		conf.enableColors = conf.isColoringSupported()
	}

	return
}

// getColor returns the matching color for the level
func (l Level) getColor() func(str string) string {
	switch l {
	case LevelTrace:
		return colPurpleLight
	case LevelDebug:
		return colGreen
	case LevelInfo:
		return colBlue
	case LevelWarning:
		return colYellow
	default:
		return colRed
	}
}
