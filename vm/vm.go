package vm

import (
	"compiler01/code"
	"compiler01/compiler"
	"compiler01/object"
	"fmt"
	"github.com/fatih/color"
)

const StackSize = 2048
const GlobalSize = 65536

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Null = &object.Null{}

type StackItem []byte

// StackMonitor /**
/*
用来监控VM stack的数据结构，用于记录stack在每条指令执行之后的运行状态
 */
type StackMonitor struct {
	Len int64
	Sp int64
	CurItems []StackItem
}

type VM struct {
	constants [] object.Object
	instructions code.Instructions

	stack []object.Object
	sp int

	globals []object.Object
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants: bytecode.Constants,

		stack: make([]object.Object, StackSize),
		sp :0,
		globals: make([]object.Object, GlobalSize),
	}
}

func NewWithGlobalsStore(bytecode *compiler.Bytecode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s

	return vm
}

func (vm *VM) ShowReadableInstructions() {

	color.Set(color.FgCyan)
	fmt.Println("Instructions:")
	color.Unset()
	color.Set(color.FgGreen)
	fmt.Println(vm.instructions.String())
	color.Unset()
}

func (vm *VM) ShowReadableConstants() {
	constants := vm.constants
	color.Set(color.FgCyan)
	fmt.Println("Data:")
	color.Unset()
	color.Set(color.FgGreen)
	for i, c := range constants {
		fmt.Println(i,": ",c.Inspect())
	}
	color.Unset()
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}

	return vm.stack[vm.sp - 1]
}

func (vm *VM) Run() error {
	//VM进入运行状态，ip是指令指针，每次向前移动一个字节，
	//每条指令的宽度可能不同，所以每条指令执行完后必须jump over合适的operands，如果碰到无法识别的指令，就会造成指令的执行出错，因为
	//可能会导致IP指向操作数，会导致后续所有指令的执行不可预测，出一些奇怪的错误
	for ip := 0; ip < len(vm.instructions); ip++ {
		//取出当前ip指向的指令，指令长度一字节，直接强制类型转换为Opcode类型，不会有信息丢失
		op := code.Opcode(vm.instructions[ip])

		switch op {
		case code.OpConstant:

			constIndex := code.ReadUint16(vm.instructions[ip + 1:])
			// 因为OpConstant指令带有1个Operands,且操作数宽度为2自己所以取出指令后让指令指针跳过操作数
			// 下面对所有带操作数的指令做同样的处理，但是不用指令操作数数量和宽度可能不同，具体情况可通过指令的Definition查找
			ip +=2

			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		case code.OpPop:
			vm.pop()
		case code.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}
		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}
		case code.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}

		case code.OpMinus:
			vo := vm.pop()
			vi := vo.(*object.Integer)
			v := vi.Value
			vm.push(&object.Integer{Value: -v})
		case code.OpJump:
			// 取出OpJump指令的操作数，也就是跳转的目的地址（是一个相对于指令序列0位置的绝对偏移量）
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))
			//减一是因为指令是在for循环中取出的，每次取出指令后循环会自动让ip加一，所以这里要减一，取下一条指令
			//会自动令ip增加1，否则就多走了一步
			ip = pos - 1
		case code.OpJumpNotTruthy:
			//跳转偏移量是在编译时确定的，通过back-patching的方式把偏移量回填到占位指令上
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2 // 跟OpConstant指令的处理相同，跳过操作数占用的字节

			condition := vm.pop()
			// 因为此指令的作用是：条件不成立的时跳转，条件成立就跳过本条指令，什么也不做，继续取后面的指令
			if !isTruthy(condition) {
				ip = pos -1
			}
		case code.OpNull:
			err := vm.push(Null)
			if err != nil {
				return err
			}
		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			vm.globals[globalIndex] = vm.pop()
		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}
		case code.OpArray:
			// OpArray 用来计算（构造）数组字面量，没错，数组的构造过程是在运行时进行的，因为数组每个元素可能是某个表达式的
			// 计算结果，所以只能无法在编译期确定数组的具体元素内容
			// 但是编译器可以做一些优化，比如对常量表达式直接算出结果，优化暂且不做，后续在再做
			numElements := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2
			//这里有个顺序，数组原始是根据指令的顺序执行压栈的，所以前面的元素会先压栈，因为指令先生成
			array := vm.buildArray(vm.sp-numElements, vm.sp)
			vm.sp = vm.sp - numElements // 数组构造完成后让栈指针下移数组长度的距离，代表数组出栈了

			err := vm.push(array) // 把构造好的数组对象压栈
			if err != nil {
				return err
			}
		case code.OpHash:
			numElements := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2
			hash, err := vm.buildHash(vm.sp-numElements, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - numElements
			err = vm.push(hash)
			if err != nil {
				return err
			}
		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return err
			}

		}
	}

	return nil
}

