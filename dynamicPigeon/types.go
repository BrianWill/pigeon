package dynamicPigeon

import (
	"errors"
	"fmt"
	"strconv"
)

type TokenType int

const (
	ReservedWord TokenType = iota
	OperatorWord
	IdentifierWord
	Newline
	Indentation
	OpenParen
	CloseParen
	NumberLiteral
	StringLiteral
	BooleanLiteral
	NilLiteral
	OpenSquare
	CloseSquare
	Dot
	Space
	TypeName
	OpenAngle
	CloseAngle
	Comma
)

const indentationSpaces = 4

var reservedWords = []string{
	"func",
	"global",
	"foreach",
	"while",
	"foreach",
	"forinc",
	"fordec",
	"break",
	"continue",
	"if",
	"else",
	"elif",
	"return",
	"as",
	"locals",
	"_p",
	"_main",
	"_break",
	"_breakpoints",
	"_validBreakpoints",
}

var operators = []string{
	"add",
	"sub",
	"mul",
	"div",
	"inc",
	"dec",
	"mod",
	"eq",
	"neq",
	"not",
	"lt",
	"gt",
	"lte",
	"gte",
	"get",
	"set",
	"list",
	"map",
	"push",
	"or",
	"and",
	"print",
	"println",
	"prompt",
	"concat",
	"len",
	"randNum",
	"parseInt",
	"formatInt",
	"parseFloat",
	"formatFloat",
	"timeNow",
	"formatTime",
	"getchar",
	"getrune",
	"charlist",
	"runelist",
}

type Token struct {
	Type       TokenType
	Content    string // the token itself, e.g. a number 3.7 is stored here as "3.7"
	LineNumber int    // first line is line 1
	Column     int    // first character of a line is in column 1
}

type Statement interface {
	Statement()
	Line() int
}

type Expression interface {
	Expression()
	Line() int
}

type Definition interface {
	Definition()
	Line() int
}

func (t Token) Expression()        {}
func (t FunctionCall) Expression() {}
func (t Operation) Expression()    {}

func (t FunctionDefinition) Definition() {}
func (t GlobalDefinition) Definition()   {}

func (t LocalsStatement) Statement()     {}
func (t IfStatement) Statement()         {}
func (t WhileStatement) Statement()      {}
func (t ForeachStatement) Statement()    {}
func (t ForincStatement) Statement()     {}
func (t AssignmentStatement) Statement() {}
func (t ReturnStatement) Statement()     {}
func (t FunctionCall) Statement()        {}
func (t Operation) Statement()           {}
func (t BreakStatement) Statement()      {}
func (t ContinueStatement) Statement()   {}

func (t LocalsStatement) Line() int {
	return t.LineNumber
}
func (t IfStatement) Line() int {
	return t.LineNumber
}
func (t WhileStatement) Line() int {
	return t.LineNumber
}
func (t ForeachStatement) Line() int {
	return t.LineNumber
}
func (t ForincStatement) Line() int {
	return t.LineNumber
}
func (t AssignmentStatement) Line() int {
	return t.LineNumber
}
func (t ReturnStatement) Line() int {
	return t.LineNumber
}
func (t BreakStatement) Line() int {
	return t.LineNumber
}
func (t ContinueStatement) Line() int {
	return t.LineNumber
}
func (t FunctionCall) Line() int {
	return t.LineNumber
}
func (t Operation) Line() int {
	return t.LineNumber
}
func (t FunctionDefinition) Line() int {
	return t.LineNumber
}
func (t GlobalDefinition) Line() int {
	return t.LineNumber
}

func (t Token) Line() int {
	return t.LineNumber
}

type FunctionDefinition struct {
	LineNumber int
	Column     int
	Name       string
	Parameters []string
	Body       []Statement
	Pkg        *Package
}

type GlobalDefinition struct {
	LineNumber int
	Column     int
	Name       string
	Value      Expression
	Pkg        *Package
}

type FunctionCall struct {
	LineNumber int
	Column     int
	Function   Expression // either an identifier or another function/operator call
	Arguments  []Expression
}

type MethodCall struct {
	LineNumber int
	Column     int
	MethodName string // either an identifier or another function/operator call
	Receiver   Expression
	Arguments  []Expression
}

type Operation struct {
	LineNumber int
	Column     int
	Operator   string
	Operands   []Expression
}

type IfStatement struct {
	LineNumber int
	Column     int
	Condition  Expression
	Body       []Statement
	Elifs      []ElseifClause
	Else       ElseClause
}

type ElseifClause struct {
	LineNumber int
	Column     int
	Condition  Expression
	Body       []Statement
}

type ElseClause struct {
	LineNumber int
	Column     int
	Body       []Statement
}

type LocalsStatement struct {
	LineNumber int
	Column     int
	Vars       []string
}

type WhileStatement struct {
	LineNumber int
	Column     int
	Condition  Expression
	Body       []Statement
}

type ForeachStatement struct {
	LineNumber int
	Column     int
	IndexName  string
	ValName    string
	Collection Expression
	Body       []Statement
}

type ForincStatement struct {
	LineNumber int
	Column     int
	IndexName  string
	StartVal   Expression
	EndVal     Expression
	Body       []Statement
	Dec        bool
}

type ReturnStatement struct {
	LineNumber int
	Column     int
	Value      Expression
}

type BreakStatement struct {
	LineNumber int
	Column     int
}

type ContinueStatement struct {
	LineNumber int
	Column     int
}

type AssignmentStatement struct {
	LineNumber int
	Column     int
	Target     string
	Value      Expression
}

type Package struct {
	Globals          map[string]GlobalDefinition
	ValidBreakpoints map[string]bool
	Funcs            map[string]FunctionDefinition
	Code             string
}

func msg(line int, column int, s string) error {
	return errors.New("Line " + strconv.Itoa(line) + ", column " +
		strconv.Itoa(column) + ": " + s)
}

func debug(args ...interface{}) {
	fmt.Print("DEBUG: ")
	fmt.Println(args...)
}
