// Package eval implements the tree-walking evaluator for GeoFlow.
package eval

import (
	"fmt"
	"math"
	"strings"

	"github.com/rogue780/geoflow/internal/ast"
	"github.com/rogue780/geoflow/internal/object"
	"github.com/rogue780/geoflow/internal/stdlib"
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
		return evalImportStatement(node, env)
	case *ast.TypeStatement:
		return evalTypeStatement(node, env)
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
		geom, err := stdlib.ParseWKT(node.Value)
		if err != nil {
			return newError("WKT parse error: %s", err)
		}
		return geom
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
	case *ast.ErrorPropagation:
		return evalErrorPropagation(node, env)
	case *ast.TryCatchExpression:
		return evalTryCatchExpression(node, env)
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
	case *object.Vector:
		elems := make([]float64, len(obj.Elements))
		for i, e := range obj.Elements {
			elems[i] = -e
		}
		return &object.Vector{Elements: elems}
	case *object.Complex:
		return &object.Complex{Real: -obj.Real, Imag: -obj.Imag}
	default:
		return newError("unknown operator: -%s", right.Type())
	}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.FLOAT_OBJ || right.Type() == object.FLOAT_OBJ:
		// Check for Vector * scalar or scalar * Vector
		if left.Type() == object.VECTOR_OBJ || right.Type() == object.VECTOR_OBJ {
			return evalVectorInfixExpression(operator, left, right)
		}
		return evalFloatInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ:
		return evalBooleanInfixExpression(operator, left, right)
	case left.Type() == object.VECTOR_OBJ || right.Type() == object.VECTOR_OBJ:
		return evalVectorInfixExpression(operator, left, right)
	case left.Type() == object.COMPLEX_OBJ || right.Type() == object.COMPLEX_OBJ:
		return evalComplexInfixExpression(operator, left, right)
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
	if node.Reverse {
		// f <| x  means f(x)
		return evalReversePipeline(node, env)
	}

	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}

	// x |> f  means f(x)
	// x |> f(y) means f(x, y)
	switch right := node.Right.(type) {
	case *ast.CallExpression:
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

func evalReversePipeline(node *ast.PipelineExpression, env *object.Environment) object.Object {
	// f <| x  means f(x)
	fn := Eval(node.Left, env)
	if isError(fn) {
		return fn
	}
	right := Eval(node.Right, env)
	if isError(right) {
		return right
	}
	return applyFunction(fn, []object.Object{right}, env)
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
		// Partial application: if fewer args than required params (that lack defaults), return closure
		requiredParams := 0
		for _, p := range fn.Parameters {
			if p.Default == nil {
				requiredParams++
			}
		}
		if len(args) < requiredParams {
			return partialApply(fn, args)
		}
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		if fn.Fn != nil {
			result := fn.Fn(args...)
			if result != nil {
				return result
			}
			return object.NIL
		}
		if fn.Name == "map" || fn.Name == "filter" || fn.Name == "reduce" {
			return evalHigherOrderBuiltin(fn.Name, args, env)
		}
		return newError("builtin function %s not implemented", fn.Name)
	case *object.ComposedFunction:
		// (f . g)(x) = f(g(x))
		innerResult := applyFunction(fn.Inner, args, env)
		if isError(innerResult) {
			return innerResult
		}
		return applyFunction(fn.Outer, []object.Object{innerResult}, env)
	case *object.StructDef:
		return applyStructConstructor(fn, args)
	case *object.EnumDef:
		// EnumDef itself isn't callable; variant constructors are registered as builtins
		return newError("cannot call enum type directly; use variant constructors")
	default:
		return newError("not a function: %s", fn.Type())
	}
}

func partialApply(fn *object.Function, args []object.Object) *object.Function {
	closureEnv := object.NewEnclosedEnvironment(fn.Env)
	for i, arg := range args {
		if i < len(fn.Parameters) {
			closureEnv.Set(fn.Parameters[i].Name.Value, arg, false)
		}
	}
	remaining := fn.Parameters[len(args):]
	return &object.Function{
		Name:       fn.Name,
		Parameters: remaining,
		Body:       fn.Body,
		Env:        closureEnv,
	}
}

