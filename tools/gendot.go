package main

import (
	"bufio"
	"bytes"
	"fmt"
	"glue/ast"
	"glue/lexer"
	"glue/parser"
	"os"
	"strconv"
	"strings"
)

func main() {
	input :=`
let a=10;
let b=fn(x,y){ return x+y;}
let c= a+b;
`
	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	fmt.Println("the whole program:\n", program.String())

	var dotSrc bytes.Buffer
	dotSrc.WriteString("digraph ast {\n")
	dotSrc.WriteString(`label = "statement list";`)
	dotSrc.WriteString("\n")
	dotSrc.WriteString("root;\n")
	for i, s := range program.Statements {
		order :=strconv.Itoa(i)
		fmt.Println(s.String()+order)
		dotSrc.WriteString("root->"+s.TokenLiteral()+order)
		dotSrc.WriteString(";")
	}
	dotSrc.WriteString("\n}")

	genFile(dotSrc.String())
}

func genFile(src string) {
	filePath := "./ast.dot"
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Println("文件打开错误：%v \n", err)
		return
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	writer.WriteString(src)

	writer.Flush()
}

type handler func(v ...interface{})

func gen(start, end ast.Node) {
	b := strings.Builder{}
	switch node := start.(type) {
	case ast.Expression:
		b.WriteString(node.String())
	}

	b.WriteString("->")

	switch node := end.(type) {
	case ast.Expression:
		b.WriteString(node.String())
	}

}


func walk(program ast.Program, handle handler) {

}
