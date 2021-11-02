package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"fmt"
	"glue/ast"
	"glue/lexer"
	"glue/parser"
	"os"
	"strings"
)
var Index int64 = 0

func main() {
	input :=`
let a=10;
let d=a+b+c*6/9;
if(a>b){
	let x=10;
}
let add=fn(a, b){
return a+b;
}
`
	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	var lines []string

	fmt.Println("the whole program:\n", program.String())

	var dotSrc bytes.Buffer
	dotSrc.WriteString("digraph ast {\n")
	dotSrc.WriteString(`label = "program";`)
	dotSrc.WriteString("\n")

	walk(program, &lines)

	body := strings.Join(lines, "\n")
	dotSrc.WriteString(body)

	dotSrc.WriteString("\n}")

	genFile(dotSrc.String())
}

func genFile(src string) {
	filePath := "./ast.dot"
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Printf("文件打开错误：%v \n", err)
		return
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	writer.WriteString(src)

	writer.Flush()
}

func md5v1(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has)
	return md5str
}

func genEdgeToNode(start, end ast.Node) string {
	b := strings.Builder{}
	b.WriteString("\"")
	b.WriteString(start.Tag())

	b.WriteString("\"")

	b.WriteString("->")

	b.WriteString("\"")
	b.WriteString(end.Tag())
	b.WriteString("\"")
	b.WriteString(";")

	return b.String()
}

func genEdgeToLeaf(start ast.Node, leaf interface{}) string {
	b := strings.Builder{}
	b.WriteString("\"")
	b.WriteString(start.Tag())

	b.WriteString("\"")

	b.WriteString("->")

	b.WriteString("\"")
	b.WriteString(genLeaf(leaf))
	b.WriteString("\"")

	b.WriteString(";")

	return b.String()
}

func genNode(node ast.Node) string {

	return "\""+node.Tag()+"\""
}

func genLeaf(obj interface{}) string {
	switch obj.(type) {
	case int64:
		value := obj

		Index++
		return fmt.Sprintf("[%d]l%d", value, Index)
	case string:
		value := obj
		Index++
		return fmt.Sprintf("[%s]l%d", value, Index)
	case bool:
		value := obj
		Index++
		return fmt.Sprintf("[%b]l%b", value, Index)
	default:
		return ""
	}
}


func walk(node ast.Node, lines *[]string) {
	switch node := node.(type) {
	case *ast.Program:
		//*lines = append(*lines, genNode(node))
		for _, statement := range node.Statements{
			*lines = append(*lines, genEdgeToNode(node, statement))
			walk(statement,lines)
		}
	case *ast.BlockStatement:
		for _, statement := range node.Statements {
			*lines = append(*lines, genEdgeToNode(node, statement))
			walk(statement, lines)
		}

	case *ast.ExpressionStatement:
		*lines = append(*lines, genEdgeToNode(node, node.Expression))
		walk(node.Expression, lines)
	case *ast.ReturnStatement:
		*lines = append(*lines, genEdgeToLeaf(node, "return"))

		*lines = append(*lines, genEdgeToNode(node, node.ReturnValue))
		walk(node.ReturnValue, lines)
	case *ast.IfExpression:
		//*lines = append(*lines, genNode(node))

		*lines = append(*lines, genEdgeToLeaf(node, "if"))

		*lines = append(*lines, genEdgeToNode(node, node.Condition))
		walk(node.Condition, lines)

		*lines = append(*lines, genEdgeToNode(node, node.Consequence))
		walk(node.Consequence, lines)

		if node.Alternative != nil {
			*lines = append(*lines, genEdgeToLeaf(node, "else"))
			*lines = append(*lines, genEdgeToNode(node, node.Alternative))
			walk(node.Alternative, lines)
		}


	case *ast.InfixExpression:
		//*lines = append(*lines, genNode(node))


		*lines = append(*lines, genEdgeToNode(node, node.Left))
		walk(node.Left, lines)

		*lines = append(*lines, genEdgeToLeaf(node, node.Operator))

		*lines = append(*lines, genEdgeToNode(node, node.Right))
		walk(node.Right, lines)
	case *ast.PrefixExpression:
		*lines = append(*lines, genEdgeToLeaf(node, node.Operator))
		walk(node.Right, lines)


	case *ast.LetStatement:
		//*lines = append(*lines, genNode(node))
		*lines = append(*lines, genEdgeToLeaf(node, "let"))
		*lines = append(*lines, genEdgeToNode(node, node.Name))
		walk(node.Name, lines)

		*lines = append(*lines, genEdgeToLeaf(node, "="))

		*lines = append(*lines, genEdgeToNode(node, node.Value))
		walk(node.Value, lines)
	case *ast.Identifier:
		//*lines = append(*lines, genNode(node))
		//*lines = append(*lines, genLeaf(node.Value))
		*lines = append(*lines, genEdgeToLeaf(node, node.Value))
		return
	case *ast.IntegerLiteral:
		//*lines = append(*lines, genNode(node))
		//*lines = append(*lines, genLeaf(node.Value))
		*lines = append(*lines, genEdgeToLeaf(node, node.Value))
		return
	case *ast.StringLiteral:
		//*lines = append(*lines, genNode(node))
		//*lines = append(*lines, genLeaf(node.Value))
		*lines = append(*lines, genEdgeToLeaf(node, node.Value))
		return
	case *ast.FunctionLiteral:
		*lines = append(*lines, genEdgeToLeaf(node, "fn"))

		*lines = append(*lines, genEdgeToLeaf(node, "("))

		if node.Parameters != nil && len(node.Parameters)>0{
			for _, parameter := range node.Parameters {
				*lines = append(*lines, genEdgeToNode(node, parameter))
				walk(parameter, lines)
			}
		}
		*lines = append(*lines, genEdgeToLeaf(node, ")"))
		*lines = append(*lines, genEdgeToLeaf(node, "{"))

		*lines = append(*lines, genEdgeToNode(node, node.Body))
		walk(node.Body, lines)
		*lines = append(*lines, genEdgeToLeaf(node, "}"))

	}
}
