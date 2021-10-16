package vm

import (
	"compiler01/code"
	"compiler01/compiler"
	"compiler01/object"
	"fmt"
	"github.com/fatih/color"
)

const StackSize = 2048

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
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants: bytecode.Constants,

		stack: make([]object.Object, StackSize),
		sp :0,
	}
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
		}

	}

	return nil
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

	return nil
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

	if leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ {
		return vm.executeBinaryIntegerOperation(op, left, right)
	}

	return fmt.Errorf("unsupported types for binary operation: %s, %s", leftType, rightType)
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