func applyStructConstructor(sd *object.StructDef, args []object.Object) object.Object {
	if len(args) != len(sd.FieldNames) {
		return newError("struct %s expects %d fields, got %d", sd.TypeName, len(sd.FieldNames), len(args))
	}
	fields := make(map[string]object.Object, len(sd.FieldNames))
	order := make([]string, len(sd.FieldNames))
	for i, name := range sd.FieldNames {
		fields[name] = args[i]
		order[i] = name
	}
	return &object.Struct{
		TypeName: sd.TypeName,
		Fields:   fields,
		Order:    order,
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
	case *object.Module:
		if val, ok := obj.Exports[node.Field]; ok {
			return val
		}
		return newError("module '%s' has no export '%s'", obj.Name, node.Field)
	case *object.Struct:
		if val, ok := obj.Fields[node.Field]; ok {
			return val
		}
		return newError("struct '%s' has no field '%s'", obj.TypeName, node.Field)
	case *object.EnumVariant:
		return newError("no field '%s' on enum variant %s", node.Field, obj.VariantName)
	case *object.Set:
		return evalSetMethod(obj, node.Field, env)
	case *object.DateTime:
		return evalDateTimeMethod(obj, node.Field)
	case *object.Duration:
		return evalDurationMethod(obj, node.Field)
	case *object.CRS:
		return evalCRSMethod(obj, node.Field)
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
	case *object.Vector:
		return evalVectorMethod(obj, node.Field)
	case *object.Matrix:
		return evalMatrixMethod(obj, node.Field)
	case *object.Complex:
		return evalComplexMethod(obj, node.Field)
	case *object.Point:
		return evalPointMethod(obj, node.Field)
	case *object.LineString:
		return evalLineStringMethod(obj, node.Field)
	case *object.Polygon:
		return evalPolygonMethod(obj, node.Field)
	case *object.MultiPoint:
		return evalMultiPointMethod(obj, node.Field)
	case *object.MultiLineString:
		return evalMultiLineStringMethod(obj, node.Field)
	case *object.MultiPolygon:
		return evalMultiPolygonMethod(obj, node.Field)
	case *object.GeometryCollection:
		return evalGeometryCollectionMethod(obj, node.Field)
	case *object.BBox:
		return evalBBoxMethod(obj, node.Field)
	case *object.Feature:
		return evalFeatureMethod(obj, node.Field)
	case *object.FeatureCollection:
		return evalFeatureCollectionMethod(obj, node.Field)
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
	case "isBlank":
		return &object.Builtin{Name: "isBlank", Fn: func(args ...object.Object) object.Object {
			return object.NativeBoolToBooleanObject(strings.TrimSpace(s.Value) == "")
		}}
	case "capitalize":
		return &object.Builtin{Name: "capitalize", Fn: func(args ...object.Object) object.Object {
			if s.Value == "" {
				return s
			}
			r := []rune(s.Value)
			r[0] = []rune(strings.ToUpper(string(r[0])))[0]
			return &object.String{Value: string(r)}
		}}
	case "trimStart":
		return &object.Builtin{Name: "trimStart", Fn: func(args ...object.Object) object.Object {
			return &object.String{Value: strings.TrimLeft(s.Value, " \t\n\r")}
		}}
	case "trimEnd":
		return &object.Builtin{Name: "trimEnd", Fn: func(args ...object.Object) object.Object {
			return &object.String{Value: strings.TrimRight(s.Value, " \t\n\r")}
		}}
	case "indexOf":
		return &object.Builtin{Name: "indexOf", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("indexOf expects 1 argument")
			}
			sub, ok := args[0].(*object.String)
			if !ok {
				return newError("indexOf argument must be a string")
			}
			idx := strings.Index(s.Value, sub.Value)
			if idx < 0 {
				return object.NONE
			}
			return &object.Option{Value: &object.Integer{Value: int64(idx)}, IsSome: true}
		}}
	case "lastIndexOf":
		return &object.Builtin{Name: "lastIndexOf", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("lastIndexOf expects 1 argument")
			}
			sub, ok := args[0].(*object.String)
			if !ok {
				return newError("lastIndexOf argument must be a string")
			}
			idx := strings.LastIndex(s.Value, sub.Value)
			if idx < 0 {
				return object.NONE
			}
			return &object.Option{Value: &object.Integer{Value: int64(idx)}, IsSome: true}
		}}
	case "count":
		return &object.Builtin{Name: "count", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("count expects 1 argument")
			}
			sub, ok := args[0].(*object.String)
			if !ok {
				return newError("count argument must be a string")
			}
			return &object.Integer{Value: int64(strings.Count(s.Value, sub.Value))}
		}}
	case "charAt":
		return &object.Builtin{Name: "charAt", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("charAt expects 1 argument")
			}
			idx, ok := args[0].(*object.Integer)
			if !ok {
				return newError("charAt argument must be an integer")
			}
			runes := []rune(s.Value)
			i := int(idx.Value)
			if i < 0 || i >= len(runes) {
				return object.NONE
			}
			return &object.Option{Value: &object.String{Value: string(runes[i])}, IsSome: true}
		}}
	case "substring":
		return &object.Builtin{Name: "substring", Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("substring expects 2 arguments (start, end)")
			}
			start, ok1 := args[0].(*object.Integer)
			end, ok2 := args[1].(*object.Integer)
			if !ok1 || !ok2 {
				return newError("substring arguments must be integers")
			}
			runes := []rune(s.Value)
			st, en := int(start.Value), int(end.Value)
			if st < 0 {
				st = 0
			}
			if en > len(runes) {
				en = len(runes)
			}
			if st > en {
				return &object.String{Value: ""}
			}
			return &object.String{Value: string(runes[st:en])}
		}}
	case "lines":
		return &object.Builtin{Name: "lines", Fn: func(args ...object.Object) object.Object {
			parts := strings.Split(s.Value, "\n")
			elems := make([]object.Object, len(parts))
			for i, p := range parts {
				elems[i] = &object.String{Value: p}
			}
			return &object.List{Elements: elems}
		}}
	case "words":
		return &object.Builtin{Name: "words", Fn: func(args ...object.Object) object.Object {
			parts := strings.Fields(s.Value)
			elems := make([]object.Object, len(parts))
			for i, p := range parts {
				elems[i] = &object.String{Value: p}
			}
			return &object.List{Elements: elems}
		}}
	case "replaceAll":
		return &object.Builtin{Name: "replaceAll", Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("replaceAll expects 2 arguments")
			}
			old, ok1 := args[0].(*object.String)
			new, ok2 := args[1].(*object.String)
			if !ok1 || !ok2 {
				return newError("replaceAll arguments must be strings")
			}
			return &object.String{Value: strings.ReplaceAll(s.Value, old.Value, new.Value)}
		}}
	case "replaceFirst":
		return &object.Builtin{Name: "replaceFirst", Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("replaceFirst expects 2 arguments")
			}
			old, ok1 := args[0].(*object.String)
			new, ok2 := args[1].(*object.String)
			if !ok1 || !ok2 {
				return newError("replaceFirst arguments must be strings")
			}
			return &object.String{Value: strings.Replace(s.Value, old.Value, new.Value, 1)}
		}}
	case "reverse":
		return &object.Builtin{Name: "reverse", Fn: func(args ...object.Object) object.Object {
			runes := []rune(s.Value)
			for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
				runes[i], runes[j] = runes[j], runes[i]
			}
			return &object.String{Value: string(runes)}
		}}
	case "padStart":
		return &object.Builtin{Name: "padStart", Fn: func(args ...object.Object) object.Object {
			if len(args) < 1 || len(args) > 2 {
				return newError("padStart expects 1-2 arguments (length, char?)")
			}
			targetLen, ok := args[0].(*object.Integer)
			if !ok {
				return newError("padStart: first argument must be an integer")
			}
			padChar := " "
			if len(args) == 2 {
				if pc, ok := args[1].(*object.String); ok {
					padChar = pc.Value
				}
			}
			result := s.Value
			for len([]rune(result)) < int(targetLen.Value) {
				result = padChar + result
			}
			return &object.String{Value: result}
		}}
	case "padEnd":
		return &object.Builtin{Name: "padEnd", Fn: func(args ...object.Object) object.Object {
			if len(args) < 1 || len(args) > 2 {
				return newError("padEnd expects 1-2 arguments (length, char?)")
			}
			targetLen, ok := args[0].(*object.Integer)
			if !ok {
				return newError("padEnd: first argument must be an integer")
			}
			padChar := " "
			if len(args) == 2 {
				if pc, ok := args[1].(*object.String); ok {
					padChar = pc.Value
				}
			}
			result := s.Value
			for len([]rune(result)) < int(targetLen.Value) {
				result = result + padChar
			}
			return &object.String{Value: result}
		}}
	case "repeat":
		return &object.Builtin{Name: "repeat", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("repeat expects 1 argument (count)")
			}
			n, ok := args[0].(*object.Integer)
			if !ok {
				return newError("repeat: argument must be an integer")
			}
			return &object.String{Value: strings.Repeat(s.Value, int(n.Value))}
		}}
	case "toInt":
		return &object.Builtin{Name: "toInt", Fn: func(args ...object.Object) object.Object {
			v, err := fmt.Sscanf(strings.TrimSpace(s.Value), "%d", new(int64))
			if err != nil || v != 1 {
				return &object.Result{Value: &object.String{Value: fmt.Sprintf("cannot parse '%s' as int", s.Value)}, IsOk: false}
			}
			var n int64
			fmt.Sscanf(strings.TrimSpace(s.Value), "%d", &n)
			return &object.Result{Value: &object.Integer{Value: n}, IsOk: true}
		}}
	case "toFloat":
		return &object.Builtin{Name: "toFloat", Fn: func(args ...object.Object) object.Object {
			var n float64
			v, err := fmt.Sscanf(strings.TrimSpace(s.Value), "%g", &n)
			if err != nil || v != 1 {
				return &object.Result{Value: &object.String{Value: fmt.Sprintf("cannot parse '%s' as float", s.Value)}, IsOk: false}
			}
			return &object.Result{Value: &object.Float{Value: n}, IsOk: true}
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
	case "nonEmpty":
		return &object.Builtin{Name: "nonEmpty", Fn: func(args ...object.Object) object.Object {
			return object.NativeBoolToBooleanObject(len(l.Elements) > 0)
		}}
	case "head":
		return &object.Builtin{Name: "head", Fn: func(args ...object.Object) object.Object {
			if len(l.Elements) == 0 {
				return object.NONE
			}
			return &object.Option{Value: l.Elements[0], IsSome: true}
		}}
	case "tail":
		return &object.Builtin{Name: "tail", Fn: func(args ...object.Object) object.Object {
			if len(l.Elements) <= 1 {
				return &object.List{Elements: []object.Object{}}
			}
			return &object.List{Elements: l.Elements[1:]}
		}}
	case "init":
		return &object.Builtin{Name: "init", Fn: func(args ...object.Object) object.Object {
			if len(l.Elements) <= 1 {
				return &object.List{Elements: []object.Object{}}
			}
			return &object.List{Elements: l.Elements[:len(l.Elements)-1]}
		}}
	case "indexOf":
		return &object.Builtin{Name: "indexOf", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("indexOf expects 1 argument")
			}
			for i, elem := range l.Elements {
				if objectsEqual(elem, args[0]) {
					return &object.Option{Value: &object.Integer{Value: int64(i)}, IsSome: true}
				}
			}
			return object.NONE
		}}
	case "find":
		return &object.Builtin{Name: "find", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("find expects 1 argument (predicate)")
			}
			for _, elem := range l.Elements {
				result := applyFunction(args[0], []object.Object{elem}, env)
				if isError(result) {
					return result
				}
				if object.IsTruthy(result) {
					return &object.Option{Value: elem, IsSome: true}
				}
			}
			return object.NONE
		}}
	case "findIndex":
		return &object.Builtin{Name: "findIndex", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("findIndex expects 1 argument (predicate)")
			}
			for i, elem := range l.Elements {
				result := applyFunction(args[0], []object.Object{elem}, env)
				if isError(result) {
					return result
				}
				if object.IsTruthy(result) {
					return &object.Option{Value: &object.Integer{Value: int64(i)}, IsSome: true}
				}
			}
			return object.NONE
		}}
	case "any":
		return &object.Builtin{Name: "any", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("any expects 1 argument (predicate)")
			}
			for _, elem := range l.Elements {
				result := applyFunction(args[0], []object.Object{elem}, env)
				if isError(result) {
					return result
				}
				if object.IsTruthy(result) {
					return object.TRUE_OBJ
				}
			}
			return object.FALSE_OBJ
		}}
	case "all":
		return &object.Builtin{Name: "all", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("all expects 1 argument (predicate)")
			}
			for _, elem := range l.Elements {
				result := applyFunction(args[0], []object.Object{elem}, env)
				if isError(result) {
					return result
				}
				if !object.IsTruthy(result) {
					return object.FALSE_OBJ
				}
			}
			return object.TRUE_OBJ
		}}
	case "none":
		return &object.Builtin{Name: "none", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("none expects 1 argument (predicate)")
			}
			for _, elem := range l.Elements {
				result := applyFunction(args[0], []object.Object{elem}, env)
				if isError(result) {
					return result
				}
				if object.IsTruthy(result) {
					return object.FALSE_OBJ
				}
			}
			return object.TRUE_OBJ
		}}
	case "flatMap":
		return &object.Builtin{Name: "flatMap", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("flatMap expects 1 argument (function)")
			}
			var result []object.Object
			for _, elem := range l.Elements {
				val := applyFunction(args[0], []object.Object{elem}, env)
				if isError(val) {
					return val
				}
				if inner, ok := val.(*object.List); ok {
					result = append(result, inner.Elements...)
				} else {
					result = append(result, val)
				}
			}
			return &object.List{Elements: result}
		}}
	case "take":
		return &object.Builtin{Name: "take", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("take expects 1 argument")
			}
			n, ok := args[0].(*object.Integer)
			if !ok {
				return newError("take: argument must be an integer")
			}
			count := int(n.Value)
			if count > len(l.Elements) {
				count = len(l.Elements)
			}
			if count < 0 {
				count = 0
			}
			return &object.List{Elements: l.Elements[:count]}
		}}
	case "drop":
		return &object.Builtin{Name: "drop", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("drop expects 1 argument")
			}
			n, ok := args[0].(*object.Integer)
			if !ok {
				return newError("drop: argument must be an integer")
			}
			count := int(n.Value)
			if count > len(l.Elements) {
				count = len(l.Elements)
			}
			if count < 0 {
				count = 0
			}
			return &object.List{Elements: l.Elements[count:]}
		}}
	case "takeWhile":
		return &object.Builtin{Name: "takeWhile", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("takeWhile expects 1 argument (predicate)")
			}
			var result []object.Object
			for _, elem := range l.Elements {
				val := applyFunction(args[0], []object.Object{elem}, env)
				if isError(val) {
					return val
				}
				if !object.IsTruthy(val) {
					break
				}
				result = append(result, elem)
			}
			return &object.List{Elements: result}
		}}
	case "dropWhile":
		return &object.Builtin{Name: "dropWhile", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("dropWhile expects 1 argument (predicate)")
			}
			dropping := true
			var result []object.Object
			for _, elem := range l.Elements {
				if dropping {
					val := applyFunction(args[0], []object.Object{elem}, env)
					if isError(val) {
						return val
					}
					if !object.IsTruthy(val) {
						dropping = false
						result = append(result, elem)
					}
				} else {
					result = append(result, elem)
				}
			}
			return &object.List{Elements: result}
		}}
	case "slice":
		return &object.Builtin{Name: "slice", Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("slice expects 2 arguments (start, end)")
			}
			start, ok1 := args[0].(*object.Integer)
			end, ok2 := args[1].(*object.Integer)
			if !ok1 || !ok2 {
				return newError("slice arguments must be integers")
			}
			s, e := int(start.Value), int(end.Value)
			n := len(l.Elements)
			if s < 0 {
				s = n + s
			}
			if e < 0 {
				e = n + e
			}
			if s < 0 {
				s = 0
			}
			if e > n {
				e = n
			}
			if s > e {
				return &object.List{Elements: []object.Object{}}
			}
			return &object.List{Elements: l.Elements[s:e]}
		}}
	case "distinct":
		return &object.Builtin{Name: "distinct", Fn: func(args ...object.Object) object.Object {
			seen := make(map[string]bool)
			var result []object.Object
			for _, elem := range l.Elements {
				key := elem.Inspect()
				if !seen[key] {
					seen[key] = true
					result = append(result, elem)
				}
			}
			return &object.List{Elements: result}
		}}
	case "sortBy":
		return &object.Builtin{Name: "sortBy", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("sortBy expects 1 argument (key function)")
			}
			elems := make([]object.Object, len(l.Elements))
			copy(elems, l.Elements)
			keys := make([]object.Object, len(elems))
			for i, elem := range elems {
				key := applyFunction(args[0], []object.Object{elem}, env)
				if isError(key) {
					return key
				}
				keys[i] = key
			}
			// Simple insertion sort by keys
			for i := 1; i < len(elems); i++ {
				for j := i; j > 0; j-- {
					if compareForSort(keys[j], keys[j-1]) < 0 {
						elems[j], elems[j-1] = elems[j-1], elems[j]
						keys[j], keys[j-1] = keys[j-1], keys[j]
					} else {
						break
					}
				}
			}
			return &object.List{Elements: elems}
		}}
	case "sort":
		return &object.Builtin{Name: "sort", Fn: func(args ...object.Object) object.Object {
			builtins := object.GetBuiltins()
			return builtins["sort"].Fn(l)
		}}
	case "fold":
		return &object.Builtin{Name: "fold", Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("fold expects 2 arguments (initial, function)")
			}
			return evalHigherOrderBuiltin("reduce", []object.Object{l, args[0], args[1]}, env)
		}}
	case "product":
		return &object.Builtin{Name: "product", Fn: func(args ...object.Object) object.Object {
			builtins := object.GetBuiltins()
			return builtins["product"].Fn(l)
		}}
	case "min":
		return &object.Builtin{Name: "min", Fn: func(args ...object.Object) object.Object {
			if len(l.Elements) == 0 {
				return object.NONE
			}
			min := l.Elements[0]
			for _, elem := range l.Elements[1:] {
				if compareForSort(elem, min) < 0 {
					min = elem
				}
			}
			return &object.Option{Value: min, IsSome: true}
		}}
	case "max":
		return &object.Builtin{Name: "max", Fn: func(args ...object.Object) object.Object {
			if len(l.Elements) == 0 {
				return object.NONE
			}
			max := l.Elements[0]
			for _, elem := range l.Elements[1:] {
				if compareForSort(elem, max) > 0 {
					max = elem
				}
			}
			return &object.Option{Value: max, IsSome: true}
		}}
	case "prepend":
		return &object.Builtin{Name: "prepend", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("prepend expects 1 argument")
			}
			newElems := make([]object.Object, 0, len(l.Elements)+1)
			newElems = append(newElems, args[0])
			newElems = append(newElems, l.Elements...)
			return &object.List{Elements: newElems}
		}}
	case "zip":
		return &object.Builtin{Name: "zip", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("zip expects 1 argument (list)")
			}
			other, ok := args[0].(*object.List)
			if !ok {
				return newError("zip: argument must be a list")
			}
			minLen := len(l.Elements)
			if len(other.Elements) < minLen {
				minLen = len(other.Elements)
			}
			result := make([]object.Object, minLen)
			for i := 0; i < minLen; i++ {
				result[i] = &object.Tuple{Elements: []object.Object{l.Elements[i], other.Elements[i]}}
			}
			return &object.List{Elements: result}
		}}
	case "partition":
		return &object.Builtin{Name: "partition", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("partition expects 1 argument (predicate)")
			}
			var trueList, falseList []object.Object
			for _, elem := range l.Elements {
				val := applyFunction(args[0], []object.Object{elem}, env)
				if isError(val) {
					return val
				}
				if object.IsTruthy(val) {
					trueList = append(trueList, elem)
				} else {
					falseList = append(falseList, elem)
				}
			}
			return &object.Tuple{Elements: []object.Object{
				&object.List{Elements: trueList},
				&object.List{Elements: falseList},
			}}
		}}
	case "groupBy":
		return &object.Builtin{Name: "groupBy", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("groupBy expects 1 argument (key function)")
			}
			groups := make(map[string][]object.Object)
			var groupOrder []string
			var groupKeys []object.Object
			for _, elem := range l.Elements {
				key := applyFunction(args[0], []object.Object{elem}, env)
				if isError(key) {
					return key
				}
				keyStr := key.Inspect()
				if _, exists := groups[keyStr]; !exists {
					groupOrder = append(groupOrder, keyStr)
					groupKeys = append(groupKeys, key)
				}
				groups[keyStr] = append(groups[keyStr], elem)
			}
			m := &object.Map{}
			for i, keyStr := range groupOrder {
				m.Pairs = append(m.Pairs, object.MapPair{
					Key:   groupKeys[i],
					Value: &object.List{Elements: groups[keyStr]},
				})
			}
			return m
		}}
	case "chunked":
		return &object.Builtin{Name: "chunked", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("chunked expects 1 argument (size)")
			}
			size, ok := args[0].(*object.Integer)
			if !ok || size.Value <= 0 {
				return newError("chunked: argument must be a positive integer")
			}
			n := int(size.Value)
			var result []object.Object
			for i := 0; i < len(l.Elements); i += n {
				end := i + n
				if end > len(l.Elements) {
					end = len(l.Elements)
				}
				result = append(result, &object.List{Elements: l.Elements[i:end]})
			}
			return &object.List{Elements: result}
		}}
	case "forEach":
		return &object.Builtin{Name: "forEach", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("forEach expects 1 argument (function)")
			}
			for _, elem := range l.Elements {
				result := applyFunction(args[0], []object.Object{elem}, env)
				if isError(result) {
					return result
				}
			}
			return object.NIL
		}}
	case "join":
		return &object.Builtin{Name: "join", Fn: func(args ...object.Object) object.Object {
			sep := ""
			if len(args) == 1 {
				if s, ok := args[0].(*object.String); ok {
					sep = s.Value
				}
			}
			parts := make([]string, len(l.Elements))
			for i, e := range l.Elements {
				parts[i] = e.Inspect()
			}
			return &object.String{Value: strings.Join(parts, sep)}
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
	case "entries":
		return &object.Builtin{Name: "entries", Fn: func(args ...object.Object) object.Object {
			result := make([]object.Object, len(m.Pairs))
			for i, p := range m.Pairs {
				result[i] = &object.Tuple{Elements: []object.Object{p.Key, p.Value}}
			}
			return &object.List{Elements: result}
		}}
	case "getOrDefault":
		return &object.Builtin{Name: "getOrDefault", Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("getOrDefault expects 2 arguments (key, default)")
			}
			keyStr := args[0].Inspect()
			if s, ok := args[0].(*object.String); ok {
				keyStr = s.Value
			}
			val, found := m.Get(keyStr)
			if found {
				return val
			}
			return args[1]
		}}
	case "put":
		return &object.Builtin{Name: "put", Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("put expects 2 arguments (key, value)")
			}
			nm := &object.Map{}
			keyStr := ""
			if s, ok := args[0].(*object.String); ok {
				keyStr = s.Value
			}
			replaced := false
			for _, p := range m.Pairs {
				if s, ok := p.Key.(*object.String); ok && s.Value == keyStr {
					nm.Pairs = append(nm.Pairs, object.MapPair{Key: args[0], Value: args[1]})
					replaced = true
				} else {
					nm.Pairs = append(nm.Pairs, p)
				}
			}
			if !replaced {
				nm.Pairs = append(nm.Pairs, object.MapPair{Key: args[0], Value: args[1]})
			}
			return nm
		}}
	case "remove":
		return &object.Builtin{Name: "remove", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("remove expects 1 argument (key)")
			}
			nm := &object.Map{}
			keyStr := ""
			if s, ok := args[0].(*object.String); ok {
				keyStr = s.Value
			}
			for _, p := range m.Pairs {
				if s, ok := p.Key.(*object.String); ok && s.Value == keyStr {
					continue
				}
				nm.Pairs = append(nm.Pairs, p)
			}
			return nm
		}}
	case "mapValues":
		return &object.Builtin{Name: "mapValues", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("mapValues expects 1 argument (function)")
			}
			nm := &object.Map{}
			for _, p := range m.Pairs {
				val := applyFunction(args[0], []object.Object{p.Value}, env)
				if isError(val) {
					return val
				}
				nm.Pairs = append(nm.Pairs, object.MapPair{Key: p.Key, Value: val})
			}
			return nm
		}}
	case "mapKeys":
		return &object.Builtin{Name: "mapKeys", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("mapKeys expects 1 argument (function)")
			}
			nm := &object.Map{}
			for _, p := range m.Pairs {
				key := applyFunction(args[0], []object.Object{p.Key}, env)
				if isError(key) {
					return key
				}
				nm.Pairs = append(nm.Pairs, object.MapPair{Key: key, Value: p.Value})
			}
			return nm
		}}
	case "filter":
		return &object.Builtin{Name: "filter", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("filter expects 1 argument (predicate)")
			}
			nm := &object.Map{}
			for _, p := range m.Pairs {
				val := applyFunction(args[0], []object.Object{p.Key, p.Value}, env)
				if isError(val) {
					return val
				}
				if object.IsTruthy(val) {
					nm.Pairs = append(nm.Pairs, p)
				}
			}
			return nm
		}}
	case "merge":
		return &object.Builtin{Name: "merge", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("merge expects 1 argument (map)")
			}
			other, ok := args[0].(*object.Map)
			if !ok {
				return newError("merge: argument must be a Map")
			}
			nm := &object.Map{}
			nm.Pairs = append(nm.Pairs, m.Pairs...)
			for _, p := range other.Pairs {
				keyStr := ""
				if s, ok := p.Key.(*object.String); ok {
					keyStr = s.Value
				}
				found := false
				for i, existing := range nm.Pairs {
					if s, ok := existing.Key.(*object.String); ok && s.Value == keyStr {
						nm.Pairs[i] = p
						found = true
						break
					}
				}
				if !found {
					nm.Pairs = append(nm.Pairs, p)
				}
			}
			return nm
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
		// Check if identifier matches an enum unit variant
		if ev, ok := value.(*object.EnumVariant); ok {
			if p.Value == ev.VariantName && len(ev.Payload) == 0 {
				return true
			}
		}
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
	case *ast.CallExpression:
		// Enum variant with payload: Circle(r) matching Circle(5.0)
		if ident, ok := p.Function.(*ast.Identifier); ok {
			if ev, ok := value.(*object.EnumVariant); ok {
				return ident.Value == ev.VariantName && len(p.Arguments) == len(ev.Payload)
			}
		}
		return false
	case *ast.TupleLiteral:
		if t, ok := value.(*object.Tuple); ok {
			if len(p.Elements) != len(t.Elements) {
				return false
			}
			for i, elem := range p.Elements {
				if !matchPattern(elem, t.Elements[i], env) {
					return false
				}
			}
			return true
		}
		return false
	default:
		return true
	}
}

