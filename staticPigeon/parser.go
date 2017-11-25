package staticPigeon

import (
	"fmt"
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
			return nil, msg(line, column, "File improperly contains a non-ASCII character.")
		}
		if r == '\n' {
			tokens = append(tokens, Token{Newline, "\n", line, column})
			line++
			column = 1
			i++
		} else if r == '\r' {
			if runes[i+1] != '\n' {
				return nil, msg(line, column, "Improper newline: expecting LF (linefeed) after CR (carriage return).")
			}
			tokens = append(tokens, Token{Newline, "\n", line, column})
			line++
			column = 1
			i += 2
		} else if r == '/' { // start of a comment
			if runes[i+1] != '/' {
				return nil, msg(line, column, "Expected second / (slash).")
			}
			for runes[i] != '\n' && runes[i] != '\r' {
				i++
			}
			i++
			if runes[i] == '\n' { // LF after CR
				i++
			}
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
		} else if r == ',' {
			tokens = append(tokens, Token{Comma, ",", line, column})
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
			return nil, msg(line, column, "File improperly contains a tab character.")
		} else if r == '"' { // start of a string
			prev := r
			endIdx := i + 1
			for {
				current := runes[endIdx]
				// loop will never run past end of runes because \n appended to end of file
				if current == '\n' || current == '\r' {
					return nil, msg(line, column, "String literal not closed.")
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
				if strings.Contains("> \r\n)]", string(current)) {
					break
				} else if current == '.' {
					if decimalPointIdx != -1 {
						return nil, msg(line, column, "Number literal has more than one decimal point.")
					}
					decimalPointIdx = endIdx
				} else if !isNumeral(current) {
					return nil, msg(line, column, "Number literal not properly formed.")
				}
				endIdx++
			}

			if decimalPointIdx == endIdx {
				return nil, msg(line, column, "Number literal should not end with decimal point.")
			}

			tokens = append(tokens, Token{NumberLiteral, string(runes[i:endIdx]), line, column})
			column += (endIdx - i)
			i = endIdx
		} else if r == '\'' { // start of a multi-line string
			if runes[i+1] != '\'' && runes[i+2] != '\'' {
				return nil, msg(line, column, "Single quotes must come in threes to start multi-line string.")
			}
			column += 3
			endIdx := i + 3
			for {
				if endIdx >= len(runes) {
					return nil, msg(line, column, "Multi-line string is never closed.")
				}
				if runes[endIdx] == '\'' && runes[endIdx+1] == '\'' && runes[endIdx+2] == '\'' {
					endIdx += 3
					column += 3
					break
				}
				if runes[endIdx] == '\n' {
					line++
					column = 0
				} else {
					column++
				}
				endIdx++
			}
			tokens = append(tokens, Token{MultilineStringLiteral, string(runes[i:endIdx]), line, column})
			i = endIdx
		} else if isAlpha(r) { // start of a word (_ is not a valid identifier character in Pigeon)
			endIdx := i + 1
			for {
				current := runes[endIdx]
				// loop will never run past end of runes because \n appended to end of file
				// A word should always end with space, newline, <, >, ., [, or )
				if strings.Contains(" \r\n)<>.[", string(current)) {
					break
				} else if !(isAlpha(current) || isNumeral(current)) {
					return nil, msg(line, column, "Word improperly formed.")
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
			return nil, msg(line, column, "Unexpected character "+string(r)+".")
		}
	}

	// TODO filter out blank lines and spaces before newlines in one pass
	// filter out blank lines with indentation
	filteredTokens := []Token{}
	for i := 0; i < len(tokens)-1; {
		if tokens[i].Type == Indentation && tokens[i+1].Type == Newline {
			i += 2
		} else {
			filteredTokens = append(filteredTokens, tokens[i])
			i++
		}
	}
	tokens = append(filteredTokens, tokens[len(tokens)-1])

	// filter out blank lines
	filteredTokens = []Token{}
	for i := 0; i < len(tokens)-1; {
		if tokens[i].Type == Newline && tokens[i+1].Type == Newline {
			i++
		} else {
			filteredTokens = append(filteredTokens, tokens[i])
			i++
		}
	}
	tokens = append(filteredTokens, tokens[len(tokens)-1])

	// remove all spaces followed by newlines
	filteredTokens = []Token{}
	for i := 0; i < len(tokens); i++ {
		if tokens[i].Type == Space && tokens[i+1].Type == Newline {
			continue
		}
		filteredTokens = append(filteredTokens, tokens[i])
	}
	tokens = filteredTokens

	// remove all sequences of [newline -> indentation -> comma], replace with space
	if tokens[0].Type == Comma || tokens[1].Type == Comma {
		return nil, msg(line, column, "Unexpected comma at start of file.")
	}
	if tokens[len(tokens)-2].Type == Comma || tokens[len(tokens)-1].Type == Comma {
		return nil, msg(line, column, "Unexpected comma at end of file.")
	}
	filteredTokens = []Token{}
	for i := 0; i < len(tokens)-2; {
		// commas should only be encountered in this pattern
		if tokens[i].Type == Newline && tokens[i+1].Type == Indentation && tokens[i+2].Type == Comma {
			tokens[i].Type = Space
			tokens[i].Content = " "
			filteredTokens = append(filteredTokens, tokens[i])
			i += 3
			continue
		}
		if tokens[i].Type == Comma {
			return nil, msg(line, column, "Unexpected comma.")
		}
		filteredTokens = append(filteredTokens, tokens[i])
		i++
	}
	if len(tokens) > 2 {
		filteredTokens = append(filteredTokens, tokens[len(tokens)-2:]...)
	}
	return filteredTokens, nil
}

// parse the top-level definitions
func parse(tokens []Token, pkg *Package) ([]Definition, error) {
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
				definition, numTokens, err = parseImport(tokens[i:], line, pkg)
			case "nativeimport":
				definition, numTokens, err = parseNativeImport(tokens[i:], line, pkg)
			case "struct":
				definition, numTokens, err = parseStruct(tokens[i:], line, pkg)
			case "nativestruct":
				definition, numTokens, err = parseNativeStruct(tokens[i:], line, pkg)
			case "interface":
				definition, numTokens, err = parseInterface(tokens[i:], line, pkg)
			case "method":
				definition, numTokens, err = parseMethod(tokens[i:], line, pkg)
			case "func":
				definition, numTokens, err = parseFunction(tokens[i:], line, pkg)
			case "nativefunc":
				definition, numTokens, err = parseNativeFunc(tokens[i:], line, pkg)
			case "global":
				definition, numTokens, err = parseGlobal(tokens[i:], line, pkg)
			default:
				return nil, msg(t.LineNumber, t.Column, "Improper reserved word at top level of code.")
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
				return nil, msg(t.LineNumber, t.Column, "Improper indentation at top level of code.")
			}
		default:
			return nil, msg(t.LineNumber, t.Column, "Improper token at top level of code.")
		}
	}
	return definitions, nil
}

