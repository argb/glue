package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instructions []byte

type Opcode byte

const (
	OpConstant Opcode = iota
	OpAdd
	OpPop
	OpSub
	OpMul
	OpDiv
	OpTrue
	OpFalse
	OpEqual
	OpNotEqual
	OpGreaterThan
	OpMinus
	OpBang
	OpJumpNotTruthy
	OpJump
	OpNull
	OpGetGlobal
	OpSetGlobal
	OpArray
	OpHash
	OpIndex
	OpCall
	OpReturnValue
	OpReturn
	OpGetLocal
	OpSetLocal
)

type Definition struct {
	Name string
	OperandWidths [] int
}

var definitions = map[Opcode]*Definition{
	OpConstant: {"OpConstant", []int{2}}, // the operand is a index of literal value sitting on constant pool
	OpAdd: {"OpAdd", []int{}},
	OpPop: {"OpPop", []int{}},
	OpSub: {"OpSub", []int{}},
	OpMul: {"OpMul", []int{}},
	OpDiv: {"OpDiv", []int{}},
	OpTrue: {"OpTrue", []int{}},
	OpFalse: {"OpFalse", []int{}},
	OpEqual: {"OpEqual", []int{}},
	OpNotEqual: {"OpNotEqual", []int{}},
	OpGreaterThan: {"OpGreaterThan", []int{}},
	OpMinus: {"OpMinus", []int{}},
	OpBang: {"OpBang", []int{}},
	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}}, //the operand is the index of the target instruction
	OpJump: {"OpJump", []int{2}}, // same with OpJumpNotTruthy
	OpNull: {"OpNull", []int{}},
	OpGetGlobal: {"OpGetGlobal", []int{2}}, // the operand is the index of a value siting on a scope
	OpSetGlobal: {"OpSetGlobal", []int{2}},
	OpArray: {"OpArray", [] int{2}}, // the operand is the length of the array, which is used to build the array literal
	OpHash: {"OpHash", []int{2}}, // same with OpHash
	OpIndex: {"OpIndex", []int{}},
	OpCall: {"OpCall", []int{}},
	OpReturnValue: {"OpReturnValue", []int{}},
	OpReturn: {"OpReturn", []int{}},
	OpGetLocal: {"OpGetLocal", []int{1}}, //操作数宽度1字节，也就是说本地变量不能超过256个，一个函数里面应该不会写那么多变量吧
	OpSetLocal: {"OpSetLocal", []int{1}},
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}

	return def, nil
}

func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	InstructionLen := 1
	for _, w := range def.OperandWidths {
		InstructionLen += w
	}

	instruction := make([]byte, InstructionLen)
	instruction[0] = byte(op)
	offset :=1
	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		case 1:
			instruction[offset] = byte(o)
		}
		offset += width
	}

	return instruction
}

// ReadOperands /**
/*
这个参数ins可能需要调整，ins不应该是instructions，应该是oprands，现在是从指令序列里读取操作数，参数传进来要给出正确
的指令的偏移位置，否则就出错，所以语义上有点乱，读取操作数就应该单纯的读取，不要考虑偏移，控制偏移就是专门的控制偏移，有单独的函数起个合适的
函数名字来完成，图省事，暂时先这样
Todo: 后续要调整
 */
func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0
	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		case 1:
			operands[i] = int(ReadUint8(ins[offset:]))
		}
		offset += width
	}
	return operands, offset
}

func ReadUint8(ins Instructions) uint8 {
	return uint8(ins[0])
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

func (ins Instructions) String() string {
	var out bytes.Buffer
	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}
		operands, read := ReadOperands(def, ins[i+1:])
		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))
		i += 1 + read
	}
	return out.String()
}
func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)
	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n",
			len(operands), operandCount)
	}
	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	}
	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}