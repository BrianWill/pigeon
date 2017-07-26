/*
Read in the file 'example.pigeon' and lex (tokenize) its content.


If the file contains tab characters or any non-ASCII character, the lexer
returns an error. (Only spaces are allowed for indentation.)


The lexer does not produce tokens for spaces, but the lexer returns an error if spaces
are improperly missing, e.g. foo"hi" should have a space between foo and "hi".

Every line ends with a newline token.

Every line starts with an indentation token representing the spaces at the start of a line.
An unindented line starts with an indentation token with an empty string for its content. (This will make parsing a bit easier.)


For example, this line of Pigeon:

function david a b c

...is represented as seven tokens:

	Indentation ("")
	ReservedWord ("function")
	Identifier ("david")
	Identifier ("a")
	Identifier ("b")
	Identifier ("c")
	Newline ("\n")


The last line of the input file will not necessarily end with a newline, but add a newline token at the end anyway.

*/

package dynamicPigeon

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"unicode"

	highlight "github.com/BrianWill/pigeon/syntaxHighlighter"
)

// TODO use bytes.Buffer for more efficent string building
//import "bytes"

// we use arbitrary number values to designate each type of token. Rather than using straight ints, we
// create a distinct type to help avoid mistreating these values like ints.
type TokenType int

type Scope map[string]bool // set of variable names declared in scope

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
)

const indentationSpaces = 4
const outputDir = "output"

