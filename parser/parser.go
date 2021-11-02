package parser

import (
	"fmt"
	"glue/ast"
	"glue/lexer"
	"glue/token"
	"reflect"
	"strconv"
	"strings"
)

// 操作符优先级
//type Precedence int
const (
	_ int = iota
	LOWEST
	ASSIGN
	EQUALS // ==
	LESSGREATER // > or <
	SUM // +
	PRODUCT // *
	PREFIX // -x OR !x
	CALL // myFuction(x)
	INDEX
)

var precedences = map[token.TokenType]int{
	token.ASSIGN: ASSIGN,
	token.EQ: EQUALS,
	token.NOT_EQ: EQUALS,
	token.LT: LESSGREATER,
	token.GT: LESSGREATER,
	token.PLUS: SUM,
	token.MINUS: SUM,
	token.SLASH: PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN: CALL,
	token.LBRACKET: INDEX,
}

var NodeIndex int64 = 0

func getNodeIndex() int64 {
	NodeIndex++
	return NodeIndex
}

type ParseError struct {
	Token *token.Token
	LineNum int
	ColNum int
	msg string
}

func (pe *ParseError) String() string {
	var builder strings.Builder
	var lineNum, colNum int
	var tokenInfo string

	builder.Grow(200) // 初始化一下，避免过多内存分配次数

	if pe.Token != nil {
		tokenInfo = pe.Token.Literal
	}
	lineNum = pe.LineNum
	colNum = pe.ColNum
	builder.Grow(100)
	builder.WriteString(pe.msg)
	builder.WriteString("[Token:")
	builder.WriteString(tokenInfo)
	builder.WriteString("]")
	builder.WriteString("[")
	builder.WriteString(string(strconv.Itoa(lineNum)))
	builder.WriteString(":")
	builder.WriteString(strconv.Itoa(colNum))
	builder.WriteString("]")

	return builder.String()
}

func (pe *ParseError) Error() string {
	errorInfo := fmt.Sprintf("parse error, token: %#v, line: %d, error descrition: %s\n",
		pe.Token, pe.LineNum, pe.msg)

	return errorInfo
}

type Parser struct {
	l *lexer.Lexer
	errors []string

	curToken token.Token
	peekToken token.Token

	prefixParserFns map[token.TokenType]prefixParserFn
	infixParseFns map[token.TokenType]infixParseFn

	currLineNum int
	currColNum int
}

type (
	prefixParserFn func() ast.Expression
	infixParseFn func(expression ast.Expression) ast.Expression
)

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParserFn) {
	p.prefixParserFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l: l,
		errors: []string{},
	}
	p.nextToken()
	p.nextToken()

	p.prefixParserFns = make(map[token.TokenType]prefixParserFn)

	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.LBRACE, p.parseHashLiteral2)

	p.infixParseFns	  = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerInfix(token.ASSIGN, p.parseAssignExpression)

	return p
}

func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	stmt := &ast.WhileStatement{Token: p.curToken, Id: getNodeIndex()}
	if p.expectPeek(token.LPAREN) {
		p.nextToken()
		stmt.Condition = p.parseExpression(LOWEST)
	}else{
		return nil
	}

	if !p.expectPeek(token.RPAREN) {

		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	stmt.Body = p.parseBlockStatement()
	return stmt
}

func (p* Parser) parseAssignExpression(left ast.Expression) ast.Expression {
	assign := &ast.AssignExpression{Token:p.curToken, Name: left, Operator: p.curToken.Literal, Id: getNodeIndex()}

	p.nextToken()
	assign.Expression = p.parseExpression(LOWEST)

	return assign
}

func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken, Id: getNodeIndex()}
	hash.Pairs = make(map[ast.Expression]ast.Expression)
	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		key := p.parseExpression(LOWEST)
		if !p.expectPeek(token.COLON) {
			return nil
		}
		p.nextToken()
		value := p.parseExpression(LOWEST)
		hash.Pairs[key] = value
		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}
	}
	if !p.expectPeek(token.RBRACE) {
		return nil
	}
	return hash
}