func (vm *VM) executeIndexExpression(left, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return vm.executeArrayIndex(left, index)
	case left.Type() == object.HASH_OBJ:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeArrayIndex(array, index object.Object) error {
	arrayObject := array.(*object.Array)
	i := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) -1)
	if i < 0 || i > max {
		return vm.push(Null)
	}

	return vm.push(arrayObject.Elements[i])
}

func (vm *VM) executeHashIndex(hash, index object.Object) error {
	hashObject := hash.(*object.Hash)
	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}
	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}

	return vm.push(pair.Value)
}

func (vm *VM) buildHash(startIndex, endIndex int) (object.Object, error) {
	hashedPairs := make(map[object.HashKey]object.HashPair)
	for i:= startIndex; i< endIndex; i+=2 {
		key :=vm.stack[i]
		value := vm.stack[i+1]
		pair := object.HashPair{Key: key, Value: value}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}
		hashedPairs[hashKey.HashKey()] = pair
	}

	return &object.Hash{Pairs: hashedPairs}, nil
}

func (vm *VM) buildArray(startIndex, endIndex int) object.Object {
	elements := make([]object.Object, endIndex - startIndex)

	for i:= startIndex; i < endIndex; i++ {
		elements[i - startIndex] = vm.stack[i]
	}

	return &object.Array{Elements: elements}
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	default:
		return true
	}
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	if operand.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for negation: %s(type:%s)", operand.Inspect(), operand.Type())
	}
	value := operand.(*object.Integer).Value

	return vm.push(&object.Integer{
		Value: -value,
	})
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()

	switch operand {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	case Null:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()
	if left.Type() == object.INTEGER_OBJ || right.Type() == object.INTEGER_OBJ {
		return vm.executeIntegerComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(right == left))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(left != right))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())
	}
}

func (vm *VM) executeIntegerComparison(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue == rightValue))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue != rightValue))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftValue > rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}


func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return True
	}
	return False
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	switch {
	case leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ:
		return vm.executeBinaryIntegerOperation(op, left, right)
	case leftType == object.STRING_OBJ && rightType == object.STRING_OBJ:
		return vm.executeBinaryStringOperation(op, left, right)
	default:
		return fmt.Errorf("unsupported types for binary operation: %s, %s", leftType, rightType)
	}

}

func (vm *VM) executeBinaryStringOperation(op code.Opcode, left, right object.Object) error {
	if op != code.OpAdd {
		return fmt.Errorf("unknown string operator: %d", op)
	}
	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value
	//好像go的+在处理字符串相加时效率较低，后续考虑优化一下。 todo:
	return vm.push(&object.String{Value: leftValue + rightValue})
}

func (vm *VM) executeBinaryIntegerOperation(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value
	var result int64

	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSub:
		result = leftValue - rightValue
	case code.OpMul:
		result = leftValue * rightValue
	case code.OpDiv:
		result = leftValue / rightValue
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	return vm.push(&object.Integer{Value: result})
}

func (vm *VM) push(o object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

func (vm *VM) pop() object.Object {
	// 规定 sp 始终指向下一个可用槽位，所以当前槽位就是sp-1
	o := vm.stack[vm.sp-1]
	vm.sp--

	return o
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}