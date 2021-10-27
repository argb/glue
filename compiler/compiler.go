package compiler

import (
	"compiler01/ast"
	"compiler01/code"
	"compiler01/object"
	"fmt"
	"sort"
)

type Bytecode struct {
	Instructions code.Instructions
	Constants []object.Object
}

type EmittedInstruction struct {
	Opcode code.Opcode
	Position int
}

type CompilationScope struct {
	instructions code.Instructions
	lastInstruction EmittedInstruction
	previousInstruction EmittedInstruction
}

type Compiler struct {
	//instructions code.Instructions
	constants []object.Object // 行为上，相当于基于index来寻址的连续内存，用
	//lastInstruction EmittedInstruction
	//previousInstruction EmittedInstruction

	// 其实是个栈和列表的混合结构，在作用域层级上，跟随者作用域的变化，总是以LIFO方式访问，但当查询其上具体的Symbol时，
	//就会沿着整个链表从内向外递归查询一遍
	symbolTable *SymbolTable

	scopes []CompilationScope // 行为上，作用域是个栈
	scopeIndex int // 栈指针
}

func New() *Compiler {
	mainScope := CompilationScope{
		instructions: code.Instructions{},
		lastInstruction: EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	symbolTable := NewSymbolTable()
	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}
	return &Compiler{
		//instructions: code.Instructions{},
		constants: []object.Object{},
		//lastInstruction: EmittedInstruction{},
		//previousInstruction: EmittedInstruction{},
		symbolTable: symbolTable,
		scopes: []CompilationScope{mainScope},
		scopeIndex: 0,
	}
}

func NewWithState(s *SymbolTable, constants []object.Object) *Compiler {
	compiler := New()
	compiler.symbolTable = s
	compiler.constants = constants

	return compiler
}

func (c *Compiler) currentInstructions() code.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) loadSymbol(s Symbol) {
	switch s.Scope {
	case GlobalScope:
		c.emit(code.OpGetGlobal, s.Index)
	case LocalScope:
		c.emit(code.OpGetLocal, s.Index)
	case BuiltinScope:
		c.emit(code.OpGetBuiltin, s.Index)
	case FreeScope:
		c.emit(code.OpGetFree, s.Index)
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.LetStatement:
		// 生成符号，加入到当前作用域对应的符号表，当前作用域是在编译函数字面量的时候确定的，在此处处理 case *ast.FunctionLiteral:
		symbol := c.symbolTable.Define(node.Name.Value)
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		//编译器只需要确定当前处理的符号是个本地变量还是全局变量，而不需要关心嵌套了几层
		// 只要是局部变量就生成局部指令OpSetLocal, 而到底要从哪一层作用域取出绑定的数据有VM在运行时完成
		// 其实VM也不需要去特殊判断，只要按正常的指令运算流程处理就可以了，因为整个栈机制和生成的每个指令执行方式就已经可以确保
		// 从正确的作用域中取出数据
		if symbol.Scope == GlobalScope {
			c.emit(code.OpSetGlobal, symbol.Index)
		}else {
			c.emit(code.OpSetLocal, symbol.Index)
		}
	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)
	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}
		switch node.Operator {
		case "!":
			c.emit(code.OpBang)
		case "-":
			c.emit(code.OpMinus)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}
	case *ast.InfixExpression:
		if node.Operator == "<" { // reorder the operands for "<"
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}
			err = c.Compile(node.Left)
			if err != nil {
				return err
			}
			c.emit(code.OpGreaterThan)
			return nil
		}

		err := c.Compile(node.Left)
		if err != nil {
			return err
		}
		err = c.Compile(node.Right)
		if err != nil {
			return err
		}
		switch node.Operator {
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "*":
			c.emit(code.OpMul)
		case "/":
			c.emit(code.OpDiv)
		case ">":
			c.emit(code.OpGreaterThan)
		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		default:
			return fmt.Errorf("unknow operator %s", node.Operator)
		}
	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		//9999这里只是个随意占位符，小于65535就可以，因为操作数目前设置宽度为2字节，超出2字节会导致指令对齐出错
		// if not true, jump over the consequence block
		jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)
		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}
		//如果以后不把if作为表达式，而是作为语句，这里的tricky处理需要去掉
		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}

		// jump over the alternative block
		jumpPos := c.emit(code.OpJump, 9999)

		afterConsequencePos := len(c.currentInstructions())
		c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

		if node.Alternative == nil {
			c.emit(code.OpNull)
		}else {

			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			if c.lastInstructionIs(code.OpPop) {
				c.removeLastPop()
			}
		}

		afterAlternativePos := len(c.currentInstructions())
		c.changeOperand(jumpPos, afterAlternativePos)
	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))
	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		}else {
			c.emit(code.OpFalse)
		}
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}
		c.loadSymbol(symbol)
	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(str))
	case *ast.ArrayLiteral:
		for _, el := range node.Elements {
			err := c.Compile(el)
			if err != nil {
				return err
			}
		}
		c.emit(code.OpArray, len(node.Elements))
	case *ast.HashLiteral:
		var keys []ast.Expression
		for k := range node.Pairs{
			keys = append(keys, k)
		}
		// go的range循环遍历map时不会保证顺序，顺是随机的，这里是为了便于写
		// test cases, 即使不排序也不影响指令生成和vm的执行
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() <keys[j].String()
		})
		for _, k := range keys {
			err := c.Compile(k)
			if err != nil {
				return err
			}
			err = c.Compile(node.Pairs[k])
			if err != nil {
				return err
			}
		}
		c.emit(code.OpHash, len(node.Pairs)*2)
	case *ast.IndexExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}
		err = c.Compile(node.Index)
		if err != nil {
			return err
		}

		c.emit(code.OpIndex)
	case *ast.FunctionLiteral:
		c.enterScope()

		for _, p := range node.Parameters {
			c.symbolTable.Define(p.Value)
		}

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}
		if c.lastInstructionIs(code.OpPop) {
			c.replaceLastPopWithReturn()
		}
		if !c.lastInstructionIs(code.OpReturnValue) {
			c.emit(code.OpReturn)
		}
		// 暂存自由变量，因为下面c.leaveScope()后，局部作用域就释放了，也就是对应的符号表就销毁了，因为指令已经生成完毕了
		freeSymbols := c.symbolTable.FreeSymbols
		numLocals := c.symbolTable.numDefinitions // 形式参数也看做局部变量
		//对函数字面量的解析完成了，退出当前函数的作用域
		instructions := c.leaveScope()

		// 生成对应的指令，用来把上面暂存的自由变量加载的栈上，VM会执行这些指令
		for _, s := range freeSymbols {
			c.loadSymbol(s)
		}

		compiledFn := &object.CompiledFunction{
			Name: node.String(),
			Instructions: instructions,
			NumLocals: numLocals,
			NumParameters: len(node.Parameters),
		}

		// 字面量，包括函数定义，统统看做常量，常量池里存储的依旧是object.CompiledFunction对象，VM执行到OpClosure指令才把它转化成object.Closure对象
		fnIndex := c.addConstant(compiledFn)
		//c.emit(code.OpConstant, c.addConstant(compiledFn))
		c.emit(code.OpClosure, fnIndex, len(freeSymbols))
	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return err
		}
		c.emit(code.OpReturnValue)
	case *ast.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}

		for _, a := range node.Arguments {
			err := c.Compile(a)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpCall, len(node.Arguments)) // 操作数是相对于本指令在stack上的偏移量
	}
	return nil
}