func bindPattern(pattern ast.Expression, value object.Object, env *object.Environment) {
	switch p := pattern.(type) {
	case *ast.Identifier:
		// Don't re-bind if it matches an enum unit variant name
		if ev, ok := value.(*object.EnumVariant); ok {
			if p.Value == ev.VariantName && len(ev.Payload) == 0 {
				return
			}
		}
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
	case *ast.CallExpression:
		// Enum variant with payload: Circle(r) => bind r to payload
		if ev, ok := value.(*object.EnumVariant); ok {
			for i, arg := range p.Arguments {
				if i < len(ev.Payload) {
					bindPattern(arg, ev.Payload[i], env)
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

// ── Geometry method dispatch ──

func evalPointMethod(p *object.Point, method string) object.Object {
	switch method {
	case "x", "lon":
		return &object.Float{Value: p.Coord.X}
	case "y", "lat":
		return &object.Float{Value: p.Coord.Y}
	case "z", "elevation":
		if p.Coord.HasZ {
			return &object.Float{Value: p.Coord.Z}
		}
		return object.NIL
	case "hasZ":
		return object.NativeBoolToBooleanObject(p.Coord.HasZ)
	case "toWKT":
		return &object.Builtin{Name: "toWKT", Fn: func(args ...object.Object) object.Object {
			return &object.String{Value: p.ToWKT()}
		}}
	case "geomType":
		return &object.String{Value: p.GeomType()}
	case "isEmpty":
		return object.NativeBoolToBooleanObject(math.IsNaN(p.Coord.X) || math.IsNaN(p.Coord.Y))
	case "isValid":
		return object.NativeBoolToBooleanObject(!math.IsNaN(p.Coord.X) && !math.IsNaN(p.Coord.Y))
	case "dimension":
		return &object.Integer{Value: 0}
	case "srid":
		return &object.Integer{Value: int64(p.SRID)}
	case "setSRID":
		return &object.Builtin{Name: "setSRID", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("setSRID expects 1 argument (int)")
			}
			srid, ok := args[0].(*object.Integer)
			if !ok {
				return newError("setSRID expects an integer argument")
			}
			return &object.Point{Coord: p.Coord, SRID: int(srid.Value)}
		}}
	case "bounds":
		return &object.BBox{MinX: p.Coord.X, MinY: p.Coord.Y, MaxX: p.Coord.X, MaxY: p.Coord.Y}
	case "distanceTo":
		return &object.Builtin{Name: "distanceTo", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("distanceTo expects 1 argument (Point)")
			}
			other, ok := args[0].(*object.Point)
			if !ok {
				return newError("distanceTo expects a Point argument")
			}
			return &object.Float{Value: stdlib.EuclideanDistance(p.Coord, other.Coord)}
		}}
	case "azimuthTo":
		return &object.Builtin{Name: "azimuthTo", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("azimuthTo expects 1 argument (Point)")
			}
			other, ok := args[0].(*object.Point)
			if !ok {
				return newError("azimuthTo expects a Point argument")
			}
			dx := other.Coord.X - p.Coord.X
			dy := other.Coord.Y - p.Coord.Y
			return &object.Float{Value: math.Atan2(dy, dx)}
		}}
	// Spatial relationship methods
	case "intersects":
		return &object.Builtin{Name: "intersects", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("intersects expects 1 argument (geometry)")
			}
			return object.NativeBoolToBooleanObject(stdlib.SpatialIntersects(p, args[0]))
		}}
	case "contains":
		return &object.Builtin{Name: "contains", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("contains expects 1 argument (geometry)")
			}
			return object.NativeBoolToBooleanObject(stdlib.SpatialContains(p, args[0]))
		}}
	case "within":
		return &object.Builtin{Name: "within", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("within expects 1 argument (geometry)")
			}
			return object.NativeBoolToBooleanObject(stdlib.SpatialContains(args[0], p))
		}}
	case "disjoint":
		return &object.Builtin{Name: "disjoint", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("disjoint expects 1 argument (geometry)")
			}
			return object.NativeBoolToBooleanObject(!stdlib.SpatialIntersects(p, args[0]))
		}}
	// Affine transform methods
	case "translate":
		return &object.Builtin{Name: "translate", Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("translate expects 2 arguments (dx, dy)")
			}
			dx := toFloat64(args[0])
			dy := toFloat64(args[1])
			return &object.Point{Coord: object.Coordinate{X: p.Coord.X + dx, Y: p.Coord.Y + dy, Z: p.Coord.Z, HasZ: p.Coord.HasZ}, SRID: p.SRID}
		}}
	case "scale":
		return &object.Builtin{Name: "scale", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("scale expects 1 argument (factor)")
			}
			f := toFloat64(args[0])
			return &object.Point{Coord: object.Coordinate{X: p.Coord.X * f, Y: p.Coord.Y * f, Z: p.Coord.Z, HasZ: p.Coord.HasZ}, SRID: p.SRID}
		}}
	case "rotate":
		return &object.Builtin{Name: "rotate", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("rotate expects 1 argument (angle in radians)")
			}
			angle := toFloat64(args[0])
			cosA := math.Cos(angle)
			sinA := math.Sin(angle)
			nx := p.Coord.X*cosA - p.Coord.Y*sinA
			ny := p.Coord.X*sinA + p.Coord.Y*cosA
			return &object.Point{Coord: object.Coordinate{X: nx, Y: ny, Z: p.Coord.Z, HasZ: p.Coord.HasZ}, SRID: p.SRID}
		}}
	case "reflect":
		return &object.Builtin{Name: "reflect", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("reflect expects 1 argument (axis: \"x\" or \"y\")")
			}
			axis, ok := args[0].(*object.String)
			if !ok {
				return newError("reflect expects a string argument (\"x\" or \"y\")")
			}
			switch axis.Value {
			case "x":
				return &object.Point{Coord: object.Coordinate{X: p.Coord.X, Y: -p.Coord.Y, Z: p.Coord.Z, HasZ: p.Coord.HasZ}, SRID: p.SRID}
			case "y":
				return &object.Point{Coord: object.Coordinate{X: -p.Coord.X, Y: p.Coord.Y, Z: p.Coord.Z, HasZ: p.Coord.HasZ}, SRID: p.SRID}
			default:
				return newError("reflect axis must be \"x\" or \"y\"")
			}
		}}
	default:
		return newError("no field '%s' on Point", method)
	}
}

