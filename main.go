package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"glue/compiler"
	"glue/evaluator"
	"glue/lexer"
	"glue/object"
	"glue/parser"
	"glue/repl"
	"glue/tools/log"
	"glue/vm"
	"io"
	"os"
	"os/user"
	"reflect"
	"strings"
)

func main() {

	currUser, err := user.Current()
	if err != nil {
		panic(err)
	}

	//terminal.Set(terminal.FG_GREEN)
	color.Set(color.FgMagenta)
	fmt.Printf("你好 %s! 吃了吗？\n", currUser.Username)
	//terminal.Unset()
	fmt.Printf("欢迎使用【Go艹】语言！\n")
	color.Unset()

	interactive := flag.Bool("i", false, "start REPL")
	engine := flag.String("engine", "vm", "the running mode, evaluate directly or by vm")
	input := flag.String("src", "", "the input(source) file name")
	//output := flag.String("src", "", "the output file name")
	flag.Parse()
	fmt.Println(*interactive)
	if *interactive == true {
		//terminal.TT()
		repl.Start(os.Stdin, os.Stdout)
		return
	}
	args :=flag.Args()
	var iptFile string

	if *input == "" {
		if len(args) == 0 {
			fmt.Println("请指定源文件，示例：")
			fmt.Println("./glue helloworld.gl")
			//return
		}else {
			iptFile = args[0]
		}

	}else {
		iptFile = *input
	}
	iptFile = "./examples/t3.gl"
	fmt.Println(iptFile)
	log.Infof("source file: %s", iptFile)

	//fmt.Printf("input and args0: %#v, %#v\n", *input, args[0])

	l := lexer.NewFromFile(iptFile)
	//l :=lexer.New("-a * b")
	p := parser.New(l)
	program := p.ParseProgram()
	if p.HasError() {
		p.ReportParseErrors()
		os.Exit(10)
	}
	if *engine == "vm" {
		c := compiler.New()
		err = c.Compile(program)
		if err != nil {
			log.ErrorF("error %s", err)
			panic(err)
		}
		c.Info()
		//fmt.Println("instructions:", c.Bytecode().Instructions.String())
		machine := vm.New(c.Bytecode())

		err = machine.Run()
		if err != nil {
			log.ErrorF("error %s", err)
			panic(err)
		}
		result := machine.LastPoppedStackElem()
		//io.WriteString(os.Stdout, lastPopped.Inspect())
		//io.WriteString(os.Stdout, "\n")
		fmt.Println("engine: vm")
		fmt.Println(result.Inspect())
		machine.ShowReadableConstants()
		//mn := monitor.SingletonNew()
		//fmt.Println("Instruction sequence:")
		//mn.ShowInstructions()

	}else {
		env := object.NewEnvironment()
		result := evaluator.Eval(program, env)
		fmt.Println("engine: evaluating")
		fmt.Println(result)
	}


	if err != nil {
		panic(err)
	}

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
	l := lexer.NewForREPL(input)
	p := parser.New(l)
	program := p.ParseProgram()

	env := object.NewEnvironment()

	return evaluator.Eval(program, env)
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
func xxx() {
	var m map[string]string
	m = map[string]string{"name":"wg"}
	fmt.Printf("m is %p\n",m)
	i:=0
	for i<3 {
		i++
		fmt.Println("i:", i)
	}


}