func (c *Compiler) replaceLastPopWithReturn() {
	lastPos := c.scopes[c.scopeIndex].lastInstruction.Position
	c.replaceInstruction(lastPos, code.Make(code.OpReturnValue))

	c.scopes[c.scopeIndex].lastInstruction.Opcode = code.OpReturnValue
}

/*
编译时的不变量都加到这个常量池中，就是一块连续内存，借助go的数组来表示
VM启动后会直接给VM使用,vm.constants
因为vm的启动跟编译是连续的过程，所以vm可以直接使用编译过程开辟的这块内存
如果需要把字节码导出，然后可以独立的作为vm的输入去执行的话，就需要额外处理这个常量池内存的分配问题
 */
func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)

	return len(c.constants)-1
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)

	return pos
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}
	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

func (c *Compiler) removeLastPop() {
	last := c.scopes[c.scopeIndex].lastInstruction
	previous := c.scopes[c.scopeIndex].previousInstruction

	old := c.currentInstructions()
	new := old[:last.Position]

	c.scopes[c.scopeIndex].instructions = new
	c.scopes[c.scopeIndex].lastInstruction = previous
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int)  {
	previous := c.scopes[c.scopeIndex].lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}

	c.scopes[c.scopeIndex].previousInstruction = previous
	c.scopes[c.scopeIndex].lastInstruction = last
}

func (c *Compiler) addInstruction(ins []byte) int {
	currInstructions := c.currentInstructions()
	posNewInstruction := len(currInstructions)
	/*
	currInstructions = append(currInstructions, ins...)
	c.scopes[c.scopeIndex].instructions = currInstructions
	*/
	c.scopes[c.scopeIndex].instructions = append(c.scopes[c.scopeIndex].instructions, ins...)

	return posNewInstruction
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(),
		Constants: c.constants,
	}
}

/**
逐字节替换旧指令
 */
func (c *Compiler) replaceInstruction(pos int, newInstruction[]byte) {
	instructions := c.currentInstructions()
	for i:=0; i<len(newInstruction); i++ {
		instructions[pos+i] = newInstruction[i]
	}
}

/**
找到第一个参数位置的指令，单字节，取出，根据第二个参数（操作数，实际是个偏移量）生成新的指令
然后找到旧指令的位置逐字节替换旧指令的内容
 */
func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.currentInstructions()[opPos])
	newInstruction := code.Make(op, operand)

	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions: code.Instructions{},
		lastInstruction: EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	c.scopes = append(c.scopes, scope)

	c.scopeIndex++

	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

/**
同时修改作用域和符号表
 */
func (c *Compiler) leaveScope() code.Instructions {
	instructions := c.currentInstructions()
	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--

	c.symbolTable = c.symbolTable.Outer

	return instructions
}
