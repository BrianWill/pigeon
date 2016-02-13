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


package main

import "fmt"
import "io/ioutil"
import "errors"
import "strconv"
import "unicode"

// TODO use bytes.Buffer for more efficent string building
//import "bytes"


// we use arbitrary number values to designate each type of token. Rather than using straight ints, we 
// create a distinct type to help avoid mistreating these values like ints.  
type TokenType int

type Scope map[string]bool   // set of variable names declared in scope

const(
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
    NullLiteral
)

const indentationSpaces = 4

var reservedWords = []string{
	"function",
	"if",
	"while",
	"else",
	"elif",
	"return",
	"as",
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
    "list",
    "dict",
}

type Token struct {
	Type TokenType
	Content string    // the token itself, e.g. a number 3.7 is stored here as "3.7"
	LineNumber int   // first line is line 1
	Column int       // first character of a line is in column 1
}


type Statement interface {
    Statement()
}

type Expression interface {
    Expression()
}

func (t Token) Expression() {}
func (t FunctionCall) Expression() {}
func (t Operation) Expression() {}

func (t FunctionDefinition) Statement() {}
func (t IfStatement) Statement() {}
func (t WhileStatement) Statement() {}
func (t AssignmentStatement) Statement() {}
func (t ReturnStatement) Statement() {}
func (t FunctionCall) Statement() {}
func (t Operation) Statement() {}


type FunctionDefinition struct {
    FirstToken Token
    Name Token
    Parameters []Token
    Body []Statement
}

type FunctionCall struct {
    Function Expression    // either an identifier or another function/operator call
    Arguments []Expression
}

type Operation struct {
    Operator Token
    Operands []Expression
}

type IfStatement struct {
    FirstToken Token
    Condition Expression
    Body []Statement
    Elifs []ElifClause
    Else ElseClause
}

type ElifClause struct {
    FirstToken Token
    Condition Expression
    Body []Statement
}

type ElseClause struct {
    FirstToken Token
    Body []Statement
}

type WhileStatement struct {
    FirstToken Token
    Condition Expression
    Body []Statement
}

type ReturnStatement struct {
    FirstToken Token
    Value Expression
}

type AssignmentStatement struct {
    Target Token  // identifier
    Value Expression
}

// returns true if rune is a letter of the English alphabet
func isAlpha(r rune) bool {
	return (r >= 65 && r <= 90) || (r >= 97 && r <= 122)
}

// returns true if ruen is a numeral
func isNumeral(r rune) bool {
	return (r >= 48 && r <= 57)
}

