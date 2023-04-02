package logger

import (
	"fmt"
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
	colPurple      = color("\033[1;35m%s\033[0m")
	colPurpleLight = color("\033[0;35m%s\033[0m")
	colRed         = color("\033[1;31m%s\033[0m")
	colYellow      = color("\033[1;33m%s\033[0m")
	colBlue        = color("\033[1;34m%s\033[0m")
	colCyan        = color("\033[1;36m%s\033[0m")
	colGreen       = color("\033[0;32m%s\033[0m")
)

// Color returns a function that pads the string with the given color code
func color(colorString string) func(str string, parameters ...any) string {
	return func(str string, parameters ...any) string {
		return fmt.Sprintf(colorString, fmt.Sprintf(str, parameters...))
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
func (l Level) getColor() func(str string, parameters ...any) string {
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
