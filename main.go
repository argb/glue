package main

import (
	"bufio"
	"compiler01/evaluator"
	"compiler01/lexer"
	"compiler01/object"
	"compiler01/parser"
	"compiler01/repl"
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"io"
	"os"
	user2 "os/user"
	"reflect"
	"strings"
)

func main() {
	//fmt.Println((*int)(nil) == (*int)(nil))
	user, err := user2.Current()
	if err != nil {
		panic(err)
	}
	//t10()
	//t1()
	//ttt()
//tt()
	/*
	f := flag.String("src", "", "source file")
	flag.Parse()
	f, err := os.OpenFile(*f, os.O_RDONLY, 0644)
	 */

	//terminal.Set(terminal.FG_GREEN)
	color.Set(color.FgMagenta)
	fmt.Printf("你好 %s! 吃了吗？\n", user.Username)
	//terminal.Unset()
	fmt.Printf("欢迎使用【Go艹】语言！\n")
	color.Unset()

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

type Animal interface {
	walk()
}
type Dog struct {
	Name string
}

func (d Dog)walk()  {
	fmt.Println("I can walk")
}

func getNilDog() *Dog {
	return nil
}

func getNilNilDog() Animal {
	return getNilDog()
}

func ttt() {
	//var empty interface{}
	//var emptyDog Dog
	//fmt.Println(emptyDog==nil) // 无法通过编译

	//nilDog := getNilDog() // 编译时类型检查无法给出提示
	nilDog := getNilNilDog()
	fmt.Printf("I am a not existed dog %#v\n", nilDog)
	fmt.Println("the judge value is", nilDog == nil)
	if nilDog != nil {
		println("I want to say:I am not nil.")
	}
	//var d Dog
	var dp *Dog
	idp :=Animal(dp)
	fmt.Println(dp ==nil,idp==nil, reflect.ValueOf(idp).IsNil(), reflect.ValueOf(dp).IsNil())
	//var i Dog
	p := (*int)(nil)
	fmt.Println(reflect.ValueOf(p).Elem().IsValid())

	num :=0
	{
		var num = 9
		num =10
		num, n:= secret()
		fmt.Println(num, n)
	}
	fmt.Println("the value of num:", num)

}
func secret() (int,int) {
	return 100, 200
}

func t1() {
	var a = []int { 1,2,3,4,5}
	b := a[2:4]
	var b1 = make([]byte, 2)
	 binary.BigEndian.PutUint16(b1, uint16(65534))
	fmt.Println(b,b1,byte(b1[0]))
}
func n(n int) int {
	return n
}

type objk struct {
	name string
}
type pair map[*objk]string
func init() {
	var m map[string]string
	m = map[string]string{"name":"wg"}
	fmt.Printf("m is %p",m)
}
