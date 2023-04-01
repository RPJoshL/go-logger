package logger

import "os"

func (c colorConfig) isColoringSupported() bool {
	// Check if $TERM variable is set. Almost every terminal does support coloring in linux
	return os.Getenv("TERM") != ""
}
