// Package eval implements the tree-walking evaluator for GeoFlow.
package eval

import (
	"fmt"
	"math"
	"strings"

	"github.com/rogue780/geoflow/internal/ast"
	"github.com/rogue780/geoflow/internal/object"
)

// Eval evaluates an AST node in the given environment.
func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.LetStatement:
		return evalLetStatement(node, env)
	case *ast.ConstStatement:
		return evalConstStatement(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.BreakStatement:
		return object.BREAK
	case *ast.ContinueStatement:
		return object.CONT
	case *ast.ImportStatement:
		// Imports are no-ops for now
		return object.NIL
	case *ast.TypeStatement:
		// Type statements are no-ops at runtime
		return object.NIL
	case *ast.AssignStatement:
		return evalAssignStatement(node, env)

	// Expressions
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.BooleanLiteral:
		return object.NativeBoolToBooleanObject(node.Value)
	case *ast.NilLiteral:
		return object.NIL
	case *ast.WKTLiteral:
		return &object.String{Value: node.Value}
	case *ast.SymbolicLiteral:
		return &object.String{Value: node.Value}
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		if node.Operator == ":=" {
			return evalMutableAssignment(node, env)
		}
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.PipelineExpression:
		return evalPipelineExpression(node, env)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.BlockExpression:
		return evalBlockExpression(node, env)
	case *ast.FunctionLiteral:
		return evalFunctionLiteral(node, env)
	case *ast.LambdaExpression:
		return evalLambdaExpression(node, env)
	case *ast.CallExpression:
		return evalCallExpression(node, env)
	case *ast.DotExpression:
		return evalDotExpression(node, env)
	case *ast.IndexExpression:
		return evalIndexExpression(node, env)
	case *ast.ListLiteral:
		return evalListLiteral(node, env)
	case *ast.MapLiteral:
		return evalMapLiteral(node, env)
	case *ast.TupleLiteral:
		return evalTupleLiteral(node, env)
	case *ast.MatchExpression:
		return evalMatchExpression(node, env)
	case *ast.ForExpression:
		return evalForExpression(node, env)
	case *ast.WhileExpression:
		return evalWhileExpression(node, env)
	case *ast.StringInterpolation:
		return evalStringInterpolation(node, env)
	case *ast.Wildcard:
		return object.NIL
	}

	return newError("unknown node type: %T", node)
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object
	for _, stmt := range program.Statements {
		result = Eval(stmt, env)
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}
	return result
}

func evalLetStatement(node *ast.LetStatement, env *object.Environment) object.Object {
	val := Eval(node.Value, env)
	if isError(val) {
		return val
	}
	env.Set(node.Name.Value, val, node.Mutable)
	return object.NIL
}

func evalConstStatement(node *ast.ConstStatement, env *object.Environment) object.Object {
	val := Eval(node.Value, env)
	if isError(val) {
		return val
	}
	env.Set(node.Name.Value, val, false)
	return object.NIL
}

func evalAssignStatement(node *ast.AssignStatement, env *object.Environment) object.Object {
	val := Eval(node.Value, env)
	if isError(val) {
		return val
	}
	if ident, ok := node.Name.(*ast.Identifier); ok {
		err := env.Update(ident.Value, val)
		if err != nil {
			return newError("%s", err.Error())
		}
		return object.NIL
	}
	return newError("invalid assignment target")
}

