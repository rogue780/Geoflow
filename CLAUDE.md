# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GeoFlow is a composition-oriented functional language for geospatial data transformation and ETL pipelines, implemented in Go. It features first-class OGC geometry support, pipeline operators, and a tree-walking interpreter. No external dependencies — pure Go.

## Build, Test, and Run

```bash
# Build
go build -o geoflow ./cmd/geoflow

# Run all tests
go test ./...

# Run tests for a specific package
go test -v ./internal/eval/...
go test -v ./internal/stdlib/...
go test -v ./internal/parser/...
go test -v ./internal/lexer/...

# Run a GeoFlow script
./geoflow examples/hello.gf

# Start the REPL
./geoflow
```

## Architecture

The interpreter follows a classic pipeline: **Lexer → Parser → AST → Evaluator**.

- **`internal/token/`** — Token types, position tracking, keyword lookup
- **`internal/lexer/`** — UTF-8 tokenizer with WKT literal support (`#POINT(...)#` syntax)
- **`internal/parser/`** — Recursive descent parser with operator precedence climbing
- **`internal/ast/`** — AST node hierarchy (statements and expressions implement `Node` interface)
- **`internal/types/`** — Type system: `PrimitiveType`, `GenericType`, `FunctionType`, `AutoType`
- **`internal/object/`** — ~27 runtime value types (primitives, collections, geometry, math, control flow). Contains `Environment` (lexical scoping with mutability tracking) and all builtin function definitions
- **`internal/eval/`** — Tree-walking evaluator. `modules.go` holds the `ModuleRegistry` which maps `import` paths (e.g. `std.math`) to lazy-loaded `*object.Module` values
- **`internal/repl/`** — REPL and `Execute()` for running scripts programmatically
- **`internal/stdlib/`** — Core builtin implementations for geospatial (`geo.go`) and math (`math.go`) functions

### Standard Library Modules (pkg/)

Modules under `pkg/` export functions via `GetExports()` and are registered in `internal/eval/modules.go`:

- **`pkg/stdlib/core/`** — Functional utilities: compose, pipe, curry, identity, flip
- **`pkg/stdlib/collections/`** — Set operations
- **`pkg/stdlib/assert/`** — Assertion functions for testing
- **`pkg/stdlib/io/`** — File I/O, JSON, CSV parsing
- **`pkg/stdlib/log/`** — Leveled logging
- **`pkg/stdlib/time/`** — DateTime/Duration operations
- **`pkg/geo/geog/`** — Geodesic calculations (distance, bearing on ellipsoid)
- **`pkg/geo/crs/`** — Coordinate Reference System / SRID handling
- **`pkg/geo/index/`** — R-tree spatial indexing
- **`pkg/geo/io/`** — WKT, WKB, KML, GeoJSON I/O

## Key Conventions

- **Errors as objects**: Runtime errors are `*object.Error` with a `Message` field, propagated via `isError()` checks — not Go `error` values.
- **Module registration**: To add a new stdlib module, create a package under `pkg/`, implement a `GetExports() map[string]object.Object` function, and register it in `internal/eval/modules.go`.
- **Builtins**: Core builtins (println, len, map, filter, etc.) are defined in `internal/object/builtins.go`. Module-scoped builtins live in their respective `pkg/` or `internal/stdlib/` packages.
- **Environment scoping**: `object.Environment` uses a linked-list chain. Variables have a `mutable` flag — only `let mut` bindings can be reassigned with `:=`.
- **GeoFlow scripts** use the `.gf` file extension. See `examples/` for 20 working examples.
- **Comments**: `--` for single-line, `{- -}` for multi-line.
- **Language spec**: `GEOFLOW_SPECIFICATION.md` is the authoritative language reference (~96KB).
