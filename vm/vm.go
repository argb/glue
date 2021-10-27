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
const MaxFrames = 1024 // 预分配调用栈大小，要什么自行车，差不多够用了

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

type Frame struct {
	cl *object.Closure
	ip int
	basePointer int
}

// NewFrame /**
/*
每个函数会对应一个栈帧，栈帧会存储函数的指令序列，调用前的执行栈sp值信息，函数调用结束后会用来恢复调用前的状态，继续执行主指令序列
 */
func NewFrame(cl *object.Closure, basePointer int) *Frame {
	f := &Frame{
		cl: cl,
		ip: -1, // VM run loop中 ip总是先自动+1，所以初始化成-1
		basePointer: basePointer,
	}
	return f
}

func (f *Frame) Instructions() code.Instructions {
	return f.cl.Fn.Instructions
}

type VM struct {
	constants [] object.Object
	//instructions code.Instructions

	stack []object.Object
	sp int

	globals []object.Object

	frames [] *Frame // the stack for frame
	frameIndex int // always pointing to the next available frame, equals len(frames)
}

// RuntimeInfo /**
/*
For debug
 */
type RuntimeInfo struct {
	fp int
	sp int
	currIns []byte
}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &object.CompiledFunction{
		Instructions: bytecode.Instructions,
	}
	mainClosure := &object.Closure{Fn: mainFn}
	mainFrame := NewFrame(mainClosure, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame //初始化一下VM的call stack，代码默认相当于写在一个主函数里

	return &VM{
		//instructions: bytecode.Instructions,
		constants: bytecode.Constants,

		stack: make([]object.Object, StackSize),
		sp :0,
		globals: make([]object.Object, GlobalSize),

		frames: frames,
		frameIndex: 1, //总是指向下一个可用栈帧(stack frame)
	}
}

func NewWithGlobalsStore(bytecode *compiler.Bytecode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s

	return vm
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.frameIndex-1]
}
func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.frameIndex] = f
	vm.frameIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.frameIndex--
	return vm.frames[vm.frameIndex]
}

func (vm *VM) ShowReadableInstructions() {

	color.Set(color.FgCyan)
	fmt.Println("Instructions:")
	color.Unset()
	color.Set(color.FgGreen)
	fmt.Println(vm.currentFrame().Instructions().String())
	color.Unset()
}

func (vm *VM) ShowReadableConstants() {
	constants := vm.constants
	color.Set(color.FgCyan)
	fmt.Println("Constant Pool:")
	color.Unset()
	color.Set(color.FgGreen)
	for i, c := range constants {
		fmt.Println(i,": ",c.Inspect())
	}
	color.Unset()
}

func (vm *VM) ShowStack() {
	color.Set(color.FgCyan)
	fmt.Println("Stack:")
	color.Unset()
	color.Set(color.FgGreen)
	if vm.sp <= 0 {
		fmt.Printf("The stack is empty, sp: %d\n", vm.sp)
	}

	for i:=vm.sp-1;i>=0;i-- {
		fmt.Printf("sp: %d, value: %#v\n", i, vm.stack[i])
	}
	color.Unset()
}

func (vm *VM) ShowCallStack() {
	color.Set(color.FgCyan)
	fmt.Println("Call Stack:")
	color.Unset()
	color.Set(color.FgGreen)
	numFrame := vm.frameIndex
	fmt.Printf("there are %d stack frame on call stack\n", numFrame)
	color.Unset()
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}

	return vm.stack[vm.sp - 1]
}

