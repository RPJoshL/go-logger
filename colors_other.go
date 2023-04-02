//go:build !unix && !windows

package logger

func (c colorConfig) isColoringSupported() bool {
	return false
}
