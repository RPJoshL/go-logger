// logger provides basic logging support for your application.
// Supported log destinations are the console and a log file
package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

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

type Logger struct {
	PrintLevel  Level
	LogLevel    Level
	LogFilePath string
	PrintSource bool

	// Colorizes the log messages for the console.
	// Even if you set this to true, the user is able to overwrite this behaviour by
	// setting the environment variables "TERMINAL_DISABLE_COLORS" and
	// "TERMINAL_ENABLE_COLORS" (to force coloring for "unsupported" terminals)
	ColoredOutput bool

	// While logging, the file and line number of the
	// invoking (calling) line can be printed out.
	// This defines an offset that is applied to the call stack.
	// If you you are using an own wrapper function, you
	// have to set this value to one
	FuncCallIncrement int

	colorConf        colorConfig
	consoleLogger    *log.Logger
	consoleLoggerErr *log.Logger
	fileLogger       *log.Logger
	logFile          *os.File
}

// Globally available logging instance. This will be uesed if log functions
// without a Logger struct are called
var dLogger Logger

func init() {
	dLogger = Logger{
		PrintLevel:  LevelDebug,
		LogLevel:    LevelInfo,
		LogFilePath: "",
		PrintSource: false,
	}

	dLogger.setup(false)
}

// NewLogger creates a new instance of the logger with
// the given configuration.
func NewLogger(logger *Logger) *Logger {
	logger.setup(false)
	return logger
}

// NewLoggerWithFile creates a new instance with the given logger
// configuration.
// Instead of opening a new file to write the log messages to,
// the old file reference of the other logger will be used internal.
// This enables you to writhe to the same file with different log configurations.
func NewLoggerWithFile(logger *Logger, file *Logger) *Logger {
	logger.logFile = file.logFile
	logger.LogFilePath = file.LogFilePath
	logger.fileLogger = file.fileLogger

	logger.setup(true)
	return logger
}

// Log logs a message with the given level. As additional parameters you can specify
// replace values for the message. See "fmt.printf()" for more infos.
func (l *Logger) Log(level Level, message string, parameters ...any) {
	// This function is needed that "runtime.Caller(2)" is always correct (even on direct call)
	l.log(level, message, parameters...)
}

func (l *Logger) log(level Level, message string, parameters ...any) {
	pc, file, line, ok := runtime.Caller(3 + l.FuncCallIncrement)
	if !ok {
		file = "#unknown"
		line = 0
	}

	// Get the name of the level to log
	var levelName string
	switch level {
	case LevelTrace:
		levelName = "TRACE"
	case LevelDebug:
		levelName = "DEBUG"
	case LevelInfo:
		levelName = "INFO "
	case LevelWarning:
		levelName = "WARN "
	case LevelError:
		levelName = "ERROR"
	case LevelFatal:
		levelName = "FATAL"
	}

	if levelName == "" {
		message = fmt.Sprintf("Invalid level value given: %d. Original message: ", level) + message
		levelName = "WARN "
		level = LevelWarning
	}

	printMessage := "[" + levelName + "] " + time.Now().Local().Format("2006-01-02 15:04:05") +
		getSourceMessage(file, line, pc, *l) + " - " + fmt.Sprintf(message, parameters...)

	printMessageColored :=
		l.getColored("["+levelName+"] ", level.getColor()) +
			l.getColored(time.Now().Local().Format("2006-01-02 15:04:05"), colCyan) +
			l.getColored(getSourceMessage(file, line, pc, *l), colPurple) + " - " +
			l.getColored(fmt.Sprintf(message, parameters...), level.getColor())

	if l.LogLevel <= level && l.fileLogger != nil {
		l.fileLogger.Println(printMessage)
		l.logFile.Sync()

		if level == LevelFatal {
			l.CloseFile()
		}
	}

	if l.PrintLevel <= level {
		if level == LevelError {
			l.consoleLoggerErr.Println(printMessageColored)
		} else if level == LevelFatal {
			l.consoleLoggerErr.Fatal(printMessageColored)
		} else {
			l.consoleLogger.Println(printMessageColored)
		}
	}

}