var reservedWords = []string{
	"func",
	"global",
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
	"_fmt",
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
	"id",
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
func (t AssignmentStatement) Statement() {}
func (t ReturnStatement) Statement()     {}
func (t FunctionCall) Statement()        {}
func (t Operation) Statement()           {}

func (t LocalsStatement) Line() int {
	return t.Names[0].LineNumber
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
	return t.FirstToken.LineNumber
}
func (t Token) Line() int {
	return t.LineNumber
}
func (t FunctionCall) Line() int {
	return t.LineNumber
}
func (t Operation) Line() int {
	return t.Operator.LineNumber
}

func (t FunctionDefinition) Line() int {
	return t.FirstToken.LineNumber
}

func (t GlobalDefinition) Line() int {
	return t.Name.LineNumber
}

type FunctionDefinition struct {
	FirstToken Token
	Name       Token
	Parameters []Token
	Body       []Statement
}

type GlobalDefinition struct {
	Name  Token // identifier
	Value Expression
}

type FunctionCall struct {
	LineNumber int
	Function   Expression // either an identifier or another function/operator call
	Arguments  []Expression
}

type Operation struct {
	Operator Token
	Operands []Expression
}

type IfStatement struct {
	FirstToken Token
	Condition  Expression
	Body       []Statement
	Elifs      []ElseifClause
	Else       ElseClause
}

type ElseifClause struct {
	FirstToken Token
	Condition  Expression
	Body       []Statement
}

type ElseClause struct {
	FirstToken Token
	Body       []Statement
}

type LocalsStatement struct {
	Names []Token
}

type WhileStatement struct {
	FirstToken Token
	Condition  Expression
	Body       []Statement
}

type ReturnStatement struct {
	FirstToken Token
	Value      Expression
}

type AssignmentStatement struct {
	Target Expression
	Value  Expression
}

// returns true if rune is a letter of the English alphabet
func isAlpha(r rune) bool {
	return (r >= 65 && r <= 90) || (r >= 97 && r <= 122)
}

// returns true if rune is a numeral
func isNumeral(r rune) bool {
	return (r >= 48 && r <= 57)
}

// assumes the string ends with a newline (because that makes it a bit easier to lex)
func lex(text string) ([]Token, error) {
	var tokens []Token

	line := 1
	column := 1
	runes := []rune(text) // to account for unicode properly, we need to iterate through runes, not bytes

	for i := 0; i < len(runes); {
		r := runes[i]
		if r >= 128 {
			return nil, errors.New("File improperly contains a non-ASCII character at line " + strconv.Itoa(line) + " and column " + strconv.Itoa(column))
		}
		if r == '\n' {
			tokens = append(tokens, Token{Newline, "\n", line, column})
			line++
			column = 1
			i++
		} else if r == '/' { // start of a comment
			if runes[i+1] != '/' {
				return nil, errors.New("Expected second / on line " + strconv.Itoa(line) + " and column " + strconv.Itoa(column))
			}
			for runes[i] != '\n' {
				i++
			}
			i++
			if len(tokens) > 1 && tokens[len(tokens)-1].Type != Newline {
				tokens = append(tokens, Token{Newline, "\n", line, column})
			}
			line++
			column = 1
		} else if r == '(' {
			tokens = append(tokens, Token{OpenParen, "(", line, column})
			column++
			i++
		} else if r == ')' {
			tokens = append(tokens, Token{CloseParen, ")", line, column})
			column++
			i++
		} else if r == '[' {
			tokens = append(tokens, Token{OpenSquare, "[", line, column})
			column++
			i++
		} else if r == ']' {
			tokens = append(tokens, Token{CloseSquare, "]", line, column})
			column++
			i++
		} else if r == '.' {
			tokens = append(tokens, Token{Dot, ".", line, column})
			column++
			i++
		} else if r == ' ' {
			tokenType := Space
			if i > 0 && runes[i-1] == '\n' {
				tokenType = Indentation
			}
			firstIdx := i
			for i < len(runes) {
				r = runes[i]
				if r != ' ' {
					break
				}
				i++
				column++
			}
			tokens = append(tokens, Token{tokenType, string(runes[firstIdx:i]), line, column})
		} else if r == '\t' {
			return nil, errors.New("File improperly contains a tab character at line " + strconv.Itoa(line) + " and column " + strconv.Itoa(column))
		} else if r == '"' { // start of a string
			prev := r
			endIdx := i + 1
			for {
				current := runes[endIdx]
				// loop will never run past end of runes because \n appended to end of file
				if current == '\n' {
					return nil, errors.New("String literal not closed on its line at line " + strconv.Itoa(line) + " and column " + strconv.Itoa(column))
				}
				if current == '"' && prev != '\\' { // end of the string
					endIdx++
					break
				}
				prev = current
				endIdx++
			}

			tokens = append(tokens, Token{StringLiteral, string(runes[i:endIdx]), line, column})
			column += (endIdx - i)
			i = endIdx
		} else if isNumeral(r) || r == '-' { // start of a number
			decimalPointIdx := -1
			endIdx := i + 1
			for {
				current := runes[endIdx]
				// loop will never run past end of runes because \n appended to end of file
				// A number literal should always end with space, newline, or )
				if strings.Contains(" \n)]", string(current)) {
					break
				} else if current == '.' {
					if decimalPointIdx != -1 {
						return nil, errors.New("Number literal has more than one decimal point at line " + strconv.Itoa(line) + " and column " + strconv.Itoa(column))
					}
					decimalPointIdx = endIdx
				} else if !isNumeral(current) {
					return nil, errors.New("Number literal not properly formed at line " + strconv.Itoa(line) + " and column " + strconv.Itoa(column))
				}
				endIdx++
			}

			if decimalPointIdx == endIdx {
				return nil, errors.New("Number literal should not end with decimal point at line " + strconv.Itoa(line) + " and column " + strconv.Itoa(column))
			}

			tokens = append(tokens, Token{NumberLiteral, string(runes[i:endIdx]), line, column})
			column += (endIdx - i)
			i = endIdx
		} else if isAlpha(r) || r == '_' { // start of a word
			endIdx := i + 1
			for {
				current := runes[endIdx]
				// loop will never run past end of runes because \n appended to end of file
				// A word should always end with space, newline, or )
				if strings.Contains(" \n).[", string(current)) {
					break
				} else if !(isAlpha(current) || current == '_' || isNumeral(current)) {
					return nil, errors.New("Word improperly formed at line " + strconv.Itoa(line) + " and column " + strconv.Itoa(column))
				}
				endIdx++
			}

			content := string(runes[i:endIdx])

			// determine if token is ReservedWord, Operator, or Identifier
			tokenType := IdentifierWord
			for _, word := range reservedWords {
				if content == word {
					tokenType = ReservedWord
					break
				}
			}
			if tokenType == IdentifierWord {
				for _, word := range operators {
					if content == word {
						tokenType = OperatorWord
						break
					}
				}
			}
			if content == "true" || content == "false" {
				tokenType = BooleanLiteral
			}
			if content == "nil" {
				tokenType = NilLiteral
			}

			tokens = append(tokens, Token{tokenType, content, line, column})
			column += (endIdx - i)
			i = endIdx
		} else {
			return nil, errors.New("Unexpected character " + string(r) + " at line " + strconv.Itoa(line) + ", column " + strconv.Itoa(column))
		}
	}
	return tokens, nil
}

// parse the top-level statements
// (for this purpose, a top-level function definition is considered a statement)
func parse(tokens []Token) ([]Definition, error) {
	var definitions []Definition
	for i := 0; i < len(tokens); {
		t := tokens[i]
		switch t.Type {
		case ReservedWord:
			var definition Definition
			var numTokens int
			var err error
			switch t.Content {
			case "func":
				definition, numTokens, err = parseFunction(tokens[i:])
			case "global":
				definition, numTokens, err = parseGlobal(tokens[i:])
			default:
				return nil, errors.New("Improper reserved word at top level of code: line " + strconv.Itoa(t.LineNumber) + " column: " + strconv.Itoa(t.Column))
			}
			if err != nil {
				return nil, err
			}
			definitions = append(definitions, definition)
			i += numTokens
		case Newline:
			// a blank line
			i++
		case Indentation:
			// only OK at top level if line is blank
			// (don't need to check if (i + 1) in bounds because we know token stream always
			// ends with newline and so this indentation token can't be last)
			if tokens[i+1].Type != Newline {
				return nil, errors.New("Improper indentation at top level of code: line " + strconv.Itoa(t.LineNumber) + " column: " + strconv.Itoa(t.Column))
			}
		default:
			return nil, errors.New("Improper token at top level of code: line " + strconv.Itoa(t.LineNumber) + " column: " + strconv.Itoa(t.Column))
		}
	}
	return definitions, nil
}

func parseGlobal(tokens []Token) (GlobalDefinition, int, error) {
	if len(tokens) < 5 {
		return GlobalDefinition{}, 0, errors.New("Improper global statement on line " + strconv.Itoa(tokens[0].LineNumber))
	}
	if tokens[1].Type != Space {
		return GlobalDefinition{}, 0, errors.New("Expected space on line " + strconv.Itoa(tokens[0].LineNumber))
	}
	target := tokens[2]
	if target.Type != IdentifierWord {
		return GlobalDefinition{}, 0, errors.New("Improper name for a global on line " + strconv.Itoa(tokens[0].LineNumber))
	}
	if tokens[3].Type != Space {
		return GlobalDefinition{}, 0, errors.New("Expected space on line " + strconv.Itoa(tokens[0].LineNumber))
	}
	var value Expression
	var numValueTokens int
	t := tokens[4]
	switch t.Type {
	case IdentifierWord, StringLiteral, NumberLiteral, BooleanLiteral, NilLiteral:
		value = t
		numValueTokens = 1
	case OpenParen:
		var err error
		value, numValueTokens, err = parseOpenParen(tokens[4:])
		if err != nil {
			return GlobalDefinition{}, 0, err
		}
	default:
		return GlobalDefinition{}, 0, errors.New("Improper value for global on line " + strconv.Itoa(tokens[0].LineNumber))
	}

	if tokens[4+numValueTokens].Type != Newline {
		return GlobalDefinition{}, 0, errors.New("Global not terminated with newline on line " + strconv.Itoa(tokens[0].LineNumber))
	}
	return GlobalDefinition{target, value}, 5 + numValueTokens, nil
}

func parseExpression(tokens []Token, line int) (Expression, int, error) {
	lineStr := strconv.Itoa(line)
	if len(tokens) < 1 {
		return nil, 0, errors.New("Missing expression on line " + lineStr)
	}
	idx := 0
	token := tokens[idx]
	var expr Expression
	switch token.Type {
	case StringLiteral, NumberLiteral, BooleanLiteral, NilLiteral:
		return token, 1, nil
	case IdentifierWord:
		expr = token
		idx++
	case OpenParen:
		var err error
		expr, idx, err = parseOpenParen(tokens)
		if err != nil {
			return nil, 0, err
		}
	default:
		return nil, 0, errors.New("Improper expression on line " + lineStr +
			": " + fmt.Sprintf("%#v", token))
	}

Loop:
	for len(tokens) > idx {
		var err error
		var n int
		switch tokens[idx].Type {
		case Dot:
			expr, n, err = parseDot(tokens[idx:], token, line)
		case OpenSquare:
			expr, n, err = parseOpenSquare(tokens[idx:], token, line)
		default:
			break Loop
		}
		if err != nil {
			return nil, 0, err
		}
		idx += n
	}
	return expr, idx, nil
}

// assumes first token is dot
func parseDot(tokens []Token, expr Expression, line int) (Expression, int, error) {
	if len(tokens) < 2 {
		return nil, 0, errors.New("Improperly formed dot operation on line " + strconv.Itoa(line))
	}
	if tokens[1].Type != IdentifierWord {
		return nil, 0, errors.New("Identifier expected after dot line " + strconv.Itoa(line))
	}
	strLiteral := Token{StringLiteral, "\"" + tokens[1].Content + "\"", line, -1}
	getOp := Operation{
		Token{OperatorWord, "get", line, -1},
		[]Expression{expr, strLiteral},
	}
	return getOp, 2, nil
}

// assumes first token is open square
func parseOpenSquare(tokens []Token, expr Expression, line int) (Expression, int, error) {
	if len(tokens) < 3 {
		return nil, 0, errors.New("Improperly formed square brackets on line " + strconv.Itoa(line))
	}
	idx := 1
	if tokens[idx].Type == Space {
		idx++
	}
	indexExpr, nTokens, err := parseExpression(tokens[idx:], line)
	if err != nil {
		return nil, 0, err
	}
	idx += nTokens
	if tokens[idx].Type == Space {
		idx++
	}
	if len(tokens) < idx || tokens[idx].Type != CloseSquare {
		return nil, 0, errors.New("Improperly formed square brackets on line " + strconv.Itoa(line))
	}
	idx++ // account for ']'
	getOp := Operation{
		Token{OperatorWord, "get", line, -1},
		[]Expression{expr, indexExpr},
	}
	return getOp, idx, nil
}

// assumes first token is open paren.
// Returns a FunctionCall or Operation and the number of tokens that make up the Expression.
func parseOpenParen(tokens []Token) (Expression, int, error) {
	line := strconv.Itoa(tokens[0].LineNumber)
	if len(tokens) < 3 {
		return nil, 0, errors.New("Improper function call or operation on line " + line)
	}

	idx := 1
	if tokens[idx].Type == Space {
		idx++
	}

	functionCall := true
	var leadingCall Expression
	var op Token
	t := tokens[idx]
	switch t.Type {
	case OperatorWord:
		op = t
		functionCall = false
		idx++
	case IdentifierWord:
		op = t
		idx++
	case OpenParen:
		var numTokens int
		var err error
		leadingCall, numTokens, err = parseOpenParen(tokens[idx:])
		if err != nil {
			return nil, 0, err
		}
		idx += numTokens
	default:
		return nil, 0, errors.New("Improper function call or operation on line " + line)
	}

	var arguments []Expression
Loop:
	for true {
		t := tokens[idx]
		switch t.Type {
		case Space:
			idx++
		case CloseParen:
			idx++
			break Loop
		}
		expr, numTokens, err := parseExpression(tokens[idx:], tokens[0].LineNumber)
		if err != nil {
			return nil, 0, err
		}
		arguments = append(arguments, expr)
		idx += numTokens
	}

	var expr Expression
	if functionCall {
		if leadingCall == nil {
			expr = FunctionCall{tokens[0].LineNumber, op, arguments}
		} else {
			expr = FunctionCall{tokens[0].LineNumber, leadingCall, arguments}
		}
	} else {
		expr = Operation{op, arguments}
	}

Outer:
	for len(tokens) > idx {
		line := tokens[idx].LineNumber
		var err error
		var n int
		switch tokens[idx].Type {
		case Dot:
			expr, n, err = parseDot(tokens[idx:], expr, line)
		case OpenSquare:
			expr, n, err = parseOpenSquare(tokens[idx:], expr, line)
			if err != nil {
				return nil, 0, err
			}
			idx += n
		default:
			break Outer
		}
		if err != nil {
			return nil, 0, err
		}
		idx += n
	}

	return expr, idx, nil
}

func debug(args ...interface{}) {
	fmt.Print("DEBUG: ")
	fmt.Println(args...)
}

func parseFunction(tokens []Token) (FunctionDefinition, int, error) {
	if len(tokens) < 5 {
		return FunctionDefinition{}, 0, errors.New("Improper function definition on line " + strconv.Itoa(tokens[0].LineNumber))
	}
	idx := 1
	if tokens[idx].Type == Space {
		idx++
	}
	name := tokens[idx]
	if name.Type != IdentifierWord {
		return FunctionDefinition{}, 0, errors.New("Function missing name on line " + strconv.Itoa(name.LineNumber))
	}
	if name.Content == "main" {
		name.Content = "_main"
	}
	idx++
	if tokens[idx].Type == Space {
		idx++
	}
	var parameters []Token
	for idx < len(tokens) {
		token := tokens[idx]
		if token.Type == IdentifierWord {
			parameters = append(parameters, token)
			idx++
		} else if token.Type == Space {
			idx++
		} else {
			break
		}
	}
	if idx >= len(tokens) || tokens[idx].Type != Newline {
		return FunctionDefinition{}, 0, errors.New("Improper function definition on line " + strconv.Itoa(tokens[0].LineNumber))
	}
	idx++
	body, numTokens, err := parseBody(tokens[idx:], indentationSpaces)
	if err != nil {
		return FunctionDefinition{}, 0, err
	}
	return FunctionDefinition{tokens[0], name, parameters, body}, idx + numTokens, nil
}

// 'indentation' = number of spaces before 'if'
func parseIf(tokens []Token, indentation int) (IfStatement, int, error) {
	line := strconv.Itoa(tokens[0].LineNumber)
	if len(tokens) < 6 {
		return IfStatement{}, 0, errors.New("Improper if statement on line " + line)
	}
	idx := 1
	if tokens[idx].Type != Space {
		return IfStatement{}, 0, errors.New("Missing space on line " + line)
	}
	idx++
	condition, numConditionTokens, err := parseExpression(tokens[idx:], tokens[0].LineNumber)
	if err != nil {
		return IfStatement{}, 0, err
	}
	idx += numConditionTokens
	if tokens[idx].Type == Space {
		idx++
	}
	if tokens[idx].Type != Newline {
		return IfStatement{}, 0, errors.New("If statement condition not followed by newline on line " + line)
	}
	idx++
	body, numTokens, err := parseBody(tokens[idx:], indentation+indentationSpaces)
	if err != nil {
		return IfStatement{}, 0, err
	}
	idx += numTokens

	var elseifClauses []ElseifClause
	var elseClause ElseClause

	for idx+1 < len(tokens) {
		if tokens[idx].Type == Indentation && len(tokens[idx].Content) == indentation && tokens[idx+1].Content == "elseif" {
			elseifClause, numTokens, err := parseElif(tokens[idx+1:], indentation)
			if err != nil {
				return IfStatement{}, 0, err
			}
			elseifClauses = append(elseifClauses, elseifClause)
			idx += numTokens + 1 // +1 for the indentation before this elif
		} else {
			break
		}
	}

	if idx+1 < len(tokens) {
		if tokens[idx].Type == Indentation && len(tokens[idx].Content) == indentation && tokens[idx+1].Content == "else" {
			var numTokens int
			var err error
			elseClause, numTokens, err = parseElse(tokens[idx+1:], indentation)
			if err != nil {
				return IfStatement{}, 0, err
			}
			idx += numTokens + 1 // +1 for the indentation before this else
		}
	}
	return IfStatement{tokens[0], condition, body, elseifClauses, elseClause}, idx, nil
}

func parseElif(tokens []Token, indentation int) (ElseifClause, int, error) {
	line := strconv.Itoa(tokens[0].LineNumber)
	if len(tokens) < 5 {
		return ElseifClause{}, 0, errors.New("Improper elif clause on line " + line)
	}
	idx := 1
	if tokens[idx].Type != Space {
		return ElseifClause{}, 0, errors.New("Missing space on line " + line)
	}
	idx++
	condition, numConditionTokens, err := parseExpression(tokens[idx:], tokens[0].LineNumber)
	if err != nil {
		return ElseifClause{}, 0, errors.New("Improper condition in if statement on line " + line)
	}
	idx += numConditionTokens
	if tokens[idx].Type == Space {
		idx++
	}
	if tokens[idx].Type != Newline {
		return ElseifClause{}, 0, errors.New("Elseif clause condition not followed by newline on line " + line)
	}
	idx++
	body, numTokens, err := parseBody(tokens[idx:], indentation+indentationSpaces)
	if err != nil {
		return ElseifClause{}, 0, err
	}
	idx += numTokens
	return ElseifClause{tokens[0], condition, body}, idx, nil
}

func parseElse(tokens []Token, indentation int) (ElseClause, int, error) {
	line := strconv.Itoa(tokens[0].LineNumber)
	if len(tokens) < 4 {
		return ElseClause{}, 0, errors.New("Improper else clause on line " + line)
	}
	idx := 1
	if tokens[idx].Type == Space {
		idx++
	}
	if tokens[idx].Type != Newline {
		return ElseClause{}, 0, errors.New("Elif clause condition not followed by newline on line " + line)
	}
	body, numTokens, err := parseBody(tokens[idx:], indentation+indentationSpaces)
	if err != nil {
		return ElseClause{}, 0, err
	}
	idx += numTokens
	return ElseClause{tokens[0], body}, idx, nil
}

func parseWhile(tokens []Token, indentation int) (WhileStatement, int, error) {
	line := strconv.Itoa(tokens[0].LineNumber)
	if len(tokens) < 5 {
		return WhileStatement{}, 0, errors.New("Improper while statement on line " + line)
	}
	idx := 1
	if tokens[idx].Type != Space {
		return WhileStatement{}, 0, errors.New("Missing space on line " + line)
	}
	idx++
	var condition Expression
	var numConditionTokens int
	switch tokens[idx].Type {
	case IdentifierWord, StringLiteral, NumberLiteral, BooleanLiteral, NilLiteral:
		condition = tokens[idx]
		numConditionTokens = 1
	case OpenParen:
		var err error
		condition, numConditionTokens, err = parseOpenParen(tokens[idx:])
		if err != nil {
			return WhileStatement{}, 0, err
		}
	default:
		return WhileStatement{}, 0, errors.New("Improper condition in while statement on line " + line)
	}
	idx += numConditionTokens
	if tokens[idx].Type != Newline {
		return WhileStatement{}, 0, errors.New("While statement condition not followed by newline on line " + line)
	}
	idx++
	body, numTokens, err := parseBody(tokens[idx:], indentation+indentationSpaces)
	if err != nil {
		return WhileStatement{}, 0, err
	}
	idx += numTokens
	return WhileStatement{tokens[0], condition, body}, idx, nil
}

func parseReturn(tokens []Token) (ReturnStatement, int, error) {
	line := strconv.Itoa(tokens[0].LineNumber)
	if len(tokens) < 3 {
		return ReturnStatement{}, 0, errors.New("Improper return statement on line " + line)
	}
	idx := 1
	if tokens[idx].Type != Space {
		return ReturnStatement{}, 0, errors.New("Missing space on line " + line)
	}
	idx++
	value, nTokens, err := parseExpression(tokens[idx:], tokens[0].LineNumber)
	if err != nil {
		return ReturnStatement{}, 0, err
	}
	idx += nTokens
	if tokens[idx].Type != Newline {
		return ReturnStatement{}, 0, errors.New("Return statement not terminated with newline on line " + line)
	}

	return ReturnStatement{tokens[0], value}, idx, nil
}

// assume first token is reserved word "as"
// returns number of tokens (including the newline at the end)
func parseAssignment(tokens []Token) (AssignmentStatement, int, error) {
	line := strconv.Itoa(tokens[0].LineNumber)
	if len(tokens) < 4 {
		return AssignmentStatement{}, 0, errors.New("Improper assignment statement on line " + line)
	}
	idx := 1
	if tokens[idx].Type != Space {
		return AssignmentStatement{}, 0, errors.New("Missing space on line " + line)
	}
	idx++
	target, nTokens, err := parseExpression(tokens[idx:], tokens[0].LineNumber)
	if err != nil {
		return AssignmentStatement{}, 0, err
	}
	idx += nTokens
	if tokens[idx].Type != Space {
		return AssignmentStatement{}, 0, errors.New("Missing space on line " + line)
	}
	idx++
	value, nTokens, err := parseExpression(tokens[idx:], tokens[0].LineNumber)
	if err != nil {
		return AssignmentStatement{}, 0, err
	}
	idx += nTokens
	if tokens[idx].Type == Space {
		idx++
	}
	if tokens[idx].Type != Newline {
		return AssignmentStatement{}, 0, errors.New("Assignment not terminated with newline on line " + line)
	}
	idx++
	return AssignmentStatement{target, value}, idx, nil // 3 because: 'as', the target, and the newline at the end
}

func parseLocals(tokens []Token) (LocalsStatement, int, error) {
	line := strconv.Itoa(tokens[0].LineNumber)
	if len(tokens) < 4 {
		return LocalsStatement{}, 0, errors.New("Improper locals statement on line " + line)
	}

	idx := 1

	var locals []Token
	for idx < len(tokens) {
		token := tokens[idx]
		if token.Type == IdentifierWord {
			locals = append(locals, token)
			idx++
		} else if token.Type == Space {
			idx++
		} else {
			break
		}
	}

	if idx >= len(tokens) || tokens[idx].Type != Newline {
		return LocalsStatement{}, 0, errors.New("Improper locals statement on line " + line)
	}
	idx++

	return LocalsStatement{locals}, idx, nil
}

// expected to start with Indentation token.
// 'indentation' = the number of spaces indentation on which the body should be aligned
// May return zero statements if body is empty.
func parseBody(tokens []Token, indentation int) ([]Statement, int, error) {
	var statements []Statement
	i := 0
	for i < len(tokens) {
		t := tokens[i]
		if t.Type == Newline { // blank line
			i++
		} else if t.Type == Indentation && tokens[i+1].Type == Newline { // blank line
			i += 2
		} else if t.Type != Indentation { // gone past end of the body
			break
		} else {
			numSpaces := len(t.Content)
			if numSpaces < indentation { // gone past end of the body
				break
			} else if numSpaces == indentation {
				i++
				t = tokens[i]
				var statement Statement
				var numTokens int
				var err error
				switch t.Type {
				case ReservedWord:
					switch t.Content {
					case "func":
						return nil, 0, errors.New("Functions cannot be nested: line " + strconv.Itoa(t.LineNumber) + " column: " + strconv.Itoa(t.Column))
					case "as":
						statement, numTokens, err = parseAssignment(tokens[i:])
					case "if":
						statement, numTokens, err = parseIf(tokens[i:], indentation)
					case "while":
						statement, numTokens, err = parseWhile(tokens[i:], indentation)
					case "locals":
						statement, numTokens, err = parseLocals(tokens[i:])
					case "return":
						statement, numTokens, err = parseReturn(tokens[i:])
					default:
						return nil, 0, errors.New("Improper reserved word '" + t.Content + "' in body: line " + strconv.Itoa(t.LineNumber) + " column: " + strconv.Itoa(t.Column))
					}
					if err != nil {
						return nil, 0, err
					}
				case OpenParen:
					var expression Expression
					expression, numTokens, err = parseOpenParen(tokens[i:])
					if err != nil {
						return nil, 0, err
					}
					statement = expression.(Statement)

					if tokens[i+numTokens].Type != Newline {
						return nil, 0, errors.New("Statement not terminated with newline on line " + strconv.Itoa(t.LineNumber))
					}

					numTokens++ // add in the newline
				default:
					return nil, 0, errors.New("Improper token. Expected start of statement: line " + strconv.Itoa(t.LineNumber) + " column: " + strconv.Itoa(t.Column))
				}
				statements = append(statements, statement)
				i += numTokens
			} else {
				return nil, 0, errors.New("Improper indentation: line " + strconv.Itoa(t.LineNumber))
			}
		}
	}
	return statements, i, nil
}

/* All identifiers get prefixed with _ to avoid collisions with Go reserved words and predefined identifiers */
func compile(definitions []Definition) (string, error) {
	globals := make(Scope)
	globalsDone := false
	code := `package main

import _p "github.com/BrianWill/pigeon/stdlib"
import _fmt "fmt"

var _breakpoints = make(map[int]bool)

`
	validBreakpoints := make(map[int]bool)
	// TODO check for duplicate global and function names
	for _, def := range definitions {
		switch d := def.(type) {
		case GlobalDefinition:
			if globalsDone {
				return "", errors.New("All globals must be defined before all functions")
			}
			name := d.Name.Content
			c, err := compileExpression(d.Value, make(Scope), globals)
			if err != nil {
				return "", err
			}
			code += "var g_" + name + " interface{} = " + c + "\n"
			globals[name] = true
		case FunctionDefinition:
			globalsDone = true
			c, err := compileFunc(d, globals, validBreakpoints)
			if err != nil {
				return "", err
			}
			code += c
		default:
			return "", errors.New("Unrecognized definition")
		}
	}
	breakpointLinenums := make([]string, len(validBreakpoints))
	i := 0
	for k := range validBreakpoints {
		breakpointLinenums[i] = fmt.Sprintf("%d: true", k)
		i++
	}
	code += "var _validBreakpoints = map[int]bool{" + strings.Join(breakpointLinenums, ", ") + "}"
	code += `
	
func main() {
	go _p.Server(_breakpoints, _validBreakpoints)
	_main()
}	
`

	return code, nil
}

// returns code snippet ending with '\n\n'
func compileFunc(fn FunctionDefinition, globals Scope, validBreakpoints map[int]bool) (string, error) {
	locals := make(Scope)

	header := "func " + fn.Name.Content + "("
	for _, param := range fn.Parameters {
		header += param.Content + " interface{}, "
		locals[param.Content] = true
	}
	if len(fn.Parameters) > 0 {
		header = header[:len(header)-2] // drop last comma and space
	}
	header += ") interface{} {\n"
	if len(fn.Body) < 1 {
		return "", errors.New("Function should contain at least one statement.")
	}
	bodyStatements := fn.Body
	if localsStatement, ok := bodyStatements[0].(LocalsStatement); ok {
		localsStr := "var "
		nullOp := "_p.NullOp(" // supresses unused variable compile errors
		for _, name := range localsStatement.Names {
			if locals[name.Content] {
				return "", fmt.Errorf("Local variable %s on line %d is already defined as a parameter.",
					name.Content, name.LineNumber)
			}
			locals[name.Content] = true
			localsStr += name.Content + ", "
			nullOp += name.Content + ", "
		}
		localsStr = localsStr[:len(localsStr)-2] // hack off last comma
		nullOp = nullOp[:len(nullOp)-2]
		header += localsStr + " interface{}\n" + nullOp + ")\n"
		bodyStatements = bodyStatements[1:]
	}
	header += genDebugFn(globals, locals)
	body, err := compileBody(bodyStatements, locals, globals, validBreakpoints)
	if err != nil {
		return "", err
	}
	if len(fn.Body) > 0 {
		_, lastIsReturn := fn.Body[len(fn.Body)-1].(ReturnStatement)
		if lastIsReturn {
			return header + body + "}\n", nil
		}
	}
	return header + body + "return nil\n}\n", nil
}

func genDebugFn(globals, locals Scope) string {
	s := `debug := func(line int) {
	_fmt.Printf("About to execute line %v.\n", line)
	// print globals
	_fmt.Println("Globals:")
`
	for k := range globals {
		s += fmt.Sprintf("_fmt.Println(\"%s\", g_%s)\n", k, k)
	}
	s += "_fmt.Println(\"Locals:\")\n"
	for k := range locals {
		s += fmt.Sprintf("_fmt.Println(\"%s\", %s)\n", k, k)
	}
	s += "\n}\n"
	return s
}

func compileIfStatement(s IfStatement, locals, globals Scope, validBreakpoints map[int]bool) (string, error) {
	c, err := compileExpression(s.Condition, locals, globals)
	if err != nil {
		return "", err
	}
	code := "if " + c + ".(bool)"
	c, err = compileBody(s.Body, locals, globals, validBreakpoints)
	if err != nil {
		return "", nil
	}
	code += " {\n" + c + "}"
	for _, elif := range s.Elifs {
		c, err := compileExpression(elif.Condition, locals, globals)
		if err != nil {
			return "", err
		}
		code += " else if " + c + ".(bool) {\n"
		c, err = compileBody(elif.Body, locals, globals, validBreakpoints)
		if err != nil {
			return "", err
		}
		code += c + "}"
	}

	if len(s.Else.Body) > 0 {
		c, err := compileBody(s.Else.Body, locals, globals, validBreakpoints)
		if err != nil {
			return "", err
		}
		code += " else {\n" + c + "}"
	}
	return code + "\n", nil
}

func compileWhileStatement(s WhileStatement, locals, globals Scope, validBreakpoints map[int]bool) (string, error) {
	c, err := compileExpression(s.Condition, locals, globals)
	if err != nil {
		return "", err
	}
	code := "for " + c + ".(bool) {\n"
	c, err = compileBody(s.Body, locals, globals, validBreakpoints)
	if err != nil {
		return "", err
	}
	return code + c + "}\n", nil
}

func compileBody(statements []Statement, locals, globals Scope, validBreakpoints map[int]bool) (string, error) {
	var code string
	for _, s := range statements {
		line := s.Line()
		validBreakpoints[line] = true
		code += fmt.Sprintf("if _breakpoints[%d] {debug(%d)}\n", line, line)
		var c string
		var err error
		switch s := s.(type) {
		case IfStatement:
			c, err = compileIfStatement(s, locals, globals, validBreakpoints)
		case WhileStatement:
			c, err = compileWhileStatement(s, locals, globals, validBreakpoints)
		case AssignmentStatement:
			c, err = compileAssignmentStatement(s, locals, globals)
		case ReturnStatement:
			c, err = compileReturnStatement(s, locals, globals)
		case FunctionCall:
			c, err = compileFunctionCall(s, locals, globals)
			c += "\n"
		case Operation:
			c, err = compileOperation(s, locals, globals)
			c += "\n"
		}
		if err != nil {
			return "", err
		}
		code += c
	}
	return code, nil
}

func compileAssignmentStatement(s AssignmentStatement, locals, globals Scope) (string, error) {
	switch target := s.Target.(type) {
	case Token:
		if target.Type != IdentifierWord {
			return "", errors.New("Assignment to non-identifier on line " + strconv.Itoa(target.LineNumber))
		}
		name := target.Content
		if !locals[name] && !globals[name] {
			return "", errors.New("Assignment to non-existent variable on line " + strconv.Itoa(target.LineNumber))
		}
		c, err := compileExpression(s.Value, locals, globals)
		if err != nil {
			return "", err
		}
		return target.Content + " = " + c + "\n", nil
	case Operation:
		if target.Operator.Content != "get" {
			return "", errors.New("Improper target of assignment on line " + strconv.Itoa(target.Operator.LineNumber))
		}
		// turn the get op into a set op
		target.Operator.Content = "set"
		target.Operands = append(target.Operands, s.Value)
		c, err := compileExpression(target, locals, globals)
		if err != nil {
			return "", err
		}
		return c + "\n", nil
	case FunctionCall:
		return "", errors.New("Invalid target of assignment on line " + strconv.Itoa(target.LineNumber))
	default:
		// TODO give Expression LineNumber() method so we can get a line number here
		return "", errors.New("Invalid target of assignment.")
	}
}

func compileReturnStatement(s ReturnStatement, locals, globals Scope) (string, error) {
	c, err := compileExpression(s.Value, locals, globals)
	if err != nil {
		return "", err
	}
	return "return " + c + "\n", nil
}

func compileFunctionCall(s FunctionCall, locals, globals Scope) (string, error) {
	var code string
	switch s := s.Function.(type) {
	case Operation:
		c, err := compileOperation(s, locals, globals)
		if err != nil {
			return "", err
		}
		code += c
	case FunctionCall:
		c, err := compileFunctionCall(s, locals, globals)
		if err != nil {
			return "", err
		}
		code += c
		// TODO have to assert type of function
	case Token: // will always be an identifier
		code += s.Content
	}
	code += "(" // start of arguments
	for _, exp := range s.Arguments {
		c, err := compileExpression(exp, locals, globals)
		if err != nil {
			return "", err
		}
		code += c + ", " // Go is OK with comma after last arg, so don't need special case for last arg
	}
	if len(s.Arguments) > 0 {
		code = code[:len(code)-2] // drop last comma and space
	}
	return code + ")", nil
}

func compileOperation(o Operation, locals, globals Scope) (string, error) {
	operator := o.Operator.Content

	runes := []rune(operator)
	runes[0] = unicode.ToUpper(runes[0])
	operator = string(runes)

	code := "_p." + operator + "("
	for _, exp := range o.Operands {
		c, err := compileExpression(exp, locals, globals)
		if err != nil {
			return "", err
		}
		code += c + ", " // Go is OK with comma after last arg, so don't need special case for last arg
	}
	if len(o.Operands) > 0 {
		code = code[:len(code)-2] // drop last comma and space
	}
	code += ")"
	return code, nil
}

func compileExpression(e Expression, locals, globals Scope) (string, error) {
	var code string
	switch e := e.(type) {
	case Operation:
		c, err := compileOperation(e, locals, globals)
		if err != nil {
			return "", err
		}
		code = c
	case FunctionCall:
		c, err := compileFunctionCall(e, locals, globals)
		if err != nil {
			return "", err
		}
		code = c
	case Token:
		switch e.Type {
		case IdentifierWord:
			name := e.Content
			if locals[name] {
				code = name
			} else if globals[name] {
				code = "g_" + name
			} else {
				return "", fmt.Errorf("Name %s on line %d is undefined.", name, e.LineNumber)
			}
		case NumberLiteral:
			code = "float64(" + e.Content + ")"
		case StringLiteral, BooleanLiteral:
			code = e.Content
		case NilLiteral:
			code = "_p.Nil(0)"
		}
	}
	return code, nil
}

func Highlight(code []byte) ([]byte, error) {
	return highlight.AsHTML(code, highlight.OrderedList())
}

func Run(args []string) error {
	if len(os.Args) < 3 {
		fmt.Println("")
		return errors.New("Must specify a file to run.")
	}

	inputFilename := os.Args[2]

	data, err := ioutil.ReadFile(inputFilename)
	if err != nil {
		return err
	}
	tokens, err := lex(string(data) + "\n")
	if err != nil {
		return err
	}

	definitions, err := parse(tokens)
	if err != nil {
		return err
	}

	code, err := compile(definitions)
	if err != nil {
		return err
	}

	outputFilename := outputDir + "/" + inputFilename + ".go"
	err = ioutil.WriteFile(outputFilename, []byte(code), os.ModePerm)
	if err != nil {
		return err
	}

	err = exec.Command("go", "fmt", outputFilename).Run()
	if err != nil {
		return err
	}

	cmd := exec.Command("go", "run", outputFilename)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