func evalMutableAssignment(node *ast.InfixExpression, env *object.Environment) object.Object {
	val := Eval(node.Right, env)
	if isError(val) {
		return val
	}
	if ident, ok := node.Left.(*ast.Identifier); ok {
		err := env.Update(ident.Value, val)
		if err != nil {
			return newError("%s", err.Error())
		}
		return object.NIL
	}
	return newError("invalid assignment target")
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	// Handle break/continue as expression-level signals
	if node.Value == "break" {
		return object.BREAK
	}
	if node.Value == "continue" {
		return object.CONT
	}
	if val, ok := env.Get(node.Value); ok {
		return val
	}
	// Check builtins
	builtins := object.GetBuiltins()
	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}
	return newError("identifier not found: %s", node.Value)
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperator(right)
	case "-":
		return evalMinusPrefixOperator(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalBangOperator(right object.Object) object.Object {
	return object.NativeBoolToBooleanObject(!object.IsTruthy(right))
}

func evalMinusPrefixOperator(right object.Object) object.Object {
	switch obj := right.(type) {
	case *object.Integer:
		return &object.Integer{Value: -obj.Value}
	case *object.Float:
		return &object.Float{Value: -obj.Value}
	default:
		return newError("unknown operator: -%s", right.Type())
	}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.FLOAT_OBJ || right.Type() == object.FLOAT_OBJ:
		return evalFloatInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ:
		return evalBooleanInfixExpression(operator, left, right)
	case operator == "==":
		return object.NativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return object.NativeBoolToBooleanObject(left != right)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return newError("division by zero")
		}
		return &object.Integer{Value: leftVal / rightVal}
	case "%":
		if rightVal == 0 {
			return newError("modulo by zero")
		}
		return &object.Integer{Value: leftVal % rightVal}
	case "^":
		return &object.Integer{Value: intPow(leftVal, rightVal)}
	case "<":
		return object.NativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return object.NativeBoolToBooleanObject(leftVal > rightVal)
	case "<=":
		return object.NativeBoolToBooleanObject(leftVal <= rightVal)
	case ">=":
		return object.NativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return object.NativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return object.NativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalFloatInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := toFloat64(left)
	rightVal := toFloat64(right)

	switch operator {
	case "+":
		return &object.Float{Value: leftVal + rightVal}
	case "-":
		return &object.Float{Value: leftVal - rightVal}
	case "*":
		return &object.Float{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return newError("division by zero")
		}
		return &object.Float{Value: leftVal / rightVal}
	case "%":
		return &object.Float{Value: math.Mod(leftVal, rightVal)}
	case "^":
		return &object.Float{Value: math.Pow(leftVal, rightVal)}
	case "<":
		return object.NativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return object.NativeBoolToBooleanObject(leftVal > rightVal)
	case "<=":
		return object.NativeBoolToBooleanObject(leftVal <= rightVal)
	case ">=":
		return object.NativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return object.NativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return object.NativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	switch operator {
	case "+":
		return &object.String{Value: leftVal + rightVal}
	case "==":
		return object.NativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return object.NativeBoolToBooleanObject(leftVal != rightVal)
	case "<":
		return object.NativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return object.NativeBoolToBooleanObject(leftVal > rightVal)
	case "<=":
		return object.NativeBoolToBooleanObject(leftVal <= rightVal)
	case ">=":
		return object.NativeBoolToBooleanObject(leftVal >= rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalBooleanInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Boolean).Value
	rightVal := right.(*object.Boolean).Value

	switch operator {
	case "&&":
		return object.NativeBoolToBooleanObject(leftVal && rightVal)
	case "||":
		return object.NativeBoolToBooleanObject(leftVal || rightVal)
	case "==":
		return object.NativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return object.NativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalPipelineExpression(node *ast.PipelineExpression, env *object.Environment) object.Object {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}

	// The right side should be a function or a call expression
	// x |> f  means f(x)
	// x |> f(y) means f(x, y)
	switch right := node.Right.(type) {
	case *ast.CallExpression:
		// Prepend left to the arguments
		fn := Eval(right.Function, env)
		if isError(fn) {
			return fn
		}
		evalArgs := []object.Object{left}
		for _, arg := range right.Arguments {
			evaluated := Eval(arg, env)
			if isError(evaluated) {
				return evaluated
			}
			evalArgs = append(evalArgs, evaluated)
		}
		return applyFunction(fn, evalArgs, env)
	case *ast.Identifier:
		fn := Eval(right, env)
		if isError(fn) {
			return fn
		}
		return applyFunction(fn, []object.Object{left}, env)
	case *ast.DotExpression:
		fn := Eval(right, env)
		if isError(fn) {
			return fn
		}
		return applyFunction(fn, []object.Object{left}, env)
	default:
		fn := Eval(node.Right, env)
		if isError(fn) {
			return fn
		}
		return applyFunction(fn, []object.Object{left}, env)
	}
}

func evalIfExpression(node *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(node.Condition, env)
	if isError(condition) {
		return condition
	}

	if object.IsTruthy(condition) {
		return Eval(node.Consequence, env)
	} else if node.Alternative != nil {
		return Eval(node.Alternative, env)
	}
	return object.NIL
}

func evalBlockExpression(node *ast.BlockExpression, env *object.Environment) object.Object {
	var result object.Object = object.NIL
	blockEnv := object.NewEnclosedEnvironment(env)

	for _, stmt := range node.Statements {
		result = Eval(stmt, blockEnv)
		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ ||
				rt == object.BREAK_OBJ || rt == object.CONTINUE_OBJ {
				return result
			}
		}
	}

	if node.Value != nil {
		result = Eval(node.Value, blockEnv)
	}

	return result
}

func evalFunctionLiteral(node *ast.FunctionLiteral, env *object.Environment) object.Object {
	fn := &object.Function{
		Name:       node.Name,
		Parameters: node.Parameters,
		Body:       node.Body,
		Env:        env,
	}
	if node.Name != "" {
		env.Set(node.Name, fn, false)
	}
	return fn
}

func evalLambdaExpression(node *ast.LambdaExpression, env *object.Environment) object.Object {
	params := make([]*ast.FunctionParameter, len(node.Parameters))
	for i, p := range node.Parameters {
		params[i] = &ast.FunctionParameter{Name: p}
	}

	// Wrap the body expression in a block
	body := &ast.BlockExpression{
		Value: node.Body,
	}

	return &object.Function{
		Parameters: params,
		Body:       body,
		Env:        env,
	}
}

func evalCallExpression(node *ast.CallExpression, env *object.Environment) object.Object {
	fn := Eval(node.Function, env)
	if isError(fn) {
		return fn
	}

	args := evalExpressions(node.Arguments, env)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}

	return applyFunction(fn, args, env)
}

func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object
	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}
	return result
}

