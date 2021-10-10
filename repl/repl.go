package repl

import (
	"bufio"
	"compiler01/evaluator"
	"compiler01/lexer"
	"compiler01/object"
	"compiler01/parser"
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
	env := object.NewEnvironment()

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

		evaluated := evaluator.Eval(program, env)
		if evaluated != nil {
			color.Set(color.FgMagenta)
			fmt.Println("Evaluated result:")
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
			color.Unset()
		}

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