package ast

import (
	"bytes"
	"fmt"
	"glue/token"
	"strings"
)

type NodeType string

const (
	PROGRAM NodeType = "PROGRAM"
	IDENTIFIER NodeType = "IDENTIFIER"
	LETSTATEMENT NodeType = "LETSTATEMENT"
	RETURNSTATEMENT NodeType = "RETURNSTATEMENT"
	EXPRESSIONSTATEMENT NodeType = "EXPRESSIONSTATEMENT"
	INTEGERLITERAL NodeType = "INTEGERLITERAL"
	PREFIXEXPRESSION NodeType = "PREFIXEXPRESSION"
	INFIXEXPRESSION NodeType = "INFIXEXPRESSION"
	BOOLEAN NodeType = "BOOLEAN"
	IFEXPRESSION NodeType = "IFEXPRESSION"
	BLOCKSTATEMENT NodeType = "BLOCKSTATEMENT"
	FUNCTIONLITERAL NodeType = "FUNCTIONLITERAL"
	CALLEXPRESSION NodeType = "CALLEXPRESSION"
	STRINGLITERAL NodeType = "STRINGLITERAL"
	ARRAYLITERAL NodeType = "ARRAYLITERAL"
	INDEXEXPRESSION NodeType = "INDEXEXPRESSION"
	HASHLITERAL NodeType = "HASHLITERAL"
	ASSIGNSTATEMENT NodeType = "ASSIGNSTATEMENT"
	ASSIGNEXPRESSION NodeType = "ASSIGNEXPRESSION"
	WHILESTATEMENT NodeType = "WHILESTATEMENT"
	FUNCTIONDEFINITIONSTATEMENT NodeType = "FUNCTIONDEFINITIONSTATEMENT"

)

type Node interface {
	TokenLiteral() string
	String() string

	Tag() string // for drawing ast tree
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement

	Id int64 //用时间戳表示 time.Now.Unix()
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}else {
		return ""
	}
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

func (p *Program) Tag() string {
	return fmt.Sprintf("[%s]%d", PROGRAM, p.Id)
}

type Identifier struct {
	Token token.Token
	Value string
	Id int64
}

func (i *Identifier) String() string {
	return i.Value
}

func (i *Identifier) expressionNode() {

}

func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}

func (this *Identifier) Tag() string {
	return fmt.Sprintf("[%s]%d", IDENTIFIER, this.Id)
}

type LetStatement struct {
	Token token.Token
	Name *Identifier
	Value Expression
	Id int64
}

func (ls *LetStatement) statementNode() {

}

func (ls *LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	out.WriteString(";")

	return out.String()
}

func (ls *LetStatement) TokenLiteral() string {
	return ls.Token.Literal
}

func (this *LetStatement) Tag() string {
	return fmt.Sprintf("[%s]%d", LETSTATEMENT, this.Id)
}

type ReturnStatement struct {
	Token token.Token
	ReturnValue Expression
	Id int64
}

func (rs *ReturnStatement) statementNode() {

}
func (rs *ReturnStatement) TokenLiteral() string {
	return rs.Token.Literal
}

func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + " ")

	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}

	out.WriteString(";")

	return out.String()
}

func (this *ReturnStatement) Tag() string {
	return fmt.Sprintf("[%s]%d", RETURNSTATEMENT, this.Id)
}

type ExpressionStatement struct {
	Token token.Token
	Expression Expression
	Id int64
}

func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}

	return ""
}

func (es *ExpressionStatement) statementNode() {

}

func (es *ExpressionStatement) TokenLiteral() string  {
	return es.Token.Literal
}

func (this *ExpressionStatement) Tag() string {
	return fmt.Sprintf("[%s]%d", EXPRESSIONSTATEMENT, this.Id)
}

type IntegerLiteral struct {
	Token token.Token
	Value int64
	Id int64
}

func (il *IntegerLiteral) expressionNode() {

}

func (il *IntegerLiteral) TokenLiteral() string {
	return il.Token.Literal
}

func (il *IntegerLiteral) String() string {
	return il.Token.Literal
}

func (this *IntegerLiteral) Tag() string {
	return fmt.Sprintf("[%s]%d", INTEGERLITERAL, this.Id)
}