func applyFunction(fn object.Object, args []object.Object, env *object.Environment) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		// If the builtin has a concrete Fn implementation, use it directly.
		if fn.Fn != nil {
			result := fn.Fn(args...)
			if result != nil {
				return result
			}
			return object.NIL
		}
		// Standalone higher-order builtins with nil Fn (map, filter, reduce called as functions)
		if fn.Name == "map" || fn.Name == "filter" || fn.Name == "reduce" {
			return evalHigherOrderBuiltin(fn.Name, args, env)
		}
		return newError("builtin function %s not implemented", fn.Name)
	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)
	for i, param := range fn.Parameters {
		if i < len(args) {
			env.Set(param.Name.Value, args[i], false)
		} else if param.Default != nil {
			// Evaluate default in function's closure environment
			defaultVal := Eval(param.Default, fn.Env)
			env.Set(param.Name.Value, defaultVal, false)
		}
	}
	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if rv, ok := obj.(*object.ReturnValue); ok {
		return rv.Value
	}
	return obj
}

func evalHigherOrderBuiltin(name string, args []object.Object, env *object.Environment) object.Object {
	switch name {
	case "map":
		if len(args) != 2 {
			return newError("map expects 2 arguments: list and function")
		}
		list, ok := args[0].(*object.List)
		if !ok {
			return newError("first argument to map must be a list")
		}
		fn := args[1]
		result := make([]object.Object, len(list.Elements))
		for i, elem := range list.Elements {
			val := applyFunction(fn, []object.Object{elem}, env)
			if isError(val) {
				return val
			}
			result[i] = val
		}
		return &object.List{Elements: result}

	case "filter":
		if len(args) != 2 {
			return newError("filter expects 2 arguments: list and predicate")
		}
		list, ok := args[0].(*object.List)
		if !ok {
			return newError("first argument to filter must be a list")
		}
		fn := args[1]
		var result []object.Object
		for _, elem := range list.Elements {
			val := applyFunction(fn, []object.Object{elem}, env)
			if isError(val) {
				return val
			}
			if object.IsTruthy(val) {
				result = append(result, elem)
			}
		}
		return &object.List{Elements: result}

	case "reduce":
		if len(args) != 3 {
			return newError("reduce expects 3 arguments: list, initial, function")
		}
		list, ok := args[0].(*object.List)
		if !ok {
			return newError("first argument to reduce must be a list")
		}
		acc := args[1]
		fn := args[2]
		for _, elem := range list.Elements {
			acc = applyFunction(fn, []object.Object{acc, elem}, env)
			if isError(acc) {
				return acc
			}
		}
		return acc
	}
	return newError("unknown higher-order builtin: %s", name)
}

