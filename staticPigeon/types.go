package staticPigeon

import (
	"errors"
	"fmt"
	"strconv"
)

// we use arbitrary number values to designate each type of token. Rather than using straight ints, we
// create a distinct type to help avoid mistreating these values like ints.
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
	MultilineStringLiteral
	BooleanLiteral
	NilLiteral
	OpenSquare
	CloseSquare
	Dot
	Space
	TypeName
	OpenAngle
	CloseAngle
	Colon
	Comma
)

const indentationSpaces = 4

var reservedWords = []string{
	"func",
	"global",
	"struct",
	"interface",
	"import",
	"nativeimport",
	"nativefunc",
	"nativestruct",
	"method",
	"foreach",
	"go",
	"typeswitch",
	"case",
	"default",
	"break",
	"continue",
	"forinc",
	"fordec",
	"if",
	"else",
	"elseif",
	"while",
	"foreach",
	"return",
	"as",
	"locals",
	"localfunc",
	"select",
	"sending",
	"rcving",
	"asinc",
	"asdec",
	"asadd",
	"assub",
	"asmul",
	"asdiv",
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
	"append",
	"push",
	"slice",
	"ref",
	"dr",
	"or",
	"and",
	"print",
	"println",
	"prompt",
	"concat",
	"len",
	"istype",
	"send",
	"rcv",
	"band", // bitwise and
	"bor",  // bitwise or
	"bxor", // bitwise xor
	"bnot", // bitwise not
	"randInt",
	"randIntN",
	"randFloat",
	"parseInt",
	"parseFloat",
	"formatInt",
	"formatFloat",
	"timeNow",
	"formatTime",
	"getchar",
	"getrune",
	"charlist",
	"runelist",
	"charlist",
	"charslice",
}

var builtinTypes = []string{
	"I",
	"F",
	"Fn",
	"Str",
	"Bool",
	"A",   // array
	"S",   // slice
	"Ch",  // channel
	"L",   // list
	"M",   // map
	"P",   // pointer
	"Err", // error
	"Type",
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
	// Type() DataType
}

type Definition interface {
	Definition()
	Line() int
}

type DataType interface {
	DataType()
}

type ParsedDataType struct {
	LineNumber  int
	Type        string
	Params      []ParsedDataType
	ReturnTypes []ParsedDataType // non-nil only for functions with return types
}

type BuiltinType struct {
	Name   string
	Params []DataType
}

type ArrayType struct {
	Size int
	Type DataType
}

type FunctionType struct {
	Params      []DataType
	ReturnTypes []DataType
}

type SelectClause interface {
	SelectClause()
}

func (t Token) Expression()          {}
func (t FunctionCall) Expression()   {}
func (t Operation) Expression()      {}
func (t TypeExpression) Expression() {}
func (t MethodCall) Expression()     {}
func (t ParsedDataType) Expression() {}

func (t FunctionDefinition) Definition()     {}
func (t GlobalDefinition) Definition()       {}
func (t ImportDefinition) Definition()       {}
func (t NativeImportDefinition) Definition() {}
func (t StructDefinition) Definition()       {}
func (t InterfaceDefinition) Definition()    {}
func (t MethodDefinition) Definition()       {}

func (t LocalsStatement) Statement()     {}
func (t LocalFuncStatement) Statement()  {}
func (t IfStatement) Statement()         {}
func (t WhileStatement) Statement()      {}
func (t ForeachStatement) Statement()    {}
func (t ForincStatement) Statement()     {}
func (t AssignmentStatement) Statement() {}
func (t ReturnStatement) Statement()     {}
func (t FunctionCall) Statement()        {}
func (t MethodCall) Statement()          {}
func (t Operation) Statement()           {}
func (t TypeswitchStatement) Statement() {}
func (t BreakStatement) Statement()      {}
func (t ContinueStatement) Statement()   {}
func (t GoStatement) Statement()         {}
func (t SelectStatement) Statement()     {}

func (t InterfaceDefinition) DataType() {}
func (t StructDefinition) DataType()    {}
func (t BuiltinType) DataType()         {}
func (t FunctionType) DataType()        {}
func (t Struct) DataType()              {}
func (t ArrayType) DataType()           {}

func (s SelectSendClause) SelectClause() {}
func (s SelectRcvClause) SelectClause()  {}

