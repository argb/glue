package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"fmt"
	"glue/ast"
	"glue/lexer"
	"glue/parser"
	"glue/token"
	"os"
	"strings"
)

func main() {
	input :=`
let a=10;
let b= 100;
let c= a+b;
let b=100;
let a=10;
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
	dotSrc.WriteString("program;\n")

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

func gen(start, end ast.Node) string {
	b := strings.Builder{}
	b.WriteString("\"")
	b.WriteString(md5v1(start.String()))
	b.WriteString("\"")

	b.WriteString("->")

	b.WriteString("\"")
	b.WriteString(md5v1(end.String()))
	b.WriteString("\"")
	b.WriteString(";")

	return b.String()
}

func genNode(node ast.Node) string {

	return "\""+md5v1(node.String())+"\""
}


func walk(node ast.Node, lines *[]string) {
	switch node := node.(type) {
	case *ast.Program:
		*lines = append(*lines, genNode(node))
		for _, statement := range node.Statements{
			*lines = append(*lines, gen(node, statement))
			walk(statement,lines)
		}
	case *ast.IfExpression:
		*lines = append(*lines, genNode(node))

		ifToken := token.Token{Type: token.IF, Literal: "if"}
		ifIdentifier := &ast.Identifier{Token: ifToken, Value: token.IF}
		gen(node, ifIdentifier)

		*lines = append(*lines, gen(node, node.Condition))
		walk(node.Condition, lines)
	case *ast.InfixExpression:
		*lines = append(*lines, genNode(node))

		gen(node, node.Left)
		walk(node.Left, lines)

		//gen(node, node.Operator)

		*lines = append(*lines, gen(node, node.Right))
		walk(node.Right, lines)

	case *ast.Identifier:
		*lines = append(*lines, genNode(node))
		return
	case *ast.LetStatement:
		*lines = append(*lines, genNode(node))

		*lines = append(*lines, gen(node, node.Name))
		walk(node.Name, lines)

		*lines = append(*lines, gen(node, node.Value))
		walk(node.Value, lines)
	case *ast.IntegerLiteral:
		*lines = append(*lines, genNode(node))
		return
	case *ast.StringLiteral:
		*lines = append(*lines, genNode(node))
		return
	case *ast.BlockStatement:
		*lines = append(*lines, genNode(node))

	}
}