func evalDotExpression(node *ast.DotExpression, env *object.Environment) object.Object {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}

	// Method calls on built-in types
	switch obj := left.(type) {
	case *object.String:
		return evalStringMethod(obj, node.Field, env)
	case *object.List:
		return evalListMethod(obj, node.Field, env)
	case *object.Map:
		return evalMapMethod(obj, node.Field, env)
	case *object.Tuple:
		return evalTupleField(obj, node.Field)
	case *object.Option:
		return evalOptionMethod(obj, node.Field)
	case *object.Result:
		return evalResultMethod(obj, node.Field)
	}

	return newError("no field '%s' on type %s", node.Field, left.Type())
}

func evalStringMethod(s *object.String, method string, env *object.Environment) object.Object {
	switch method {
	case "length":
		return &object.Builtin{Name: "length", Fn: func(args ...object.Object) object.Object {
			return &object.Integer{Value: int64(len(s.Value))}
		}}
	case "uppercase":
		return &object.Builtin{Name: "uppercase", Fn: func(args ...object.Object) object.Object {
			return &object.String{Value: strings.ToUpper(s.Value)}
		}}
	case "lowercase":
		return &object.Builtin{Name: "lowercase", Fn: func(args ...object.Object) object.Object {
			return &object.String{Value: strings.ToLower(s.Value)}
		}}
	case "trim":
		return &object.Builtin{Name: "trim", Fn: func(args ...object.Object) object.Object {
			return &object.String{Value: strings.TrimSpace(s.Value)}
		}}
	case "contains":
		return &object.Builtin{Name: "contains", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("contains expects 1 argument")
			}
			sub, ok := args[0].(*object.String)
			if !ok {
				return newError("contains argument must be a string")
			}
			return object.NativeBoolToBooleanObject(strings.Contains(s.Value, sub.Value))
		}}
	case "split":
		return &object.Builtin{Name: "split", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("split expects 1 argument")
			}
			sep, ok := args[0].(*object.String)
			if !ok {
				return newError("split argument must be a string")
			}
			parts := strings.Split(s.Value, sep.Value)
			elems := make([]object.Object, len(parts))
			for i, p := range parts {
				elems[i] = &object.String{Value: p}
			}
			return &object.List{Elements: elems}
		}}
	case "replace":
		return &object.Builtin{Name: "replace", Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("replace expects 2 arguments")
			}
			old, ok1 := args[0].(*object.String)
			new, ok2 := args[1].(*object.String)
			if !ok1 || !ok2 {
				return newError("replace arguments must be strings")
			}
			return &object.String{Value: strings.ReplaceAll(s.Value, old.Value, new.Value)}
		}}
	case "startsWith":
		return &object.Builtin{Name: "startsWith", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("startsWith expects 1 argument")
			}
			prefix, ok := args[0].(*object.String)
			if !ok {
				return newError("startsWith argument must be a string")
			}
			return object.NativeBoolToBooleanObject(strings.HasPrefix(s.Value, prefix.Value))
		}}
	case "endsWith":
		return &object.Builtin{Name: "endsWith", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("endsWith expects 1 argument")
			}
			suffix, ok := args[0].(*object.String)
			if !ok {
				return newError("endsWith argument must be a string")
			}
			return object.NativeBoolToBooleanObject(strings.HasSuffix(s.Value, suffix.Value))
		}}
	case "isEmpty":
		return &object.Builtin{Name: "isEmpty", Fn: func(args ...object.Object) object.Object {
			return object.NativeBoolToBooleanObject(len(s.Value) == 0)
		}}
	default:
		return newError("no method '%s' on string", method)
	}
}

