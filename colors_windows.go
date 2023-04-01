package logger

import (
	"os"

	"golang.org/x/sys/windows"
)

func (c colorConfig) isColoringSupported() bool {

	// In cmd ANSI colors are not supported by default from the beggining on (>16257) → enable explicit support via
	// the flag ENABLE_VIRTUAL_TERMINAL_PROCESSING
	stdout := windows.Handle(os.Stdout.Fd())
	var originalMode uint32

	if windows.GetConsoleMode(stdout, &originalMode) == nil {
		if windows.SetConsoleMode(stdout, originalMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING) == nil {
			return true
		}
	}

	return false
}