func evalLineStringMethod(ls *object.LineString, method string) object.Object {
	switch method {
	case "numPoints":
		return &object.Integer{Value: int64(len(ls.Coords))}
	case "startPoint":
		if len(ls.Coords) > 0 {
			return &object.Point{Coord: ls.Coords[0]}
		}
		return object.NIL
	case "endPoint":
		if len(ls.Coords) > 0 {
			return &object.Point{Coord: ls.Coords[len(ls.Coords)-1]}
		}
		return object.NIL
	case "isClosed":
		if len(ls.Coords) >= 2 {
			return object.NativeBoolToBooleanObject(ls.Coords[0].Equals(ls.Coords[len(ls.Coords)-1]))
		}
		return object.FALSE_OBJ
	case "points":
		return &object.Builtin{Name: "points", Fn: func(args ...object.Object) object.Object {
			elems := make([]object.Object, len(ls.Coords))
			for i, c := range ls.Coords {
				elems[i] = &object.Point{Coord: c}
			}
			return &object.List{Elements: elems}
		}}
	case "toWKT":
		return &object.Builtin{Name: "toWKT", Fn: func(args ...object.Object) object.Object {
			return &object.String{Value: ls.ToWKT()}
		}}
	case "geomType":
		return &object.String{Value: ls.GeomType()}
	case "reverse":
		return &object.Builtin{Name: "reverse", Fn: func(args ...object.Object) object.Object {
			coords := make([]object.Coordinate, len(ls.Coords))
			for i, c := range ls.Coords {
				coords[len(ls.Coords)-1-i] = c
			}
			return &object.LineString{Coords: coords}
		}}
	case "isEmpty":
		return object.NativeBoolToBooleanObject(len(ls.Coords) == 0)
	case "isValid":
		return object.NativeBoolToBooleanObject(len(ls.Coords) >= 2)
	case "isRing":
		if len(ls.Coords) < 4 {
			return object.FALSE_OBJ
		}
		closed := ls.Coords[0].Equals(ls.Coords[len(ls.Coords)-1])
		if !closed {
			return object.FALSE_OBJ
		}
		// Check simplicity: no non-adjacent segment intersections
		simple := lineStringIsSimple(ls.Coords)
		return object.NativeBoolToBooleanObject(simple)
	case "isSimple":
		if len(ls.Coords) < 2 {
			return object.TRUE_OBJ
		}
		return object.NativeBoolToBooleanObject(lineStringIsSimple(ls.Coords))
	case "dimension":
		return &object.Integer{Value: 1}
	case "srid":
		return &object.Integer{Value: int64(ls.SRID)}
	case "setSRID":
		return &object.Builtin{Name: "setSRID", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("setSRID expects 1 argument (int)")
			}
			srid, ok := args[0].(*object.Integer)
			if !ok {
				return newError("setSRID expects an integer argument")
			}
			newCoords := make([]object.Coordinate, len(ls.Coords))
			copy(newCoords, ls.Coords)
			return &object.LineString{Coords: newCoords, SRID: int(srid.Value)}
		}}
	case "bounds":
		if len(ls.Coords) == 0 {
			return &object.BBox{}
		}
		return stdlib.Envelope(ls.Coords)
	case "length":
		return &object.Float{Value: stdlib.LineLength(ls.Coords)}
	case "lengthGeodesic":
		return &object.Float{Value: stdlib.LineLengthGeodesic(ls.Coords)}
	case "interpolate":
		return &object.Builtin{Name: "interpolate", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("interpolate expects 1 argument (fraction 0-1)")
			}
			frac := toFloat64(args[0])
			if frac < 0 || frac > 1 {
				return newError("interpolate fraction must be between 0 and 1")
			}
			if len(ls.Coords) < 2 {
				return newError("interpolate requires at least 2 points")
			}
			totalLen := stdlib.LineLength(ls.Coords)
			targetLen := totalLen * frac
			accumulated := 0.0
			for i := 1; i < len(ls.Coords); i++ {
				segLen := stdlib.EuclideanDistance(ls.Coords[i-1], ls.Coords[i])
				if accumulated+segLen >= targetLen {
					t := 0.0
					if segLen > 0 {
						t = (targetLen - accumulated) / segLen
					}
					return &object.Point{Coord: object.Coordinate{
						X: ls.Coords[i-1].X + t*(ls.Coords[i].X-ls.Coords[i-1].X),
						Y: ls.Coords[i-1].Y + t*(ls.Coords[i].Y-ls.Coords[i-1].Y),
					}}
				}
				accumulated += segLen
			}
			last := ls.Coords[len(ls.Coords)-1]
			return &object.Point{Coord: last}
		}}
	case "locatePoint":
		return &object.Builtin{Name: "locatePoint", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("locatePoint expects 1 argument (Point)")
			}
			pt, ok := args[0].(*object.Point)
			if !ok {
				return newError("locatePoint expects a Point argument")
			}
			if len(ls.Coords) < 2 {
				return newError("locatePoint requires at least 2 points")
			}
			totalLen := stdlib.LineLength(ls.Coords)
			if totalLen == 0 {
				return &object.Float{Value: 0}
			}
			minDist := math.MaxFloat64
			minFrac := 0.0
			accumulated := 0.0
			for i := 1; i < len(ls.Coords); i++ {
				segLen := stdlib.EuclideanDistance(ls.Coords[i-1], ls.Coords[i])
				// Project point onto segment
				dx := ls.Coords[i].X - ls.Coords[i-1].X
				dy := ls.Coords[i].Y - ls.Coords[i-1].Y
				t := 0.0
				if segLen > 0 {
					t = ((pt.Coord.X-ls.Coords[i-1].X)*dx + (pt.Coord.Y-ls.Coords[i-1].Y)*dy) / (dx*dx + dy*dy)
				}
				if t < 0 {
					t = 0
				}
				if t > 1 {
					t = 1
				}
				proj := object.Coordinate{
					X: ls.Coords[i-1].X + t*dx,
					Y: ls.Coords[i-1].Y + t*dy,
				}
				d := stdlib.EuclideanDistance(pt.Coord, proj)
				if d < minDist {
					minDist = d
					minFrac = (accumulated + t*segLen) / totalLen
				}
				accumulated += segLen
			}
			return &object.Float{Value: minFrac}
		}}
	case "substring":
		return &object.Builtin{Name: "substring", Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("substring expects 2 arguments (startFrac, endFrac)")
			}
			startFrac := toFloat64(args[0])
			endFrac := toFloat64(args[1])
			if startFrac < 0 {
				startFrac = 0
			}
			if endFrac > 1 {
				endFrac = 1
			}
			if startFrac > endFrac {
				startFrac, endFrac = endFrac, startFrac
			}
			if len(ls.Coords) < 2 {
				return &object.LineString{Coords: nil, SRID: ls.SRID}
			}
			totalLen := stdlib.LineLength(ls.Coords)
			startDist := totalLen * startFrac
			endDist := totalLen * endFrac
			var result []object.Coordinate
			accumulated := 0.0
			started := false
			for i := 1; i < len(ls.Coords); i++ {
				segLen := stdlib.EuclideanDistance(ls.Coords[i-1], ls.Coords[i])
				segEnd := accumulated + segLen
				if !started && segEnd >= startDist {
					t := 0.0
					if segLen > 0 {
						t = (startDist - accumulated) / segLen
					}
					result = append(result, object.Coordinate{
						X: ls.Coords[i-1].X + t*(ls.Coords[i].X-ls.Coords[i-1].X),
						Y: ls.Coords[i-1].Y + t*(ls.Coords[i].Y-ls.Coords[i-1].Y),
					})
					started = true
				}
				if started {
					if segEnd >= endDist {
						t := 0.0
						if segLen > 0 {
							t = (endDist - accumulated) / segLen
						}
						result = append(result, object.Coordinate{
							X: ls.Coords[i-1].X + t*(ls.Coords[i].X-ls.Coords[i-1].X),
							Y: ls.Coords[i-1].Y + t*(ls.Coords[i].Y-ls.Coords[i-1].Y),
						})
						break
					}
					result = append(result, ls.Coords[i])
				}
				accumulated = segEnd
			}
			if len(result) < 2 {
				// Degenerate: return a linestring with duplicated point
				if len(result) == 1 {
					result = append(result, result[0])
				}
			}
			return &object.LineString{Coords: result, SRID: ls.SRID}
		}}
	case "merge":
		return &object.Builtin{Name: "merge", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("merge expects 1 argument (LineString)")
			}
			other, ok := args[0].(*object.LineString)
			if !ok {
				return newError("merge expects a LineString argument")
			}
			newCoords := make([]object.Coordinate, 0, len(ls.Coords)+len(other.Coords))
			newCoords = append(newCoords, ls.Coords...)
			// If the last point of ls matches the first point of other, skip the duplicate
			if len(ls.Coords) > 0 && len(other.Coords) > 0 && ls.Coords[len(ls.Coords)-1].Equals(other.Coords[0]) {
				newCoords = append(newCoords, other.Coords[1:]...)
			} else {
				newCoords = append(newCoords, other.Coords...)
			}
			return &object.LineString{Coords: newCoords, SRID: ls.SRID}
		}}
	case "centroid":
		if len(ls.Coords) == 0 {
			return object.NIL
		}
		cx, cy := 0.0, 0.0
		for _, c := range ls.Coords {
			cx += c.X
			cy += c.Y
		}
		n := float64(len(ls.Coords))
		return &object.Point{Coord: object.Coordinate{X: cx / n, Y: cy / n}}
	// Spatial relationship methods
	case "intersects":
		return &object.Builtin{Name: "intersects", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("intersects expects 1 argument (geometry)")
			}
			return object.NativeBoolToBooleanObject(stdlib.SpatialIntersects(ls, args[0]))
		}}
	case "contains":
		return &object.Builtin{Name: "contains", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("contains expects 1 argument (geometry)")
			}
			return object.NativeBoolToBooleanObject(stdlib.SpatialContains(ls, args[0]))
		}}
	case "within":
		return &object.Builtin{Name: "within", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("within expects 1 argument (geometry)")
			}
			return object.NativeBoolToBooleanObject(stdlib.SpatialContains(args[0], ls))
		}}
	case "disjoint":
		return &object.Builtin{Name: "disjoint", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("disjoint expects 1 argument (geometry)")
			}
			return object.NativeBoolToBooleanObject(!stdlib.SpatialIntersects(ls, args[0]))
		}}
	// Affine transform methods
	case "translate":
		return &object.Builtin{Name: "translate", Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("translate expects 2 arguments (dx, dy)")
			}
			dx := toFloat64(args[0])
			dy := toFloat64(args[1])
			coords := make([]object.Coordinate, len(ls.Coords))
			for i, c := range ls.Coords {
				coords[i] = object.Coordinate{X: c.X + dx, Y: c.Y + dy, Z: c.Z, HasZ: c.HasZ}
			}
			return &object.LineString{Coords: coords, SRID: ls.SRID}
		}}
	case "scale":
		return &object.Builtin{Name: "scale", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("scale expects 1 argument (factor)")
			}
			f := toFloat64(args[0])
			coords := make([]object.Coordinate, len(ls.Coords))
			for i, c := range ls.Coords {
				coords[i] = object.Coordinate{X: c.X * f, Y: c.Y * f, Z: c.Z, HasZ: c.HasZ}
			}
			return &object.LineString{Coords: coords, SRID: ls.SRID}
		}}
	case "rotate":
		return &object.Builtin{Name: "rotate", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("rotate expects 1 argument (angle in radians)")
			}
			angle := toFloat64(args[0])
			cosA := math.Cos(angle)
			sinA := math.Sin(angle)
			coords := make([]object.Coordinate, len(ls.Coords))
			for i, c := range ls.Coords {
				coords[i] = object.Coordinate{X: c.X*cosA - c.Y*sinA, Y: c.X*sinA + c.Y*cosA, Z: c.Z, HasZ: c.HasZ}
			}
			return &object.LineString{Coords: coords, SRID: ls.SRID}
		}}
	case "reflect":
		return &object.Builtin{Name: "reflect", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("reflect expects 1 argument (axis: \"x\" or \"y\")")
			}
			axis, ok := args[0].(*object.String)
			if !ok {
				return newError("reflect expects a string argument (\"x\" or \"y\")")
			}
			coords := make([]object.Coordinate, len(ls.Coords))
			switch axis.Value {
			case "x":
				for i, c := range ls.Coords {
					coords[i] = object.Coordinate{X: c.X, Y: -c.Y, Z: c.Z, HasZ: c.HasZ}
				}
			case "y":
				for i, c := range ls.Coords {
					coords[i] = object.Coordinate{X: -c.X, Y: c.Y, Z: c.Z, HasZ: c.HasZ}
				}
			default:
				return newError("reflect axis must be \"x\" or \"y\"")
			}
			return &object.LineString{Coords: coords, SRID: ls.SRID}
		}}
	default:
		return newError("no field '%s' on LineString", method)
	}
}