func evalListMethod(l *object.List, method string, env *object.Environment) object.Object {
	switch method {
	case "length":
		return &object.Builtin{Name: "length", Fn: func(args ...object.Object) object.Object {
			return &object.Integer{Value: int64(len(l.Elements))}
		}}
	case "isEmpty":
		return &object.Builtin{Name: "isEmpty", Fn: func(args ...object.Object) object.Object {
			return object.NativeBoolToBooleanObject(len(l.Elements) == 0)
		}}
	case "first":
		return &object.Builtin{Name: "first", Fn: func(args ...object.Object) object.Object {
			if len(l.Elements) == 0 {
				return object.NONE
			}
			return &object.Option{Value: l.Elements[0], IsSome: true}
		}}
	case "last":
		return &object.Builtin{Name: "last", Fn: func(args ...object.Object) object.Object {
			if len(l.Elements) == 0 {
				return object.NONE
			}
			return &object.Option{Value: l.Elements[len(l.Elements)-1], IsSome: true}
		}}
	case "get":
		return &object.Builtin{Name: "get", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("get expects 1 argument")
			}
			idx, ok := args[0].(*object.Integer)
			if !ok {
				return newError("get index must be an integer")
			}
			i := int(idx.Value)
			if i < 0 || i >= len(l.Elements) {
				return object.NONE
			}
			return &object.Option{Value: l.Elements[i], IsSome: true}
		}}
	case "contains":
		return &object.Builtin{Name: "contains", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("contains expects 1 argument")
			}
			for _, elem := range l.Elements {
				if objectsEqual(elem, args[0]) {
					return object.TRUE_OBJ
				}
			}
			return object.FALSE_OBJ
		}}
	case "map":
		return &object.Builtin{Name: "map", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("map expects 1 argument (function)")
			}
			return evalHigherOrderBuiltin("map", []object.Object{l, args[0]}, env)
		}}
	case "filter":
		return &object.Builtin{Name: "filter", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("filter expects 1 argument (predicate)")
			}
			return evalHigherOrderBuiltin("filter", []object.Object{l, args[0]}, env)
		}}
	case "reduce":
		return &object.Builtin{Name: "reduce", Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("reduce expects 2 arguments (initial, function)")
			}
			return evalHigherOrderBuiltin("reduce", []object.Object{l, args[0], args[1]}, env)
		}}
	case "append":
		return &object.Builtin{Name: "append", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("append expects 1 argument")
			}
			newElems := make([]object.Object, len(l.Elements)+1)
			copy(newElems, l.Elements)
			newElems[len(l.Elements)] = args[0]
			return &object.List{Elements: newElems}
		}}
	case "reverse":
		return &object.Builtin{Name: "reverse", Fn: func(args ...object.Object) object.Object {
			newElems := make([]object.Object, len(l.Elements))
			for i, e := range l.Elements {
				newElems[len(l.Elements)-1-i] = e
			}
			return &object.List{Elements: newElems}
		}}
	case "sum":
		return &object.Builtin{Name: "sum", Fn: func(args ...object.Object) object.Object {
			builtins := object.GetBuiltins()
			return builtins["sum"].Fn(l)
		}}
	case "enumerate":
		return &object.Builtin{Name: "enumerate", Fn: func(args ...object.Object) object.Object {
			result := make([]object.Object, len(l.Elements))
			for i, elem := range l.Elements {
				result[i] = &object.Tuple{Elements: []object.Object{
					&object.Integer{Value: int64(i)},
					elem,
				}}
			}
			return &object.List{Elements: result}
		}}
	case "concat":
		return &object.Builtin{Name: "concat", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("concat expects 1 argument")
			}
			other, ok := args[0].(*object.List)
			if !ok {
				return newError("concat argument must be a list")
			}
			newElems := make([]object.Object, len(l.Elements)+len(other.Elements))
			copy(newElems, l.Elements)
			copy(newElems[len(l.Elements):], other.Elements)
			return &object.List{Elements: newElems}
		}}
	case "flatten":
		return &object.Builtin{Name: "flatten", Fn: func(args ...object.Object) object.Object {
			var result []object.Object
			for _, elem := range l.Elements {
				if inner, ok := elem.(*object.List); ok {
					result = append(result, inner.Elements...)
				} else {
					result = append(result, elem)
				}
			}
			return &object.List{Elements: result}
		}}
	default:
		return newError("no method '%s' on List", method)
	}
}