func parseImport(tokens []Token, line int, pkg *Package) (ImportDefinition, int, error) {
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type != Space {
		return ImportDefinition{}, 0, msg(line, column, "Expected space.")
	}
	idx++
	pathToken := tokens[idx]
	if pathToken.Type != StringLiteral {
		return ImportDefinition{}, 0, msg(line, column, "Expected string literal.")
	}
	path := strings.Trim(pathToken.Content, "\"")
	idx++
	if tokens[idx].Type == Space {
		idx++
	}
	if tokens[idx].Type != Newline {
		return ImportDefinition{}, 0, msg(line, column, "Expected newline.")
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
			return ImportDefinition{}, 0, msg(line, column, "Expected name to import.")
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
			return ImportDefinition{}, 0, msg(line, column, "Expected newline.")
		}
		idx++
	}
	if len(importedNames) == 0 {
		return ImportDefinition{}, 0, msg(line, column, "Import statement has no imported names.")
	}
	return ImportDefinition{line, column, path, importedNames, importedAliases, pkg}, idx, nil
}

func parseNativeImport(tokens []Token, line int, pkg *Package) (NativeImportDefinition, int, error) {
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type != Space {
		return NativeImportDefinition{}, 0, msg(line, column, "Expected space.")
	}
	idx++
	pathToken := tokens[idx]
	if pathToken.Type != StringLiteral {
		return NativeImportDefinition{}, 0, msg(line, column, "Expected string literal.")
	}
	path := strings.Trim(pathToken.Content, "\"")
	idx++
	alias := ""
	if tokens[idx].Type == Space && tokens[idx+1].Type == IdentifierWord {
		alias = tokens[idx+1].Content
		idx += 2
	}
	if tokens[idx].Type != Newline {
		return NativeImportDefinition{}, 0, msg(line, column, "Expected newline.")
	}
	idx++
	return NativeImportDefinition{line, column, path, alias, pkg}, idx, nil
}

func parseNativeStruct(tokens []Token, line int, pkg *Package) (StructDefinition, int, error) {
	column := tokens[0].Column
	st, idx, err := parseStruct(tokens, line, pkg)
	if err != nil {
		return StructDefinition{}, 0, err
	}
	if tokens[idx].Type != MultilineStringLiteral {
		return StructDefinition{}, 0, msg(line, column, "Expected multiline string.")
	}
	native := tokens[idx].Content
	st.NativeCode = native[3 : len(native)-3]
	idx++
	if tokens[idx].Type != Newline {
		return StructDefinition{}, 0, msg(line, column, "Expecting newline at end of nativestruct.")
	}
	return st, idx, nil
}

func parseStruct(tokens []Token, line int, pkg *Package) (StructDefinition, int, error) {
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type != Space {
		return StructDefinition{}, 0, msg(line, column, "Expected space.")
	}
	idx++
	name := tokens[idx]
	if name.Type != TypeName {
		return StructDefinition{}, 0, msg(line, column, "Expected name for struct.")
	}
	idx++
	for _, v := range builtinTypes {
		if name.Content == v {
			return StructDefinition{}, 0, msg(line, column, "Invalid struct name: cannot redefine builtin type "+
				name.Content+".")
		}
	}
	if tokens[idx].Type == Space {
		idx++
	}
	if tokens[idx].Type != Newline {
		return StructDefinition{}, 0, msg(line, tokens[idx].Column, "Expected newline.")
	}
	idx++

	members := []Variable{}
	for {
		line = tokens[idx].LineNumber
		if tokens[idx].Type != Indentation {
			break
		}
		idx++
		memberName := tokens[idx]
		if memberName.Type != IdentifierWord {
			return StructDefinition{}, 0, msg(line, column, "Expected struct member name.")
		}
		idx++
		if tokens[idx].Type != Space {
			return StructDefinition{}, 0, msg(line, column, "Expected space.")
		}
		idx++
		memberType, numTypeTokens, err := parseType(tokens[idx:], line)
		if err != nil {
			return StructDefinition{}, 0, err
		}
		idx += numTypeTokens
		members = append(members, Variable{memberName.LineNumber, memberName.Column, memberName.Content, memberType})
		if tokens[idx].Type == Space {
			idx++
		}
		if tokens[idx].Type != Newline {
			return StructDefinition{}, 0, msg(line, column, "Expected newline.")
		}
		idx++
	}
	return StructDefinition{line, column, name.Content, members, "", pkg}, idx, nil
}

