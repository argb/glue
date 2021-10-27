package lexer

import (
	"bufio"
	"bytes"
	"fmt"
	"glue/token"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

type Lexer struct {
	input string
	position int
	readPosition int
	ch byte
	char string // for debug, the string format of ch

	currLineNum int
	currColNm int

	lines []string
}


func NewForREPL(input string) *Lexer {
	l := &Lexer{input: input}
	l.lines = append(l.lines, input)

	l.readChar() // init the cursor

	return l
}

func New(input string) *Lexer {
	l := &Lexer{input: input}

	l.readChar()

	return l
}

func NewFromFile(filename string) *Lexer {
	l := &Lexer{}
	err := l.loadSrc(filename)
	//fmt.Println(l.lines)
	if err != nil {
		panic(err)
	}
	l.currLineNum = 0
	l.currColNm = 0

	l.readChar() // init read cursor

	return l
}

/**
该函数的正确性非常重要，是后面一切处理的基石。
 */
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input){
		err := l.readLine()
		if err != nil {
			// 这种情况依然向后移动光标位置主要是因为现在readIdentifier()这个函数处理标志符时候依赖这个位置，不继续向前的话position跟l.position
			// 会变成相等，input[position:l.position]就会取到空字符串，而不是最后的字符，最后一个字符会丢掉了。
			// 暂且这样
			l.ch =0
		}else {
			// 因为加载了新内容，所以重置读取光标位置
			l.position = 0
			l.readPosition = 0

			l.ch = l.input[l.readPosition]
			l.char = string(l.ch) // for debug

		}
	}else {
		l.ch = l.input[l.readPosition]
		l.char = string(l.ch) // for debug

	}

	l.position = l.readPosition
	l.readPosition +=1

}

func (l *Lexer) peakChar() byte {
	if l.readPosition >= len(l.input) {
		err := l.readLine()
		if err != nil {
			return 0
		}else {
			// 新加载内容的开头，因为希望只让readChar修改读取位置，所以这里直接取索引0，而不是让l.readPosition = 0, 然后再l.input[l.readPosition]
			// 如果不止一个地方会修改对字符流的读取位置，很容出错，读取合适的下一个字符是后面一切分析的基石，所以保证读取位置的正确非常重要！
			return l.input[0]
		}

	}else {
		return l.input[l.readPosition]
	}
}
/**
只负责加载，不改变对字符流的读取位置，改变位置由readChar函数负责
*/
func (l *Lexer)readLine() error {
	if len(l.lines) == 0 {
		return fmt.Errorf("没有可加载内容") // 抛出一个报错信号，通知readChar整个多行文件的数据读取完毕，类似EOF的作用
	}
	if l.currLineNum >= len(l.lines) {
		return fmt.Errorf("没有可加载内容")
	}else {
		l.input = l.lines[l.currLineNum]
	}

	l.currLineNum++

	return nil
}

// NextToken /**
/*
该函数驱动词法解析器向前读取字符流，所以skipWhitespace会过滤掉所有出现的'空白',因为这里是整个字符流处理的最开始位置
 */
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	switch l.ch {
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '/':
		tok = newToken(token.SLASH, l.ch)
	case '*':
		tok = newToken(token.ASTERISK, l.ch)
	case '=':
		if l.peakChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: string(ch)+string(l.ch)}
		}else {
			tok = newToken(token.ASSIGN, l.ch)
		}
	case '!':
		if l.peakChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: string(ch)+string(l.ch)}
		}else {
			tok = newToken(token.BANG, l.ch)
		}
	case '<':
		tok = newToken(token.LT, l.ch)
	case '>':
		tok = newToken(token.GT, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case '"':
		//fmt.Println("xxxxx")
		tok.StartChar = '"'
		tok.Type = token.STRING
		tok.Literal = l.readString2(&tok)
	case '[':
		tok = newToken(token.LBRACKET, l.ch)
	case ']':
		tok = newToken(token.RBRACKET, l.ch)
	case ':':
		tok = newToken(token.COLON, l.ch)

	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
		return tok
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		}else if isDigit(l.ch) {
			tok.Literal = l.readNumber()
			tok.Type = token.INT
			return tok
		}else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) readString2(tok *token.Token) string {
	var out bytes.Buffer
	for {
		ch := l.peakChar()
		if ch == '"'{
			break
		}else if ch == 0 {
			tok.Type = token.ILLEGAL
			break
		}
		l.readChar()

		out.Write([]byte{l.ch})

	}
	out.Write([]byte{})
	l.readChar() //吃掉第二个'"'
	return out.String()
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' {
			break
		}
	}
	return l.input[position:l.position]
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func isLetter(ch byte) bool {
	return 'a'<=ch && ch <='z'||'A'<=ch && ch<='Z'||ch =='_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch){
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' '|| l.ch =='\t' || l.ch == '\n' || l.ch == '\r'{
		l.readChar()
	}
}

func newToken(tokenType token.TokenType, char byte) token.Token {
	return token.Token{
		Type: tokenType,
		Literal: string(char),
	}
}

type TokenError struct {
	startCh byte

}


func (l *Lexer) loadSrc(file string) error {
	fmt.Println("file:", file)
	var err error
	file, err = filepath.Abs(file)
	if err != nil {
		panic(err)
	}
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				if line != "" { // 跳过空行
					l.lines = append(l.lines, line) // 最后一行
				}

				return nil
			}else {
				fmt.Printf("error reading file %s", err)
				return err
			}

		}
		//fmt.Println("line:",line)
		if !isEmptyLine(line) { // 跳过空行
			l.lines = append(l.lines, line)
		}

	}
	//fmt.Println(l.lines)
	return nil
}

func isEmptyLine(line string) bool {
	if line == "" || line == "\n" {
		return true
	}
	reg, err := regexp.Compile(`[\s\t]*\r\n`)
	if err != nil {
		panic(err)
	}
	matched := reg.MatchString(line)

	return matched
}
