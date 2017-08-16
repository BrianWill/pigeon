package staticPigeon

// TODO use bytes.Buffer for more efficent string building
//import "bytes"

// we use arbitrary number values to designate each type of token. Rather than using straight ints, we
// create a distinct type to help avoid mistreating these values like ints.
type TokenType int

const (
	// every constant is assigned the same expression, but the value of iota is zero in
	// the first line, then 1 in the second, then 2 in the third, and so forth
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
	Colon
)

const indentationSpaces = 4
const outputDir = "output"

var reservedWords = []string{
	"func",
	"global",
	"struct",
	"interface",
	"import",
	"method",
	"foreach",
	"typeswitch",
	"break",
	"continue",
	"if",
	"while",
	"else",
	"elseif",
	"return",
	"as",
	"locals",
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
	"or",
	"and",
	"print",
	"prompt",
	"concat",
	"list",
	"map",
	"len",
}

var builtinTypes = []string{
	"N",
	"Str",
	"Bool",
	"L",
	"M",
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
	Type        string
	Params      []ParsedDataType
	ReturnTypes []ParsedDataType // non-nil only for functions with return types
}

// N, Str, Bool, L<>, M<>
type BuiltinType struct {
	Name   string
	Params []DataType
}

type FunctionType struct {
	Params      []DataType
	ReturnTypes []DataType
}

func (t Token) Expression() {}

func (t FunctionCall) Expression() {}
func (t Operation) Expression()    {}

func (t FunctionDefinition) Definition()  {}
func (t GlobalDefinition) Definition()    {}
func (t ImportDefinition) Definition()    {}
func (t StructDefinition) Definition()    {}
func (t InterfaceDefinition) Definition() {}
func (t MethodDefinition) Definition()    {}

func (t LocalsStatement) Statement()     {}
func (t IfStatement) Statement()         {}
func (t WhileStatement) Statement()      {}
func (t AssignmentStatement) Statement() {}
func (t ReturnStatement) Statement()     {}
func (t FunctionCall) Statement()        {}
func (t Operation) Statement()           {}
func (t BreakStatement) Statement()      {}
func (t ContinueStatement) Statement()   {}

func (t InterfaceDefinition) DataType() {}
func (t StructDefinition) DataType()    {}
func (t BuiltinType) DataType()         {}
func (t FunctionType) DataType()        {}
func (t Struct) DataType()              {}

func (t LocalsStatement) Line() int {
	return t.Vars[0].LineNumber
}
func (t IfStatement) Line() int {
	return t.Condition.Line()
}
func (t WhileStatement) Line() int {
	return t.Condition.Line()
}
func (t AssignmentStatement) Line() int {
	return t.Target.Line()
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
func (t Variable) Line() int {
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

func (t ImportDefinition) Line() int {
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
}

type GlobalDefinition struct {
	LineNumber int
	Column     int
	Name       string
	Value      Expression
	Type       ParsedDataType
}

type ImportDefinition struct {
	LineNumber int
	Column     int
	Path       string
	Names      []string
	Aliases    []string
}

type StructDefinition struct {
	LineNumber int
	Column     int
	Name       string
	Members    []Variable
}

type Struct struct {
	LineNumber  int
	Column      int
	Name        string
	MemberNames []string
	MemberTypes []DataType
	Implements  []string // names of the interfaces this struct implements
}

type MethodDefinition struct {
	LineNumber  int
	Column      int
	Name        string
	Receiver    Variable
	Parameters  []Variable
	ReturnTypes []ParsedDataType
	Body        []Statement
}

type InterfaceDefinition struct {
	LineNumber int
	Column     int
	Name       string
	Methods    []Signature
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
	Vars       []Variable
}

type WhileStatement struct {
	LineNumber int
	Column     int
	Condition  Expression
	Body       []Statement
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
	Target     Expression
	Value      Expression
}

type CodeContext struct {
	Globals          map[string]GlobalDefinition
	Locals           map[string]Variable
	Types            map[string]DataType
	FuncTypes        map[string]FunctionType
	ValidBreakpoints map[string]bool
}
