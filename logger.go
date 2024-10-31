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

type Logger struct {

	// Minimum log level for printing to the console (stdout and stderr)
	Level Level

	// Colorizes the log messages for the console.
	// Even if you set this to true the user is able to overwrite this behaviour by
	// setting the environment variables "TERMINAL_DISABLE_COLORS" and
	// "TERMINAL_ENABLE_COLORS" (to force coloring for "unsupported" terminals)
	ColoredOutput bool

	// Whether to print the file and line number of the invoking (calling line)
	PrintSource bool

	// Only print the log message without any additional info. This property will ignore other options linke
	// PrintSource or FuncCallIncrement
	OnlyPrintMessage bool

	// While logging, the file and line number of the
	// invoking (calling) line can be printed out.
	// This defines an offset that is applied to the call stack.
	// If you are using an own wrapper function, you
	// have to set this value to one
	FuncCallIncrement int

	// Prefix is applied as a prefix for all log messages.
	// It's positioned after all other information:
	//  [INFO ] 2024-04-10 19:00:00 (file:1)PREFIX - Message
	Prefix string

	// Configuration options for logging into a file
	File *FileLogger

	colorConf        colorConfig
	consoleLogger    *log.Logger
	consoleLoggerErr *log.Logger
}

// Globally available logging instance. This will be uesed if log functions
// without a Logger struct are called
var dLogger Logger

func init() {
	dLogger = Logger{
		Level: LevelDebug,
		File: &FileLogger{
			Level: LevelInfo,
			Path:  "",
		},
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
// This enables you to write to the same file with different log configurations.
func NewLoggerWithFile(logger *Logger, file *Logger) *Logger {
	logger.File.file = file.File.file
	logger.File.Path = file.File.Path
	logger.File.logger = file.File.logger
	logger.File.fileSync = file.File.fileSync
	logger.File.fileSyncWrite = file.File.fileSyncWrite

	logger.setup(true)
	return logger
}

// CloneLogger creates a copy of the provided logger with it's
// file reference.
// All configuration options are cloned from "logger" to the new one
func CloneLogger(logger *Logger) *Logger {
	// Copy by dereference the pointer
	copyIn := *logger
	copy := &copyIn
	fileIn := *copy.File
	file := &fileIn

	copy.File = file
	copy.consoleLogger = nil
	copy.consoleLoggerErr = nil

	return NewLoggerWithFile(copy, logger)
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
	var levelName = fmt.Sprintf("%-5s", level)

	// Build the message to print
	printMessage := message
	if len(parameters) > 0 {
		printMessage = fmt.Sprintf(message, parameters...)
	}
	if !l.OnlyPrintMessage {
		printMessage = "[" + levelName + "] " + time.Now().Local().Format("2006-01-02 15:04:05") +
			getSourceMessage(file, line, pc, l) + l.Prefix + " - " + printMessage
	}

	// Build the colored message to print
	printMessageColored := l.getColored(printMessage, level.getColor())
	if !l.OnlyPrintMessage {
		printMessageColored =
			l.getColored("["+levelName+"] ", level.getColor()) +
				l.getColored(time.Now().Local().Format("2006-01-02 15:04:05"), colCyan) +
				l.getColored(getSourceMessage(file, line, pc, l), colPurple) +
				l.getColored(l.Prefix, colBlueLight) +
				" - " + printMessageColored
	}

	if l.File.Level <= level && l.File.logger != nil {
		l.File.writeToFile(printMessage, level)
	}

	if l.Level <= level {
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
func (l *Logger) getColored(message string, color func(str string) string) string {
	if l.colorConf.enableColors {
		return color(message)
	}
	return message
}

func getSourceMessage(file string, line int, _ uintptr, l *Logger) string {
	if !l.PrintSource {
		return ""
	}

	fileName := file[strings.LastIndex(file, "/")+1:] + ":" + strconv.Itoa(line)

	return " (" + fileName + ")"
}

// setup setups the provided logger.
// This function has to be called before you can use the logger
// struct!
func (l *Logger) setup(keepFile bool) {

	// Setup reference for file logger
	l.File.rootLogger = l

	// log.Ldate|log.Ltime|log.Lshortfile
	l.consoleLogger = log.New(os.Stdout, "", 0)
	l.consoleLoggerErr = log.New(os.Stderr, "", 0)

	if strings.TrimSpace(l.File.Path) != "" && !keepFile {
		l.File.openFile()
	} else if !keepFile {
		l.File.CloseFile()
	}

	// Functions that could produce a panic
	defer func() {
		if err := recover(); err != nil {
			l.log(LevelDebug, "Panic occured: %s", err)
		}
	}()
	l.colorConf = *newColorConfig(l.ColoredOutput)
}

// SetGlobalLogger updates the global default logger with a custom one.
// You can create one via the Logger struct.
func SetGlobalLogger(l *Logger) {
	dLogger = *l // nolint: golint
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

// Available methods for each logger per logging level

func (l *Logger) Trace(message string, parameters ...any) {
	l.Log(LevelTrace, message, parameters...)
}
func (l *Logger) Debug(message string, parameters ...any) {
	l.Log(LevelDebug, message, parameters...)
}
func (l *Logger) Info(message string, parameters ...any) {
	l.Log(LevelInfo, message, parameters...)
}
func (l *Logger) Warning(message string, parameters ...any) {
	l.Log(LevelWarning, message, parameters...)
}
func (l *Logger) Error(message string, parameters ...any) {
	l.Log(LevelError, message, parameters...)
}
func (l *Logger) Fatal(message string, parameters ...any) {
	l.Log(LevelFatal, message, parameters...)
}

// CloseFile closes the underlaying file to which the logger messages are written.
func CloseFile() {
	dLogger.File.CloseFile()
}

// GetLoggerFromEnv returns a logging instance configured
// from the available environment variables.
//
// The environment variables have to be named like the struct
// fields in upper case with the prefix "LOGGER_".
// Sub structs are divided also by an underscore. Example:
// "LOGGER_SUBCONFIG_DISABLED"
//
// If no env variable was found the default value of the given
// logger struct will be used.
//
// Note that only generic options can be set like:
// - Print and Log Level
// - Log path
// - ColoredOutput
// - Tracing disabled
func GetLoggerFromEnv(defaultLogger *Logger) *Logger {
	defaultLogger.ColoredOutput = getEnvBool("LOGGER_COLOREDOUTPUT", defaultLogger.ColoredOutput)
	defaultLogger.Level = GetLevelByName(getEnvString("LOGGER_LEVEL", defaultLogger.Level.String()))
	defaultLogger.OnlyPrintMessage = getEnvBool("LOGGER_ONLYPRINTMESSAGE", defaultLogger.OnlyPrintMessage)
	defaultLogger.File.Level = GetLevelByName(getEnvString("LOGGER_FILE_LEVEL", defaultLogger.File.Level.String()))
	defaultLogger.File.Path = getEnvString("LOGGER_FILE_PATH", defaultLogger.File.Path)
	defaultLogger.File.AppendDate = getEnvBool("LOGGER_FILE_APPENDDATE", defaultLogger.File.AppendDate)
	defaultLogger.PrintSource = getEnvBool("LOGGER_PRINTSOURCE", defaultLogger.PrintSource)
	return NewLogger(defaultLogger)
}