type PrefixExpression struct {
	Token token.Token
	Operator string
	Right Expression
	Id int64
}

func (pe *PrefixExpression) expressionNode() {

}

func (pe *PrefixExpression) TokenLiteral() string {
	return pe.Token.Literal
}

func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

func (this *PrefixExpression) Tag() string {
	return fmt.Sprintf("[%s]%d", PREFIXEXPRESSION, this.Id)
}

type InfixExpression struct {
	Token token.Token
	Left Expression
	Operator string
	Right Expression
	Id int64
}

func (ie *InfixExpression) expressionNode() {

}

func (ie *InfixExpression) TokenLiteral() string {
	return ie.Token.Literal
}

func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

func (this *InfixExpression) Tag() string {
	return fmt.Sprintf("[%s]%d", INFIXEXPRESSION, this.Id)
}

type Boolean struct {
	Token token.Token
	Value bool
	Id int64
}

func (b *Boolean) expressionNode() {

}

func (b *Boolean) TokenLiteral() string {
	return b.Token.Literal
}

func (b *Boolean) String() string {
	return b.Token.Literal
}

func (this *Boolean) Tag() string {
	return fmt.Sprintf("[%s]%d", BOOLEAN, this.Id)
}

type IfExpression struct {
	Token token.Token
	Condition Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
	Id int64
}

func (ie *IfExpression) expressionNode() {

}
func (ie *IfExpression) TokenLiteral() string {
	return ie.Token.Literal
}

func (ie *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())

	if ie.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(ie.Alternative.String())
	}

	return out.String()
}

func (this *IfExpression) Tag() string {
	return fmt.Sprintf("[%s]%d", IFEXPRESSION, this.Id)
}

type BlockStatement struct {
	Token token.Token
	Statements []Statement
	Id int64
}

func (bs *BlockStatement) statementNode() {

}

func (bs *BlockStatement) TokenLiteral() string {
	return bs.Token.Literal
}

func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

func (this *BlockStatement) Tag() string {
	return fmt.Sprintf("[%s]%d", BLOCKSTATEMENT, this.Id)
}

const (
	EXPRESSION int64 = iota
	STATEMENT
)

type FunctionLiteral struct {
	Token token.Token // The 'fn' token
	Parameters [] *Identifier
	Body *BlockStatement

	Name *Identifier // function name,which mainly used to handle the closure-recursion problem

	Id int64

	From int64
}

func (fl *FunctionLiteral) expressionNode() {

}

func (fl *FunctionLiteral) TokenLiteral() string {
	return fl.Token.Literal
}

func (this *FunctionLiteral) Tag() string {
	return fmt.Sprintf("[%s]%d", FUNCTIONLITERAL, this.Id)
}

func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer
	var params []string
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(fl.TokenLiteral())
	if fl.Name != nil {
		out.WriteString(fmt.Sprintf("<%s>", fl.Name))
	}
	out.WriteString("(")
	out.WriteString(strings.Join(params, ","))
	out.WriteString(")")
	out.WriteString(fl.Body.String())

	return out.String()
}

type CallExpression struct {
	Token token.Token  // the '(' token
	Function Expression // Identifier or FunctionLiteral
	Arguments []Expression
	Id int64
}

func (ce *CallExpression) expressionNode()  {

}

func (ce *CallExpression) TokenLiteral() string {
	return ce.Token.Literal
}

