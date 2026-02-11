// Package gftime provides date/time functions for the GeoFlow standard library.
package gftime

import (
	"fmt"
	"time"

	"github.com/rogue780/geoflow/internal/object"
)

// GetExports returns all exported functions for the time module.
func GetExports() map[string]object.Object {
	return map[string]object.Object{
		"now":      &object.Builtin{Name: "time.now", Fn: now},
		"parse":    &object.Builtin{Name: "time.parse", Fn: parseTime},
		"fromUnix": &object.Builtin{Name: "time.fromUnix", Fn: fromUnix},
		"date":     &object.Builtin{Name: "time.date", Fn: date},
		"datetime": &object.Builtin{Name: "time.datetime", Fn: datetime},
		"seconds":  &object.Builtin{Name: "time.seconds", Fn: seconds},
		"minutes":  &object.Builtin{Name: "time.minutes", Fn: minutes},
		"hours":    &object.Builtin{Name: "time.hours", Fn: hours},
		"days":     &object.Builtin{Name: "time.days", Fn: days},
		"millis":   &object.Builtin{Name: "time.millis", Fn: millis},
		"since":    &object.Builtin{Name: "time.since", Fn: since},
		"until":    &object.Builtin{Name: "time.until", Fn: until},
	}
}

// geoflowFormatToGo converts common format specifiers to Go's reference time.
// Supported patterns:
//
//	"RFC3339"    -> time.RFC3339
//	"RFC822"     -> time.RFC822
//	"UnixDate"   -> time.UnixDate
//	"YYYY-MM-DD" -> "2006-01-02"
//	"YYYY-MM-DD HH:mm:ss" -> "2006-01-02 15:04:05"
//
// Otherwise the string is passed through as a Go time layout.
func geoflowFormatToGo(format string) string {
	switch format {
	case "RFC3339":
		return time.RFC3339
	case "RFC822":
		return time.RFC822
	case "UnixDate":
		return time.UnixDate
	case "YYYY-MM-DD":
		return "2006-01-02"
	case "YYYY-MM-DD HH:mm:ss":
		return "2006-01-02 15:04:05"
	default:
		return format
	}
}

func now(args ...object.Object) object.Object {
	if len(args) != 0 {
		return &object.Error{Message: "time.now expects 0 arguments"}
	}
	return &object.DateTime{Value: time.Now()}
}

func parseTime(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "time.parse expects 2 arguments (string, format)"}
	}
	s, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "time.parse: first argument must be a string"}
	}
	format, ok := args[1].(*object.String)
	if !ok {
		return &object.Error{Message: "time.parse: second argument must be a format string"}
	}
	goLayout := geoflowFormatToGo(format.Value)
	t, err := time.Parse(goLayout, s.Value)
	if err != nil {
		return &object.Result{
			Value: &object.String{Value: fmt.Sprintf("time.parse: %s", err.Error())},
			IsOk:  false,
		}
	}
	return &object.Result{
		Value: &object.DateTime{Value: t},
		IsOk:  true,
	}
}

func fromUnix(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "time.fromUnix expects 1 argument (seconds)"}
	}
	var sec int64
	switch v := args[0].(type) {
	case *object.Integer:
		sec = v.Value
	case *object.Float:
		sec = int64(v.Value)
	default:
		return &object.Error{Message: "time.fromUnix: argument must be a number"}
	}
	return &object.DateTime{Value: time.Unix(sec, 0).UTC()}
}

func date(args ...object.Object) object.Object {
	if len(args) != 3 {
		return &object.Error{Message: "time.date expects 3 arguments (year, month, day)"}
	}
	year, ok := args[0].(*object.Integer)
	if !ok {
		return &object.Error{Message: "time.date: year must be an integer"}
	}
	month, ok := args[1].(*object.Integer)
	if !ok {
		return &object.Error{Message: "time.date: month must be an integer"}
	}
	day, ok := args[2].(*object.Integer)
	if !ok {
		return &object.Error{Message: "time.date: day must be an integer"}
	}
	t := time.Date(int(year.Value), time.Month(month.Value), int(day.Value), 0, 0, 0, 0, time.UTC)
	return &object.DateTime{Value: t}
}

func datetime(args ...object.Object) object.Object {
	if len(args) != 6 {
		return &object.Error{Message: "time.datetime expects 6 arguments (year, month, day, hour, min, sec)"}
	}
	ints := make([]int64, 6)
	names := []string{"year", "month", "day", "hour", "min", "sec"}
	for i := 0; i < 6; i++ {
		v, ok := args[i].(*object.Integer)
		if !ok {
			return &object.Error{Message: fmt.Sprintf("time.datetime: %s must be an integer", names[i])}
		}
		ints[i] = v.Value
	}
	t := time.Date(int(ints[0]), time.Month(ints[1]), int(ints[2]),
		int(ints[3]), int(ints[4]), int(ints[5]), 0, time.UTC)
	return &object.DateTime{Value: t}
}

func toDurationArg(name string, args []object.Object) (float64, *object.Error) {
	if len(args) != 1 {
		return 0, &object.Error{Message: fmt.Sprintf("time.%s expects 1 argument", name)}
	}
	switch v := args[0].(type) {
	case *object.Integer:
		return float64(v.Value), nil
	case *object.Float:
		return v.Value, nil
	default:
		return 0, &object.Error{Message: fmt.Sprintf("time.%s: argument must be a number", name)}
	}
}

func seconds(args ...object.Object) object.Object {
	n, err := toDurationArg("seconds", args)
	if err != nil {
		return err
	}
	return &object.Duration{Value: time.Duration(n * float64(time.Second))}
}

func minutes(args ...object.Object) object.Object {
	n, err := toDurationArg("minutes", args)
	if err != nil {
		return err
	}
	return &object.Duration{Value: time.Duration(n * float64(time.Minute))}
}

func hours(args ...object.Object) object.Object {
	n, err := toDurationArg("hours", args)
	if err != nil {
		return err
	}
	return &object.Duration{Value: time.Duration(n * float64(time.Hour))}
}

func days(args ...object.Object) object.Object {
	n, err := toDurationArg("days", args)
	if err != nil {
		return err
	}
	return &object.Duration{Value: time.Duration(n * 24 * float64(time.Hour))}
}

func millis(args ...object.Object) object.Object {
	n, err := toDurationArg("millis", args)
	if err != nil {
		return err
	}
	return &object.Duration{Value: time.Duration(n * float64(time.Millisecond))}
}

func since(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "time.since expects 1 argument (DateTime)"}
	}
	dt, ok := args[0].(*object.DateTime)
	if !ok {
		return &object.Error{Message: "time.since: argument must be a DateTime"}
	}
	return &object.Duration{Value: time.Since(dt.Value)}
}

func until(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "time.until expects 1 argument (DateTime)"}
	}
	dt, ok := args[0].(*object.DateTime)
	if !ok {
		return &object.Error{Message: "time.until: argument must be a DateTime"}
	}
	return &object.Duration{Value: time.Until(dt.Value)}
}
