// main contains a simple example of using this logger package
package main

import (
	"git.rpjosh.de/RPJosh/go-logger"
)

func main() {
	defer logger.CloseFile()

	// Create a logger configuration
	l := &logger.Logger{
		ColoredOutput: true,
		PrintSource:   true,
		LogFilePath:   "./logs",
		PrintLevel:    logger.LevelTrace,
		LogLevel:      logger.LevelWarning,
	}
	logger.SetGlobalLogger(l)

	// Printing to the different levels
	logger.Trace("You can't find me within %d hours", 5)
	logger.Debug("Im a bunny hunter")
	logger.Info("That should be a feature.\nOf course!")
	logger.Warning("But it would not be safe to use it")
	logger.Error("Now it happend")

	// New logger with the same file to log in
	lOther := &logger.Logger{
		ColoredOutput: false,
		PrintSource:   false,
		PrintLevel:    logger.LevelDebug,
		LogLevel:      logger.LevelDebug,
	}
	logger.NewLoggerWithFile(lOther, logger.GetGlobalLogger())

	lOther.Log(logger.LevelDebug, "Greetings from your brother")
	logger.Info("It's a Me, Mario")
	lOther.Log(logger.LevelError, "And im your brother luigi")

	logger.Fatal("Bowser enters the room...")
}