func evalPolygonMethod(p *object.Polygon, method string) object.Object {
	switch method {
	case "exteriorRing":
		return &object.Builtin{Name: "exteriorRing", Fn: func(args ...object.Object) object.Object {
			return &object.LineString{Coords: p.ExteriorRing}
		}}
	case "numInteriorRings":
		return &object.Integer{Value: int64(len(p.InteriorRings))}
	case "interiorRing":
		return &object.Builtin{Name: "interiorRing", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("interiorRing expects 1 argument (index)")
			}
			idx, ok := args[0].(*object.Integer)
			if !ok {
				return newError("interiorRing expects an integer index")
			}
			i := int(idx.Value)
			if i < 0 || i >= len(p.InteriorRings) {
				return newError("interior ring index out of bounds: %d", i)
			}
			return &object.LineString{Coords: p.InteriorRings[i]}
		}}
	case "toWKT":
		return &object.Builtin{Name: "toWKT", Fn: func(args ...object.Object) object.Object {
			return &object.String{Value: p.ToWKT()}
		}}
	case "geomType":
		return &object.String{Value: p.GeomType()}
	case "isEmpty":
		return object.NativeBoolToBooleanObject(len(p.ExteriorRing) == 0)
	case "isValid":
		// Valid polygon: exterior ring has >=4 points and is closed
		valid := len(p.ExteriorRing) >= 4 &&
			p.ExteriorRing[0].Equals(p.ExteriorRing[len(p.ExteriorRing)-1])
		return object.NativeBoolToBooleanObject(valid)
	case "dimension":
		return &object.Integer{Value: 2}
	case "srid":
		return &object.Integer{Value: int64(p.SRID)}
	case "setSRID":
		return &object.Builtin{Name: "setSRID", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("setSRID expects 1 argument (int)")
			}
			srid, ok := args[0].(*object.Integer)
			if !ok {
				return newError("setSRID expects an integer argument")
			}
			extCopy := make([]object.Coordinate, len(p.ExteriorRing))
			copy(extCopy, p.ExteriorRing)
			intCopy := make([][]object.Coordinate, len(p.InteriorRings))
			for i, r := range p.InteriorRings {
				intCopy[i] = make([]object.Coordinate, len(r))
				copy(intCopy[i], r)
			}
			return &object.Polygon{ExteriorRing: extCopy, InteriorRings: intCopy, SRID: int(srid.Value)}
		}}
	case "bounds":
		if len(p.ExteriorRing) == 0 {
			return &object.BBox{}
		}
		return stdlib.Envelope(p.ExteriorRing)
	case "area":
		a := math.Abs(stdlib.ShoelaceArea(p.ExteriorRing))
		for _, ring := range p.InteriorRings {
			a -= math.Abs(stdlib.ShoelaceArea(ring))
		}
		return &object.Float{Value: a}
	case "signedArea":
		return &object.Float{Value: stdlib.ShoelaceArea(p.ExteriorRing)}
	case "perimeter":
		return &object.Float{Value: stdlib.LineLength(p.ExteriorRing)}
	case "centroid":
		if len(p.ExteriorRing) == 0 {
			return object.NIL
		}
		c := stdlib.CentroidOfRing(p.ExteriorRing)
		return &object.Point{Coord: c}
	case "removeHoles":
		extCopy := make([]object.Coordinate, len(p.ExteriorRing))
		copy(extCopy, p.ExteriorRing)
		return &object.Polygon{ExteriorRing: extCopy, SRID: p.SRID}
	case "orientExteriorCCW":
		// Positive signed area = CCW, negative = CW
		area := stdlib.ShoelaceArea(p.ExteriorRing)
		if area < 0 {
			// Reverse exterior ring to make it CCW
			ext := make([]object.Coordinate, len(p.ExteriorRing))
			for i, c := range p.ExteriorRing {
				ext[len(p.ExteriorRing)-1-i] = c
			}
			intCopy := make([][]object.Coordinate, len(p.InteriorRings))
			for i, r := range p.InteriorRings {
				intCopy[i] = make([]object.Coordinate, len(r))
				copy(intCopy[i], r)
			}
			return &object.Polygon{ExteriorRing: ext, InteriorRings: intCopy, SRID: p.SRID}
		}
		// Already CCW, return copy
		extCopy := make([]object.Coordinate, len(p.ExteriorRing))
		copy(extCopy, p.ExteriorRing)
		intCopy := make([][]object.Coordinate, len(p.InteriorRings))
		for i, r := range p.InteriorRings {
			intCopy[i] = make([]object.Coordinate, len(r))
			copy(intCopy[i], r)
		}
		return &object.Polygon{ExteriorRing: extCopy, InteriorRings: intCopy, SRID: p.SRID}
	case "numPoints":
		total := len(p.ExteriorRing)
		for _, ring := range p.InteriorRings {
			total += len(ring)
		}
		return &object.Integer{Value: int64(total)}
	// Spatial relationship methods
	case "intersects":
		return &object.Builtin{Name: "intersects", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("intersects expects 1 argument (geometry)")
			}
			return object.NativeBoolToBooleanObject(stdlib.SpatialIntersects(p, args[0]))
		}}
	case "contains":
		return &object.Builtin{Name: "contains", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("contains expects 1 argument (geometry)")
			}
			return object.NativeBoolToBooleanObject(stdlib.SpatialContains(p, args[0]))
		}}
	case "within":
		return &object.Builtin{Name: "within", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("within expects 1 argument (geometry)")
			}
			return object.NativeBoolToBooleanObject(stdlib.SpatialContains(args[0], p))
		}}
	case "disjoint":
		return &object.Builtin{Name: "disjoint", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("disjoint expects 1 argument (geometry)")
			}
			return object.NativeBoolToBooleanObject(!stdlib.SpatialIntersects(p, args[0]))
		}}
	// Affine transform methods
	case "translate":
		return &object.Builtin{Name: "translate", Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("translate expects 2 arguments (dx, dy)")
			}
			dx := toFloat64(args[0])
			dy := toFloat64(args[1])
			return transformPolygon(p, func(c object.Coordinate) object.Coordinate {
				return object.Coordinate{X: c.X + dx, Y: c.Y + dy, Z: c.Z, HasZ: c.HasZ}
			})
		}}
	case "scale":
		return &object.Builtin{Name: "scale", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("scale expects 1 argument (factor)")
			}
			f := toFloat64(args[0])
			return transformPolygon(p, func(c object.Coordinate) object.Coordinate {
				return object.Coordinate{X: c.X * f, Y: c.Y * f, Z: c.Z, HasZ: c.HasZ}
			})
		}}
	case "rotate":
		return &object.Builtin{Name: "rotate", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("rotate expects 1 argument (angle in radians)")
			}
			angle := toFloat64(args[0])
			cosA := math.Cos(angle)
			sinA := math.Sin(angle)
			return transformPolygon(p, func(c object.Coordinate) object.Coordinate {
				return object.Coordinate{X: c.X*cosA - c.Y*sinA, Y: c.X*sinA + c.Y*cosA, Z: c.Z, HasZ: c.HasZ}
			})
		}}
	case "reflect":
		return &object.Builtin{Name: "reflect", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("reflect expects 1 argument (axis: \"x\" or \"y\")")
			}
			axis, ok := args[0].(*object.String)
			if !ok {
				return newError("reflect expects a string argument (\"x\" or \"y\")")
			}
			switch axis.Value {
			case "x":
				return transformPolygon(p, func(c object.Coordinate) object.Coordinate {
					return object.Coordinate{X: c.X, Y: -c.Y, Z: c.Z, HasZ: c.HasZ}
				})
			case "y":
				return transformPolygon(p, func(c object.Coordinate) object.Coordinate {
					return object.Coordinate{X: -c.X, Y: c.Y, Z: c.Z, HasZ: c.HasZ}
				})
			default:
				return newError("reflect axis must be \"x\" or \"y\"")
			}
		}}
	default:
		return newError("no field '%s' on Polygon", method)
	}
}