func parseMethod(tokens []Token, line int, pkg *Package) (MethodDefinition, int, error) {
	column := tokens[0].Column
	funcDef, numTokens, err := parseFunction(tokens, line, pkg)
	if err != nil {
		return MethodDefinition{}, 0, err
	}
	if len(funcDef.Parameters) == 0 {
		return MethodDefinition{}, 0, msg(line, column, "Method must have a receiver parameter.")
	}
	return MethodDefinition{
		funcDef.LineNumber,
		funcDef.Column,
		funcDef.Name,
		funcDef.Parameters[0],
		funcDef.Parameters[1:],
		funcDef.ReturnTypes,
		funcDef.Body,
		pkg,
	}, numTokens, nil
}

// used by parseFunction
// consumes all tokens through end of line
func parseParameters(tokens []Token, line int) ([]Variable, []ParsedDataType, int, error) {
	column := tokens[0].Column
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
				return nil, nil, 0, msg(line, column, "Expecting space.")
			}
			idx++
			dataType, n, err := parseType(tokens[idx:], line)
			if err != nil {
				return nil, nil, 0, err
			}
			idx += n
			expectingSpace = true
			params = append(params, Variable{line, t.Column, t.Content, dataType})
		case Colon:
			if expectingSpace {
				return nil, nil, 0, msg(line, t.Column, "Expecting space.")
			}
			// don't inc idx
			break Loop
		case Newline:
			// don't inc idx
			break Loop
		default:
			return nil, nil, 0, msg(t.LineNumber, t.Column, "Unexpected token.")
		}
	}

	// optional colon and return types
	var returnTypes []ParsedDataType
	if tokens[idx].Type == Colon {
		idx++
		if tokens[idx].Type != Space {
			return nil, nil, 0, msg(line, tokens[idx].Column, "Expecting space.")
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
				return nil, nil, 0, msg(tokens[idx].LineNumber, tokens[idx].Column, "Expecting space.")
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
		return nil, nil, 0, msg(tokens[idx].LineNumber, tokens[idx].Column, "Expecting newline.")
	}
	idx++
	return params, returnTypes, idx, nil
}

// used by parseInterface, not parseFunction
// consumes all tokens through end of line
func parseSignature(tokens []Token, line int) (Signature, int, error) {
	column := tokens[0].Column
	paramTypes := []ParsedDataType{}
	returnTypes := []ParsedDataType{}
	idx := 0
	methodName := tokens[idx]
	if methodName.Type != IdentifierWord {
		return Signature{}, 0, msg(line, column, "Expecting method name.")
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
					return Signature{}, 0, msg(line, column, "Expecting space.")
				}
				// don't inc idx
				break Loop
			case Newline:
				break Loop
			default:
				return Signature{}, 0, msg(line, column, "Unexpected token.")
			}
		}
		// optional colon and return types
		if tokens[idx].Type == Colon {
			idx++
			if tokens[idx].Type != Space {
				return Signature{}, 0, msg(line, column, "Expecting space.")
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
					return Signature{}, 0, msg(line, column, "Expecting space.")
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
		return Signature{}, 0, msg(line, column, "Expecting newline.")
	}
	idx++
	return Signature{line, column, methodName.Content, paramTypes, returnTypes}, idx, nil
}

func parseInterface(tokens []Token, line int, pkg *Package) (InterfaceDefinition, int, error) {
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type != Space {
		return InterfaceDefinition{}, 0, msg(line, column, "Expected space.")
	}
	idx++
	name := tokens[idx]
	if name.Type != TypeName {
		return InterfaceDefinition{}, 0, msg(line, column, "Expected name for interface.")
	}
	idx++
	if tokens[idx].Type == Space {
		idx++
	}
	if tokens[idx].Type != Newline {
		return InterfaceDefinition{}, 0, msg(line, column, "Expected newline.")
	}
	idx++

	methods := []Signature{}
	for {
		if tokens[idx].Type != Indentation {
			break
		}
		idx++
		signature, numTokens, err := parseSignature(tokens[idx:], line)
		if err != nil {
			return InterfaceDefinition{}, 0, err
		}
		methods = append(methods, signature)
		idx += numTokens
	}
	if len(methods) == 0 {
		return InterfaceDefinition{}, 0, msg(line, column, "Interface definition has no method signatures.")
	}
	return InterfaceDefinition{line, column, name.Content, methods, pkg}, idx, nil
}

func parseType(tokens []Token, line int) (ParsedDataType, int, error) {
	column := tokens[0].Column
	idx := 0
	primary := tokens[idx]
	if primary.Type != TypeName {
		return ParsedDataType{}, 0, msg(line, column, "Expecting type name.")
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
			return ParsedDataType{line, tokens[idx].Column, primary.Content, paramTypes, returnTypes}, idx, nil
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
			if tokens[idx].Type == NumberLiteral {
				// special case for arrays (we expect a number literal, not just a constant expression)
				paramTypes = append(paramTypes, ParsedDataType{line, tokens[idx].Column, tokens[idx].Content, nil, nil})
				idx++
			} else {
				dataType, n, err := parseType(tokens[idx:], line)
				if err != nil {
					return ParsedDataType{}, 0, err
				}
				idx += n
				paramTypes = append(paramTypes, dataType)
			}
		}
		if tokens[idx].Type != CloseAngle {
			return ParsedDataType{}, 0, msg(line, column, "Expecting closing angle bracket.")
		}
		idx++
	}
	return ParsedDataType{line, tokens[idx].Column, primary.Content, paramTypes, returnTypes}, idx, nil
}