func (p *Parser) parseHashLiteral2() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken, Id: getNodeIndex()}
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	if p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		return hash
	}

	label:
		p.nextToken()
		key := p.parseExpression(LOWEST)
		if !p.expectPeek(token.COLON) {
			return nil
		}
		p.nextToken()
		value := p.parseExpression(LOWEST)
		hash.Pairs[key] = value
	if p.peekTokenIs(token.COMMA) {
		p.nextToken()
		goto label
	}else if p.expectPeek(token.RBRACE) {
		return hash
	}else {
		return nil
	}

	//return hash
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left, Id: getNodeIndex()}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)
	if !p.expectPeek(token.RBRACKET) {
		return nil
	}
	return exp
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken, Id: getNodeIndex()}

	array.Elements = p.parseExpressionList(token.RBRACKET)

	return array
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	var list []ast.Expression

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}
	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) HasError() bool {
	if len(p.errors) > 0 {
		return true
	}

	return false
}

func (p *Parser) ReportParseErrors() {
	for _, err := range p.errors {
		fmt.Println(err)
	}
}

func (p *Parser) addParseError(pErr *ParseError) {
	pErr.LineNum = p.l.CurrLineNum
	pErr.ColNum = p.l.CurrColNum

	p.errors = append(p.errors, pErr.String())
}

func (p *Parser) peekError(t token.TokenType) {
	pErr := new(ParseError)
	pErr.Token = &p.curToken
	pErr.LineNum = p.l.CurrLineNum
	pErr.ColNum = p.l.CurrColNum
	msg := fmt.Sprintf("expected next token to be %s, got %s instead.", t, p.peekToken.Type)
	pErr.msg = msg

	p.errors = append(p.errors, pErr.String())
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

/**
向前看一步
 */
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

/**
向前走一步
*/
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

/*
向前看一步，符合预期就向前走一步，否则原地不动, 可以看做向前看与向前走的合体逻辑
函数名约定：以expect开头，比如expectXXX，表示必须出现期望的token，否则就认为发生了错误。 如果单纯的是探测一下下一个token是啥，
不要用这个函数，可以使用peekTokenIs。
这样是为了进行正确的错误处理，以及给出正确错误位置。
 */
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t){
		p.nextToken()
		return true
	}else {
		// 不是预期token就报错
		p.peekError(t)
		return false
	}
}

// ParseProgram /**
/**
Parser执行语法解析的入口方法
 */
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{Id: getNodeIndex()}
	program.Statements = [] ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil && !reflect.ValueOf(stmt).IsNil(){
			//fmt.Printf("I am not nil, my value type is:%T,\n my value is %#v\n", stmt, stmt)
			//fmt.Printf("I am not nil, my value type is:%T,\n my value is %#v\n", stmt, stmt)
			//fmt.Println("I am nil", stmt)

			//fmt.Printf("I am not nil, I am %#v,my value is %s",stmt, stmt.String())
			program.Statements = append(program.Statements, stmt)

		}else{
			pErr := new(ParseError)
			pErr.msg = fmt.Sprintf("can't parse statement.")
			p.addParseError(pErr)

			//fmt.Println("I am nil", stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	//case token.IDENT:
		//return p.parseAssignStatement()
	case token.WHILE:
		return p.parseWhileStatement()
	case token.FUNCTION:
		if p.peekTokenIs(token.IDENT) {
			return p.parseFunctionDefinitionStatement()
		}
		return p.parseExpressionStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseAssignStatement() *ast.AssignStatement {
	stmt := &ast.AssignStatement{Token: p.curToken, Id: getNodeIndex()}
	/*
	left := p.parseIdentifier()
	as := p.parseAssignExpression(left)

	if p.peekTokenIs(token.ASSIGN) {
		p.nextToken()
		p.nextToken()
		stmt.Expression = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(token.SEMICOLON) {
		return nil
	}
	 */
	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken, Id: getNodeIndex()}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		// 表达式后的;不是必须的，有的话就吃掉，没有也不会报错
		p.nextToken() // consume the current ';'
	}
	return stmt
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %q found.", t)
	//fmt.Printf("token:%#v, token-literal:%s\n", t, p.curToken.Literal)
	//fmt.Println(msg)
	pErr := new(ParseError)

	pErr.Token = &p.curToken
	pErr.LineNum = p.l.CurrLineNum
	pErr.ColNum = p.l.CurrColNum

	pErr.msg = msg

	p.errors = append(p.errors, pErr.String())
}

