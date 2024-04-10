package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// FileLogger contains configuration options specific to logging into a file.
// It is enabled if the FilePath != ""
type FileLogger struct {

	// Minimum log level for logging into a file
	Level Level

	// Absolute or relative path to log files to
	Path string

	// With this option the path of the log file will be appended with the current date
	// so that an own log file for each day is used. The format of the date is 'YYYYMMDD'
	AppendDate bool

	// Internal dependency used to synchronize the access to the log file
	fileSync *sync.RWMutex
	// Additional file sync that is used during writing to the log file
	fileSyncWrite *sync.RWMutex

	logger *log.Logger
	file   *os.File

	// Upper logger struct
	rootLogger *Logger
}

// CloseFile closes the file that is currently used for logging messages to
// a file
func (l *FileLogger) CloseFile() {
	if l.file != nil {
		l.fileSync.Lock()
		l.file.Close()
		l.file = nil
		l.logger = nil
		l.fileSync.Unlock()
	}
}

// openFile tries to open the file that is configured inside the loggers fild
// "LogFilePath" and initializes the mutex
func (l *FileLogger) openFile() {
	// Initialize new mutex
	if l.fileSync == nil {
		l.fileSync = &sync.RWMutex{}
		l.fileSyncWrite = &sync.RWMutex{}
	}

	l.fileSync.Lock()

	path := l.getFilePath()
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		l.logger = log.New(file, "", 0)
		l.file = file
	} else {
		l.rootLogger.Log(LevelError, fmt.Sprintf("Cannot access the log file '%s'\n%s", path, err.Error()))
	}

	l.fileSync.Unlock()
}

// writeToFile writes the given message to the opened log file
func (l *FileLogger) writeToFile(message string, level Level) {
	l.fileSync.RLock()
	l.fileSyncWrite.RLock()

	// When append date is enabled we need to check if file path is still accurate
	if l.AppendDate {
		currentPath := l.getFilePath()

		if l.file.Name() != currentPath {
			// The file path is not up-to-date anymore → update the log file
			l.fileSync.RUnlock()
			l.fileSyncWrite.RUnlock()

			l.fileSyncWrite.Lock()
			// The syncWriter is now locked. So check again the file name against the current path because the file could already be changed
			// in the time framew between locking and checking
			if l.file.Name() != currentPath {
				l.CloseFile()
				l.openFile()
			}
			l.fileSyncWrite.Unlock()

			// Lock previous locks again
			l.fileSync.RLock()
			l.fileSyncWrite.RLock()
		}
	}

	l.logger.Println(message)
	l.file.Sync()

	l.fileSync.RUnlock()
	l.fileSyncWrite.RUnlock()

	// Close the file because for fatal log level the program is going to be exited
	if level == LevelFatal {
		l.CloseFile()
	}
}

// getFilePath returns the path to use for the log file
func (l *FileLogger) getFilePath() string {
	path := strings.ReplaceAll(l.Path, "\\", "/")

	// Append the current date to the log path when enabled
	if l.AppendDate {
		lastSlash := strings.LastIndex(path, "/")
		if lastSlash != -1 && (lastSlash+1) < len(path) {
			path = path + "." + getFileDate()
		} else if lastSlash == -1 {
			path = path + "." + getFileDate()
		} else {
			path += getFileDate()
		}
	}

	return path
}

// getFileDate returns the current date formatted as the log files path name
func getFileDate() string {
	return time.Now().Format("2006-01-02")
}