func evalMapMethod(m *object.Map, method string, env *object.Environment) object.Object {
	switch method {
	case "size":
		return &object.Builtin{Name: "size", Fn: func(args ...object.Object) object.Object {
			return &object.Integer{Value: int64(len(m.Pairs))}
		}}
	case "isEmpty":
		return &object.Builtin{Name: "isEmpty", Fn: func(args ...object.Object) object.Object {
			return object.NativeBoolToBooleanObject(len(m.Pairs) == 0)
		}}
	case "keys":
		return &object.Builtin{Name: "keys", Fn: func(args ...object.Object) object.Object {
			keys := make([]object.Object, len(m.Pairs))
			for i, p := range m.Pairs {
				keys[i] = p.Key
			}
			return &object.List{Elements: keys}
		}}
	case "values":
		return &object.Builtin{Name: "values", Fn: func(args ...object.Object) object.Object {
			vals := make([]object.Object, len(m.Pairs))
			for i, p := range m.Pairs {
				vals[i] = p.Value
			}
			return &object.List{Elements: vals}
		}}
	case "get":
		return &object.Builtin{Name: "get", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("get expects 1 argument")
			}
			keyStr := args[0].Inspect()
			if s, ok := args[0].(*object.String); ok {
				keyStr = s.Value
			}
			val, found := m.Get(keyStr)
			if !found {
				return object.NONE
			}
			return &object.Option{Value: val, IsSome: true}
		}}
	case "contains":
		return &object.Builtin{Name: "contains", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("contains expects 1 argument")
			}
			keyStr := ""
			if s, ok := args[0].(*object.String); ok {
				keyStr = s.Value
			}
			_, found := m.Get(keyStr)
			return object.NativeBoolToBooleanObject(found)
		}}
	default:
		return newError("no method '%s' on Map", method)
	}
}

func evalTupleField(t *object.Tuple, field string) object.Object {
	// Support .0, .1, etc for tuple access
	var idx int
	if _, err := fmt.Sscanf(field, "%d", &idx); err == nil {
		if idx >= 0 && idx < len(t.Elements) {
			return t.Elements[idx]
		}
		return newError("tuple index out of bounds: %d", idx)
	}
	return newError("no field '%s' on Tuple", field)
}

func evalOptionMethod(o *object.Option, method string) object.Object {
	switch method {
	case "isSome":
		return &object.Builtin{Name: "isSome", Fn: func(args ...object.Object) object.Object {
			return object.NativeBoolToBooleanObject(o.IsSome)
		}}
	case "isNone":
		return &object.Builtin{Name: "isNone", Fn: func(args ...object.Object) object.Object {
			return object.NativeBoolToBooleanObject(!o.IsSome)
		}}
	case "unwrap":
		return &object.Builtin{Name: "unwrap", Fn: func(args ...object.Object) object.Object {
			if !o.IsSome {
				return newError("called unwrap on None")
			}
			return o.Value
		}}
	case "unwrapOr":
		return &object.Builtin{Name: "unwrapOr", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("unwrapOr expects 1 argument")
			}
			if o.IsSome {
				return o.Value
			}
			return args[0]
		}}
	default:
		return newError("no method '%s' on Option", method)
	}
}

func evalResultMethod(r *object.Result, method string) object.Object {
	switch method {
	case "isOk":
		return &object.Builtin{Name: "isOk", Fn: func(args ...object.Object) object.Object {
			return object.NativeBoolToBooleanObject(r.IsOk)
		}}
	case "isErr":
		return &object.Builtin{Name: "isErr", Fn: func(args ...object.Object) object.Object {
			return object.NativeBoolToBooleanObject(!r.IsOk)
		}}
	case "unwrap":
		return &object.Builtin{Name: "unwrap", Fn: func(args ...object.Object) object.Object {
			if !r.IsOk {
				return newError("called unwrap on Err: %s", r.Value.Inspect())
			}
			return r.Value
		}}
	default:
		return newError("no method '%s' on Result", method)
	}
}

