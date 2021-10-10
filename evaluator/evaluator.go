package evaluator

import (
	"compiler01/ast"
	"compiler01/object"
	"fmt"
)

var (
	NULL = &object.Null{}
	TRUE = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		//return &object.Boolean{Value: node.Value}
		//对上面代码的优化，不必每次使用布尔值都创建新对象来表示，所有布尔值引用同一个true/false Object
		return nativeBoolToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		//fmt.Printf("right: %#v",right)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.ReturnStatement:
		return evalReturnStatement(node, env)
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body

		return &object.Function{Parameters: params, Env: env, Body: body}
	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env)

		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		//fmt.Println("CallExpression, node.Function:", node.Function.String())
		return applyFunction(function, args)
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	}
	return nil
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated :=Eval(fn.Body, extendedEnv)

		return unwrapReturnValue(evaluated)
	case *object.Builtin:

		return fn.Fn(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}

	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}

	return obj
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

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return newError("identifier not found: " + node.Value)
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}
	return result
}

/*
这个处理貌似也有点不合适，还是重复解析的问题，如果没嵌套语句没问题，
如果有嵌套语句就会出现重复解析。根本问题还是在于，所有的节点没有长在一棵树上，如果不在ast，就不应该被evaluate取值。
算是个补丁，后面要改一下
 */
func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ{
				return result
			}
		}
	}

	return result
}

func evalReturnStatement(rs *ast.ReturnStatement, env *object.Environment) object.Object {
	evaluated := Eval(rs.ReturnValue, env)

	if isError(evaluated) {
		return evaluated
	}
	return &object.ReturnValue{Value: evaluated}
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	}else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	}else {
		return NULL
	}
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

/*
这个写法好像有点问题，因为这里存储了整个代码文件的全部语句，是并列关系
比如:
A ::= statement|statement_list
statement_list ::=b|c|d
假如a = {b,c,d}
那么这个statements数组就是[a,b,c,d],因为b,c,d是从a里面解析出来的，
语法上来说也是句子，就塞到这个顶层数组了，但是b,c,d已经存在于基于a所构造的ast上了
也就是这个statements数组里的b,c,d应该只能看做是一种快捷方式，而不能作为ast的一部分
进行解析，否则就会重复解析，可能会导致解析逻辑混乱和错误
暂时先这么着吧，打补丁弥补逻辑问题，现在能正常跑起来，但是逻辑上不合理，后续优化吧
[todo]
 */
func evalStatements(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range stmts {
		/*
		if i>=1 {return result}
		fmt.Println(i,statement.String())
		 */
		result = Eval(statement, env)
		if returnValue, ok := result.(*object.ReturnValue); ok {
			return returnValue.Value
		}
	}
	return result
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		//fmt.Println(1)
		return FALSE
	case FALSE:
		//fmt.Println(2)
		return TRUE
	case NULL:
		//fmt.Println(3)
		return TRUE
	default:
		//fmt.Println(4)
		return FALSE
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}
	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	//fmt.Printf("left:%#v, right:%#v\n", left,right)
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case operator == "==":
		// 这里的相等和不等是直接比较两个Object 类型left和right，不必提取native value(原生值）：Object.Value来比较，
		//因为我们所有的布尔类型都只共用两个TRUE和FALSE的实例，都是这两个实例的引用，所以直接比较对象就可以。
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpress(operator, left, right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator,right.Type())
	}
}

func evalStringInfixExpress(operater string, left, right object.Object) object.Object {
	if operater != "+" {
		return newError("unknown operator: %s %s %s", left.Type(), operater, right.Type())
	}

	leftVal :=left.(*object.String).Value
	rightVal := right.(*object.String).Value

	return &object.String{Value: leftVal + rightVal}
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
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)

	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}

	return false
}