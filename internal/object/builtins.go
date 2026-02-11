package object

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// GetBuiltins returns all built-in functions.
func GetBuiltins() map[string]*Builtin {
	return map[string]*Builtin{
		"println": {
			Name: "println",
			Fn: func(args ...Object) Object {
				parts := make([]string, len(args))
				for i, arg := range args {
					parts[i] = arg.Inspect()
				}
				fmt.Println(strings.Join(parts, " "))
				return NIL
			},
		},
		"print": {
			Name: "print",
			Fn: func(args ...Object) Object {
				parts := make([]string, len(args))
				for i, arg := range args {
					parts[i] = arg.Inspect()
				}
				fmt.Print(strings.Join(parts, " "))
				return NIL
			},
		},
		"typeof": {
			Name: "typeof",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return &Error{Message: "typeof expects 1 argument"}
				}
				return &String{Value: string(args[0].Type())}
			},
		},
		"len": {
			Name: "len",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return &Error{Message: "len expects 1 argument"}
				}
				switch arg := args[0].(type) {
				case *String:
					return &Integer{Value: int64(len(arg.Value))}
				case *List:
					return &Integer{Value: int64(len(arg.Elements))}
				case *Map:
					return &Integer{Value: int64(len(arg.Pairs))}
				case *Tuple:
					return &Integer{Value: int64(len(arg.Elements))}
				default:
					return &Error{Message: fmt.Sprintf("len not supported for %s", arg.Type())}
				}
			},
		},
		"toString": {
			Name: "toString",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return &Error{Message: "toString expects 1 argument"}
				}
				return &String{Value: args[0].Inspect()}
			},
		},
		"int": {
			Name: "int",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return &Error{Message: "int expects 1 argument"}
				}
				switch arg := args[0].(type) {
				case *Integer:
					return arg
				case *Float:
					return &Integer{Value: int64(arg.Value)}
				case *Boolean:
					if arg.Value {
						return &Integer{Value: 1}
					}
					return &Integer{Value: 0}
				case *String:
					var v int64
					_, err := fmt.Sscanf(arg.Value, "%d", &v)
					if err != nil {
						return &Error{Message: fmt.Sprintf("cannot convert '%s' to int", arg.Value)}
					}
					return &Integer{Value: v}
				default:
					return &Error{Message: fmt.Sprintf("cannot convert %s to int", arg.Type())}
				}
			},
		},
		"float": {
			Name: "float",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return &Error{Message: "float expects 1 argument"}
				}
				switch arg := args[0].(type) {
				case *Integer:
					return &Float{Value: float64(arg.Value)}
				case *Float:
					return arg
				case *Boolean:
					if arg.Value {
						return &Float{Value: 1.0}
					}
					return &Float{Value: 0.0}
				case *String:
					var v float64
					_, err := fmt.Sscanf(arg.Value, "%f", &v)
					if err != nil {
						return &Error{Message: fmt.Sprintf("cannot convert '%s' to float", arg.Value)}
					}
					return &Float{Value: v}
				default:
					return &Error{Message: fmt.Sprintf("cannot convert %s to float", arg.Type())}
				}
			},
		},
		"string": {
			Name: "string",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return &Error{Message: "string expects 1 argument"}
				}
				return &String{Value: args[0].Inspect()}
			},
		},
		"append": {
			Name: "append",
			Fn: func(args ...Object) Object {
				if len(args) != 2 {
					return &Error{Message: "append expects 2 arguments: list and element"}
				}
				list, ok := args[0].(*List)
				if !ok {
					return &Error{Message: "first argument to append must be a list"}
				}
				newElems := make([]Object, len(list.Elements)+1)
				copy(newElems, list.Elements)
				newElems[len(list.Elements)] = args[1]
				return &List{Elements: newElems}
			},
		},
		"head": {
			Name: "head",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return &Error{Message: "head expects 1 argument"}
				}
				list, ok := args[0].(*List)
				if !ok {
					return &Error{Message: "argument to head must be a list"}
				}
				if len(list.Elements) == 0 {
					return NONE
				}
				return &Option{Value: list.Elements[0], IsSome: true}
			},
		},
		"tail": {
			Name: "tail",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return &Error{Message: "tail expects 1 argument"}
				}
				list, ok := args[0].(*List)
				if !ok {
					return &Error{Message: "argument to tail must be a list"}
				}
				if len(list.Elements) == 0 {
					return &List{Elements: []Object{}}
				}
				newElems := make([]Object, len(list.Elements)-1)
				copy(newElems, list.Elements[1:])
				return &List{Elements: newElems}
			},
		},
		"map": {
			Name: "map",
			Fn:   nil, // handled specially in evaluator
		},
		"filter": {
			Name: "filter",
			Fn:   nil, // handled specially in evaluator
		},
		"reduce": {
			Name: "reduce",
			Fn:   nil, // handled specially in evaluator
		},
		"sum": {
			Name: "sum",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return &Error{Message: "sum expects 1 argument (a list)"}
				}
				list, ok := args[0].(*List)
				if !ok {
					return &Error{Message: "argument to sum must be a list"}
				}
				var intSum int64
				var floatSum float64
				isFloat := false
				for _, elem := range list.Elements {
					switch v := elem.(type) {
					case *Integer:
						intSum += v.Value
						floatSum += float64(v.Value)
					case *Float:
						isFloat = true
						floatSum += v.Value
					default:
						return &Error{Message: fmt.Sprintf("sum: unsupported element type %s", elem.Type())}
					}
				}
				if isFloat {
					return &Float{Value: floatSum}
				}
				return &Integer{Value: intSum}
			},
		},
		"product": {
			Name: "product",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return &Error{Message: "product expects 1 argument (a list)"}
				}
				list, ok := args[0].(*List)
				if !ok {
					return &Error{Message: "argument to product must be a list"}
				}
				var intProd int64 = 1
				var floatProd float64 = 1
				isFloat := false
				for _, elem := range list.Elements {
					switch v := elem.(type) {
					case *Integer:
						intProd *= v.Value
						floatProd *= float64(v.Value)
					case *Float:
						isFloat = true
						floatProd *= v.Value
					default:
						return &Error{Message: fmt.Sprintf("product: unsupported element type %s", elem.Type())}
					}
				}
				if isFloat {
					return &Float{Value: floatProd}
				}
				return &Integer{Value: intProd}
			},
		},
		"reverse": {
			Name: "reverse",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return &Error{Message: "reverse expects 1 argument"}
				}
				switch arg := args[0].(type) {
				case *List:
					newElems := make([]Object, len(arg.Elements))
					for i, e := range arg.Elements {
						newElems[len(arg.Elements)-1-i] = e
					}
					return &List{Elements: newElems}
				case *String:
					runes := []rune(arg.Value)
					for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
						runes[i], runes[j] = runes[j], runes[i]
					}
					return &String{Value: string(runes)}
				default:
					return &Error{Message: fmt.Sprintf("reverse not supported for %s", arg.Type())}
				}
			},
		},
		"sort": {
			Name: "sort",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return &Error{Message: "sort expects 1 argument"}
				}
				list, ok := args[0].(*List)
				if !ok {
					return &Error{Message: "argument to sort must be a list"}
				}
				newElems := make([]Object, len(list.Elements))
				copy(newElems, list.Elements)
				sort.Slice(newElems, func(i, j int) bool {
					return compareObjects(newElems[i], newElems[j]) < 0
				})
				return &List{Elements: newElems}
			},
		},
		"range": {
			Name: "range",
			Fn: func(args ...Object) Object {
				if len(args) < 1 || len(args) > 3 {
					return &Error{Message: "range expects 1-3 arguments: end or start, end[, step]"}
				}
				var start, end, step int64
				step = 1
				switch len(args) {
				case 1:
					e, ok := args[0].(*Integer)
					if !ok {
						return &Error{Message: "range arguments must be integers"}
					}
					end = e.Value
				case 2:
					s, ok1 := args[0].(*Integer)
					e, ok2 := args[1].(*Integer)
					if !ok1 || !ok2 {
						return &Error{Message: "range arguments must be integers"}
					}
					start = s.Value
					end = e.Value
				case 3:
					s, ok1 := args[0].(*Integer)
					e, ok2 := args[1].(*Integer)
					st, ok3 := args[2].(*Integer)
					if !ok1 || !ok2 || !ok3 {
						return &Error{Message: "range arguments must be integers"}
					}
					start = s.Value
					end = e.Value
					step = st.Value
				}
				if step == 0 {
					return &Error{Message: "range step cannot be 0"}
				}
				var elems []Object
				if step > 0 {
					for i := start; i < end; i += step {
						elems = append(elems, &Integer{Value: i})
					}
				} else {
					for i := start; i > end; i += step {
						elems = append(elems, &Integer{Value: i})
					}
				}
				return &List{Elements: elems}
			},
		},
		"Some": {
			Name: "Some",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return &Error{Message: "Some expects 1 argument"}
				}
				return &Option{Value: args[0], IsSome: true}
			},
		},
		"None": {
			Name: "None",
			Fn: func(args ...Object) Object {
				return NONE
			},
		},
		"Ok": {
			Name: "Ok",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return &Error{Message: "Ok expects 1 argument"}
				}
				return &Result{Value: args[0], IsOk: true}
			},
		},
		"Err": {
			Name: "Err",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return &Error{Message: "Err expects 1 argument"}
				}
				return &Result{Value: args[0], IsOk: false}
			},
		},
		"assert": {
			Name: "assert",
			Fn: func(args ...Object) Object {
				if len(args) < 1 || len(args) > 2 {
					return &Error{Message: "assert expects 1-2 arguments"}
				}
				if !IsTruthy(args[0]) {
					msg := "assertion failed"
					if len(args) == 2 {
						if s, ok := args[1].(*String); ok {
							msg = s.Value
						}
					}
					return &Error{Message: msg}
				}
				return NIL
			},
		},
		// Math builtins
		"abs": {
			Name: "abs",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return &Error{Message: "abs expects 1 argument"}
				}
				switch arg := args[0].(type) {
				case *Integer:
					if arg.Value < 0 {
						return &Integer{Value: -arg.Value}
					}
					return arg
				case *Float:
					return &Float{Value: math.Abs(arg.Value)}
				default:
					return &Error{Message: fmt.Sprintf("abs not supported for %s", arg.Type())}
				}
			},
		},
		"min": {
			Name: "min",
			Fn: func(args ...Object) Object {
				if len(args) != 2 {
					return &Error{Message: "min expects 2 arguments"}
				}
				if compareObjects(args[0], args[1]) <= 0 {
					return args[0]
				}
				return args[1]
			},
		},
		"max": {
			Name: "max",
			Fn: func(args ...Object) Object {
				if len(args) != 2 {
					return &Error{Message: "max expects 2 arguments"}
				}
				if compareObjects(args[0], args[1]) >= 0 {
					return args[0]
				}
				return args[1]
			},
		},
		"sqrt": {
			Name: "sqrt",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return &Error{Message: "sqrt expects 1 argument"}
				}
				val := toFloat(args[0])
				if val == nil {
					return &Error{Message: "sqrt expects a numeric argument"}
				}
				return &Float{Value: math.Sqrt(val.Value)}
			},
		},
	}
}

func compareObjects(a, b Object) int {
	switch av := a.(type) {
	case *Integer:
		switch bv := b.(type) {
		case *Integer:
			if av.Value < bv.Value {
				return -1
			} else if av.Value > bv.Value {
				return 1
			}
			return 0
		case *Float:
			af := float64(av.Value)
			if af < bv.Value {
				return -1
			} else if af > bv.Value {
				return 1
			}
			return 0
		}
	case *Float:
		switch bv := b.(type) {
		case *Integer:
			bf := float64(bv.Value)
			if av.Value < bf {
				return -1
			} else if av.Value > bf {
				return 1
			}
			return 0
		case *Float:
			if av.Value < bv.Value {
				return -1
			} else if av.Value > bv.Value {
				return 1
			}
			return 0
		}
	case *String:
		if bv, ok := b.(*String); ok {
			if av.Value < bv.Value {
				return -1
			} else if av.Value > bv.Value {
				return 1
			}
			return 0
		}
	}
	return 0
}

func toFloat(o Object) *Float {
	switch v := o.(type) {
	case *Float:
		return v
	case *Integer:
		return &Float{Value: float64(v.Value)}
	default:
		return nil
	}
}