func evalMultiPointMethod(mp *object.MultiPoint, method string) object.Object {
	switch method {
	case "numGeometries":
		return &object.Integer{Value: int64(len(mp.Points))}
	case "geometries":
		return &object.Builtin{Name: "geometries", Fn: func(args ...object.Object) object.Object {
			elems := make([]object.Object, len(mp.Points))
			for i, p := range mp.Points {
				elems[i] = p
			}
			return &object.List{Elements: elems}
		}}
	case "toWKT":
		return &object.Builtin{Name: "toWKT", Fn: func(args ...object.Object) object.Object {
			return &object.String{Value: mp.ToWKT()}
		}}
	case "geomType":
		return &object.String{Value: mp.GeomType()}
	case "isEmpty":
		return object.NativeBoolToBooleanObject(len(mp.Points) == 0)
	case "dimension":
		return &object.Integer{Value: 0}
	case "srid":
		return &object.Integer{Value: int64(mp.SRID)}
	case "setSRID":
		return &object.Builtin{Name: "setSRID", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("setSRID expects 1 argument (int)")
			}
			srid, ok := args[0].(*object.Integer)
			if !ok {
				return newError("setSRID expects an integer argument")
			}
			pts := make([]*object.Point, len(mp.Points))
			for i, p := range mp.Points {
				pts[i] = &object.Point{Coord: p.Coord, SRID: p.SRID}
			}
			return &object.MultiPoint{Points: pts, SRID: int(srid.Value)}
		}}
	case "bounds":
		coords := mp.Coordinates()
		if len(coords) == 0 {
			return &object.BBox{}
		}
		return stdlib.Envelope(coords)
	default:
		return newError("no field '%s' on MultiPoint", method)
	}
}

func evalMultiLineStringMethod(ml *object.MultiLineString, method string) object.Object {
	switch method {
	case "numGeometries":
		return &object.Integer{Value: int64(len(ml.Lines))}
	case "geometries":
		return &object.Builtin{Name: "geometries", Fn: func(args ...object.Object) object.Object {
			elems := make([]object.Object, len(ml.Lines))
			for i, l := range ml.Lines {
				elems[i] = l
			}
			return &object.List{Elements: elems}
		}}
	case "toWKT":
		return &object.Builtin{Name: "toWKT", Fn: func(args ...object.Object) object.Object {
			return &object.String{Value: ml.ToWKT()}
		}}
	case "geomType":
		return &object.String{Value: ml.GeomType()}
	case "isEmpty":
		return object.NativeBoolToBooleanObject(len(ml.Lines) == 0)
	case "dimension":
		return &object.Integer{Value: 1}
	case "srid":
		return &object.Integer{Value: int64(ml.SRID)}
	case "setSRID":
		return &object.Builtin{Name: "setSRID", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("setSRID expects 1 argument (int)")
			}
			srid, ok := args[0].(*object.Integer)
			if !ok {
				return newError("setSRID expects an integer argument")
			}
			lines := make([]*object.LineString, len(ml.Lines))
			for i, l := range ml.Lines {
				coords := make([]object.Coordinate, len(l.Coords))
				copy(coords, l.Coords)
				lines[i] = &object.LineString{Coords: coords, SRID: l.SRID}
			}
			return &object.MultiLineString{Lines: lines, SRID: int(srid.Value)}
		}}
	case "bounds":
		coords := ml.Coordinates()
		if len(coords) == 0 {
			return &object.BBox{}
		}
		return stdlib.Envelope(coords)
	default:
		return newError("no field '%s' on MultiLineString", method)
	}
}

func evalMultiPolygonMethod(mp *object.MultiPolygon, method string) object.Object {
	switch method {
	case "numGeometries":
		return &object.Integer{Value: int64(len(mp.Polygons))}
	case "geometries":
		return &object.Builtin{Name: "geometries", Fn: func(args ...object.Object) object.Object {
			elems := make([]object.Object, len(mp.Polygons))
			for i, p := range mp.Polygons {
				elems[i] = p
			}
			return &object.List{Elements: elems}
		}}
	case "toWKT":
		return &object.Builtin{Name: "toWKT", Fn: func(args ...object.Object) object.Object {
			return &object.String{Value: mp.ToWKT()}
		}}
	case "geomType":
		return &object.String{Value: mp.GeomType()}
	case "isEmpty":
		return object.NativeBoolToBooleanObject(len(mp.Polygons) == 0)
	case "dimension":
		return &object.Integer{Value: 2}
	case "srid":
		return &object.Integer{Value: int64(mp.SRID)}
	case "setSRID":
		return &object.Builtin{Name: "setSRID", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("setSRID expects 1 argument (int)")
			}
			srid, ok := args[0].(*object.Integer)
			if !ok {
				return newError("setSRID expects an integer argument")
			}
			polys := make([]*object.Polygon, len(mp.Polygons))
			for i, p := range mp.Polygons {
				extCopy := make([]object.Coordinate, len(p.ExteriorRing))
				copy(extCopy, p.ExteriorRing)
				intCopy := make([][]object.Coordinate, len(p.InteriorRings))
				for j, r := range p.InteriorRings {
					intCopy[j] = make([]object.Coordinate, len(r))
					copy(intCopy[j], r)
				}
				polys[i] = &object.Polygon{ExteriorRing: extCopy, InteriorRings: intCopy, SRID: p.SRID}
			}
			return &object.MultiPolygon{Polygons: polys, SRID: int(srid.Value)}
		}}
	case "bounds":
		coords := mp.Coordinates()
		if len(coords) == 0 {
			return &object.BBox{}
		}
		return stdlib.Envelope(coords)
	default:
		return newError("no field '%s' on MultiPolygon", method)
	}
}

func evalGeometryCollectionMethod(gc *object.GeometryCollection, method string) object.Object {
	switch method {
	case "numGeometries":
		return &object.Integer{Value: int64(len(gc.Geometries))}
	case "geometries":
		return &object.Builtin{Name: "geometries", Fn: func(args ...object.Object) object.Object {
			elems := make([]object.Object, len(gc.Geometries))
			for i, g := range gc.Geometries {
				elems[i] = g
			}
			return &object.List{Elements: elems}
		}}
	case "toWKT":
		return &object.Builtin{Name: "toWKT", Fn: func(args ...object.Object) object.Object {
			return &object.String{Value: gc.ToWKT()}
		}}
	case "geomType":
		return &object.String{Value: gc.GeomType()}
	case "isEmpty":
		return object.NativeBoolToBooleanObject(len(gc.Geometries) == 0)
	case "bounds":
		coords := gc.Coordinates()
		if len(coords) == 0 {
			return &object.BBox{}
		}
		return stdlib.Envelope(coords)
	default:
		return newError("no field '%s' on GeometryCollection", method)
	}
}

func evalBBoxMethod(bb *object.BBox, method string) object.Object {
	switch method {
	case "minX":
		return &object.Float{Value: bb.MinX}
	case "minY":
		return &object.Float{Value: bb.MinY}
	case "maxX":
		return &object.Float{Value: bb.MaxX}
	case "maxY":
		return &object.Float{Value: bb.MaxY}
	case "width":
		return &object.Float{Value: bb.MaxX - bb.MinX}
	case "height":
		return &object.Float{Value: bb.MaxY - bb.MinY}
	case "center":
		return &object.Point{Coord: object.Coordinate{X: (bb.MinX + bb.MaxX) / 2, Y: (bb.MinY + bb.MaxY) / 2}}
	default:
		return newError("no field '%s' on BBox", method)
	}
}

