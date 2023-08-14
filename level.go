package logger

import "strings"

// Level of the log message
type Level uint8

const (
	LevelTrace Level = iota
	LevelDebug
	LevelInfo
	LevelWarning
	LevelError
	LevelFatal
)

// String returns a string expression of the level
func (lvl Level) String() string {
	switch lvl {
	case LevelTrace:
		return "TRACE"
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarning:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	}

	return "DEBUG"
}

// GetLevelByName tries to convert the given level name to the represented level code.
// Allowed values are: 'trace', 'debug', 'info', 'warn', 'warning', 'error', 'panic' and 'fatal'
// If an incorrect level name was given a warning is logged and info will be returned
func GetLevelByName(levelName string) Level {
	levelName = strings.ToLower(levelName)
	switch levelName {
	case "trace":
		return LevelTrace
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarning
	case "error":
		return LevelError
	case "panic", "fatal":
		return LevelFatal

	default:
		{
			Warning("Unable to parse the level name '%s'. Expected 'debug', 'info', 'warn', 'error' or 'fatal'", levelName)
			return LevelWarning
		}
	}
}