// assumes the string ends with a newline (because that makes it a bit easier to lex)
func lex(text string) ([]Token, error) {
    tokens := make([]Token, 0)

    line := 1
    column := 1
    runes := []rune(text)       // to account for unicode properly, we need to iterate through runes, not bytes
    
    for i := 0; i < len(runes); {
        r := runes[i]
        if r >= 128 {
            return nil, errors.New("File improperly contains a non-ASCII character at line " + strconv.Itoa(line) + " and column " + strconv.Itoa(column))
        }
        if r == '\n' {
            tokens = append(tokens, Token{Newline, "\n", line, column})
            line += 1
            column = 1
            i++
        } else if r == '#' {  // start of a comment
            for runes[i] != '\n' {
                i++
            }
            i++
            line += 1
            column = 1
        } else if r == '(' {
            tokens = append(tokens, Token{OpenParen, "(", line, column})
            column++
            i++
        } else if r == ')' {
            tokens = append(tokens, Token{CloseParen, ")", line, column})
            column++
            i++
        } else if r == ' ' {
            if i > 0 && runes[i - 1] == '\n' {
                firstIdx := i
                for i < len(runes) {
                    r = runes[i]
                    if r != ' ' {
                        break
                    }
                    i++
                    column++
                }
                tokens = append(tokens, Token{Indentation, string(runes[firstIdx:i]), line, column})
            } else {
                column++
                i++
            }
        } else if r == '\t' {
            return nil, errors.New("File improperly contains a tab character at line " + strconv.Itoa(line) + " and column " + strconv.Itoa(column))
        } else if r == '"' {  // start of a string
            prev := r
            endIdx := i + 1
            for {  
                current := runes[endIdx]
                // loop will never run past end of runes because \n appended to end of file
                if current == '\n' {
                    return nil, errors.New("String literal not closed on its line at line " + strconv.Itoa(line) + " and column " + strconv.Itoa(column))
                }
                if current == '"' && prev != '\\' {  // end of the string
                    endIdx++
                    break
                }
                prev = current
                endIdx++
            }

            tokens = append(tokens, Token{StringLiteral, string(runes[i: endIdx]), line, column})
            column += (endIdx - i)
            i = endIdx
        } else if isNumeral(r) || r == '-' {   // start of a number
            decimalPointIdx := -1
            endIdx := i + 1
            for {  
                current := runes[endIdx]
                // loop will never run past end of runes because \n appended to end of file
                // A number literal should always end with space, newline, or )
                if current == ' ' || current == '\n' || current == ')' {  
                    break
                } else if current == '.' {
                    if decimalPointIdx != -1 {
                        return nil, errors.New("Number literal has more than one decimal point at line " + strconv.Itoa(line) + " and column " + strconv.Itoa(column))
                    }
                    decimalPointIdx = endIdx
                } else if !isNumeral(current){
                    return nil, errors.New("Number literal not properly formed at line " + strconv.Itoa(line) + " and column " + strconv.Itoa(column))
                }
                endIdx++
            }

            if decimalPointIdx == endIdx {
                return nil, errors.New("Number literal should not end with decimal point at line " + strconv.Itoa(line) + " and column " + strconv.Itoa(column))
            }

            tokens = append(tokens, Token{NumberLiteral, string(runes[i: endIdx]), line, column})
            column += (endIdx - i)
            i = endIdx
        } else if isAlpha(r) || r == '_' {  // start of a word
            endIdx := i + 1
            for {  
                current := runes[endIdx]
                // loop will never run past end of runes because \n appended to end of file
                // A word should always end with space, newline, or )
                if current == ' ' || current == '\n' || current == ')' {
                    break
                } else if !(isAlpha(current) || current == '_' || isNumeral(current)) {
                    return nil, errors.New("Word not properly formed at line " + strconv.Itoa(line) + " and column " + strconv.Itoa(column))
                }
                endIdx++
            }

            content := string(runes[i: endIdx])

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
            if content == "null" {
                tokenType = NullLiteral
            }

            tokens = append(tokens, Token{tokenType, content, line, column})            
            column += (endIdx - i)
            i = endIdx
        }
    }
    return tokens, nil
}

// parse the top-level statements
// (for this purpose, a top-level function definition is considered a statement)
func parse(tokens []Token) ([]Statement, error) {
    statements := make([]Statement, 0)
    for i := 0; i < len(tokens); {
        t := tokens[i]
        switch t.Type {
        case ReservedWord:
            var statement Statement
            var numTokens int
            var err error
            switch t.Content {
            case "function":
                statement, numTokens, err = parseFunction(tokens[i:])
            case "as":
                statement, numTokens, err = parseAssignment(tokens[i:])
            case "if":
                statement, numTokens, err = parseIf(tokens[i:], 0)
            case "while":
                statement, numTokens, err = parseWhile(tokens[i:], 0)
            default:
                return nil, errors.New("Improper reserved word at top level of code: line " + strconv.Itoa(t.LineNumber) + " column: " + strconv.Itoa(t.Column))
            }
            if err != nil {
                return nil, err
            }
            statements = append(statements, statement)
            i += numTokens
        case Newline:
            // a blank line
            i++
        case Indentation:
            // only OK at top level if line is blank
            // (don't need to check if (i + 1) in bounds because we know token stream always 
            // ends with newline and so this indentation token can't be last)
            if (tokens[i + 1].Type != Newline) {
                return nil, errors.New("Improper indentation at top level of code: line " + strconv.Itoa(t.LineNumber) + " column: " + strconv.Itoa(t.Column))
            }
        case OpenParen:
            expression, numTokens, err := parseOpenParen(tokens[i:])
            if err != nil {
                return nil, err
            }
            i += numTokens
            statements = append(statements, expression.(Statement))
        default:
            return nil, errors.New("Improper token at top level of code: line " + strconv.Itoa(t.LineNumber) + " column: " + strconv.Itoa(t.Column))
        }
    }
    return statements, nil
}