func evalFeatureMethod(f *object.Feature, method string) object.Object {
	switch method {
	case "geometry":
		return f.Geom
	case "properties":
		return f.Properties
	case "id":
		if f.ID != nil {
			return f.ID
		}
		return object.NIL
	case "geomType":
		return &object.String{Value: f.Geom.GeomType()}
	case "get":
		return &object.Builtin{Name: "get", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("get expects 1 argument (property name)")
			}
			key, ok := args[0].(*object.String)
			if !ok {
				return newError("get expects a string key")
			}
			val, found := f.Properties.Get(key.Value)
			if !found {
				return object.NIL
			}
			return val
		}}
	default:
		// Try property lookup
		val, found := f.Properties.Get(method)
		if found {
			return val
		}
		return newError("no field '%s' on Feature", method)
	}
}

func evalFeatureCollectionMethod(fc *object.FeatureCollection, method string) object.Object {
	switch method {
	case "numFeatures", "length":
		return &object.Integer{Value: int64(len(fc.Features))}
	case "features":
		return &object.Builtin{Name: "features", Fn: func(args ...object.Object) object.Object {
			elems := make([]object.Object, len(fc.Features))
			for i, f := range fc.Features {
				elems[i] = f
			}
			return &object.List{Elements: elems}
		}}
	default:
		return newError("no field '%s' on FeatureCollection", method)
	}
}

// ── Vector/Matrix/Complex methods and operators ──

func evalVectorMethod(v *object.Vector, method string) object.Object {
	switch method {
	case "length":
		return &object.Builtin{Name: "length", Fn: func(args ...object.Object) object.Object {
			return &object.Integer{Value: int64(len(v.Elements))}
		}}
	case "magnitude":
		return &object.Builtin{Name: "magnitude", Fn: func(args ...object.Object) object.Object {
			sum := 0.0
			for _, e := range v.Elements {
				sum += e * e
			}
			return &object.Float{Value: math.Sqrt(sum)}
		}}
	case "normalize":
		return &object.Builtin{Name: "normalize", Fn: func(args ...object.Object) object.Object {
			mag := 0.0
			for _, e := range v.Elements {
				mag += e * e
			}
			mag = math.Sqrt(mag)
			if mag == 0 {
				return newError("cannot normalize zero vector")
			}
			elems := make([]float64, len(v.Elements))
			for i, e := range v.Elements {
				elems[i] = e / mag
			}
			return &object.Vector{Elements: elems}
		}}
	case "toList":
		return &object.Builtin{Name: "toList", Fn: func(args ...object.Object) object.Object {
			elems := make([]object.Object, len(v.Elements))
			for i, e := range v.Elements {
				elems[i] = &object.Float{Value: e}
			}
			return &object.List{Elements: elems}
		}}
	case "x":
		if len(v.Elements) >= 1 {
			return &object.Float{Value: v.Elements[0]}
		}
		return newError("vector has no x component")
	case "y":
		if len(v.Elements) >= 2 {
			return &object.Float{Value: v.Elements[1]}
		}
		return newError("vector has no y component")
	case "z":
		if len(v.Elements) >= 3 {
			return &object.Float{Value: v.Elements[2]}
		}
		return newError("vector has no z component")
	default:
		return newError("no method '%s' on Vector", method)
	}
}

func evalMatrixMethod(m *object.Matrix, method string) object.Object {
	switch method {
	case "rows":
		return &object.Integer{Value: int64(m.Rows)}
	case "cols":
		return &object.Integer{Value: int64(m.Cols)}
	case "toList":
		return &object.Builtin{Name: "toList", Fn: func(args ...object.Object) object.Object {
			rows := make([]object.Object, m.Rows)
			for i, row := range m.Data {
				elems := make([]object.Object, len(row))
				for j, e := range row {
					elems[j] = &object.Float{Value: e}
				}
				rows[i] = &object.List{Elements: elems}
			}
			return &object.List{Elements: rows}
		}}
	case "get":
		return &object.Builtin{Name: "get", Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("get expects 2 arguments (row, col)")
			}
			r, ok1 := args[0].(*object.Integer)
			c, ok2 := args[1].(*object.Integer)
			if !ok1 || !ok2 {
				return newError("get expects integer arguments")
			}
			ri, ci := int(r.Value), int(c.Value)
			if ri < 0 || ri >= m.Rows || ci < 0 || ci >= m.Cols {
				return newError("matrix index out of bounds: [%d, %d]", ri, ci)
			}
			return &object.Float{Value: m.Data[ri][ci]}
		}}
	case "row":
		return &object.Builtin{Name: "row", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("row expects 1 argument")
			}
			idx, ok := args[0].(*object.Integer)
			if !ok {
				return newError("row expects an integer argument")
			}
			i := int(idx.Value)
			if i < 0 || i >= m.Rows {
				return newError("row index out of bounds: %d", i)
			}
			return &object.Vector{Elements: append([]float64{}, m.Data[i]...)}
		}}
	case "col":
		return &object.Builtin{Name: "col", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("col expects 1 argument")
			}
			idx, ok := args[0].(*object.Integer)
			if !ok {
				return newError("col expects an integer argument")
			}
			j := int(idx.Value)
			if j < 0 || j >= m.Cols {
				return newError("column index out of bounds: %d", j)
			}
			elems := make([]float64, m.Rows)
			for i := 0; i < m.Rows; i++ {
				elems[i] = m.Data[i][j]
			}
			return &object.Vector{Elements: elems}
		}}
	case "isSquare":
		return object.NativeBoolToBooleanObject(m.Rows == m.Cols)
	default:
		return newError("no method '%s' on Matrix", method)
	}
}

func evalComplexMethod(c *object.Complex, method string) object.Object {
	switch method {
	case "real":
		return &object.Float{Value: c.Real}
	case "imag":
		return &object.Float{Value: c.Imag}
	case "magnitude":
		return &object.Builtin{Name: "magnitude", Fn: func(args ...object.Object) object.Object {
			return &object.Float{Value: math.Sqrt(c.Real*c.Real + c.Imag*c.Imag)}
		}}
	case "conjugate":
		return &object.Builtin{Name: "conjugate", Fn: func(args ...object.Object) object.Object {
			return &object.Complex{Real: c.Real, Imag: -c.Imag}
		}}
	case "phase":
		return &object.Builtin{Name: "phase", Fn: func(args ...object.Object) object.Object {
			return &object.Float{Value: math.Atan2(c.Imag, c.Real)}
		}}
	default:
		return newError("no method '%s' on Complex", method)
	}
}

func evalVectorInfixExpression(operator string, left, right object.Object) object.Object {
	// Vector op Vector
	if lv, ok := left.(*object.Vector); ok {
		if rv, ok := right.(*object.Vector); ok {
			if len(lv.Elements) != len(rv.Elements) {
				return newError("vector length mismatch: %d vs %d", len(lv.Elements), len(rv.Elements))
			}
			elems := make([]float64, len(lv.Elements))
			switch operator {
			case "+":
				for i := range elems {
					elems[i] = lv.Elements[i] + rv.Elements[i]
				}
			case "-":
				for i := range elems {
					elems[i] = lv.Elements[i] - rv.Elements[i]
				}
			case "*":
				// Element-wise multiplication
				for i := range elems {
					elems[i] = lv.Elements[i] * rv.Elements[i]
				}
			case "==":
				for i := range lv.Elements {
					if lv.Elements[i] != rv.Elements[i] {
						return object.FALSE_OBJ
					}
				}
				return object.TRUE_OBJ
			case "!=":
				for i := range lv.Elements {
					if lv.Elements[i] != rv.Elements[i] {
						return object.TRUE_OBJ
					}
				}
				return object.FALSE_OBJ
			default:
				return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
			}
			return &object.Vector{Elements: elems}
		}
		// Vector op Scalar
		if scalar, ok := numericToFloat(right); ok {
			elems := make([]float64, len(lv.Elements))
			switch operator {
			case "*":
				for i := range elems {
					elems[i] = lv.Elements[i] * scalar
				}
			case "/":
				if scalar == 0 {
					return newError("division by zero")
				}
				for i := range elems {
					elems[i] = lv.Elements[i] / scalar
				}
			default:
				return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
			}
			return &object.Vector{Elements: elems}
		}
	}
	// Scalar op Vector
	if rv, ok := right.(*object.Vector); ok {
		if scalar, ok := numericToFloat(left); ok {
			if operator == "*" {
				elems := make([]float64, len(rv.Elements))
				for i := range elems {
					elems[i] = scalar * rv.Elements[i]
				}
				return &object.Vector{Elements: elems}
			}
		}
	}
	return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
}

