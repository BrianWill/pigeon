/*
If the file contains tab characters or any non-ASCII character, the lexer
returns an error. (Only spaces are allowed for indentation.)

Every line ends with a newline token.

Every line starts with an indentation token representing the spaces at the start of a line.
An unindented line starts with an indentation token with an empty string for its content. (This will make parsing a bit easier.)

For example, this line of Pigeon:

function david a b c

...is represented as eleven tokens:

	Indentation ("")
	ReservedWord ("function")
	Space ("")
	Identifier ("david")
	Space ("")
	Identifier ("a")
	Space ("")
	Identifier ("b")
	Space ("")
	Identifier ("c")
	Newline ("\n")


The last line of the input file will not necessarily end with a newline, but we add a newline token at the end anyway.

*/

package staticPigeon

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

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
		} else if r == '<' {
			tokens = append(tokens, Token{OpenAngle, "<", line, column})
			column++
			i++
		} else if r == '>' {
			tokens = append(tokens, Token{CloseAngle, ">", line, column})
			column++
			i++
		} else if r == '.' {
			tokens = append(tokens, Token{Dot, ".", line, column})
			column++
			i++
		} else if r == ':' {
			tokens = append(tokens, Token{Colon, ":", line, column})
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
				// A word should always end with space, newline, <, >, ., [, or )
				if strings.Contains(" \n)<>.[", string(current)) {
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
			if tokenType == IdentifierWord {
				if content[0] >= 65 && content[0] <= 90 {
					tokenType = TypeName
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
	// filter out blank lines
	filteredTokens := []Token{}
	for i := 0; i < len(tokens)-1; {
		if tokens[i].Type == Indentation && tokens[i+1].Type == Newline {
			i += 2
		} else if tokens[i].Type == Newline && tokens[i+1].Type == Newline {
			i++
		} else {
			filteredTokens = append(filteredTokens, tokens[i])
			i++
		}
	}
	filteredTokens = append(filteredTokens, tokens[len(tokens)-1])
	return filteredTokens, nil
}

// parse the top-level definitions
func parse(tokens []Token) ([]Definition, error) {
	var definitions []Definition
	for i := 0; i < len(tokens); {
		t := tokens[i]
		line := t.LineNumber
		switch t.Type {
		case ReservedWord:
			var definition Definition
			var numTokens int
			var err error
			switch t.Content {
			case "import":
				definition, numTokens, err = parseImport(tokens[i:], line)
			case "struct":
				definition, numTokens, err = parseStruct(tokens[i:], line)
			case "interface":
				definition, numTokens, err = parseInterface(tokens[i:], line)
			case "method":
				definition, numTokens, err = parseMethod(tokens[i:], line)
			case "func":
				definition, numTokens, err = parseFunction(tokens[i:], line)
			case "global":
				definition, numTokens, err = parseGlobal(tokens[i:], line)
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

func parseImport(tokens []Token, line int) (ImportDefinition, int, error) {
	lineStr := strconv.Itoa(line)
	idx := 1
	if tokens[idx].Type != Space {
		return ImportDefinition{}, 0, errors.New("Expected space on line " + lineStr)
	}
	idx++
	path := tokens[idx]
	if path.Type != StringLiteral {
		return ImportDefinition{}, 0, errors.New("Expected string literal on line " + lineStr)
	}
	idx++
	if tokens[idx].Type == Space {
		idx++
	}
	if tokens[idx].Type != Newline {
		return ImportDefinition{}, 0, errors.New("Expected newline on line " + lineStr)
	}
	idx++

	importedNames := []string{}
	importedAliases := []string{}
	for {
		if tokens[idx].Type != Indentation {
			break
		}
		idx++
		t := tokens[idx]
		if t.Type != IdentifierWord {
			return ImportDefinition{}, 0, errors.New("Expected name to import on line " + lineStr)
		}
		idx++
		importedNames = append(importedNames, t.Content)
		if tokens[idx].Type == Space {
			idx++
		}
		t = tokens[idx]
		if t.Type == IdentifierWord {
			importedAliases = append(importedAliases, t.Content)
			idx++
		} else {
			importedAliases = append(importedAliases, "")
		}
		if tokens[idx].Type == Space {
			idx++
		}
		if tokens[idx].Type != Newline {
			return ImportDefinition{}, 0, errors.New("Expected newline on line " + lineStr)
		}
		idx++
	}
	if len(importedNames) == 0 {
		return ImportDefinition{}, 0, errors.New("Import statement has no imported names on line " + lineStr)
	}
	return ImportDefinition{line, tokens[0].Column, path.Content, importedNames, importedAliases}, idx, nil
}

func parseStruct(tokens []Token, line int) (StructDefinition, int, error) {
	lineStr := strconv.Itoa(line)
	idx := 1
	if tokens[idx].Type != Space {
		return StructDefinition{}, 0, errors.New("Expected space on line " + lineStr)
	}
	idx++
	name := tokens[idx]
	if name.Type != TypeName {
		return StructDefinition{}, 0, errors.New("Expected name for struct on line " + lineStr)
	}
	idx++
	for _, v := range builtinTypes {
		if name.Content == v {
			return StructDefinition{}, 0, errors.New("Invalid struct name: cannot redefine builtin type " +
				name.Content + " on line " + lineStr)
		}
	}
	if tokens[idx].Type == Space {
		idx++
	}
	if tokens[idx].Type != Newline {
		return StructDefinition{}, 0, errors.New("Expected newline on line " + lineStr)
	}
	idx++

	members := []Variable{}
	for {
		lineStr := strconv.Itoa(tokens[idx].LineNumber)
		if tokens[idx].Type != Indentation {
			break
		}
		idx++
		memberName := tokens[idx]
		if memberName.Type != IdentifierWord {
			return StructDefinition{}, 0, errors.New("Expected struct member name on line " + lineStr)
		}
		idx++
		if tokens[idx].Type != Space {
			return StructDefinition{}, 0, errors.New("Expected space on line " + lineStr)
		}
		idx++
		memberType, numTypeTokens, err := parseType(tokens[idx:], tokens[idx].LineNumber)
		if err != nil {
			return StructDefinition{}, 0, err
		}
		idx += numTypeTokens
		members = append(members, Variable{memberName.LineNumber, memberName.Column, memberName.Content, memberType})
		if tokens[idx].Type == Space {
			idx++
		}
		if tokens[idx].Type != Newline {
			return StructDefinition{}, 0, errors.New("Expected newline on line " + lineStr)
		}
		idx++
	}
	if len(members) == 0 {
		return StructDefinition{}, 0, errors.New("Struct definition has no members on line " + lineStr)
	}
	return StructDefinition{line, tokens[0].Column, name.Content, members}, idx, nil
}

func parseMethod(tokens []Token, line int) (MethodDefinition, int, error) {
	lineStr := strconv.Itoa(line)
	funcDef, numTokens, err := parseFunction(tokens, line)
	if err != nil {
		return MethodDefinition{}, 0, err
	}
	if len(funcDef.Parameters) == 0 {
		return MethodDefinition{}, 0, errors.New("Method must have a receiver parameter on line " + lineStr)
	}
	return MethodDefinition{
		funcDef.LineNumber,
		funcDef.Column,
		funcDef.Name,
		funcDef.Parameters[0],
		funcDef.Parameters[1:],
		funcDef.ReturnTypes,
		funcDef.Body,
	}, numTokens, nil
}

// used by parseFunction
// consumes all tokens through end of line
func parseParameters(tokens []Token, line int) ([]Variable, []ParsedDataType, int, error) {
	lineStr := strconv.Itoa(line)
	params := []Variable{}
	idx := 0
	expectingSpace := false
Loop:
	for {
		t := tokens[idx]
		switch t.Type {
		case Space:
			expectingSpace = false
			idx++
		case IdentifierWord:
			idx++
			if tokens[idx].Type != Space {
				return nil, nil, 0, errors.New("Expecting space on line " + lineStr)
			}
			idx++
			dataType, n, err := parseType(tokens[idx:], line)
			if err != nil {
				return nil, nil, 0, err
			}
			idx += n
			expectingSpace = true
			params = append(params, Variable{t.LineNumber, t.Column, t.Content, dataType})
		case Colon:
			if expectingSpace {
				return nil, nil, 0, errors.New("Expecting space on line " + lineStr)
			}
			// don't inc idx
			break Loop
		case Newline:
			// don't inc idx
			break Loop
		default:
			return nil, nil, 0, errors.New("Unexpected token on line " + lineStr)
		}
	}

	// optional colon and return types
	var returnTypes []ParsedDataType
	if tokens[idx].Type == Colon {
		idx++
		if tokens[idx].Type != Space {
			return nil, nil, 0, errors.New("Expecting space on line " + lineStr)
		}
		idx++
		t, n, err := parseType(tokens[idx:], line)
		if err != nil {
			return nil, nil, 0, err
		}
		idx += n
		returnTypes = append(returnTypes, t)
		for {
			if tokens[idx].Type == Newline || tokens[idx].Type == Space && tokens[idx+1].Type == Newline {
				break
			}
			if tokens[idx].Type != Space {
				return nil, nil, 0, errors.New("Expecting space on line " + lineStr)
			}
			idx++
			t, n, err := parseType(tokens[idx:], line)
			if err != nil {
				return nil, nil, 0, err
			}
			idx += n
			returnTypes = append(returnTypes, t)
		}
	}

	// newline
	if tokens[idx].Type == Space {
		idx++
	}
	if tokens[idx].Type != Newline {
		return nil, nil, 0, errors.New("Expecting newline on line " + lineStr)
	}
	idx++

	return params, returnTypes, idx, nil
}

// used by parseInterface, not parseFunction
// consumes all tokens through end of line
func parseSignature(tokens []Token, line int) (Signature, int, error) {
	lineStr := strconv.Itoa(line)
	paramTypes := []ParsedDataType{}
	returnTypes := []ParsedDataType{}
	idx := 0
	methodName := tokens[idx]
	if methodName.Type != IdentifierWord {
		return Signature{}, 0, errors.New("Expecting method name on line " + lineStr)
	}
	idx++
	if tokens[idx].Type == Space && tokens[idx+1].Type != Newline {
		expectingSpace := false
	Loop:
		for {
			t := tokens[idx]
			switch t.Type {
			case Space:
				expectingSpace = false
				idx++
			case TypeName:
				dataType, n, err := parseType(tokens[idx:], line)
				if err != nil {
					return Signature{}, 0, err
				}
				paramTypes = append(paramTypes, dataType)
				idx += n
				expectingSpace = true
			case Colon:
				if expectingSpace {
					return Signature{}, 0, errors.New("Expecting space on line " + lineStr)
				}
				// don't inc idx
				break Loop
			case Newline:
				break Loop
			default:
				return Signature{}, 0, errors.New("Unexpected token on line " + lineStr)
			}
		}
		// optional colon and return types
		if tokens[idx].Type == Colon {
			idx++
			if tokens[idx].Type != Space {
				return Signature{}, 0, errors.New("Expecting space on line " + lineStr)
			}
			idx++
			t, n, err := parseType(tokens[idx:], line)
			if err != nil {
				return Signature{}, 0, err
			}
			idx += n
			returnTypes = append(returnTypes, t)
			for {
				if tokens[idx].Type == Newline || tokens[idx].Type == Space && tokens[idx+1].Type == Newline {
					break
				}
				if tokens[idx].Type != Space {
					return Signature{}, 0, errors.New("Expecting space on line " + lineStr)
				}
				idx++
				t, n, err := parseType(tokens[idx:], line)
				if err != nil {
					return Signature{}, 0, err
				}
				idx += n
				returnTypes = append(returnTypes, t)
			}
		}
	}
	// newline
	if tokens[idx].Type == Space {
		idx++
	}
	if tokens[idx].Type != Newline {
		return Signature{}, 0, errors.New("Expecting newline on line " + lineStr)
	}
	idx++
	return Signature{line, tokens[0].Column, methodName.Content, paramTypes, returnTypes}, idx, nil
}

func parseInterface(tokens []Token, line int) (InterfaceDefinition, int, error) {
	lineStr := strconv.Itoa(line)
	idx := 1
	if tokens[idx].Type != Space {
		return InterfaceDefinition{}, 0, errors.New("Expected space on line " + lineStr)
	}
	idx++
	name := tokens[idx]
	if name.Type != TypeName {
		return InterfaceDefinition{}, 0, errors.New("Expected name for interface on line " + lineStr)
	}
	idx++
	if tokens[idx].Type == Space {
		idx++
	}
	if tokens[idx].Type != Newline {
		return InterfaceDefinition{}, 0, errors.New("Expected newline on line " + lineStr)
	}
	idx++

	methods := []Signature{}
	for {
		if tokens[idx].Type != Indentation {
			break
		}
		idx++
		signature, numTokens, err := parseSignature(tokens[idx:], tokens[idx].LineNumber)
		if err != nil {
			return InterfaceDefinition{}, 0, err
		}
		methods = append(methods, signature)
		idx += numTokens
	}
	if len(methods) == 0 {
		return InterfaceDefinition{}, 0, errors.New("Interface definition has no method signatures on line " + lineStr)
	}
	return InterfaceDefinition{line, tokens[0].Column, name.Content, methods}, idx, nil

}

func parseType(tokens []Token, line int) (ParsedDataType, int, error) {
	lineStr := strconv.Itoa(line)
	idx := 0
	primary := tokens[idx]
	if primary.Type != TypeName {
		return ParsedDataType{}, 0, errors.New("Expecting type name on line " + lineStr)
	}
	idx++
	paramTypes := []ParsedDataType{}
	returnTypes := []ParsedDataType{}
	if tokens[idx].Type == OpenAngle {
		idx++
		if tokens[idx].Type == Space {
			idx++
		}
		if tokens[idx].Type == Colon {
			returnTypes, n, err := parseReturnTypes(tokens[idx:], line)
			if err != nil {
				return ParsedDataType{}, 0, err
			}
			idx += n
			return ParsedDataType{primary.Content, paramTypes, returnTypes}, idx, nil
		}
		dataType, n, err := parseType(tokens[idx:], line)
		if err != nil {
			return ParsedDataType{}, 0, err
		}
		idx += n
		paramTypes = append(paramTypes, dataType)

		for {
			if tokens[idx].Type != Space {
				break
			}
			idx++
			if tokens[idx].Type == Colon {
				var n int
				var err error
				returnTypes, n, err = parseReturnTypes(tokens[idx:], line)
				if err != nil {
					return ParsedDataType{}, 0, err
				}
				idx += n
				break
			}
			dataType, n, err := parseType(tokens[idx:], line)
			if err != nil {
				return ParsedDataType{}, 0, err
			}
			idx += n
			paramTypes = append(paramTypes, dataType)
		}
		if tokens[idx].Type != CloseAngle {
			return ParsedDataType{}, 0, errors.New("Expecting closing angle bracket on line " + lineStr)
		}
		idx++
	}
	return ParsedDataType{primary.Content, paramTypes, returnTypes}, idx, nil
}

// expects to end with newline or >, but does not consume the newline or >
func parseReturnTypes(tokens []Token, line int) ([]ParsedDataType, int, error) {
	lineStr := strconv.Itoa(line)
	idx := 1
	if tokens[idx].Type != Space {
		return nil, 0, errors.New("Expecting space on line " + lineStr)
	}
	returnTypes := []ParsedDataType{}
	dataType, n, err := parseType(tokens[idx:], line)
	if err != nil {
		return nil, 0, err
	}
	returnTypes = append(returnTypes, dataType)
	idx += n
	for {
		if tokens[idx].Type == CloseAngle || tokens[idx].Type == Newline {
			break
		}
		if tokens[idx].Type == Space && (tokens[idx+1].Type == CloseAngle || tokens[idx+1].Type == Newline) {
			idx++
			break
		}
		if tokens[idx].Type != Space {
			return nil, 0, errors.New("Expecting space on line " + lineStr)
		}
		idx++
		dataType, n, err := parseType(tokens[idx:], line)
		if err != nil {
			return nil, 0, err
		}
		idx += n
		returnTypes = append(returnTypes, dataType)
	}
	return returnTypes, idx, nil
}

func parseGlobal(tokens []Token, line int) (GlobalDefinition, int, error) {
	lineStr := strconv.Itoa(line)
	idx := 1
	if tokens[idx].Type != Space {
		return GlobalDefinition{}, 0, errors.New("Expected space on line " + lineStr)
	}
	idx++
	target := tokens[idx]
	if target.Type != IdentifierWord {
		return GlobalDefinition{}, 0, errors.New("Improper name for a global on line " + lineStr)
	}
	idx++
	if tokens[idx].Type != Space {
		return GlobalDefinition{}, 0, errors.New("Expected space on line " + lineStr)
	}
	idx++
	globalType, numTypeTokens, err := parseType(tokens[idx:], line)
	if err != nil {
		return GlobalDefinition{}, 0, err
	}
	idx += numTypeTokens
	if tokens[idx].Type != Space {
		return GlobalDefinition{}, 0, errors.New("Expected space on line " + lineStr)
	}
	idx++
	value, numValueTokens, err := parseExpression(tokens[idx:], line)
	if err != nil {
		return GlobalDefinition{}, 0, err
	}
	idx += numValueTokens
	if tokens[idx].Type != Newline {
		return GlobalDefinition{}, 0, errors.New("Global not terminated with newline on line " + lineStr)
	}
	idx++
	return GlobalDefinition{line, tokens[0].Column, target.Content, value, globalType}, idx, nil
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
	if tokens[1].Type != IdentifierWord {
		return nil, 0, errors.New("Identifier expected after dot line " + strconv.Itoa(line))
	}
	strLiteral := Token{StringLiteral, "\"" + tokens[1].Content + "\"", line, -1}
	getOp := Operation{
		line,
		tokens[0].Column,
		"get",
		[]Expression{expr, strLiteral},
	}
	return getOp, 2, nil
}

// assumes first token is open square
func parseOpenSquare(tokens []Token, expr Expression, line int) (Expression, int, error) {
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
		line,
		tokens[0].Column,
		"get",
		[]Expression{expr, indexExpr},
	}
	return getOp, idx, nil
}

// assumes first token is open paren.
// Returns a FunctionCall or Operation and the number of tokens that make up the Expression.
func parseOpenParen(tokens []Token) (Expression, int, error) {
	line := tokens[0].LineNumber
	lineStr := strconv.Itoa(line)
	idx := 1
	if tokens[idx].Type == Space {
		idx++
	}

	functionCall := true
	typeExpression := false
	var leadingCall Expression
	var op Token
	var dt ParsedDataType
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
	case TypeName:
		typeExpression = true
		var n int
		var err error
		dt, n, err = parseType(tokens[idx:], line)
		if err != nil {
			return nil, 0, err
		}
		idx += n
	default:
		return nil, 0, errors.New("Improper function call or operation on line " + lineStr)
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
	if typeExpression {
		expr = TypeExpression{tokens[0].LineNumber, tokens[0].Column, dt, arguments}
	} else if functionCall {
		if leadingCall == nil {
			expr = FunctionCall{tokens[0].LineNumber, tokens[0].Column, op, arguments}
		} else {
			expr = FunctionCall{tokens[0].LineNumber, tokens[0].Column, leadingCall, arguments}
		}
	} else {
		expr = Operation{tokens[0].LineNumber, tokens[0].Column, op.Content, arguments}
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

func parseFunction(tokens []Token, line int) (FunctionDefinition, int, error) {
	lineStr := strconv.Itoa(tokens[0].LineNumber)
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
	var params []Variable
	var returnTypes []ParsedDataType
	var err error
	if tokens[idx].Type == Newline {
		idx++
	} else if tokens[idx].Type == Space && tokens[idx+1].Type == Newline {
		idx += 2
	} else {
		if tokens[idx].Type != Space {
			return FunctionDefinition{}, 0, errors.New("Expecting space on line " + lineStr)
		}
		idx++
		var nTokens int
		params, returnTypes, nTokens, err = parseParameters(tokens[idx:], line)
		if err != nil {
			return FunctionDefinition{}, 0, err
		}
		idx += nTokens
	}
	body, nTokens, err := parseBody(tokens[idx:], indentationSpaces)
	if err != nil {
		return FunctionDefinition{}, 0, err
	}
	idx += nTokens
	return FunctionDefinition{tokens[0].LineNumber, tokens[0].Column, name.Content,
		params, returnTypes, body}, idx, nil
}

// 'indentation' = number of spaces before 'if'
func parseIf(tokens []Token, indentation int) (IfStatement, int, error) {
	line := strconv.Itoa(tokens[0].LineNumber)
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
	return IfStatement{tokens[0].LineNumber, tokens[0].Column, condition, body, elseifClauses, elseClause}, idx, nil
}

func parseElif(tokens []Token, indentation int) (ElseifClause, int, error) {
	line := strconv.Itoa(tokens[0].LineNumber)
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
	return ElseifClause{tokens[0].LineNumber, tokens[0].Column, condition, body}, idx, nil
}

func parseElse(tokens []Token, indentation int) (ElseClause, int, error) {
	line := strconv.Itoa(tokens[0].LineNumber)
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
	return ElseClause{tokens[0].LineNumber, tokens[0].Column, body}, idx, nil
}

func parseWhile(tokens []Token, indentation int) (WhileStatement, int, error) {
	line := strconv.Itoa(tokens[0].LineNumber)
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
	return WhileStatement{tokens[0].LineNumber, tokens[0].Column, condition, body}, idx, nil
}

func parseReturn(tokens []Token) (ReturnStatement, int, error) {
	lineStr := strconv.Itoa(tokens[0].LineNumber)
	idx := 1
	if tokens[idx].Type != Space {
		return ReturnStatement{}, 0, errors.New("Missing space on line " + lineStr)
	}
	idx++
	value, nTokens, err := parseExpression(tokens[idx:], tokens[0].LineNumber)
	if err != nil {
		return ReturnStatement{}, 0, err
	}
	idx += nTokens
	values := []Expression{value}
	for {
		if tokens[idx].Type == Newline {
			idx++
			break
		}
		if tokens[idx].Type == Space && tokens[idx+1].Type == Newline {
			idx += 2
			break
		}
		if tokens[idx].Type != Space {
			return ReturnStatement{}, 0, errors.New("Missing space on line " + lineStr)
		}
		idx++
		value, nTokens, err := parseExpression(tokens[idx:], tokens[0].LineNumber)
		if err != nil {
			return ReturnStatement{}, 0, err
		}
		idx += nTokens
		values = append(values, value)
	}
	return ReturnStatement{tokens[0].LineNumber, tokens[0].Column, values}, idx, nil
}

func parseBreak(tokens []Token) (BreakStatement, int, error) {
	line := strconv.Itoa(tokens[0].LineNumber)
	idx := 1
	if tokens[idx].Type == Space {
		idx++
	}
	if tokens[idx].Type != Newline {
		return BreakStatement{}, 0, errors.New("Break statement not terminated with newline on line " + line)
	}
	return BreakStatement{tokens[0].LineNumber, tokens[0].Column}, idx, nil
}

func parseContinue(tokens []Token) (ContinueStatement, int, error) {
	line := strconv.Itoa(tokens[0].LineNumber)
	idx := 1
	if tokens[idx].Type == Space {
		idx++
	}
	if tokens[idx].Type != Newline {
		return ContinueStatement{}, 0, errors.New("Continue statement not terminated with newline on line " + line)
	}
	return ContinueStatement{tokens[0].LineNumber, tokens[0].Column}, idx, nil
}

// assume first token is reserved word "as"
// returns number of tokens (including the newline at the end)
func parseAssignment(tokens []Token) (AssignmentStatement, int, error) {
	line := strconv.Itoa(tokens[0].LineNumber)
	idx := 1
	if tokens[idx].Type != Space {
		return AssignmentStatement{}, 0, errors.New("Missing space on line " + line)
	}
	idx++
	exprs := []Expression{}
	for {
		expr, nTokens, err := parseExpression(tokens[idx:], tokens[0].LineNumber)
		if err != nil {
			return AssignmentStatement{}, 0, err
		}
		idx += nTokens
		exprs = append(exprs, expr)
		if tokens[idx].Type == Newline {
			idx++
			break
		}
		if tokens[idx].Type == Space && tokens[idx+1].Type == Newline {
			idx += 2
			break
		}
		if tokens[idx].Type != Space {
			return AssignmentStatement{}, 0, errors.New("Missing space on line " + line)
		}
		idx++
	}
	if len(exprs) < 2 {
		return AssignmentStatement{}, 0, errors.New("Invalid assignment statement on line " + line)
	}
	return AssignmentStatement{
		tokens[0].LineNumber,
		tokens[0].Column,
		exprs[:len(exprs)-1],
		exprs[len(exprs)-1]}, idx, nil
}

func parseLocals(tokens []Token) (LocalsStatement, int, error) {
	line := tokens[0].LineNumber
	lineStr := strconv.Itoa(tokens[0].LineNumber)
	idx := 1
	if tokens[idx].Type != Space {
		return LocalsStatement{}, 0, errors.New("Expecting space on line " + lineStr)
	}
	idx++
	var locals []Variable
	for idx < len(tokens) {
		token := tokens[idx]
		if token.Type == IdentifierWord {
			idx++
			if tokens[idx].Type != Space {
				return LocalsStatement{}, 0, errors.New("Expecting space on line " + lineStr)
			}
			idx++
			dataType, n, err := parseType(tokens[idx:], line)
			if err != nil {
				return LocalsStatement{}, 0, err
			}
			idx += n
			locals = append(locals, Variable{line, token.Column, token.Content, dataType})
		} else if token.Type == Space && tokens[idx+1].Type != Newline {
			idx++
		} else {
			break
		}
	}
	if tokens[idx].Type == Space {
		idx++
	}
	if tokens[idx].Type != Newline {
		return LocalsStatement{}, 0, errors.New("Expecting newline in locals statement on line " + lineStr)
	}
	idx++
	return LocalsStatement{tokens[0].LineNumber, tokens[0].Column, locals}, idx, nil
}

// expected to start with Indentation token.
// 'indentation' = the number of spaces indentation on which the body should be aligned
// May return zero statements if body is empty.
func parseBody(tokens []Token, indentation int) ([]Statement, int, error) {
	var statements []Statement
	i := 0
	for i < len(tokens) {
		t := tokens[i]
		if t.Type != Indentation { // gone past end of the body
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
					case "break":
						statement, numTokens, err = parseBreak(tokens[i:])
					case "continue":
						statement, numTokens, err = parseContinue(tokens[i:])
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
