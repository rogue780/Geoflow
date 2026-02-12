// Package gflog provides a simple logging system for the GeoFlow standard library.
package gflog

import (
	"fmt"
	"os"
	"strings"
	"time"

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

// currentFormat is the package-level log format string.
var currentFormat = "[%level] %message"

// currentOutput is the name of the current output target (for stub purposes).
var currentOutput = "stderr"

// additionalOutputs holds extra output target names (for stub purposes).
var additionalOutputs []string

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

// formatStructuredData formats a Map of structured data as key=value pairs.
func formatStructuredData(m *object.Map) string {
	if len(m.Pairs) == 0 {
		return ""
	}
	parts := make([]string, len(m.Pairs))
	for i, p := range m.Pairs {
		parts[i] = fmt.Sprintf("%s=%s", p.Key.Inspect(), p.Value.Inspect())
	}
	return strings.Join(parts, " ")
}

// logAtLevel prints a log message to stderr if the given level meets the threshold.
// It supports an optional trailing Map argument for structured data.
func logAtLevel(level int, label string, args []object.Object) object.Object {
	if level < currentLevel {
		return object.NIL
	}

	// Separate message args from optional trailing structured data Map.
	var msgArgs []object.Object
	var structuredData *object.Map

	if len(args) > 0 {
		if m, ok := args[len(args)-1].(*object.Map); ok && len(args) > 1 {
			structuredData = m
			msgArgs = args[:len(args)-1]
		} else {
			msgArgs = args
		}
	}

	parts := make([]string, len(msgArgs))
	for i, arg := range msgArgs {
		parts[i] = arg.Inspect()
	}
	msg := strings.Join(parts, " ")

	// Build the formatted output from the format string.
	output := currentFormat
	output = strings.ReplaceAll(output, "%level", label)
	output = strings.ReplaceAll(output, "%message", msg)

	if structuredData != nil {
		output += " " + formatStructuredData(structuredData)
	}

	fmt.Fprintf(os.Stderr, "%s\n", output)
	return object.NIL
}

// GetExports returns all exported functions for the log module.
func GetExports() map[string]object.Object {
	return map[string]object.Object{
		"trace":      &object.Builtin{Name: "log.trace", Fn: trace},
		"debug":      &object.Builtin{Name: "log.debug", Fn: debug},
		"info":       &object.Builtin{Name: "log.info", Fn: info},
		"warn":       &object.Builtin{Name: "log.warn", Fn: warn},
		"error":      &object.Builtin{Name: "log.error", Fn: logError},
		"fatal":      &object.Builtin{Name: "log.fatal", Fn: fatal},
		"setLevel":   &object.Builtin{Name: "log.setLevel", Fn: setLevel},
		"setFormat":  &object.Builtin{Name: "log.setFormat", Fn: setFormat},
		"setOutput":  &object.Builtin{Name: "log.setOutput", Fn: setOutput},
		"addOutput":  &object.Builtin{Name: "log.addOutput", Fn: addOutput},
		"withContext": &object.Builtin{Name: "log.withContext", Fn: withContext},
		"timed":      &object.Builtin{Name: "log.timed", Fn: timed},
		"debugIf":    &object.Builtin{Name: "log.debugIf", Fn: debugIf},
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

func setFormat(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "log.setFormat expects 1 argument (format string)"}
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "log.setFormat: argument must be a string"}
	}
	currentFormat = s.Value
	return object.NIL
}

func setOutput(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "log.setOutput expects 1 argument (output target name)"}
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "log.setOutput: argument must be a string"}
	}
	currentOutput = s.Value
	_ = currentOutput // stub: store but don't act on it
	return object.NIL
}

func addOutput(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "log.addOutput expects 1 argument (output target name)"}
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "log.addOutput: argument must be a string"}
	}
	additionalOutputs = append(additionalOutputs, s.Value)
	return object.NIL
}

func withContext(args ...object.Object) object.Object {
	if len(args) == 0 {
		return &object.Error{Message: "log.withContext expects at least 1 argument (key-value pairs or a Map)"}
	}
	// If a single Map is passed, return it directly as the context.
	if len(args) == 1 {
		if m, ok := args[0].(*object.Map); ok {
			return m
		}
	}
	// Otherwise, treat args as alternating key-value pairs.
	if len(args)%2 != 0 {
		return &object.Error{Message: "log.withContext: expected even number of arguments (key-value pairs)"}
	}
	pairs := make([]object.MapPair, 0, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		pairs = append(pairs, object.MapPair{Key: args[i], Value: args[i+1]})
	}
	return &object.Map{Pairs: pairs}
}

func timed(args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 {
		return &object.Error{Message: "log.timed expects 1-2 arguments (function, label?)"}
	}
	label := "timed"
	if len(args) == 2 {
		if s, ok := args[1].(*object.String); ok {
			label = s.Value
		} else if s, ok := args[0].(*object.String); ok {
			// Allow timed("label", fn) ordering too
			label = s.Value
			args[0] = args[1]
		}
	}
	fn := args[0]
	start := time.Now()
	result := object.CallFunction(fn)
	elapsed := time.Since(start)

	// Log the elapsed time at INFO level.
	logAtLevel(levelInfo, "TIMED", []object.Object{
		&object.String{Value: fmt.Sprintf("%s completed in %s", label, elapsed)},
	})

	return result
}

func debugIf(args ...object.Object) object.Object {
	if len(args) < 2 {
		return &object.Error{Message: "log.debugIf expects at least 2 arguments (condition, message, ...)"}
	}
	if !object.IsTruthy(args[0]) {
		return object.NIL
	}
	return logAtLevel(levelDebug, "DEBUG", args[1:])
}