func (t LocalsStatement) Line() int {
	return t.Vars[0].LineNumber
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
	return t.Targets[0].Line()
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
func (t GoStatement) Line() int {
	return t.LineNumber
}
func (t SelectStatement) Line() int {
	return t.LineNumber
}
func (t LocalFuncStatement) Line() int {
	return t.LineNumber
}
func (t Variable) Line() int {
	return t.LineNumber
}
func (t FunctionCall) Line() int {
	return t.LineNumber
}
func (t TypeswitchStatement) Line() int {
	return t.LineNumber
}
func (t ParsedDataType) Line() int {
	return t.LineNumber
}
func (t TypeExpression) Line() int {
	return t.LineNumber
}
func (t MethodCall) Line() int {
	return t.LineNumber
}
func (t Operation) Line() int {
	return t.LineNumber
}

func (t FunctionDefinition) Line() int {
	return t.LineNumber
}

func (t ImportDefinition) Line() int {
	return t.LineNumber
}

func (t NativeImportDefinition) Line() int {
	return t.LineNumber
}

func (t GlobalDefinition) Line() int {
	return t.LineNumber
}

func (t StructDefinition) Line() int {
	return t.LineNumber
}

func (t InterfaceDefinition) Line() int {
	return t.LineNumber
}

func (t Token) Line() int {
	return t.LineNumber
}

func (t MethodDefinition) Line() int {
	return t.LineNumber
}

type FunctionDefinition struct {
	LineNumber  int
	Column      int
	Name        string
	Parameters  []Variable
	ReturnTypes []ParsedDataType
	Body        []Statement
	NativeCode  string // a native function has a string of native code and an empty body
	Pkg         *Package
}

type GlobalDefinition struct {
	LineNumber int
	Column     int
	Name       string
	Value      Expression
	Type       ParsedDataType
	Pkg        *Package
}

type ImportDefinition struct {
	LineNumber int
	Column     int
	Path       string
	Names      []string
	Aliases    []string
	Pkg        *Package
}

type NativeImportDefinition struct {
	LineNumber int
	Column     int
	Path       string
	Alias      string
	Pkg        *Package
}

type StructDefinition struct {
	LineNumber int
	Column     int
	Name       string
	Members    []Variable
	NativeCode string
	Pkg        *Package
}

type Struct struct {
	LineNumber  int
	Column      int
	Name        string
	MemberNames []string
	MemberTypes []DataType
	Implements  map[string]bool // names of the interfaces this struct implements
	Methods     map[string]FunctionType
	NativeCode  string
	Pkg         *Package
}

type MethodDefinition struct {
	LineNumber  int
	Column      int
	Name        string
	Receiver    Variable
	Parameters  []Variable
	ReturnTypes []ParsedDataType
	Body        []Statement
	Pkg         *Package
}

type InterfaceDefinition struct {
	LineNumber int
	Column     int
	Name       string
	Methods    []Signature
	Pkg        *Package
}

type Signature struct {
	LineNumber  int
	Column      int
	Name        string
	ParamTypes  []ParsedDataType
	ReturnTypes []ParsedDataType
}

type Variable struct {
	LineNumber int
	Column     int
	Name       string
	Type       ParsedDataType
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

type TypeExpression struct {
	LineNumber int
	Column     int
	Type       ParsedDataType
	Operands   []Expression
}

type LocalFuncStatement struct {
	LineNumber  int
	Column      int
	Name        string
	Parameters  []Variable
	ReturnTypes []ParsedDataType
	Body        []Statement
}

type TypeswitchStatement struct {
	LineNumber int
	Column     int
	Value      Expression
	Cases      []TypeswitchCase
	Default    []Statement
}

type TypeswitchCase struct {
	LineNumber int
	Column     int
	Variable   Variable
	Body       []Statement
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

type SelectStatement struct {
	LineNumber int
	Column     int
	Clauses    []SelectClause
	Default    SelectDefaultClause
}

type SelectSendClause struct {
	LineNumber int
	Column     int
	Channel    Expression
	Value      Expression
	Body       []Statement
}

type SelectRcvClause struct {
	LineNumber int
	Column     int
	Target     Variable
	Channel    Expression
	Body       []Statement
}

type SelectDefaultClause struct {
	LineNumber int
	Column     int
	Body       []Statement
}

type GoStatement struct {
	LineNumber int
	Column     int
	Call       Expression // FunctionCall or MethodCall
}

type LocalsStatement struct {
	LineNumber int
	Column     int
	Vars       []Variable
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
	IndexType  ParsedDataType
	ValName    string
	ValType    ParsedDataType
	Collection Expression
	Body       []Statement
}

type ForincStatement struct {
	LineNumber int
	Column     int
	IndexName  string
	IndexType  ParsedDataType
	StartVal   Expression
	EndVal     Expression
	Body       []Statement
	Dec        bool
}

type ReturnStatement struct {
	LineNumber int
	Column     int
	Values     []Expression
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
	Targets    []Expression
	Value      Expression
}

type Package struct {
	Globals          map[string]GlobalDefinition
	Types            map[string]DataType // all types of current package and those brought into current namespace
	ImportedTypes    map[DataType]bool   // all types of imported packages (directly and indirectly)
	ValidBreakpoints map[string]bool
	StructDefs       map[string]StructDefinition // parsed form of struct
	Structs          map[string]Struct           // processed form of struct
	Funcs            map[string]FunctionDefinition
	Methods          map[string]map[string]MethodDefinition // method name:, struct name:, def
	Interfaces       map[string]InterfaceDefinition
	FullPath         string
	Prefix           string
	ImportDefs       map[string]ImportDefinition
	ImportedPackages map[string]*Package
	Code             string
	NativeImports    map[string]string
}

func (p *Package) getExportedDefinition(name string) Definition {
	g, ok := p.Globals[name]
	if ok {
		return g
	}
	st, ok := p.StructDefs[name]
	if ok {
		return st
	}
	f, ok := p.Funcs[name]
	if ok {
		return f
	}
	inter, ok := p.Interfaces[name]
	if ok {
		return inter
	}
	return nil
}

func msg(line int, column int, s string) error {
	return errors.New("Line " + strconv.Itoa(line) + ", column " +
		strconv.Itoa(column) + " . " + s)
}

func debug(args ...interface{}) {
	fmt.Print("DEBUG: ")
	fmt.Println(args...)
}
