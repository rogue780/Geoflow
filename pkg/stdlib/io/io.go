// Package io provides file and path I/O functions for the GeoFlow standard library.
package io

import (
	"fmt"
	goio "io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rogue780/geoflow/internal/object"
)

// GetExports returns all exported functions for the io module.
func GetExports() map[string]object.Object {
	return map[string]object.Object{
		"readFile":     &object.Builtin{Name: "io.readFile", Fn: readFile},
		"readLines":    &object.Builtin{Name: "io.readLines", Fn: readLines},
		"writeFile":    &object.Builtin{Name: "io.writeFile", Fn: writeFile},
		"appendFile":   &object.Builtin{Name: "io.appendFile", Fn: appendFile},
		"exists":       &object.Builtin{Name: "io.exists", Fn: exists},
		"isFile":       &object.Builtin{Name: "io.isFile", Fn: isFile},
		"isDir":        &object.Builtin{Name: "io.isDir", Fn: isDir},
		"fileSize":     &object.Builtin{Name: "io.fileSize", Fn: fileSize},
		"listDir":      &object.Builtin{Name: "io.listDir", Fn: listDir},
		"createDir":    &object.Builtin{Name: "io.createDir", Fn: createDir},
		"createDirAll": &object.Builtin{Name: "io.createDirAll", Fn: createDirAll},
		"removeFile":   &object.Builtin{Name: "io.removeFile", Fn: removeFile},
		"copy":         &object.Builtin{Name: "io.copy", Fn: copyFile},
		"move":         &object.Builtin{Name: "io.move", Fn: moveFile},
		"join":         &object.Builtin{Name: "io.join", Fn: joinPath},
		"dirname":      &object.Builtin{Name: "io.dirname", Fn: dirname},
		"basename":     &object.Builtin{Name: "io.basename", Fn: basename},
		"extension":    &object.Builtin{Name: "io.extension", Fn: extension},
		"absolute":     &object.Builtin{Name: "io.absolute", Fn: absolute},
	}
}

// okResult wraps a value in a Result{IsOk: true}.
func okResult(val object.Object) *object.Result {
	return &object.Result{Value: val, IsOk: true}
}

// errResult wraps an error message in a Result{IsOk: false}.
func errResult(msg string) *object.Result {
	return &object.Result{Value: &object.String{Value: msg}, IsOk: false}
}

func readFile(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "io.readFile expects 1 argument (path)"}
	}
	path, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "io.readFile: path must be a string"}
	}
	data, err := os.ReadFile(path.Value)
	if err != nil {
		return errResult(fmt.Sprintf("io.readFile: %s", err.Error()))
	}
	return okResult(&object.String{Value: string(data)})
}

func readLines(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "io.readLines expects 1 argument (path)"}
	}
	path, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "io.readLines: path must be a string"}
	}
	data, err := os.ReadFile(path.Value)
	if err != nil {
		return errResult(fmt.Sprintf("io.readLines: %s", err.Error()))
	}
	content := string(data)
	lines := strings.Split(content, "\n")
	// Remove trailing empty string if the file ends with a newline.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	elements := make([]object.Object, len(lines))
	for i, line := range lines {
		elements[i] = &object.String{Value: line}
	}
	return okResult(&object.List{Elements: elements})
}

func writeFile(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "io.writeFile expects 2 arguments (path, content)"}
	}
	path, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "io.writeFile: path must be a string"}
	}
	content, ok := args[1].(*object.String)
	if !ok {
		return &object.Error{Message: "io.writeFile: content must be a string"}
	}
	err := os.WriteFile(path.Value, []byte(content.Value), 0644)
	if err != nil {
		return errResult(fmt.Sprintf("io.writeFile: %s", err.Error()))
	}
	return okResult(object.NIL)
}

func appendFile(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "io.appendFile expects 2 arguments (path, content)"}
	}
	path, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "io.appendFile: path must be a string"}
	}
	content, ok := args[1].(*object.String)
	if !ok {
		return &object.Error{Message: "io.appendFile: content must be a string"}
	}
	f, err := os.OpenFile(path.Value, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errResult(fmt.Sprintf("io.appendFile: %s", err.Error()))
	}
	defer f.Close()
	if _, err := f.WriteString(content.Value); err != nil {
		return errResult(fmt.Sprintf("io.appendFile: %s", err.Error()))
	}
	return okResult(object.NIL)
}

func exists(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "io.exists expects 1 argument (path)"}
	}
	path, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "io.exists: path must be a string"}
	}
	_, err := os.Stat(path.Value)
	return object.NativeBoolToBooleanObject(err == nil)
}

func isFile(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "io.isFile expects 1 argument (path)"}
	}
	path, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "io.isFile: path must be a string"}
	}
	info, err := os.Stat(path.Value)
	if err != nil {
		return object.FALSE_OBJ
	}
	return object.NativeBoolToBooleanObject(!info.IsDir())
}