func evalComplexInfixExpression(operator string, left, right object.Object) object.Object {
	lc := toComplex(left)
	rc := toComplex(right)
	if lc == nil || rc == nil {
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
	a := complex(lc.Real, lc.Imag)
	b := complex(rc.Real, rc.Imag)
	switch operator {
	case "+":
		r := a + b
		return &object.Complex{Real: real(r), Imag: imag(r)}
	case "-":
		r := a - b
		return &object.Complex{Real: real(r), Imag: imag(r)}
	case "*":
		r := a * b
		return &object.Complex{Real: real(r), Imag: imag(r)}
	case "/":
		if b == 0 {
			return newError("division by zero")
		}
		r := a / b
		return &object.Complex{Real: real(r), Imag: imag(r)}
	case "==":
		return object.NativeBoolToBooleanObject(a == b)
	case "!=":
		return object.NativeBoolToBooleanObject(a != b)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func toComplex(obj object.Object) *object.Complex {
	switch o := obj.(type) {
	case *object.Complex:
		return o
	case *object.Integer:
		return &object.Complex{Real: float64(o.Value), Imag: 0}
	case *object.Float:
		return &object.Complex{Real: o.Value, Imag: 0}
	default:
		return nil
	}
}

func numericToFloat(obj object.Object) (float64, bool) {
	switch o := obj.(type) {
	case *object.Float:
		return o.Value, true
	case *object.Integer:
		return float64(o.Value), true
	default:
		return 0, false
	}
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

// ── Import System ──

func evalImportStatement(node *ast.ImportStatement, env *object.Environment) object.Object {
	path := strings.Join(node.Path, ".")
	mod, ok := DefaultRegistry.Get(path)
	if !ok {
		return newError("unknown module: %s", path)
	}
	// Determine binding name: alias, or last segment of path
	bindName := node.Alias
	if bindName == "" {
		bindName = node.Path[len(node.Path)-1]
	}
	env.Set(bindName, mod, false)
	return object.NIL
}

// ── Type Statement (Struct + Enum instantiation) ──

func evalTypeStatement(node *ast.TypeStatement, env *object.Environment) object.Object {
	switch def := node.TypeExpr.(type) {
	case *ast.StructDefinition:
		fieldNames := make([]string, len(def.Fields))
		for i, f := range def.Fields {
			fieldNames[i] = f.Name
		}
		sd := &object.StructDef{
			TypeName:   node.Name,
			FieldNames: fieldNames,
		}
		env.Set(node.Name, sd, false)
		return object.NIL
	case *ast.EnumDefinition:
		variantNames := make([]string, len(def.Variants))
		variantArity := make(map[string]int, len(def.Variants))
		for i, v := range def.Variants {
			variantNames[i] = v.Name
			variantArity[v.Name] = len(v.Fields)
		}
		ed := &object.EnumDef{
			TypeName:     node.Name,
			VariantNames: variantNames,
			VariantArity: variantArity,
		}
		env.Set(node.Name, ed, false)
		// Register each variant as a constructor or constant
		for _, v := range def.Variants {
			vName := v.Name
			arity := len(v.Fields)
			typeName := node.Name
			if arity == 0 {
				// Unit variant: register as a constant
				env.Set(vName, &object.EnumVariant{
					TypeName:    typeName,
					VariantName: vName,
					Payload:     nil,
				}, false)
			} else {
				// Variant with payload: register as a constructor function
				expectedArity := arity
				env.Set(vName, &object.Builtin{
					Name: vName,
					Fn: func(args ...object.Object) object.Object {
						if len(args) != expectedArity {
							return newError("%s expects %d argument(s), got %d", vName, expectedArity, len(args))
						}
						return &object.EnumVariant{
							TypeName:    typeName,
							VariantName: vName,
							Payload:     args,
						}
					},
				}, false)
			}
		}
		return object.NIL
	default:
		// Type alias or other - no-op for now
		return object.NIL
	}
}

// ── Error Propagation ? ──

func evalErrorPropagation(node *ast.ErrorPropagation, env *object.Environment) object.Object {
	val := Eval(node.Expression, env)
	if isError(val) {
		return val
	}
	switch v := val.(type) {
	case *object.Result:
		if !v.IsOk {
			return &object.ReturnValue{Value: val}
		}
		return v.Value
	case *object.Option:
		if !v.IsSome {
			return &object.ReturnValue{Value: val}
		}
		return v.Value
	default:
		return val
	}
}

// ── Try-Catch ──

func evalTryCatchExpression(node *ast.TryCatchExpression, env *object.Environment) object.Object {
	result := Eval(node.TryBody, env)
	if errObj, ok := result.(*object.Error); ok {
		catchEnv := object.NewEnclosedEnvironment(env)
		if node.CatchParam != "" {
			catchEnv.Set(node.CatchParam, &object.String{Value: errObj.Message}, false)
		}
		return Eval(node.CatchBody, catchEnv)
	}
	return result
}

// ── Set Methods ──

func evalSetMethod(s *object.Set, method string, env *object.Environment) object.Object {
	switch method {
	case "size":
		return &object.Builtin{Name: "size", Fn: func(args ...object.Object) object.Object {
			return &object.Integer{Value: int64(len(s.Elements))}
		}}
	case "isEmpty":
		return &object.Builtin{Name: "isEmpty", Fn: func(args ...object.Object) object.Object {
			return object.NativeBoolToBooleanObject(len(s.Elements) == 0)
		}}
	case "contains":
		return &object.Builtin{Name: "contains", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("contains expects 1 argument")
			}
			return object.NativeBoolToBooleanObject(s.Contains(args[0]))
		}}
	case "add":
		return &object.Builtin{Name: "add", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("add expects 1 argument")
			}
			return s.Add(args[0])
		}}
	case "remove":
		return &object.Builtin{Name: "remove", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("remove expects 1 argument")
			}
			ns := object.NewSet()
			key := args[0].Inspect()
			for k, v := range s.Elements {
				if k != key {
					ns.Elements[k] = v
				}
			}
			return ns
		}}
	case "toList":
		return &object.Builtin{Name: "toList", Fn: func(args ...object.Object) object.Object {
			return s.ToList()
		}}
	case "union":
		return &object.Builtin{Name: "union", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("union expects 1 argument (Set)")
			}
			other, ok := args[0].(*object.Set)
			if !ok {
				return newError("union: argument must be a Set")
			}
			ns := object.NewSet()
			for k, v := range s.Elements {
				ns.Elements[k] = v
			}
			for k, v := range other.Elements {
				ns.Elements[k] = v
			}
			return ns
		}}
	case "intersection":
		return &object.Builtin{Name: "intersection", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("intersection expects 1 argument (Set)")
			}
			other, ok := args[0].(*object.Set)
			if !ok {
				return newError("intersection: argument must be a Set")
			}
			ns := object.NewSet()
			for k, v := range s.Elements {
				if _, ok := other.Elements[k]; ok {
					ns.Elements[k] = v
				}
			}
			return ns
		}}
	case "difference":
		return &object.Builtin{Name: "difference", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("difference expects 1 argument (Set)")
			}
			other, ok := args[0].(*object.Set)
			if !ok {
				return newError("difference: argument must be a Set")
			}
			ns := object.NewSet()
			for k, v := range s.Elements {
				if _, ok := other.Elements[k]; !ok {
					ns.Elements[k] = v
				}
			}
			return ns
		}}
	case "isSubset":
		return &object.Builtin{Name: "isSubset", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("isSubset expects 1 argument (Set)")
			}
			other, ok := args[0].(*object.Set)
			if !ok {
				return newError("isSubset: argument must be a Set")
			}
			for k := range s.Elements {
				if _, ok := other.Elements[k]; !ok {
					return object.FALSE_OBJ
				}
			}
			return object.TRUE_OBJ
		}}
	case "isSuperset":
		return &object.Builtin{Name: "isSuperset", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("isSuperset expects 1 argument (Set)")
			}
			other, ok := args[0].(*object.Set)
			if !ok {
				return newError("isSuperset: argument must be a Set")
			}
			for k := range other.Elements {
				if _, ok := s.Elements[k]; !ok {
					return object.FALSE_OBJ
				}
			}
			return object.TRUE_OBJ
		}}
	default:
		return newError("no method '%s' on Set", method)
	}
}

// ── DateTime Methods ──

func evalDateTimeMethod(dt *object.DateTime, method string) object.Object {
	switch method {
	case "year":
		return &object.Integer{Value: int64(dt.Value.Year())}
	case "month":
		return &object.Integer{Value: int64(dt.Value.Month())}
	case "day":
		return &object.Integer{Value: int64(dt.Value.Day())}
	case "hour":
		return &object.Integer{Value: int64(dt.Value.Hour())}
	case "minute":
		return &object.Integer{Value: int64(dt.Value.Minute())}
	case "second":
		return &object.Integer{Value: int64(dt.Value.Second())}
	case "dayOfWeek":
		return &object.Integer{Value: int64(dt.Value.Weekday())}
	case "dayOfYear":
		return &object.Integer{Value: int64(dt.Value.YearDay())}
	case "toUnix":
		return &object.Integer{Value: dt.Value.Unix()}
	case "toISO8601":
		return &object.Builtin{Name: "toISO8601", Fn: func(args ...object.Object) object.Object {
			return &object.String{Value: dt.Value.Format("2006-01-02T15:04:05Z07:00")}
		}}
	case "format":
		return &object.Builtin{Name: "format", Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("format expects 1 argument (pattern)")
			}
			pat, ok := args[0].(*object.String)
			if !ok {
				return newError("format: argument must be a string")
			}
			return &object.String{Value: dt.Value.Format(pat.Value)}
		}}
	default:
		return newError("no method '%s' on DateTime", method)
	}
}

// ── Duration Methods ──

func evalDurationMethod(d *object.Duration, method string) object.Object {
	switch method {
	case "toSeconds":
		return &object.Float{Value: d.Value.Seconds()}
	case "toMinutes":
		return &object.Float{Value: d.Value.Minutes()}
	case "toHours":
		return &object.Float{Value: d.Value.Hours()}
	case "toDays":
		return &object.Float{Value: d.Value.Hours() / 24}
	case "toMillis":
		return &object.Integer{Value: d.Value.Milliseconds()}
	default:
		return newError("no method '%s' on Duration", method)
	}
}

// ── CRS Methods ──

func evalCRSMethod(c *object.CRS, method string) object.Object {
	switch method {
	case "epsg":
		return &object.Integer{Value: int64(c.EPSGCode)}
	case "name":
		return &object.String{Value: c.CRSName}
	case "isGeographic":
		return object.NativeBoolToBooleanObject(c.IsGeographic)
	case "isProjected":
		return object.NativeBoolToBooleanObject(c.IsProjected)
	case "units":
		return &object.String{Value: c.Units}
	default:
		return newError("no method '%s' on CRS", method)
	}
}

// ── Enhanced Match Pattern for Enums/Structs ──

// compareForSort compares two objects for sorting. Returns -1, 0, or 1.
func compareForSort(a, b object.Object) int {
	af, aOk := numericToFloat(a)
	bf, bOk := numericToFloat(b)
	if aOk && bOk {
		if af < bf {
			return -1
		}
		if af > bf {
			return 1
		}
		return 0
	}
	// Fall back to string comparison
	as, bs := a.Inspect(), b.Inspect()
	if as < bs {
		return -1
	}
	if as > bs {
		return 1
	}
	return 0
}

func matchEnumPattern(pattern ast.Expression, value object.Object, env *object.Environment) bool {
	if call, ok := pattern.(*ast.CallExpression); ok {
		if ident, ok := call.Function.(*ast.Identifier); ok {
			if ev, ok := value.(*object.EnumVariant); ok {
				if ident.Value == ev.VariantName && len(call.Arguments) == len(ev.Payload) {
					return true
				}
			}
		}
	}
	return false
}

func bindEnumPattern(pattern ast.Expression, value object.Object, env *object.Environment) {
	if call, ok := pattern.(*ast.CallExpression); ok {
		if ev, ok := value.(*object.EnumVariant); ok {
			for i, arg := range call.Arguments {
				if ident, ok := arg.(*ast.Identifier); ok && i < len(ev.Payload) {
					if ident.Value != "_" {
						env.Set(ident.Value, ev.Payload[i], false)
					}
				}
			}
		}
	}
}

// ── Geometry helper functions ──

// lineStringIsSimple checks if a linestring has no self-intersections
// (non-adjacent segments do not cross each other).
func lineStringIsSimple(coords []object.Coordinate) bool {
	n := len(coords)
	if n < 4 {
		return true
	}
	for i := 0; i < n-1; i++ {
		for j := i + 2; j < n-1; j++ {
			// Skip adjacent segments (they share a vertex)
			if i == 0 && j == n-2 {
				continue
			}
			if stdlib.SegmentsIntersect(coords[i], coords[i+1], coords[j], coords[j+1]) {
				return false
			}
		}
	}
	return true
}

// transformPolygon applies a coordinate transformation function to all rings of a polygon.
func transformPolygon(p *object.Polygon, transform func(object.Coordinate) object.Coordinate) *object.Polygon {
	ext := make([]object.Coordinate, len(p.ExteriorRing))
	for i, c := range p.ExteriorRing {
		ext[i] = transform(c)
	}
	intRings := make([][]object.Coordinate, len(p.InteriorRings))
	for i, ring := range p.InteriorRings {
		intRings[i] = make([]object.Coordinate, len(ring))
		for j, c := range ring {
			intRings[i][j] = transform(c)
		}
	}
	return &object.Polygon{ExteriorRing: ext, InteriorRings: intRings, SRID: p.SRID}
}