func (vm *VM) Run() error {
	var ip int
	var instructions code.Instructions
	var op code.Opcode

	//mn := monitor.SingletonNew()

	//VM进入运行状态，ip是指令指针，每次向前移动一个字节，
	//每条指令的宽度可能不同，所以每条指令执行完后必须jump over合适的operands，如果碰到无法识别的指令，就会造成指令的执行出错，因为
	//可能会导致IP指向操作数，会导致后续所有指令的执行不可预测，出一些奇怪的错误
	//for ip := 0; ip < len(vm.instructions); ip++ {
	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++ // 这里就是为什么上面for条件中指令长度-1的原因

		ip = vm.currentFrame().ip
		instructions = vm.currentFrame().Instructions()
		//取出当前ip指向的指令，指令长度一字节，直接强制类型转换为Opcode类型，不会有信息丢失
		op = code.Opcode(instructions[ip])

		// for debug
		//fmt.Print(code.FmtInstruction(instructions, ip))

		//mn.AddInstruction(code.FmtInstruction(instructions, ip)) // for debug

		switch op {
		case code.OpConstant:

			constIndex := code.ReadUint16(instructions[ip + 1:])
			// 因为OpConstant指令带有1个Operands,且操作数宽度为2自己所以取出指令后让指令指针跳过操作数
			// 下面对所有带操作数的指令做同样的处理，但是不用指令操作数数量和宽度可能不同，具体情况可通过指令的Definition查找
			vm.currentFrame().ip +=2

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
			pos := int(code.ReadUint16(instructions[ip+1:]))
			//减一是因为指令是在for循环中取出的，每次取出指令后循环会自动让ip加一，所以这里要减一，取下一条指令
			//会自动令ip增加1，否则就多走了一步
			vm.currentFrame().ip = pos - 1
		case code.OpJumpNotTruthy:
			//跳转偏移量是在编译时确定的，通过back-patching的方式把偏移量回填到占位指令上
			pos := int(code.ReadUint16(instructions[ip+1:]))
			vm.currentFrame().ip += 2 // 跟OpConstant指令的处理相同，跳过操作数占用的字节

			condition := vm.pop()
			// 因为此指令的作用是：条件不成立的时跳转，条件成立就跳过本条指令，什么也不做，继续取后面的指令
			if !isTruthy(condition) {
				vm.currentFrame().ip = pos -1
			}
		case code.OpNull:
			err := vm.push(Null)
			if err != nil {
				return err
			}
		case code.OpSetGlobal: //全局变量存在一个全局变量池中，实际上就是个go的切片，在整个程序运行的生命周期中都可以被访问
			globalIndex := code.ReadUint16(instructions[ip+1:])
			vm.currentFrame().ip += 2

			vm.globals[globalIndex] = vm.pop()
		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(instructions[ip+1:])
			vm.currentFrame().ip += 2

			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}
		case code.OpSetLocal: //本地变量直接存在stack上，注意跟全局变量的处理不同，没有专门用一个数据结构存储，而是直接放在了stack上，随着函数调用结束会被销毁。
			localIndex := code.ReadUint8(instructions[ip+1:])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()

			vm.stack[frame.basePointer + int(localIndex)] = vm.pop()
		case code.OpGetLocal:
			localIndex := code.ReadUint8(instructions[ip+1:])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()

			err := vm.push(vm.stack[frame.basePointer+int(localIndex)])
			if err != nil {
				return err
			}
		case code.OpGetBuiltin:
			builtinIndex := code.ReadUint8(instructions[ip+1:])
			vm.currentFrame().ip += 1

			definition := object.Builtins[builtinIndex]
			err := vm.push(definition.Builtin)
			if err != nil {
				return err
			}
		case code.OpGetFree:
			freeIndex := code.ReadUint8(instructions[ip+1:])
			vm.currentFrame().ip += 1

			currentClosure := vm.currentFrame().cl
			err := vm.push(currentClosure.Free[freeIndex])
			if err != nil {
				return err
			}
		case code.OpArray:
			// OpArray 用来计算（构造）数组字面量，没错，数组的构造过程是在运行时进行的，因为数组每个元素可能是某个表达式的
			// 计算结果，所以只能无法在编译期确定数组的具体元素内容
			// 但是编译器可以做一些优化，比如对常量表达式直接算出结果，优化暂且不做，后续在再做
			numElements := int(code.ReadUint16(instructions[ip+1:]))
			vm.currentFrame().ip += 2
			//这里有个顺序，数组原始是根据指令的顺序执行压栈的，所以前面的元素会先压栈，因为指令先生成
			array := vm.buildArray(vm.sp-numElements, vm.sp)
			vm.sp = vm.sp - numElements // 数组构造完成后让栈指针下移数组长度的距离，代表数组出栈了

			err := vm.push(array) // 把构造好的数组对象压栈
			if err != nil {
				return err
			}
		case code.OpHash:
			numElements := int(code.ReadUint16(instructions[ip+1:]))
			vm.currentFrame().ip += 2
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
		case code.OpClosure:
			constIndex := code.ReadUint16(instructions[ip+1:])
			numFree := code.ReadUint8(instructions[ip+3:])
			vm.currentFrame().ip += 3

			err := vm.pushClosure(int(constIndex), int(numFree))
			if err != nil {
				return err
			}
		case code.OpCall:
			numArgs := int(code.ReadUint8(instructions[ip+1:]))

			vm.currentFrame().ip += 1
			err := vm.executeCall(numArgs)
			if err != nil {
				return err
			}
			/*
			// 参数arguments已经被加载到stack上了
			fn, ok := vm.stack[vm.sp - 1 - numArgs].(*object.CompiledFunction)
			if !ok {
				return fmt.Errorf("calling non-function")
			}
			if numArgs != fn.NumParameters {
				return fmt.Errorf("wrong number of arguments: want=%d, got=%d", fn.NumParameters, numArgs)
			}
			frame := NewFrame(fn, vm.sp - numArgs)
			vm.pushFrame(frame) // 下一轮循环，指令就切换到新栈帧下的指令序列
			vm.sp = frame.basePointer + fn.NumLocals //sp增长NumLocals，形成一个stack hole空洞，用来加载函数的局部变量


			 */
		case code.OpReturnValue:
			returnValue := vm.pop()

			// 这相当于OpPop指令，但是这条指令是VM的内建指令，也就是说他不是OpReturnValue的响应,
			//每当VM有栈帧出栈的时候就会执行一次vm.pop()操作,目的是清理掉stack上的*object.CompiledFunction这个对象，因为我们已经完成对其内部指令的执行
			//vm.pop()

			frame := vm.popFrame()
			// 小优化， -1 代替之前的vm.pop, 一步到位，相当于直接把函数对象也从栈上移除了，函数对象哪来的？编译的时候绑定到函数名上的，所以
			// 函数调用的上一条指令一定是OpGetGlobal或者OpGetLocal, 这是有编译器决定的
			//vm.sp = frame.basePointer - 1 // 恢复函数调用前的sp位置
			vm.sp = frame.basePointer
			vm.pop() // 把函数对象出栈

			err := vm.push(returnValue)
			if err != nil {
				return err
			}
		case code.OpReturn:
			frame := vm.popFrame()
			//vm.pop()
			//vm.sp = frame.basePointer - 1
			vm.sp = frame.basePointer
			vm.pop()

			err := vm.push(Null)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// pushClosure /**
/*
构造Closure对象，然后压栈
 */
func (vm *VM) pushClosure(constIndex int, numFree int) error {
	constant := vm.constants[constIndex]
	function, ok := constant.(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("not a function: %+v", constant)
	}

	free := make([]object.Object, numFree)
	for i:=0; i < numFree; i++ { //重要！顺序！要跟放到栈上的顺序保持一致，也就是跟自由变量在函数体中被引用的顺序保持一致
		free[i] = vm.stack[vm.sp - numFree + i]
	}
	vm.sp = vm.sp - numFree

	closure := &object.Closure{Fn: function, Free: free}

	return vm.push(closure)
}

func (vm *VM) executeCall(numArgs int) error {
	callee := vm.stack[vm.sp-1-numArgs]
	fmt.Println(callee.Inspect())
	/*
	for _, item := range vm.stack[0:3] {
		println(item.Inspect())
	}
	 */
	//fmt.Println(vm.stack[0:10])
	switch callee := callee.(type) {
	case *object.Closure:
		return vm.callClosure(callee, numArgs)
	case *object.Builtin:
		return vm.callBuiltin(callee, numArgs)
	default:
		fmt.Printf("debug, callee: %#v\n", callee)
		return fmt.Errorf("calling non-function and non-builtin")
	}
}
/**
函数调用（所有函数都封装在一个Closure对象里），把closure对象挂到当前的栈帧上，指令可以通过stack frame-->closure-->free variables 这个引用链
取到本次调用所需的free variables
 */
func (vm *VM) callClosure(cl *object.Closure, numArgs int) error {
	if numArgs != cl.Fn.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d", cl.Fn.NumParameters, numArgs)
	}
	frame := NewFrame(cl, vm.sp - numArgs)
	vm.pushFrame(frame)

	vm.sp = frame.basePointer + cl.Fn.NumLocals // NumLocals =（包括局部变量+形参）两者的数量，所以这里不必在加上numArgs

	return nil
}

func (vm *VM) callBuiltin(builtin *object.Builtin, numArgs int) error {
	args := vm.stack[vm.sp-numArgs : vm.sp]

	result := builtin.Fn(args...)
	vm.sp = vm.sp - numArgs -1
	var err error
	if result != nil {
		err = vm.push(result)
	}else {
		err = vm.push(Null)
	}

	if err != nil {
		return err
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
	//vm.stack[vm.sp-1] = nil // 不是必须的，但是把出栈后清空数据方便调试
	vm.sp--

	return o
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}