func (p *Parser) parseE() ast.Expression {
	var leftEpsilon ast.Expression //
	initOp := "epsilon"
	precedence := LOWEST
	exp := &ast.InfixExpression{Left: leftEpsilon, Operator: initOp}
	right := p.parseExpression(precedence)
	exp.Right = right

	return exp
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	//fmt.Printf("token: %#v \n", p.curToken)
	prefix := p.prefixParserFns[p.curToken.Type]

	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}

	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()

		leftExp = infix(leftExp)

	}

	return leftExp
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{Token: p.curToken, Operator: p.curToken.Literal, Id: getNodeIndex()}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token: p.curToken,
		Operator: p.curToken.Literal,
		Left: left,
		Id: getNodeIndex(),
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal, Id: getNodeIndex()}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken, Id: getNodeIndex()}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("Could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken, Id: getNodeIndex()}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	for p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken, Id: getNodeIndex()}

	if !p.expectPeek(token.IDENT) {

		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal, Id: getNodeIndex()}

	if !p.peekTokenIs(token.ASSIGN) {
		if p.peekTokenIs(token.SEMICOLON) { // 此处看做是声明 let a;用户没有赋值。
			stmt.Value = nil
			p.nextToken()
			p.nextToken()

			return stmt
		}
		//fmt.Println("从这里返回了")

		p.peekError(token.ASSIGN)

		return nil
	}

	p.nextToken()
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if fl, ok := stmt.Value.(*ast.FunctionLiteral); ok {
		//fl.Name = stmt.Name.Value
		//复制一份，更新Id，否则画图出错
		name := &ast.Identifier{Token: stmt.Name.Token, Value: stmt.Name.Value, Id: getNodeIndex()}
		fl.Name = name
		fl.From = ast.EXPRESSION
	}

	for p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}


func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE), Id: getNodeIndex()}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken, Id: getNodeIndex()}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()
		if !p.expectPeek(token.LBRACE){
			return nil
		}
		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken, Id: getNodeIndex()}
	block.Statements = []ast.Statement{}
	p.nextToken()
	for !p.curTokenIs(token.RBRACE) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit :=&ast.FunctionLiteral{Token: p.curToken, Id: getNodeIndex()} // The 'fn' token

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	lit.Parameters = p.parseFunctionParameters()
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	//identifiers := [] *ast.Identifier{}
	var identifiers []*ast.Identifier

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()
	ident := &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
		Id: getNodeIndex(),
	}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
			Id: getNodeIndex(),
		}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers

}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function, Id: getNodeIndex()}
	exp.Arguments = p.parseExpressionList(token.RPAREN)

	return exp
}

/*
过渡方法，暂时不用了，parseExpressionList是其对应的升级版
 */
func (p *Parser) parseCallArguments() []ast.Expression {
	var args []ast.Expression
	if p.peekTokenIs(token.RPAREN){
		p.nextToken()
		return args
	}

	p.nextToken()

	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return args
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal, Id: getNodeIndex()}
}

func (p *Parser) parseFunctionDefinitionStatement() *ast.FunctionDefinitionStatement {
	fnLiteral := &ast.FunctionLiteral{
		Token: p.curToken,
		Id: getNodeIndex(),
		From: ast.STATEMENT,
	}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	fnLiteral.Name = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
		Id: getNodeIndex(),
	}
	p.nextToken()

	params := p.parseFunctionParameters()
	fnLiteral.Parameters = params

	p.nextToken()
	body := p.parseBlockStatement()
	fnLiteral.Body = body

	fnStatement := &ast.FunctionDefinitionStatement{
		Id: getNodeIndex(),
		Token: p.curToken,
		FnLiteral: fnLiteral,
	}

	return fnStatement
}