// getColored returns a message padded by with a color code if coloring is supported and specified
func (l *Logger) getColored(message string, color func(str string, parameters ...any) string) string {
	if l.colorConf.enableColors {
		return color(message)
	}
	return message
}

func getSourceMessage(file string, line int, pc uintptr, l Logger) string {
	if !l.PrintSource {
		return ""
	}

	fileName := file[strings.LastIndex(file, "/")+1:] + ":" + strconv.Itoa(line)

	return " (" + fileName + ")"
}

func (l *Logger) setup(keepFile bool) {
	// log.Ldate|log.Ltime|log.Lshortfile
	l.consoleLogger = log.New(os.Stdout, "", 0)
	l.consoleLoggerErr = log.New(os.Stderr, "", 0)

	if strings.TrimSpace(l.LogFilePath) != "" && !keepFile {
		file, err := os.OpenFile(l.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			l.fileLogger = log.New(file, "", 0)
			l.logFile = file
		} else {
			l.Log(LevelError, fmt.Sprintf("Cannot access the log file '%s'\n%s", l.LogFilePath, err.Error()))
		}
	} else if !keepFile {
		l.fileLogger = nil
		if l.logFile != nil {
			l.logFile.Close()
			l.logFile = nil
		}
	}

	// Functions that could produce a panic
	defer func() {
		if err := recover(); err != nil {
			l.log(LevelDebug, "Panic occured: %s", err)
		}
	}()
	l.colorConf = *newColorConfig(l.ColoredOutput)
}

func (l *Logger) CloseFile() {
	if l.logFile != nil {
		l.logFile.Close()
		l.logFile = nil
		l.fileLogger = nil
	}
}

// SetGlobalLogger updates the global default logger with a custom one.
// You can create one via the Logger struct.
func SetGlobalLogger(l *Logger) {
	dLogger = *l
	dLogger.setup(false)
}
func GetGlobalLogger() *Logger {
	return &dLogger
}

// Global available methods per logging levels //

func Trace(message string, parameters ...any) {
	dLogger.Log(LevelTrace, message, parameters...)
}
func Debug(message string, parameters ...any) {
	dLogger.Log(LevelDebug, message, parameters...)
}
func Info(message string, parameters ...any) {
	dLogger.Log(LevelInfo, message, parameters...)
}
func Warning(message string, parameters ...any) {
	dLogger.Log(LevelWarning, message, parameters...)
}
func Error(message string, parameters ...any) {
	dLogger.Log(LevelError, message, parameters...)
}
func Fatal(message string, parameters ...any) {
	dLogger.Log(LevelFatal, message, parameters...)
}

// Available methods for each logger per logging level //

func (l Logger) Trace(message string, parameters ...any) {
	l.Log(LevelTrace, message, parameters...)
}
func (l Logger) Debug(message string, parameters ...any) {
	l.Log(LevelDebug, message, parameters...)
}
func (l Logger) Info(message string, parameters ...any) {
	l.Log(LevelInfo, message, parameters...)
}
func (l Logger) Warning(message string, parameters ...any) {
	l.Log(LevelWarning, message, parameters...)
}
func (l Logger) Error(message string, parameters ...any) {
	l.Log(LevelError, message, parameters...)
}
func (l Logger) Fatal(message string, parameters ...any) {
	l.Log(LevelFatal, message, parameters...)
}

// CloseFile closes the underlaying file to which the logger messages are written.
func CloseFile() {
	dLogger.CloseFile()
}

// GetLevelByName tries to convert the given level name to the represented level code.
// Allowed values are: 'trace', 'debug', 'info', 'warn', 'warning', 'error', 'panic' and 'fatal'
// If an incorrect level name was given an warning is logged and info will be returned
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
			return LevelInfo
		}
	}
}
