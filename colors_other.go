//go:build !unix

package logger

func (c colorConfig) isColoringSupported() bool {
	return false
}