func evalIndexExpression(node *ast.IndexExpression, env *object.Environment) object.Object {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}
	index := Eval(node.Index, env)
	if isError(index) {
		return index
	}

	switch {
	case left.Type() == object.LIST_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalListIndexExpression(left, index)
	case left.Type() == object.MAP_OBJ:
		return evalMapIndexExpression(left, index)
	case left.Type() == object.STRING_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalStringIndexExpression(left, index)
	case left.Type() == object.TUPLE_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalTupleIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s[%s]", left.Type(), index.Type())
	}
}

func evalListIndexExpression(left, index object.Object) object.Object {
	list := left.(*object.List)
	idx := index.(*object.Integer).Value
	max := int64(len(list.Elements) - 1)

	if idx < 0 {
		idx = int64(len(list.Elements)) + idx
	}

	if idx < 0 || idx > max {
		return object.NIL
	}
	return list.Elements[idx]
}

func evalMapIndexExpression(left, index object.Object) object.Object {
	m := left.(*object.Map)
	key := ""
	if s, ok := index.(*object.String); ok {
		key = s.Value
	} else {
		key = index.Inspect()
	}
	val, found := m.Get(key)
	if !found {
		return object.NIL
	}
	return val
}

func evalStringIndexExpression(left, index object.Object) object.Object {
	s := left.(*object.String)
	idx := index.(*object.Integer).Value
	runes := []rune(s.Value)
	if idx < 0 {
		idx = int64(len(runes)) + idx
	}
	if idx < 0 || idx >= int64(len(runes)) {
		return object.NIL
	}
	return &object.String{Value: string(runes[idx])}
}

func evalTupleIndexExpression(left, index object.Object) object.Object {
	t := left.(*object.Tuple)
	idx := index.(*object.Integer).Value
	if idx < 0 || idx >= int64(len(t.Elements)) {
		return newError("tuple index out of bounds: %d", idx)
	}
	return t.Elements[idx]
}

func evalListLiteral(node *ast.ListLiteral, env *object.Environment) object.Object {
	elements := evalExpressions(node.Elements, env)
	if len(elements) == 1 && isError(elements[0]) {
		return elements[0]
	}
	return &object.List{Elements: elements}
}

func evalMapLiteral(node *ast.MapLiteral, env *object.Environment) object.Object {
	m := &object.Map{}
	for _, keyNode := range node.Order {
		valNode := node.Pairs[keyNode]
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}
		val := Eval(valNode, env)
		if isError(val) {
			return val
		}
		m.Pairs = append(m.Pairs, object.MapPair{Key: key, Value: val})
	}
	return m
}

func evalTupleLiteral(node *ast.TupleLiteral, env *object.Environment) object.Object {
	elements := evalExpressions(node.Elements, env)
	if len(elements) == 1 && isError(elements[0]) {
		return elements[0]
	}
	return &object.Tuple{Elements: elements}
}

func evalMatchExpression(node *ast.MatchExpression, env *object.Environment) object.Object {
	subject := Eval(node.Subject, env)
	if isError(subject) {
		return subject
	}

	for _, arm := range node.Arms {
		// Check if pattern matches
		if matchPattern(arm.Pattern, subject, env) {
			// Check guard if present
			if arm.Guard != nil {
				guardEnv := object.NewEnclosedEnvironment(env)
				bindPattern(arm.Pattern, subject, guardEnv)
				guardVal := Eval(arm.Guard, guardEnv)
				if !object.IsTruthy(guardVal) {
					continue
				}
			}
			// Execute body with pattern bindings
			bodyEnv := object.NewEnclosedEnvironment(env)
			bindPattern(arm.Pattern, subject, bodyEnv)
			return Eval(arm.Body, bodyEnv)
		}
	}

	return object.NIL
}

func matchPattern(pattern ast.Expression, value object.Object, env *object.Environment) bool {
	switch p := pattern.(type) {
	case *ast.Wildcard:
		return true
	case *ast.Identifier:
		return true // Identifiers always match and bind
	case *ast.IntegerLiteral:
		if v, ok := value.(*object.Integer); ok {
			return v.Value == p.Value
		}
		return false
	case *ast.FloatLiteral:
		if v, ok := value.(*object.Float); ok {
			return v.Value == p.Value
		}
		return false
	case *ast.StringLiteral:
		if v, ok := value.(*object.String); ok {
			return v.Value == p.Value
		}
		return false
	case *ast.BooleanLiteral:
		if v, ok := value.(*object.Boolean); ok {
			return v.Value == p.Value
		}
		return false
	case *ast.NilLiteral:
		_, ok := value.(*object.Nil)
		return ok
	default:
		return true
	}
}

