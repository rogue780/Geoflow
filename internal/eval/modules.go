package eval

import (
	"strings"

	"github.com/rogue780/geoflow/internal/object"
	"github.com/rogue780/geoflow/internal/stdlib"
	"github.com/rogue780/geoflow/pkg/geo/crs"
	"github.com/rogue780/geoflow/pkg/geo/geog"
	"github.com/rogue780/geoflow/pkg/geo/index"
	geoio "github.com/rogue780/geoflow/pkg/geo/io"
	gfassert "github.com/rogue780/geoflow/pkg/stdlib/assert"
	"github.com/rogue780/geoflow/pkg/stdlib/collections"
	"github.com/rogue780/geoflow/pkg/stdlib/core"
	gfdata "github.com/rogue780/geoflow/pkg/stdlib/data"
	gfio "github.com/rogue780/geoflow/pkg/stdlib/io"
	gflog "github.com/rogue780/geoflow/pkg/stdlib/log"
	gftime "github.com/rogue780/geoflow/pkg/stdlib/time"
)

// ModuleRegistry maps module paths (e.g. "std.math") to lazy loader functions
// that produce a *object.Module when first requested.
type ModuleRegistry struct {
	modules map[string]func() *object.Module
}

// Get retrieves a module by its dot-separated path.
func (r *ModuleRegistry) Get(path string) (*object.Module, bool) {
	loader, ok := r.modules[path]
	if !ok {
		return nil, false
	}
	return loader(), true
}

// Register adds or replaces a module loader for the given path.
func (r *ModuleRegistry) Register(path string, loader func() *object.Module) {
	r.modules[path] = loader
}

// DefaultRegistry is the global module registry populated with all built-in
// standard library modules at init time.
var DefaultRegistry *ModuleRegistry

func init() {
	DefaultRegistry = &ModuleRegistry{
		modules: make(map[string]func() *object.Module),
	}

	registerCoreModule(DefaultRegistry)
	registerMathModules(DefaultRegistry)
	registerStringModule(DefaultRegistry)
	registerCollectionsModule(DefaultRegistry)
	registerIOModule(DefaultRegistry)
	registerTimeModule(DefaultRegistry)
	registerLogModule(DefaultRegistry)
	registerAssertModule(DefaultRegistry)
	registerDataModule(DefaultRegistry)
	registerGeoModules(DefaultRegistry)
}

// isCallable reports whether an object can be called as a function.
func isCallable(obj object.Object) bool {
	switch obj.(type) {
	case *object.Function, *object.Builtin, *object.ComposedFunction, *object.StructDef, *object.EnumDef:
		return true
	default:
		return false
	}
}

// ── std.core ──

func registerCoreModule(r *ModuleRegistry) {
	r.Register("std.core", func() *object.Module {
		return &object.Module{Name: "core", Exports: core.GetExports()}
	})
}

// ── std.math (and sub-modules) ──

var statsNames = map[string]bool{
	"mean": true, "median": true, "mode": true,
	"variance": true, "stddev": true, "percentile": true,
	"correlation": true, "covariance": true, "linreg": true,
	"sem": true, "skewness": true, "kurtosis": true,
	"iqr": true, "zscore": true, "movingMean": true,
}

var linalgNames = map[string]bool{
	"vec": true, "mat": true, "dot": true, "cross": true,
	"magnitude": true, "normalize": true, "matMul": true,
	"transpose": true, "determinant": true, "inverse": true,
	"identity": true, "zeros": true,
	"vecAdd": true, "vecSub": true, "vecScale": true,
	"matVecMul": true,
}

var complexNames = map[string]bool{
	"complex": true, "real": true, "imag": true,
	"conjugate": true, "complexMag": true, "complexPhase": true,
	"complexAdd": true, "complexSub": true, "complexMul": true,
	"complexDiv": true, "complexSqrt": true, "complexExp": true,
}

var numericNames = map[string]bool{
	"linspace": true, "arange": true,
}

func registerMathModules(r *ModuleRegistry) {
	r.Register("std.math", func() *object.Module {
		all := stdlib.GetMathBuiltins()
		exports := make(map[string]object.Object, len(all))
		for k, v := range all {
			exports[k] = v
		}
		return &object.Module{Name: "math", Exports: exports}
	})

	r.Register("std.math.stats", func() *object.Module {
		all := stdlib.GetMathBuiltins()
		exports := make(map[string]object.Object)
		for k, v := range all {
			if statsNames[k] {
				exports[k] = v
			}
		}
		return &object.Module{Name: "stats", Exports: exports}
	})

	r.Register("std.math.linalg", func() *object.Module {
		all := stdlib.GetMathBuiltins()
		exports := make(map[string]object.Object)
		for k, v := range all {
			if linalgNames[k] {
				exports[k] = v
			}
		}
		return &object.Module{Name: "linalg", Exports: exports}
	})

	r.Register("std.math.complex", func() *object.Module {
		all := stdlib.GetMathBuiltins()
		exports := make(map[string]object.Object)
		for k, v := range all {
			if complexNames[k] {
				exports[k] = v
			}
		}
		return &object.Module{Name: "complex", Exports: exports}
	})

	r.Register("std.math.numeric", func() *object.Module {
		all := stdlib.GetMathBuiltins()
		exports := make(map[string]object.Object)
		for k, v := range all {
			if numericNames[k] {
				exports[k] = v
			}
		}
		return &object.Module{Name: "numeric", Exports: exports}
	})
}

