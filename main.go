package main

import (
	"compiler01/evaluator"
	"compiler01/lexer"
	"compiler01/object"
	"compiler01/parser"
	"compiler01/repl"
	"fmt"
	"os"
	user2 "os/user"
)

func main() {
	user, err := user2.Current()
	if err != nil {
		panic(err)
	}

	//terminal.Set(terminal.FG_GREEN)
	fmt.Printf("Hello %s! This is the Monkey programming language!\n", user.Username)
	//terminal.Unset()
	fmt.Printf("Feel free to type in commands\n")

	//terminal.TT()
	repl.Start(os.Stdin, os.Stdout)
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	env := object.NewEnvironment()

	return evaluator.Eval(program, env)
}