func isDir(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "io.isDir expects 1 argument (path)"}
	}
	path, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "io.isDir: path must be a string"}
	}
	info, err := os.Stat(path.Value)
	if err != nil {
		return object.FALSE_OBJ
	}
	return object.NativeBoolToBooleanObject(info.IsDir())
}

func fileSize(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "io.fileSize expects 1 argument (path)"}
	}
	path, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "io.fileSize: path must be a string"}
	}
	info, err := os.Stat(path.Value)
	if err != nil {
		return errResult(fmt.Sprintf("io.fileSize: %s", err.Error()))
	}
	return okResult(&object.Integer{Value: info.Size()})
}

func listDir(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "io.listDir expects 1 argument (path)"}
	}
	path, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "io.listDir: path must be a string"}
	}
	entries, err := os.ReadDir(path.Value)
	if err != nil {
		return errResult(fmt.Sprintf("io.listDir: %s", err.Error()))
	}
	elements := make([]object.Object, len(entries))
	for i, entry := range entries {
		elements[i] = &object.String{Value: entry.Name()}
	}
	return okResult(&object.List{Elements: elements})
}

func createDir(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "io.createDir expects 1 argument (path)"}
	}
	path, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "io.createDir: path must be a string"}
	}
	err := os.Mkdir(path.Value, 0755)
	if err != nil {
		return errResult(fmt.Sprintf("io.createDir: %s", err.Error()))
	}
	return okResult(object.NIL)
}

func createDirAll(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "io.createDirAll expects 1 argument (path)"}
	}
	path, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "io.createDirAll: path must be a string"}
	}
	err := os.MkdirAll(path.Value, 0755)
	if err != nil {
		return errResult(fmt.Sprintf("io.createDirAll: %s", err.Error()))
	}
	return okResult(object.NIL)
}

func removeFile(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "io.removeFile expects 1 argument (path)"}
	}
	path, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "io.removeFile: path must be a string"}
	}
	err := os.Remove(path.Value)
	if err != nil {
		return errResult(fmt.Sprintf("io.removeFile: %s", err.Error()))
	}
	return okResult(object.NIL)
}

func copyFile(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "io.copy expects 2 arguments (src, dst)"}
	}
	src, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "io.copy: src must be a string"}
	}
	dst, ok := args[1].(*object.String)
	if !ok {
		return &object.Error{Message: "io.copy: dst must be a string"}
	}

	srcFile, err := os.Open(src.Value)
	if err != nil {
		return errResult(fmt.Sprintf("io.copy: %s", err.Error()))
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return errResult(fmt.Sprintf("io.copy: %s", err.Error()))
	}

	dstFile, err := os.OpenFile(dst.Value, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return errResult(fmt.Sprintf("io.copy: %s", err.Error()))
	}
	defer dstFile.Close()

	if _, err := goio.Copy(dstFile, srcFile); err != nil {
		return errResult(fmt.Sprintf("io.copy: %s", err.Error()))
	}
	return okResult(object.NIL)
}

func moveFile(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "io.move expects 2 arguments (src, dst)"}
	}
	src, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "io.move: src must be a string"}
	}
	dst, ok := args[1].(*object.String)
	if !ok {
		return &object.Error{Message: "io.move: dst must be a string"}
	}
	err := os.Rename(src.Value, dst.Value)
	if err != nil {
		return errResult(fmt.Sprintf("io.move: %s", err.Error()))
	}
	return okResult(object.NIL)
}

func joinPath(args ...object.Object) object.Object {
	if len(args) == 0 {
		return &object.Error{Message: "io.join expects at least 1 argument"}
	}
	parts := make([]string, len(args))
	for i, arg := range args {
		s, ok := arg.(*object.String)
		if !ok {
			return &object.Error{Message: fmt.Sprintf("io.join: argument %d must be a string", i+1)}
		}
		parts[i] = s.Value
	}
	return &object.String{Value: filepath.Join(parts...)}
}

func dirname(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "io.dirname expects 1 argument (path)"}
	}
	path, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "io.dirname: path must be a string"}
	}
	return &object.String{Value: filepath.Dir(path.Value)}
}

func basename(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "io.basename expects 1 argument (path)"}
	}
	path, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "io.basename: path must be a string"}
	}
	return &object.String{Value: filepath.Base(path.Value)}
}

func extension(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "io.extension expects 1 argument (path)"}
	}
	path, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "io.extension: path must be a string"}
	}
	return &object.String{Value: filepath.Ext(path.Value)}
}

func absolute(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "io.absolute expects 1 argument (path)"}
	}
	path, ok := args[0].(*object.String)
	if !ok {
		return &object.Error{Message: "io.absolute: path must be a string"}
	}
	abs, err := filepath.Abs(path.Value)
	if err != nil {
		return &object.String{Value: path.Value}
	}
	return &object.String{Value: abs}
}
