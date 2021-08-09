package logger

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type Level = logrus.Level

var (
	TraceLevel = logrus.TraceLevel
	DebugLevel = logrus.DebugLevel
	InfoLevel  = logrus.InfoLevel
	WarnLevel  = logrus.WarnLevel
	ErrorLevel = logrus.ErrorLevel
	FatalLevel = logrus.FatalLevel
	PanicLevel = logrus.PanicLevel
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		PadLevelText:  true,
	})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logrus.SetOutput(os.Stdout)

	/*
		logrus.SetFormatter(&logrus.JSONFormatter{})

		// Only log the warning severity or above.
		logrus.SetLevel(logrus.WarnLevel)

		// to log also the func
		logrus.SetReportCaller(true)

		logger.Info("log this message", fmt.Sprintf("test1:%s",test1text), "test2:test2text")
	*/
}

func Trace(msg string, tags ...string) {
	logrus.WithFields(parseFields(tags...)).Trace(msg)
}

func Debug(msg string, tags ...string) {
	logrus.WithFields(parseFields(tags...)).Debug(msg)
}

func Info(msg string, tags ...string) {
	logrus.WithFields(parseFields(tags...)).Info(msg)
}

func Warn(msg string, tags ...string) {
	logrus.WithFields(parseFields(tags...)).Warn(msg)
}

func Error(msg string, err error, tags ...string) {
	logrus.WithFields(parseFields(tags...)).Error(msg)
}

func Fatal(msg string, tags ...string) {
	// Calls os.Exit(1) after logging
	logrus.WithFields(parseFields(tags...)).Fatal(msg)
}

func Panic(msg string, tags ...string) {
	// Calls panic() after logging
	logrus.WithFields(parseFields(tags...)).Panic(msg)
}

func SetLevel(level Level) {
	logrus.SetLevel(level)
}

func parseFields(tags ...string) logrus.Fields {
	result := make(logrus.Fields, len(tags))
	for _, tag := range tags {
		t := strings.SplitAfterN(tag, ":", 2)
		result[strings.TrimSpace(t[0])] = strings.TrimSpace(t[1])
	}
	return result
}