// assumes first token is open paren. 
// Returns a FunctionCall or Operation and the number of tokens that make up the Expression.
func parseOpenParen(tokens []Token) (Expression, int, error) {
    if len(tokens) < 3 {
        return nil, 0, errors.New("Improper function call or operation on line " + strconv.Itoa(tokens[0].LineNumber))
    }

    idx := 1
    functionCall := true
    var leadingCall Expression
    t := tokens[idx]

    switch t.Type {
    case OperatorWord:
        functionCall = false
        idx++
    case IdentifierWord:
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
        return nil, 0, errors.New("Improper function call or operation on line " + strconv.Itoa(tokens[0].LineNumber))
    }

    arguments := make([]Expression, 0)
    Loop:
    for true {
        t := tokens[idx]
        switch t.Type {
        case IdentifierWord, StringLiteral, NumberLiteral, BooleanLiteral, NullLiteral:
            arguments = append(arguments, t)
            idx++
        case OpenParen:
            expression, numTokens, err := parseOpenParen(tokens[idx:])
            if err != nil {
                return nil, 0, err
            }
            arguments = append(arguments, expression)
            idx += numTokens
        case CloseParen:
            idx++
            break Loop
        default:
            return nil, 0, errors.New("Improper token in function call or operation on line " + strconv.Itoa(t.LineNumber))
        }
    }

    if functionCall {
        if (leadingCall == nil) {
            return FunctionCall{tokens[1], arguments}, idx, nil
        } else {
            return FunctionCall{leadingCall, arguments}, idx, nil
        }
    } else {
        return Operation{tokens[1], arguments}, idx, nil
    }

}

