package main

import (
	"bufio"
	"compiler01/evaluator"
	"compiler01/lexer"
	"compiler01/object"
	"compiler01/parser"
	"compiler01/repl"
	"flag"
	"fmt"
	"io"
	"os"
	user2 "os/user"
	"strings"
)

func main() {
	user, err := user2.Current()
	if err != nil {
		panic(err)
	}
tt()
	/*
	f := flag.String("src", "", "source file")
	flag.Parse()
	f, err := os.OpenFile(*f, os.O_RDONLY, 0644)
	 */

	//terminal.Set(terminal.FG_GREEN)
	fmt.Printf("Hello %s! This is the Monkey programming language!\n", user.Username)
	//terminal.Unset()
	fmt.Printf("Feel free to type in commands\n")

	//terminal.TT()
	repl.Start(os.Stdin, os.Stdout)
/*
	input :=`
let map = fn(arr, f) {
let iter = fn(arr, accumulated) {
if (len(arr) == 0) {
accumulated
} else {
iter(rest(arr), push(accumulated, f(first(arr))));
}
};
iter(arr, []);
};

let a = [1, 2, 3, 4];
let double = fn(x) { x * 2 };
map(a, double);
`
	//obj := testEval(doRead())
	/*
	obj := testEval(input)
	if obj == nil {
		fmt.Println("no valid ast node to evaluate")
		return
	}
	fmt.Println(obj.Inspect())

	 */
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	env := object.NewEnvironment()

	return evaluator.Eval(program, env)
}

func doRead() string{
	flag.Parse()
	if len(flag.Args()) == 0 {
		fmt.Printf("usage: byLine <file1> [<file2> ...]\n")
		return ""
	}
	return lineByLine(flag.Args()[0])
}

func lineByLine(file string) string {
	var err error
	f, err := os.Open(file)
	if err != nil {
		return ""
	}
	defer f.Close()
	r := bufio.NewReader(f)
	var lines []string
	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Printf("error reading file %s", err)
			break
		}
		fmt.Print(line)
		lines = append(lines, line)
	}
	srcText := strings.Join(lines, "\n")
	return srcText
}

type P interface {
	say() string
}
type Person struct {
	Name string
}

func (p *Person) say() string {
	fmt.Println("hello")
	return ""
}

func tt() {
	a := &Person{Name: "wg"}
	b := &Person{Name: "wg"}
	//c := Person{Name :"wg"}

	xx(b)

	fmt.Println(a == b)
}

func xx(i P) interface{} {
	p := i.(*Person)
	fmt.Println(p.Name)
	p.say()

	return p
}