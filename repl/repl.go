package repl

import (
	"bufio"
	"glue/compiler"
	"glue/lexer"
	"glue/object"
	"glue/parser"
	"glue/vm"
	"fmt"
	"github.com/fatih/color"
	"io"
)

const PROMPT = ">>"
const POEM = `
鹅，鹅，鹅，曲项向天歌，白毛浮绿水，红掌拨清波。
`
func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	//env := object.NewEnvironment()

	var constants []object.Object
	globals := make([]object.Object, vm.GlobalSize)
	symbolTable := compiler.NewSymbolTable()
	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	color.Green(POEM)

	for {
		color.Set(color.FgCyan)
		fmt.Println(PROMPT)
		color.Unset()
		scanned := scanner.Scan()
		if !scanned {
			return
		}
		line := scanner.Text()
/*
		line = `let newAdder = fn(a) {
let adder = fn(b) { a + b; };
return adder;
};
let addTwo = newAdder(2);
addTwo(3);`
*/
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		for _, s := range program.Statements {
			fmt.Println(s.String())
		}

		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}
		color.Set(color.FgGreen)
		fmt.Println("Parsing result:")
		io.WriteString(out, program.String())
		io.WriteString(out, "\n")
		color.Unset()

		/*
		evaluated := evaluator.Eval(program, env)
		if evaluated != nil {
			color.Set(color.FgMagenta)
			fmt.Println("Evaluated result:")
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
			color.Unset()
		}
		 */

		comp :=compiler.NewWithState(symbolTable, constants)
		err := comp.Compile(program)
		if err != nil {
			fmt.Fprintf(out, "Woops! Compilation failed:\n %s\n", err)
			continue
		}

		byteCode := comp.Bytecode()
		constants = byteCode.Constants

		machine := vm.NewWithGlobalsStore(byteCode, globals)
		err = machine.Run()
		if err != nil {
			fmt.Fprintf(out, "Woops! Executing bytecode failed:\n %s\n", err)
			continue
		}

		machine.ShowReadableInstructions()
		machine.ShowReadableConstants()
		machine.ShowStack()
		machine.ShowCallStack()

		lastPopped := machine.LastPoppedStackElem()
		io.WriteString(out, lastPopped.Inspect())
		io.WriteString(out, "\n")



		/*
		for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken(){
			fmt.Printf("%+v\n", tok)
		}
		 */
	}
}

func printParserErrors(out io.Writer, errors []string) {
	color.Set(color.FgRed)
	io.WriteString(out, "错了！傻逼！\r\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
	defer color.Unset()
}