func bindPattern(pattern ast.Expression, value object.Object, env *object.Environment) {
	switch p := pattern.(type) {
	case *ast.Identifier:
		if p.Value != "_" {
			env.Set(p.Value, value, false)
		}
	case *ast.TupleLiteral:
		if t, ok := value.(*object.Tuple); ok {
			for i, elem := range p.Elements {
				if i < len(t.Elements) {
					bindPattern(elem, t.Elements[i], env)
				}
			}
		}
	}
}

func evalForExpression(node *ast.ForExpression, env *object.Environment) object.Object {
	iterable := Eval(node.Iterable, env)
	if isError(iterable) {
		return iterable
	}

	var elements []object.Object
	switch iter := iterable.(type) {
	case *object.List:
		elements = iter.Elements
	case *object.String:
		for _, ch := range iter.Value {
			elements = append(elements, &object.String{Value: string(ch)})
		}
	default:
		return newError("cannot iterate over %s", iterable.Type())
	}

	var lastResult object.Object = object.NIL
	for _, elem := range elements {
		loopEnv := object.NewEnclosedEnvironment(env)

		// Bind pattern
		switch p := node.Pattern.(type) {
		case *ast.Identifier:
			loopEnv.Set(p.Value, elem, false)
		case *ast.TupleLiteral:
			if t, ok := elem.(*object.Tuple); ok {
				for i, pe := range p.Elements {
					if ident, ok := pe.(*ast.Identifier); ok && i < len(t.Elements) {
						loopEnv.Set(ident.Value, t.Elements[i], false)
					}
				}
			}
		}

		result := Eval(node.Body, loopEnv)
		if result != nil {
			if result.Type() == object.BREAK_OBJ {
				break
			}
			if result.Type() == object.CONTINUE_OBJ {
				continue
			}
			if result.Type() == object.RETURN_VALUE_OBJ || result.Type() == object.ERROR_OBJ {
				return result
			}
			lastResult = result
		}
	}
	return lastResult
}

func evalWhileExpression(node *ast.WhileExpression, env *object.Environment) object.Object {
	var lastResult object.Object = object.NIL
	for {
		condition := Eval(node.Condition, env)
		if isError(condition) {
			return condition
		}
		if !object.IsTruthy(condition) {
			break
		}

		result := Eval(node.Body, env)
		if result != nil {
			if result.Type() == object.BREAK_OBJ {
				break
			}
			if result.Type() == object.CONTINUE_OBJ {
				continue
			}
			if result.Type() == object.RETURN_VALUE_OBJ || result.Type() == object.ERROR_OBJ {
				return result
			}
			lastResult = result
		}
	}
	return lastResult
}

func evalStringInterpolation(node *ast.StringInterpolation, env *object.Environment) object.Object {
	var sb strings.Builder
	for _, part := range node.Parts {
		val := Eval(part, env)
		if isError(val) {
			return val
		}
		sb.WriteString(val.Inspect())
	}
	return &object.String{Value: sb.String()}
}

// Helpers

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func toFloat64(obj object.Object) float64 {
	switch o := obj.(type) {
	case *object.Float:
		return o.Value
	case *object.Integer:
		return float64(o.Value)
	default:
		return 0
	}
}

func intPow(base, exp int64) int64 {
	if exp < 0 {
		return 0
	}
	result := int64(1)
	for exp > 0 {
		if exp%2 == 1 {
			result *= base
		}
		base *= base
		exp /= 2
	}
	return result
}

func objectsEqual(a, b object.Object) bool {
	if a.Type() != b.Type() {
		return false
	}
	switch av := a.(type) {
	case *object.Integer:
		return av.Value == b.(*object.Integer).Value
	case *object.Float:
		return av.Value == b.(*object.Float).Value
	case *object.String:
		return av.Value == b.(*object.String).Value
	case *object.Boolean:
		return av.Value == b.(*object.Boolean).Value
	case *object.Nil:
		return true
	default:
		return a == b
	}
}