// expects to end with newline or >, but does not consume the newline or >
func parseReturnTypes(tokens []Token, line int) ([]ParsedDataType, int, error) {
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type != Space {
		return nil, 0, msg(line, column, "Expecting space.")
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
			return nil, 0, msg(line, column, "Expecting space.")
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

func parseGlobal(tokens []Token, line int, pkg *Package) (GlobalDefinition, int, error) {
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type != Space {
		return GlobalDefinition{}, 0, msg(line, column, "Expected space.")
	}
	idx++
	target := tokens[idx]
	if target.Type != IdentifierWord {
		return GlobalDefinition{}, 0, msg(line, column, "Improper name for a global.")
	}
	idx++
	if tokens[idx].Type != Space {
		return GlobalDefinition{}, 0, msg(line, column, "Expected space.")
	}
	idx++
	globalType, numTypeTokens, err := parseType(tokens[idx:], line)
	if err != nil {
		return GlobalDefinition{}, 0, err
	}
	idx += numTypeTokens
	if tokens[idx].Type != Space {
		return GlobalDefinition{}, 0, msg(line, column, "Expected space.")
	}
	idx++
	value, numValueTokens, err := parseExpression(tokens[idx:], line)
	if err != nil {
		return GlobalDefinition{}, 0, err
	}
	idx += numValueTokens
	if tokens[idx].Type != Newline {
		return GlobalDefinition{}, 0, msg(line, column, "Global not terminated with newline.")
	}
	idx++
	return GlobalDefinition{line, column, target.Content, value, globalType, pkg}, idx, nil
}

func parseGoStatement(tokens []Token) (GoStatement, int, error) {
	line := tokens[0].LineNumber
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type != Space {
		return GoStatement{}, 0, msg(line, column, "Expecting space in go statement.")
	}
	idx++
	expr, n, err := parseExpression(tokens[idx:], line)
	if err != nil {
		return GoStatement{}, 0, err
	}
	idx += n
	switch expr.(type) {
	case FunctionCall, MethodCall:
	default:
		return GoStatement{}, 0, msg(line, column, "Expecting function call or method call in go statement.")
	}
	if tokens[idx].Type != Newline {
		return GoStatement{}, 0, msg(line, column, "Expecting newline after go statement.")
	}
	idx++
	return GoStatement{line, column, expr}, idx, nil
}

func parseExpression(tokens []Token, line int) (Expression, int, error) {
	column := tokens[0].Column
	if len(tokens) < 1 {
		return nil, 0, msg(line, column, "Missing expression.")
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
	case TypeName:
		var err error
		var nTokens int
		expr, nTokens, err = parseType(tokens, line)
		if err != nil {
			return nil, 0, err
		}
		idx += nTokens
	case OpenParen:
		var err error
		expr, idx, err = parseOpenParen(tokens)
		if err != nil {
			return nil, 0, err
		}
	default:
		return nil, 0, msg(line, column, "Improper expression: "+fmt.Sprintf("%#v", token))
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
	column := tokens[0].Column
	if tokens[1].Type != IdentifierWord {
		return nil, 0, msg(line, column, "Identifier expected after dot.")
	}
	strLiteral := Token{StringLiteral, "\"" + tokens[1].Content + "\"", line, -1}
	getOp := Operation{
		line,
		column,
		"get",
		[]Expression{expr, strLiteral},
	}
	return getOp, 2, nil
}

// assumes first token is open square
func parseOpenSquare(tokens []Token, expr Expression, line int) (Expression, int, error) {
	column := tokens[0].Column
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
		return nil, 0, msg(line, column, "Improperly formed square brackets.")
	}
	idx++ // account for ']'
	getOp := Operation{
		line,
		column,
		"get",
		[]Expression{expr, indexExpr},
	}
	return getOp, idx, nil
}

// assumes first token is open paren.
// Returns a FunctionCall or Operation and the number of tokens that make up the Expression.
func parseOpenParen(tokens []Token) (Expression, int, error) {
	line := tokens[0].LineNumber
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type == Space {
		idx++
	}
	functionCall := true
	typeExpression := false
	methodCall := false
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
	case Dot:
		idx++
		if tokens[idx].Type != IdentifierWord {
			return nil, 0, msg(line, column, "Method call expects a method name after the dot.")
		}
		op = tokens[idx]
		methodCall = true
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
		return nil, 0, msg(line, column, "Improper function call or operation.")
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
		expr, numTokens, err := parseExpression(tokens[idx:], line)
		if err != nil {
			return nil, 0, err
		}
		arguments = append(arguments, expr)
		idx += numTokens
	}

	var expr Expression
	if methodCall {
		if len(arguments) < 1 {
			return nil, 0, msg(line, column, "Method call must have a receiver.")
		}
		expr = MethodCall{line, column, op.Content, arguments[0], arguments[1:]}
	} else if typeExpression {
		expr = TypeExpression{line, column, dt, arguments}
	} else if functionCall {
		if leadingCall == nil {
			expr = FunctionCall{line, column, op, arguments}
		} else {
			expr = FunctionCall{line, column, leadingCall, arguments}
		}
	} else {
		expr = Operation{line, column, op.Content, arguments}
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

func parseFunction(tokens []Token, line int, pkg *Package) (FunctionDefinition, int, error) {
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type == Space {
		idx++
	}
	name := tokens[idx]
	if name.Type != IdentifierWord {
		return FunctionDefinition{}, 0, msg(line, column, "Function missing name.")
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
			return FunctionDefinition{}, 0, msg(line, column, "Expecting space.")
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
	return FunctionDefinition{
		line, column,
		name.Content,
		params, returnTypes,
		body,
		"",
		pkg,
	}, idx, nil
}

func parseNativeFunc(tokens []Token, line int, pkg *Package) (FunctionDefinition, int, error) {
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type == Space {
		idx++
	}
	name := tokens[idx]
	if name.Type != IdentifierWord {
		return FunctionDefinition{}, 0, msg(line, column, "Native function missing name.")
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
			return FunctionDefinition{}, 0, msg(line, column, "Missing space.")
		}
		idx++
		var nTokens int
		params, returnTypes, nTokens, err = parseParameters(tokens[idx:], line)
		if err != nil {
			return FunctionDefinition{}, 0, err
		}
		idx += nTokens
	}
	if tokens[idx].Type != MultilineStringLiteral {
		return FunctionDefinition{}, 0, msg(line, column, "Native function expecting multi-line string.")
	}
	body := tokens[idx].Content
	body = body[3 : len(body)-3]
	idx++
	if tokens[idx].Type != Newline {
		return FunctionDefinition{}, 0, msg(line, column, "Native function expecting newline.")
	}
	idx++
	return FunctionDefinition{
		line, column,
		name.Content,
		params, returnTypes,
		nil,
		body,
		pkg,
	}, idx, nil
}

func parseTypeswitch(tokens []Token, indentation int) (TypeswitchStatement, int, error) {
	line := tokens[0].LineNumber
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type != Space {
		return TypeswitchStatement{}, 0, msg(line, column, "Missing space.")
	}
	idx++
	value, nTokens, err := parseExpression(tokens[idx:], line)
	if err != nil {
		return TypeswitchStatement{}, 0, err
	}
	idx += nTokens
	if tokens[idx].Type == Space {
		idx++
	}
	if tokens[idx].Type != Newline {
		return TypeswitchStatement{}, 0, msg(line, column, "Typeswitch expected newline.")
	}
	idx++
	var cases []TypeswitchCase
	for idx+1 < len(tokens) {
		if tokens[idx].Type == Indentation &&
			len(tokens[idx].Content) == indentation &&
			tokens[idx+1].Content == "case" {
			idx++
			c, nTokens, err := parseTypeswitchCase(tokens[idx:], indentation)
			if err != nil {
				return TypeswitchStatement{}, 0, err
			}
			cases = append(cases, c)
			idx += nTokens
		} else {
			break
		}
	}
	var defaultBody []Statement
	if idx+1 < len(tokens) {
		if tokens[idx].Type == Indentation &&
			len(tokens[idx].Content) == indentation &&
			tokens[idx+1].Content == "default" {
			idx++
			var nTokens int
			var err error
			defaultBody, nTokens, err = parseDefaultCase(tokens[idx:], indentation)
			if err != nil {
				return TypeswitchStatement{}, 0, err
			}
			idx += nTokens
		}
	}
	return TypeswitchStatement{line, column, value, cases, defaultBody}, idx, nil
}

func parseTypeswitchCase(tokens []Token, indentation int) (TypeswitchCase, int, error) {
	line := tokens[0].LineNumber
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type != Space {
		return TypeswitchCase{}, 0, msg(line, column, "Missing space.")
	}
	idx++
	if tokens[idx].Type != IdentifierWord {
		return TypeswitchCase{}, 0, msg(line, column, "Expecting identifier.")
	}
	name := tokens[idx].Content
	idx++
	if tokens[idx].Type != Space {
		return TypeswitchCase{}, 0, msg(line, column, "Missing space.")
	}
	idx++
	dt, nTokens, err := parseType(tokens[idx:], line)
	if err != nil {
		return TypeswitchCase{}, 0, err
	}
	idx += nTokens
	if tokens[idx].Type == Space {
		idx++
	}
	if tokens[idx].Type != Newline {
		return TypeswitchCase{}, 0, msg(line, column, "typeswitch case type not followed by newline.")
	}
	idx++
	body, numTokens, err := parseBody(tokens[idx:], indentation+indentationSpaces)
	if err != nil {
		return TypeswitchCase{}, 0, err
	}
	idx += numTokens
	v := Variable{line, column, name, dt}
	return TypeswitchCase{line, column, v, body}, idx, nil
}

func parseDefaultCase(tokens []Token, indentation int) ([]Statement, int, error) {
	line := tokens[0].LineNumber
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type == Space {
		idx++
	}
	if tokens[idx].Type != Newline {
		return nil, 0, msg(line, column, "Default case not followed by newline.")
	}
	idx++
	body, numTokens, err := parseBody(tokens[idx:], indentation+indentationSpaces)
	if err != nil {
		return nil, 0, err
	}
	idx += numTokens
	return body, idx, nil
}

func parseIf(tokens []Token, indentation int) (IfStatement, int, error) {
	line := tokens[0].LineNumber
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type != Space {
		return IfStatement{}, 0, msg(line, column, "Missing space.")
	}
	idx++
	condition, numConditionTokens, err := parseExpression(tokens[idx:], line)
	if err != nil {
		return IfStatement{}, 0, err
	}
	idx += numConditionTokens
	if tokens[idx].Type == Space {
		idx++
	}
	if tokens[idx].Type != Newline {
		return IfStatement{}, 0, msg(line, column, "If statement condition not followed by newline.")
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
		if tokens[idx].Type == Indentation && len(tokens[idx].Content) == indentation &&
			tokens[idx+1].Content == "elseif" {
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
	return IfStatement{line, column, condition, body, elseifClauses, elseClause}, idx, nil
}

func parseElif(tokens []Token, indentation int) (ElseifClause, int, error) {
	line := tokens[0].LineNumber
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type != Space {
		return ElseifClause{}, 0, msg(line, column, "Missing space.")
	}
	idx++
	condition, numConditionTokens, err := parseExpression(tokens[idx:], line)
	if err != nil {
		return ElseifClause{}, 0, msg(line, column, "Improper condition in if statement.")
	}
	idx += numConditionTokens
	if tokens[idx].Type == Space {
		idx++
	}
	if tokens[idx].Type != Newline {
		return ElseifClause{}, 0, msg(line, column, "Elseif clause condition not followed by newline.")
	}
	idx++
	body, numTokens, err := parseBody(tokens[idx:], indentation+indentationSpaces)
	if err != nil {
		return ElseifClause{}, 0, err
	}
	idx += numTokens
	return ElseifClause{line, column, condition, body}, idx, nil
}

func parseElse(tokens []Token, indentation int) (ElseClause, int, error) {
	line := tokens[0].LineNumber
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type == Space {
		idx++
	}
	if tokens[idx].Type != Newline {
		return ElseClause{}, 0, msg(line, column, "Else clause not followed by newline.")
	}
	idx++
	body, numTokens, err := parseBody(tokens[idx:], indentation+indentationSpaces)
	if err != nil {
		return ElseClause{}, 0, err
	}
	idx += numTokens
	return ElseClause{line, column, body}, idx, nil
}

func parseSelect(tokens []Token, indentation int) (SelectStatement, int, error) {
	line := tokens[0].LineNumber
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type != Newline {
		return SelectStatement{}, 0, msg(line, column, "Expecting newline.")
	}
	idx++
	var clauses []SelectClause
	var defaultClause SelectDefaultClause
loop:
	for idx < len(tokens) {
		if tokens[idx].Type == Indentation && len(tokens[idx].Content) == indentation {
			idx++
			if tokens[idx].Type != ReservedWord {
				return SelectStatement{}, 0, msg(line, column, "Expecting 'send' or 'rcv', or 'default'.")
			}
			switch tokens[idx].Content {
			case "sending":
				clause, n, err := parseSendClause(tokens[idx:], indentation)
				if err != nil {
					return SelectStatement{}, 0, err
				}
				clauses = append(clauses, clause)
				idx += n
			case "rcving":
				clause, n, err := parseRcvClause(tokens[idx:], indentation)
				if err != nil {
					return SelectStatement{}, 0, err
				}
				clauses = append(clauses, clause)
				idx += n
			case "default":
				var n int
				var err error
				defaultClause, n, err = parseSelectDefault(tokens[idx:], indentation)
				if err != nil {
					return SelectStatement{}, 0, err
				}
				idx += n
				break loop
			}
		} else {
			break
		}
	}
	return SelectStatement{line, column, clauses, defaultClause}, idx, nil
}

func parseSendClause(tokens []Token, indentation int) (SelectSendClause, int, error) {
	line := tokens[0].LineNumber
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type != Space {
		return SelectSendClause{}, 0, msg(line, column, "Missing space.")
	}
	idx++
	channelExpr, numConditionTokens, err := parseExpression(tokens[idx:], line)
	if err != nil {
		return SelectSendClause{}, 0, msg(line, column, "Improper channel expression in select send clause.")
	}
	idx += numConditionTokens
	if tokens[idx].Type != Space {
		return SelectSendClause{}, 0, msg(line, column, "Missing space.")
	}
	idx++
	valExpr, numConditionTokens, err := parseExpression(tokens[idx:], line)
	if err != nil {
		return SelectSendClause{}, 0, msg(line, column, "Improper value expression in select send clause.")
	}
	idx += numConditionTokens
	if tokens[idx].Type != Newline {
		return SelectSendClause{}, 0, msg(line, column, "Expecting newline in select send clause.")
	}
	idx++
	body, numTokens, err := parseBody(tokens[idx:], indentation+indentationSpaces)
	if err != nil {
		return SelectSendClause{}, 0, err
	}
	idx += numTokens
	return SelectSendClause{line, column, channelExpr, valExpr, body}, idx, nil
}

func parseRcvClause(tokens []Token, indentation int) (SelectRcvClause, int, error) {
	line := tokens[0].LineNumber
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type != Space {
		return SelectRcvClause{}, 0, msg(line, column, "Missing space.")
	}
	idx++
	if tokens[idx].Type != IdentifierWord {
		return SelectRcvClause{}, 0, msg(line, column, "Expecting variable name for target of select rcv clause.")
	}
	targetName := tokens[idx].Content
	idx++
	if tokens[idx].Type != Space {
		return SelectRcvClause{}, 0, msg(line, column, "Missing space.")
	}
	idx++
	dt, n, err := parseType(tokens[idx:], line)
	if err != nil {
		return SelectRcvClause{}, 0, err
	}
	idx += n
	if tokens[idx].Type != Space {
		return SelectRcvClause{}, 0, msg(line, column, "Missing space.")
	}
	idx++
	expr, n, err := parseExpression(tokens[idx:], line)
	if err != nil {
		return SelectRcvClause{}, 0, err
	}
	idx += n
	if tokens[idx].Type != Newline {
		return SelectRcvClause{}, 0, msg(line, column, "Expecting newline in select rcv clause.")
	}
	idx++
	body, numTokens, err := parseBody(tokens[idx:], indentation+indentationSpaces)
	if err != nil {
		return SelectRcvClause{}, 0, err
	}
	idx += numTokens
	return SelectRcvClause{
		line, column,
		Variable{line, column, targetName, dt},
		expr,
		body,
	}, idx, nil
}

func parseSelectDefault(tokens []Token, indentation int) (SelectDefaultClause, int, error) {
	line := tokens[0].LineNumber
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type != Newline {
		return SelectDefaultClause{}, 0, msg(line, column, "Elseif clause condition not followed by newline.")
	}
	idx++
	body, numTokens, err := parseBody(tokens[idx:], indentation+indentationSpaces)
	if err != nil {
		return SelectDefaultClause{}, 0, err
	}
	idx += numTokens
	return SelectDefaultClause{line, column, body}, idx, nil
}

func parseForeach(tokens []Token, indentation int) (ForeachStatement, int, error) {
	line := tokens[0].LineNumber
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type != Space {
		return ForeachStatement{}, 0, msg(line, column, "Missing space.")
	}
	idx++
	if tokens[idx].Type != IdentifierWord {
		return ForeachStatement{}, 0, msg(line, column, "Expecting identifier for the indexes in foreach.")
	}
	indexName := tokens[idx].Content
	idx++
	if tokens[idx].Type != Space {
		return ForeachStatement{}, 0, msg(line, column, "Missing space.")
	}
	idx++
	indexType, nTokens, err := parseType(tokens[idx:], line)
	if err != nil {
		return ForeachStatement{}, 0, err
	}
	idx += nTokens
	if tokens[idx].Type != Space {
		return ForeachStatement{}, 0, msg(line, column, "Missing space.")
	}
	idx++
	if tokens[idx].Type != IdentifierWord {
		return ForeachStatement{}, 0, msg(line, column, "Expecting identifier for the values in foreach.")
	}
	valName := tokens[idx].Content
	idx++
	if tokens[idx].Type != Space {
		return ForeachStatement{}, 0, msg(line, column, "Missing space.")
	}
	idx++
	valType, nTokens, err := parseType(tokens[idx:], line)
	if err != nil {
		return ForeachStatement{}, 0, err
	}
	idx += nTokens
	if tokens[idx].Type != Space {
		return ForeachStatement{}, 0, msg(line, column, "Missing space.")
	}
	idx++
	collection, nTokens, err := parseExpression(tokens[idx:], line)
	if err != nil {
		return ForeachStatement{}, 0, err
	}
	idx += nTokens
	if tokens[idx].Type != Newline {
		return ForeachStatement{}, 0, msg(line, column,
			"Foreach statement collection expression not followed by newline.")
	}
	idx++
	body, numTokens, err := parseBody(tokens[idx:], indentation+indentationSpaces)
	if err != nil {
		return ForeachStatement{}, 0, err
	}
	idx += numTokens
	return ForeachStatement{
		line, column,
		indexName, indexType,
		valName, valType,
		collection, body}, idx, nil
}

func parseForinc(tokens []Token, indentation int, isDec bool) (ForincStatement, int, error) {
	line := tokens[0].LineNumber
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type != Space {
		return ForincStatement{}, 0, msg(line, column, "Missing space.")
	}
	idx++
	if tokens[idx].Type != IdentifierWord {
		return ForincStatement{}, 0, msg(line, column, "Expecting identifier for the index in forinc.")
	}
	indexName := tokens[idx].Content
	idx++
	if tokens[idx].Type != Space {
		return ForincStatement{}, 0, msg(line, column, "Missing space.")
	}
	idx++
	indexType, nTokens, err := parseType(tokens[idx:], line)
	if err != nil {
		return ForincStatement{}, 0, err
	}
	idx += nTokens
	if tokens[idx].Type != Space {
		return ForincStatement{}, 0, msg(line, column, "Missing space.")
	}
	idx++
	startExpr, nTokens, err := parseExpression(tokens[idx:], line)
	if err != nil {
		return ForincStatement{}, 0, err
	}
	idx += nTokens
	if tokens[idx].Type != Space {
		return ForincStatement{}, 0, msg(line, column, "Missing space.")
	}
	idx++
	endExpr, nTokens, err := parseExpression(tokens[idx:], line)
	if err != nil {
		return ForincStatement{}, 0, err
	}
	idx += nTokens
	if tokens[idx].Type != Newline {
		return ForincStatement{}, 0, msg(line, column,
			"Foreach statement collection expression not followed by newline.")
	}
	idx++
	body, numTokens, err := parseBody(tokens[idx:], indentation+indentationSpaces)
	if err != nil {
		return ForincStatement{}, 0, err
	}
	idx += numTokens
	return ForincStatement{
		line, column,
		indexName, indexType,
		startExpr, endExpr,
		body, isDec}, idx, nil
}

func parseWhile(tokens []Token, indentation int) (WhileStatement, int, error) {
	line := tokens[0].LineNumber
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type != Space {
		return WhileStatement{}, 0, msg(line, column, "Missing space.")
	}
	idx++
	condition, nTokens, err := parseExpression(tokens[idx:], line)
	if err != nil {
		return WhileStatement{}, 0, err
	}
	idx += nTokens
	if tokens[idx].Type != Newline {
		return WhileStatement{}, 0, msg(line, column, "While statement condition not followed by newline.")
	}
	idx++
	body, numTokens, err := parseBody(tokens[idx:], indentation+indentationSpaces)
	if err != nil {
		return WhileStatement{}, 0, err
	}
	idx += numTokens
	return WhileStatement{line, column, condition, body}, idx, nil
}

func parseReturn(tokens []Token) (ReturnStatement, int, error) {
	line := tokens[0].LineNumber
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type != Space {
		return ReturnStatement{}, 0, msg(line, column, "Missing space.")
	}
	idx++
	value, nTokens, err := parseExpression(tokens[idx:], line)
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
			return ReturnStatement{}, 0, msg(line, column, "Missing space.")
		}
		idx++
		value, nTokens, err := parseExpression(tokens[idx:], line)
		if err != nil {
			return ReturnStatement{}, 0, err
		}
		idx += nTokens
		values = append(values, value)
	}
	return ReturnStatement{line, column, values}, idx, nil
}

func parseBreak(tokens []Token) (BreakStatement, int, error) {
	line := tokens[0].LineNumber
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type == Space {
		idx++
	}
	if tokens[idx].Type != Newline {
		return BreakStatement{}, 0, msg(line, column, "Break statement not terminated with newline.")
	}
	idx++
	return BreakStatement{line, column}, idx, nil
}

func parseContinue(tokens []Token) (ContinueStatement, int, error) {
	line := tokens[0].LineNumber
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type == Space {
		idx++
	}
	if tokens[idx].Type != Newline {
		return ContinueStatement{}, 0, msg(line, column, "Continue statement not terminated with newline.")
	}
	idx++
	return ContinueStatement{line, column}, idx, nil
}

// assume first token is reserved word "as"
// returns number of tokens (including the newline at the end)
func parseAssignment(tokens []Token) (AssignmentStatement, int, error) {
	line := tokens[0].LineNumber
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type != Space {
		return AssignmentStatement{}, 0, msg(line, column, "Missing space.")
	}
	idx++
	exprs := []Expression{}
	for {
		expr, nTokens, err := parseExpression(tokens[idx:], line)
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
			return AssignmentStatement{}, 0, msg(line, column, "Missing space.")
		}
		idx++
	}
	if len(exprs) < 2 {
		return AssignmentStatement{}, 0, msg(line, column, "Invalid assignment statement.")
	}
	return AssignmentStatement{
		tokens[0].LineNumber,
		tokens[0].Column,
		exprs[:len(exprs)-1],
		exprs[len(exprs)-1]}, idx, nil
}

func parseLocals(tokens []Token) (LocalsStatement, int, error) {
	line := tokens[0].LineNumber
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type != Space {
		return LocalsStatement{}, 0, msg(line, column, "Expecting space.")
	}
	idx++
	var locals []Variable
	for idx < len(tokens) {
		token := tokens[idx]
		if token.Type == IdentifierWord {
			idx++
			if tokens[idx].Type != Space {
				return LocalsStatement{}, 0, msg(line, column, "Expecting space.")
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
		return LocalsStatement{}, 0, msg(line, column, "Expecting newline in locals statement.")
	}
	idx++
	return LocalsStatement{tokens[0].LineNumber, tokens[0].Column, locals}, idx, nil
}

// exactly like parseFunction but with more indentation
func parseLocalFunc(tokens []Token, indentation int) (LocalFuncStatement, int, error) {
	line := tokens[0].LineNumber
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type == Space {
		idx++
	}
	name := tokens[idx]
	if name.Type != IdentifierWord {
		return LocalFuncStatement{}, 0, msg(line, column, "Local function missing name.")
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
			return LocalFuncStatement{}, 0, msg(line, column, "Expecting space.")
		}
		idx++
		var nTokens int
		params, returnTypes, nTokens, err = parseParameters(tokens[idx:], line)
		if err != nil {
			return LocalFuncStatement{}, 0, err
		}
		idx += nTokens
	}
	body, nTokens, err := parseBody(tokens[idx:], indentation+indentationSpaces)
	if err != nil {
		return LocalFuncStatement{}, 0, err
	}
	idx += nTokens
	return LocalFuncStatement{
		tokens[0].LineNumber, tokens[0].Column,
		name.Content,
		params, returnTypes,
		body,
	}, idx, nil
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
						return nil, 0, msg(t.LineNumber, t.Column, "Functions cannot be nested.")
					case "as":
						statement, numTokens, err = parseAssignment(tokens[i:])
					case "if":
						statement, numTokens, err = parseIf(tokens[i:], indentation)
					case "while":
						statement, numTokens, err = parseWhile(tokens[i:], indentation)
					case "foreach":
						statement, numTokens, err = parseForeach(tokens[i:], indentation)
					case "locals":
						statement, numTokens, err = parseLocals(tokens[i:])
					case "localfunc":
						statement, numTokens, err = parseLocalFunc(tokens[i:], indentation)
					case "return":
						statement, numTokens, err = parseReturn(tokens[i:])
					case "typeswitch":
						statement, numTokens, err = parseTypeswitch(tokens[i:], indentation)
					case "select":
						statement, numTokens, err = parseSelect(tokens[i:], indentation)
					case "forinc":
						statement, numTokens, err = parseForinc(tokens[i:], indentation, false)
					case "fordec":
						statement, numTokens, err = parseForinc(tokens[i:], indentation, true)
					case "break":
						statement, numTokens, err = parseBreak(tokens[i:])
					case "continue":
						statement, numTokens, err = parseContinue(tokens[i:])
					case "go":
						statement, numTokens, err = parseGoStatement(tokens[i:])
					default:
						return nil, 0, msg(t.LineNumber, t.Column, "Improper reserved word '"+t.Content+"' in body.")
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
						return nil, 0, msg(t.LineNumber, t.Column, "Statement not terminated with newline.")
					}
					numTokens++ // add in the newline
				default:
					return nil, 0, msg(t.LineNumber, t.Column, "Improper token. Expected start of statement.")
				}
				statements = append(statements, statement)
				i += numTokens
			} else {
				return nil, 0, msg(t.LineNumber, t.Column, "Improper indentation.")
			}
		}
	}
	return statements, i, nil
}
