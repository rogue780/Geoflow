// Package gflog provides a simple logging system for the GeoFlow standard library.
package gflog

import (
	"fmt"
	"os"
	"strings"

	"github.com/rogue780/geoflow/internal/object"
)

// Log levels ordered by severity.
const (
	levelTrace = iota
	levelDebug
	levelInfo
	levelWarn
	levelError
	levelFatal
)

// currentLevel is the package-level log threshold. Messages below this level are suppressed.
var currentLevel = levelInfo

// levelFromString converts a level name to its numeric value.
func levelFromString(s string) (int, bool) {
	switch strings.ToLower(s) {
	case "trace":
		return levelTrace, true
	case "debug":
		return levelDebug, true
	case "info":
		return levelInfo, true
	case "warn":
		return levelWarn, true
	case "error":
		return levelError, true
	case "fatal":
		return levelFatal, true
	default:
		return 0, false
	}
}

// logAtLevel prints a log message to stderr if the given level meets the threshold.
func logAtLevel(level int, label string, args []object.Object) object.Object {
	if level < currentLevel {
		return object.NIL
	}
	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = arg.Inspect()
	}
	msg := strings.Join(parts, " ")
	fmt.Fprintf(os.Stderr, "[%s] %s\n", label, msg)
	return object.NIL
}

// GetExports returns all exported functions for the log module.
func GetExports() map[string]object.Object {
	return map[string]object.Object{
		"trace": &object.Builtin{Name: "log.trace", Fn: trace},
		"debug": &object.Builtin{Name: "log.debug", Fn: debug},
		"info":  &object.Builtin{Name: "log.info", Fn: info},
		"warn":  &object.Builtin{Name: "log.warn", Fn: warn},
		"error": &object.Builtin{Name: "log.error", Fn: logError},
		"fatal": &object.Builtin{Name: "log.fatal", Fn: fatal},
		"setLevel": &object.Builtin{Name: "log.setLevel", Fn: setLevel},
	}
}

func trace(args ...object.Object) object.Object {
	return logAtLevel(levelTrace, "TRACE", args)
}

func debug(args ...object.Object) object.Object {
	return logAtLevel(levelDebug, "DEBUG", args)
}

func info(args ...object.Object) object.Object {
	return logAtLevel(levelInfo, "INFO", args)
}

func warn(args ...object.Object) object.Object {
	return logAtLevel(levelWarn, "WARN", args)
}

func logError(args ...object.Object) object.Object {
	return logAtLevel(levelError, "ERROR", args)
}

func fatal(args ...object.Object) object.Object {
	return logAtLevel(levelFatal, "FATAL", args)
}

func setLevel(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "log.setLevel expects 1 argument (level string)"}
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "log.setLevel: argument must be a string"}
	}
	lvl, valid := levelFromString(s.Value)
	if !valid {
		return &object.Error{
			Message: fmt.Sprintf("log.setLevel: unknown level %q, expected one of: trace, debug, info, warn, error, fatal", s.Value),
		}
	}
	currentLevel = lvl
	return object.NIL
}