func parseFunction(tokens []Token) (FunctionDefinition, int, error) {
    if len(tokens) < 5 {
        return FunctionDefinition{}, 0, errors.New("Improper function definition on line " + strconv.Itoa(tokens[0].LineNumber))
    }

    idx := 1
    name := tokens[idx]
    if name.Type != IdentifierWord {
        return FunctionDefinition{}, 0, errors.New("Function missing name on line " + strconv.Itoa(name.LineNumber))
    }

    idx++
    parameters := make([]Token, 0)
    for _, token := range tokens[idx:] {
        if (token.Type == IdentifierWord) {
            parameters = append(parameters, token)
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
    if len(tokens) < 5 {
        return IfStatement{}, 0, errors.New("Improper if statement on line " + strconv.Itoa(tokens[0].LineNumber))
    }

    var condition Expression
    var numConditionTokens int
    switch tokens[1].Type {
    case IdentifierWord, StringLiteral, NumberLiteral, BooleanLiteral, NullLiteral:
        condition = tokens[1]
        numConditionTokens = 1
    case OpenParen:
        var err error
        condition, numConditionTokens, err = parseOpenParen(tokens[1:])
        if err != nil {
            return IfStatement{}, 0, err
        }
    default:
        return IfStatement{}, 0, errors.New("Improper condition in if statement on line " + strconv.Itoa(tokens[0].LineNumber))
    }
    newline := tokens[1 + numConditionTokens]
    if newline.Type != Newline {
        return IfStatement{}, 0, errors.New("If statement condition not followed by newline on line " + strconv.Itoa(tokens[0].LineNumber))
    }

    numHeaderTokens := 2 + numConditionTokens
    body, numTokens, err := parseBody(tokens[numHeaderTokens:], indentation + indentationSpaces)
    if err != nil {
        return IfStatement{}, 0, err
    }

    elifClauses := make([]ElifClause, 0)
    var elseClause ElseClause
    idx := numHeaderTokens + numTokens

    if indentation == 0 {
        for idx < len(tokens) {
            if tokens[idx].Content == "elif" {
                elifClause, numTokens, err := parseElif(tokens[idx:], indentation)
                if err != nil {
                    return IfStatement{}, 0, err
                }
                elifClauses = append(elifClauses, elifClause)
                idx += numTokens
            } else {
                break
            }
        }
        
        if idx < len(tokens) {
            if tokens[idx].Content == "else" {
                var numTokens int
                var err error
                elseClause, numTokens, err = parseElse(tokens[idx:], indentation)
                if err != nil {
                    return IfStatement{}, 0, err
                }
                idx += numTokens
            }
        }
    } else {
        for idx + 1 < len(tokens) {
            if tokens[idx].Type == Indentation && len(tokens[idx].Content) == indentation && tokens[idx + 1].Content == "elif" {
                elifClause, numTokens, err := parseElif(tokens[idx + 1:], indentation)
                if err != nil {
                    return IfStatement{}, 0, err
                }
                elifClauses = append(elifClauses, elifClause)
                idx += numTokens + 1   // +1 for the indentation before this elif
            } else {
                break
            }
        }
        
        if idx + 1 < len(tokens) {
            if tokens[idx].Type == Indentation && len(tokens[idx].Content) == indentation && tokens[idx + 1].Content == "else" {
                var numTokens int
                var err error
                elseClause, numTokens, err = parseElse(tokens[idx + 1:], indentation)
                if err != nil {
                    return IfStatement{}, 0, err
                }
                idx += numTokens + 1    // +1 for the indentation before this else
            }
        }
    }
    

    return IfStatement{tokens[0], condition, body, elifClauses, elseClause}, idx, nil
}

func parseElif(tokens []Token, indentation int) (ElifClause, int, error) {
    if len(tokens) < 5 {
        return ElifClause{}, 0, errors.New("Improper elif clause on line " + strconv.Itoa(tokens[0].LineNumber))
    }

    var condition Expression
    var numConditionTokens int
    switch tokens[1].Type {
    case IdentifierWord, StringLiteral, NumberLiteral, BooleanLiteral, NullLiteral:
        condition = tokens[1]
        numConditionTokens = 1
    case OpenParen:
        var err error
        condition, numConditionTokens, err = parseOpenParen(tokens[1:])
        if err != nil {
            return ElifClause{}, 0, err
        }
    default:
        return ElifClause{}, 0, errors.New("Improper condition in elif clause on line " + strconv.Itoa(tokens[0].LineNumber))
    }
    newline := tokens[1 + numConditionTokens]
    if newline.Type != Newline {
        return ElifClause{}, 0, errors.New("Elif clause condition not followed by newline on line " + strconv.Itoa(tokens[0].LineNumber))
    }

    numHeaderTokens := 2 + numConditionTokens
    body, numTokens, err := parseBody(tokens[numHeaderTokens:], indentation + indentationSpaces)
    if err != nil {
        return ElifClause{}, 0, err
    }

    return ElifClause{tokens[0], condition, body}, numHeaderTokens + numTokens, nil
}

func parseElse(tokens []Token, indentation int) (ElseClause, int, error) {
    if len(tokens) < 4 {
        return ElseClause{}, 0, errors.New("Improper else clause on line " + strconv.Itoa(tokens[0].LineNumber))
    }

    newline := tokens[1]
    if newline.Type != Newline {
        return ElseClause{}, 0, errors.New("Elif clause condition not followed by newline on line " + strconv.Itoa(tokens[0].LineNumber))
    }

    body, numTokens, err := parseBody(tokens[2:], indentation + indentationSpaces)
    if err != nil {
        return ElseClause{}, 0, err
    }

    return ElseClause{tokens[0], body}, 2 + numTokens, nil
}

func parseWhile(tokens []Token, indentation int) (WhileStatement, int, error) {
    if len(tokens) < 5 {
        return WhileStatement{}, 0, errors.New("Improper while statement on line " + strconv.Itoa(tokens[0].LineNumber))
    }

    var condition Expression
    var numConditionTokens int
    switch tokens[1].Type {
    case IdentifierWord, StringLiteral, NumberLiteral, BooleanLiteral, NullLiteral:
        condition = tokens[1]
        numConditionTokens = 1
    case OpenParen:
        var err error
        condition, numConditionTokens, err = parseOpenParen(tokens[1:])
        if err != nil {
            return WhileStatement{}, 0, err
        }
    default:
        return WhileStatement{}, 0, errors.New("Improper condition in while statement on line " + strconv.Itoa(tokens[0].LineNumber))
    }
    newline := tokens[1 + numConditionTokens]
    if newline.Type != Newline {
        return WhileStatement{}, 0, errors.New("While statement condition not followed by newline on line " + strconv.Itoa(tokens[0].LineNumber))
    }

    numHeaderTokens := 2 + numConditionTokens
    body, numTokens, err := parseBody(tokens[numHeaderTokens:], indentation + indentationSpaces)
    if err != nil {
        return WhileStatement{}, 0, err
    }

    return WhileStatement{tokens[0], condition, body}, numHeaderTokens + numTokens, nil
}

func parseReturn(tokens []Token) (ReturnStatement, int, error) {
    if len(tokens) < 3 {
        return ReturnStatement{}, 0, errors.New("Improper return statement on line " + strconv.Itoa(tokens[0].LineNumber))
    }
    var value Expression
    var numValueTokens int
    switch tokens[1].Type {
    case IdentifierWord, StringLiteral, NumberLiteral, BooleanLiteral, NullLiteral:
        value = tokens[1]
        numValueTokens = 1
    case OpenParen:
        var err error
        value, numValueTokens, err = parseOpenParen(tokens[1:])
        if err != nil {
            return ReturnStatement{}, 0, err
        }
    default:
        return ReturnStatement{}, 0, errors.New("Improper value in return statement on line " + strconv.Itoa(tokens[0].LineNumber))
    }
    newline := tokens[1 + numValueTokens]
    if newline.Type != Newline {
        return ReturnStatement{}, 0, errors.New("Return statement not terminated with newline on line " + strconv.Itoa(tokens[0].LineNumber))
    }

    return ReturnStatement{tokens[0], value}, 2 + numValueTokens, nil
}

// assume first token is reeserved word "as"
// returns number of tokens (includes the newline at the end but not indentation)
func parseAssignment(tokens []Token) (AssignmentStatement, int, error) {
    if len(tokens) < 4 {
        return AssignmentStatement{}, 0, errors.New("Improper assignment statement on line " + strconv.Itoa(tokens[0].LineNumber))
    }
    target := tokens[1]
    if target.Type != IdentifierWord {
        return AssignmentStatement{}, 0, errors.New("Improper target of assignment on line " + strconv.Itoa(tokens[0].LineNumber))
    }
    var value Expression
    var numValueTokens int
    switch tokens[2].Type {
    case IdentifierWord, StringLiteral, NumberLiteral, BooleanLiteral, NullLiteral:
        value = tokens[2]
        numValueTokens = 1
    case OpenParen:
        var err error
        value, numValueTokens, err = parseOpenParen(tokens[2:])
        if err != nil {
            return AssignmentStatement{}, 0, err
        }
    default:
        return AssignmentStatement{}, 0, errors.New("Improper value in assignment on line " + strconv.Itoa(tokens[0].LineNumber))
    }
    newline := tokens[2 + numValueTokens]
    if newline.Type != Newline {
        return AssignmentStatement{}, 0, errors.New("Assignment not terminated with newline on line " + strconv.Itoa(tokens[0].LineNumber))
    }

    return AssignmentStatement{target, value}, 3 + numValueTokens, nil  // 3 because: 'as', the target, and the newline at the end
}

// expected to start with Indentation token. 
// 'indentation' = the number of spaces indentation on which the body should be aligned
// May return zero statements if body is empty.
func parseBody(tokens []Token, indentation int) ([]Statement, int, error) {
    statements := make([]Statement, 0)
    i := 0
    for i < len(tokens) {
        t := tokens[i]
        if t.Type == Newline {  // blank line
            i++
        } else if t.Type == Indentation && tokens[i + 1].Type == Newline {   // blank line
            i += 2
        } else if t.Type != Indentation { // gone past end of the body
            break
        } else {
            numSpaces := len(t.Content)
            if numSpaces < indentation {  // gone past end of the body
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
                    case "function":
                        return nil, 0, errors.New("Functions cannot be nested: line " + strconv.Itoa(t.LineNumber) + " column: " + strconv.Itoa(t.Column))
                    case "as":
                        statement, numTokens, err = parseAssignment(tokens[i:])
                    case "if":
                        statement, numTokens, err = parseIf(tokens[i:], indentation)
                    case "while":
                        statement, numTokens, err = parseWhile(tokens[i:], indentation)
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
                    
                    if tokens[i + numTokens].Type != Newline {
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

const inputFilename = "example.pigeon"


/* All identifiers get prefixed with _ to avoid collisions with Go reserved words and predefined identifiers */
func compile(statements []Statement) (string, error) {
    globals := make(Scope)

    var body string
    nFunctions := 0
    funcAllowed := true    // check that all function statements come before any other statements
    for _, s := range statements {
        fn, ok := s.(FunctionDefinition)
        if ok {
            if funcAllowed {
                nFunctions++
            } else {
                return "", errors.New("Functions must all be defined before rest of code. Improperly placed function on line " +
                    strconv.Itoa(fn.FirstToken.LineNumber))
            }
        } else {
           funcAllowed = false   // once we encounter a non-function, can't have subsequent functions 
        }
        var c string
        var err error
        switch s := s.(type) {
        case IfStatement:
            c, err = compileIfStatement(s, globals, nil)
        case WhileStatement:
            c, err = compileWhileStatement(s, globals, nil)
        case AssignmentStatement:
            c, err = compileAssignmentStatement(s, globals, nil)
        case ReturnStatement:
            return "", errors.New("Cannot have a return statement at top level of code. Line: " + strconv.Itoa(s.FirstToken.LineNumber))
        case FunctionCall:
            c, err = compileFunctionCall(s)
            c += "\n"
        case Operation:
            c, err = compileOperation(s)
            c += "\n"
        }
        if err != nil {
            return "", err
        }
        body += c
    }

    var functions string
    for _, s := range statements[:nFunctions] {
        c, err := compileFunc(s.(FunctionDefinition), globals)
        if err != nil {
            return "", err
        }
        functions += c
    }

    var declarations string
    for g, _ := range globals {
        declarations += "var _" + g + " interface{}\n"
    }

    header := `package main

import _pigeon "github.com/BrianWill/pigeon/stdlib/"

`
    body = "func main() {\n" + body + "}\n"

    return header + declarations + functions + body, nil
}

// returns code snippet ending with '\n\n'
func compileFunc(fn FunctionDefinition, globals Scope) (string, error) {
    locals := make(Scope)

    code := "func " + fn.Name.Content + "("
    for _, param := range fn.Parameters {
        code += "_" + param.Content + " interface{}, "
    }
    if len(fn.Parameters) > 0 {
        code = code[:len(code) - 2]  // drop last comma and space
    }
    code += ") interface{} {\n"
    c, err := compileBody(fn.Body, locals, globals)
    if err != nil {
        return "", err
    }
    if len(fn.Body) > 0 {
        _, lastIsReturn := fn.Body[len(fn.Body) - 1].(ReturnStatement)
        if lastIsReturn {
            return code + c + "}\n", nil
        }
    }

    var declarations string
    for l, _ := range locals {
        declarations += "var _" + l + " interface{}\n"
    }

    return code + declarations + c + "return nil\n}\n", nil   
}

func compileIfStatement(s IfStatement, this, enclosing Scope) (string, error) {
    c, err := compileExpression(s.Condition)
    if err != nil {
        return "", err
    }
    code := "if " + c
    c, err = compileBody(s.Body, this, enclosing)
    if err != nil {
        return "", nil
    }
    code += " {\n" + c + "}"
    for _, elif := range s.Elifs {
        c, err := compileExpression(elif.Condition)
        if err != nil {
            return "", err
        }
        code += " else if " + c + " {\n"
        c, err = compileBody(elif.Body, this, enclosing)
        if err != nil {
            return "", err
        }
        code += c + "}"
    }

    if len(s.Else.Body) > 0 {
        c, err := compileBody(s.Else.Body, this, enclosing)
        if err != nil {
            return "", err
        }
        code += " else {\n" + c + "}"
    }
    return code + "\n", nil
}

func compileWhileStatement(s WhileStatement, this, enclosing Scope) (string, error) {
    c, err := compileExpression(s.Condition)
    if err != nil {
        return "", err
    }
    code := "for " + c + " {\n"
    c, err = compileBody(s.Body, this, enclosing)
    if err != nil {
        return "", err
    }
    return code + c + "}\n", nil
}

func compileBody(statements []Statement, this, enclosing Scope) (string, error) {
    var code string
    for _, s := range statements {
        var c string
        var err error
        switch s := s.(type) {
        case IfStatement:
            c, err = compileIfStatement(s, this, enclosing)
        case WhileStatement:
            c, err = compileWhileStatement(s, this, enclosing)
        case AssignmentStatement:
            c, err = compileAssignmentStatement(s, this, enclosing)
        case ReturnStatement:
            c, err = compileReturnStatement(s)
        case FunctionCall:
            c, err = compileFunctionCall(s)
            c += "\n"
        case Operation:
            c, err = compileOperation(s)
            c += "\n"
        }
        if err != nil {
            return "", err
        }
        code += c
    }
    return code, nil
}

// 'scope' = map of variable names assigned to in the current scope
func compileAssignmentStatement(s AssignmentStatement, this, enclosing Scope) (string, error) {
    target := s.Target.Content
    if (enclosing == nil) || !enclosing[target] {
        this[target] = true
    }
    c, err := compileExpression(s.Value)
    if err != nil {
        return "", err
    }
    return "_" + target + " = " + c + "\n", nil
}

func compileReturnStatement(s ReturnStatement) (string, error) {
    c, err := compileExpression(s.Value)
    if err != nil {
        return "", err
    }
    return "return " + c + "\n", nil
}

func compileFunctionCall(s FunctionCall) (string, error) {
    var code string
    switch s:= s.Function.(type) {
    case Operation:
        c, err := compileOperation(s)
        if err != nil {
            return "", err
        }
        code += c
    case FunctionCall:
        c, err := compileFunctionCall(s)
        if err != nil {
            return "", err
        }
        code += c
    case Token:   // will always be an identifier
        code += "_" + s.Content
    }
    code += "("  // start of arguments
    for _, exp := range s.Arguments {
        c, err := compileExpression(exp)
        if err != nil {
            return "", err
        }
        code += c + ", " // Go is OK with comma after last arg, so don't need special case for last arg
    }
    if len(s.Arguments) > 0 {
        code = code[:len(code) - 2]  // drop last comma and space
    }
    return code + ")", nil
}

func compileOperation(o Operation) (string, error) {
    operator := o.Operator.Content

    runes := []rune(operator)
    runes[0] = unicode.ToUpper(runes[0])
    operator = string(runes)

    code := "_pigeon." + operator + "("
    for _, exp := range o.Operands {
        c, err := compileExpression(exp)
        if err != nil {
            return "", err
        }
        code += c + ", " // Go is OK with comma after last arg, so don't need special case for last arg
    }
    if len(o.Operands) > 0 {
        code = code[:len(code) - 2]  // drop last comma and space
    }
    code += ")"
    return code, nil
}

func compileExpression(e Expression) (string, error) {
    var code string
    switch e := e.(type) {
    case Operation:
        c, err := compileOperation(e)
        if err != nil {
            return "", err
        }
        code = c
    case FunctionCall:
        c, err := compileFunctionCall(e)
        if err != nil {
            return "", err
        }
        code = c
    case Token:
        switch e.Type {
        case IdentifierWord:
            code = "_" + e.Content
        case NumberLiteral:
            code = "float64(" + e.Content + ")"
        case StringLiteral, BooleanLiteral:
            code = e.Content
        case NullLiteral:
            code = "_pigeon.Null(0)"
        }

    }
    return code, nil
}

func main() {
	data, err := ioutil.ReadFile(inputFilename)
    if err != nil {
    	fmt.Println(err)
    	return
    }
    tokens, err := lex(string(data) + "\n")
    if err != nil {
    	fmt.Println(err)
    	return
    }
    statements, err := parse(tokens)
    if err != nil {
        fmt.Println(err)
        return
    }
    code, err := compile(statements)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println(code)
}