// ── std.string ──

func registerStringModule(r *ModuleRegistry) {
	r.Register("std.string", func() *object.Module {
		exports := map[string]object.Object{
			"from": &object.Builtin{Name: "string.from", Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Error{Message: "string.from expects 1 argument"}
				}
				return &object.String{Value: args[0].Inspect()}
			}},
			"repeat": &object.Builtin{Name: "string.repeat", Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Error{Message: "string.repeat expects 2 arguments (string, count)"}
				}
				s, ok := args[0].(*object.String)
				if !ok {
					return &object.Error{Message: "string.repeat: first argument must be a string"}
				}
				n, ok := args[1].(*object.Integer)
				if !ok {
					return &object.Error{Message: "string.repeat: second argument must be an integer"}
				}
				return &object.String{Value: strings.Repeat(s.Value, int(n.Value))}
			}},
			"join": &object.Builtin{Name: "string.join", Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return &object.Error{Message: "string.join expects 2 arguments (list, separator)"}
				}
				list, ok := args[0].(*object.List)
				if !ok {
					return &object.Error{Message: "string.join: first argument must be a list"}
				}
				sep, ok := args[1].(*object.String)
				if !ok {
					return &object.Error{Message: "string.join: second argument must be a string"}
				}
				parts := make([]string, len(list.Elements))
				for i, e := range list.Elements {
					parts[i] = e.Inspect()
				}
				return &object.String{Value: strings.Join(parts, sep.Value)}
			}},
		}
		return &object.Module{Name: "string", Exports: exports}
	})
}

// ── std.collections ──

func registerCollectionsModule(r *ModuleRegistry) {
	r.Register("std.collections", func() *object.Module {
		return &object.Module{Name: "collections", Exports: collections.GetSetExports()}
	})
}

// ── std.io ──

func registerIOModule(r *ModuleRegistry) {
	r.Register("std.io", func() *object.Module {
		return &object.Module{Name: "io", Exports: gfio.GetExports()}
	})
}

// ── std.time ──

func registerTimeModule(r *ModuleRegistry) {
	r.Register("std.time", func() *object.Module {
		return &object.Module{Name: "time", Exports: gftime.GetExports()}
	})
}

// ── std.log ──

func registerLogModule(r *ModuleRegistry) {
	r.Register("std.log", func() *object.Module {
		return &object.Module{Name: "log", Exports: gflog.GetExports()}
	})
}

// ── std.assert ──

func registerAssertModule(r *ModuleRegistry) {
	r.Register("std.assert", func() *object.Module {
		return &object.Module{Name: "assert", Exports: gfassert.GetExports()}
	})
}

// ── std.data ──

func registerDataModule(r *ModuleRegistry) {
	r.Register("std.data", func() *object.Module {
		return &object.Module{Name: "data", Exports: gfdata.GetExports()}
	})
}

// ── std.geo.* ──

var geoIONames = map[string]bool{
	"parseWKT": true, "toWKT": true,
	"parseGeoJSON": true, "toGeoJSON": true,
}

func registerGeoModules(r *ModuleRegistry) {
	r.Register("std.geo.geom", func() *object.Module {
		all := stdlib.GetGeoBuiltins()
		exports := make(map[string]object.Object, len(all))
		for k, v := range all {
			exports[k] = v
		}
		return &object.Module{Name: "geom", Exports: exports}
	})

	r.Register("std.geo.crs", func() *object.Module {
		return &object.Module{Name: "crs", Exports: crs.GetExports()}
	})

	r.Register("std.geo.io", func() *object.Module {
		all := stdlib.GetGeoBuiltins()
		exports := make(map[string]object.Object)
		for k, v := range all {
			if geoIONames[k] {
				exports[k] = v
			}
		}
		// Merge in additional exports from pkg/geo/io (WKB, KML, etc.)
		for k, v := range geoio.GetExports() {
			exports[k] = v
		}
		return &object.Module{Name: "geoio", Exports: exports}
	})

	r.Register("std.geo.index", func() *object.Module {
		return &object.Module{Name: "index", Exports: index.GetExports()}
	})

	r.Register("std.geo.geog", func() *object.Module {
		return &object.Module{Name: "geog", Exports: geog.GetExports()}
	})
}