func (ce *CallExpression) String() string {
	var out bytes.Buffer
	var args []string

	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

func (this *CallExpression) Tag() string {
	return fmt.Sprintf("[%s]%d", CALLEXPRESSION, this.Id)
}

type StringLiteral struct {
	Token token.Token
	Value string
	Id int64
}

func (sl *StringLiteral) expressionNode() {

}
func (sl *StringLiteral) TokenLiteral() string {
	return sl.Token.Literal
}

func (sl *StringLiteral) String() string {
	return sl.Token.Literal
}

func (this *StringLiteral) Tag() string {
	return fmt.Sprintf("[%s]%d", STRINGLITERAL, this.Id)
}

type ArrayLiteral struct {
	Token token.Token // the '[' token
	Elements []Expression
	Id int64
}

func (al *ArrayLiteral) expressionNode() {

}

func (al *ArrayLiteral) TokenLiteral() string {
	return al.Token.Literal
}

func (al *ArrayLiteral) String() string {
	var out bytes.Buffer

	elements := []string{}

	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

func (this *ArrayLiteral) Tag() string {
	return fmt.Sprintf("[%s]%d", ARRAYLITERAL, this.Id)
}

type IndexExpression struct {
	Token token.Token // the [ token
	Left Expression
	Index Expression
	Id int64
}

func (ie *IndexExpression) expressionNode() {

}

func (ie *IndexExpression) TokenLiteral() string {
	return ie.Token.Literal
}

func (ie *IndexExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")
	// todo:生成dot代码来画出ast树

	return out.String()
}

func (this *IndexExpression) Tag() string {
	return fmt.Sprintf("[%s]%d", INDEXEXPRESSION, this.Id)
}

type HashLiteral struct {
	Token token.Token
	Pairs map[Expression]Expression
	Id int64
}

func (hl *HashLiteral) expressionNode() {

}

func (hl *HashLiteral) TokenLiteral() string {
	return hl.Token.Literal
}

func (hl *HashLiteral) String() string {
	var out bytes.Buffer

	var pairs []string

	for key, value := range hl.Pairs {
		pairs = append(pairs, key.String()+":"+value.String())
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

func (this *HashLiteral) Tag() string {
	return fmt.Sprintf("[%s]%d", HASHLITERAL, this.Id)
}

type AssignStatement struct {
	Token token.Token
	Expressions [] Expression
	Id int64
}

func (as *AssignStatement) statementNode()  {

}

func (as *AssignStatement) TokenLiteral() string {
	return as.Token.Literal
}

func (as *AssignStatement) String() string {
	var out bytes.Buffer

	for _, exp := range as.Expressions {
		out.WriteString(exp.String())
		out.WriteString(", ")
	}
	out.WriteString(";")
	return out.String()
}

func (this *AssignStatement) Tag() string {
	return fmt.Sprintf("[%s]%d", ASSIGNSTATEMENT, this.Id)
}

type AssignExpression struct {
	Token token.Token
	Name Expression
	Operator string
	Expression Expression
	Id int64
}

func (ae *AssignExpression) expressionNode() {

}

func (ae *AssignExpression) TokenLiteral() string {
	return ae.Token.Literal
}

func (ae *AssignExpression) String() string {
	var out bytes.Buffer

	out.WriteString(ae.Name.String())
	out.WriteString(" "+ae.Operator+" ")
	out.WriteString(ae.Expression.String())
	out.WriteString(";")

	return out.String()
}

func (this *AssignExpression) Tag() string {
	return fmt.Sprintf("[%s]%d", ASSIGNEXPRESSION, this.Id)
}

type WhileStatement struct {
	Token token.Token
	Condition Expression
	Body *BlockStatement
	Id int64
}

func (ws *WhileStatement) statementNode() {

}

func (ws *WhileStatement) TokenLiteral() string {
	return ws.Token.Literal
}

func (ws *WhileStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ws.Token.Literal)
	out.WriteString("( ")
	out.WriteString(ws.Condition.String())
	out.WriteString(" ) {\n")
	out.WriteString(ws.Body.String())
	out.WriteString("\n}")

	return out.String()
}

func (this *WhileStatement) Tag() string {
	return fmt.Sprintf("[%s]%d", WHILESTATEMENT, this.Id)
}

type FunctionDefinitionStatement struct {
	Name *Identifier
	Token token.Token
	Parameters [] *Identifier
	Body *BlockStatement
	Id int64
	FnLiteral *FunctionLiteral
}

func (this *FunctionDefinitionStatement) statementNode()  {

}

func (this *FunctionDefinitionStatement) TokenLiteral() string {
	return this.Token.Literal
}

func (this *FunctionDefinitionStatement) String() string {
	var out bytes.Buffer

	out.WriteString(this.FnLiteral.String())

	return out.String()
}

func (this *FunctionDefinitionStatement) Tag() string {

	return fmt.Sprintf("[%s]%d", FUNCTIONDEFINITIONSTATEMENT, this.Id)
}
