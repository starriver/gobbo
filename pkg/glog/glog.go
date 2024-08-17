package glog

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

// Simple color logger, outputs to stderr.

type Level int

const (
	LevelNone Level = iota - 1
	LevelError
	LevelWarn
	LevelInfo
	LevelDebug
)

var levelData = []struct {
	str   string
	color color.Attribute
}{
	{"error", color.FgHiRed},
	{"warn", color.FgHiYellow},
	{"info", color.FgHiGreen},
	{"debug", color.FgHiCyan},
}

// No log output by default. This is changed by main() immediately.
var CurrentLevel = LevelNone

// Set the level from a string. Noop if an invalid level is supplied.
func ParseLevel(str string) error {
	for i, data := range levelData {
		if str != data.str {
			continue
		}

		CurrentLevel = Level(i)
		return nil
	}

	return fmt.Errorf("invalid log level")
}

// Log an error.
func Error(e ...any) {
	log(LevelError, e...)
}

// Log an error.
func Errorf(format string, a ...any) {
	logf(LevelError, format, a...)
}

// Log a warning.
func Warn(a ...any) {
	log(LevelWarn, a...)
}

// Log a warning.
func Warnf(format string, a ...any) {
	logf(LevelWarn, format, a...)
}

// Log an informational message.
func Info(a ...any) {
	log(LevelInfo, a...)
}

// Log an informational message.
func Infof(format string, a ...any) {
	logf(LevelInfo, format, a...)
}

// Log a debugging message.
func Debug(a ...any) {
	log(LevelDebug, a...)
}

// Log a debugging message.
func Debugf(format string, a ...any) {
	logf(LevelDebug, format, a...)
}

func prefix(level Level) {
	data := levelData[level]
	fmt.Fprintf(
		os.Stderr,
		"[%s] ",
		color.New(data.color).Sprint(data.str),
	)
}

func log(level Level, a ...any) {
	if level > CurrentLevel {
		return
	}

	prefix(level)
	fmt.Fprintln(os.Stderr, a...)
}

func logf(level Level, format string, a ...any) {
	if level > CurrentLevel {
		return
	}

	s := fmt.Sprintf(format, a...)
	for _, line := range strings.Split(s, "\n") {
		prefix(level)
		fmt.Fprintln(os.Stderr, line)
	}
}
