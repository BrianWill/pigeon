package dynamicPigeon

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
				if strings.Contains(" \r\n)", string(current)) {
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
		} else if isAlpha(r) { // start of a word (_ is not a valid identifier character in Pigeon)
			endIdx := i + 1
			for {
				current := runes[endIdx]
				// loop will never run past end of runes because \n appended to end of file
				// A word should always end with space, newline, ., or )
				if strings.Contains(" \r\n).", string(current)) {
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
			case "func":
				definition, numTokens, err = parseFunction(tokens[i:], line, pkg)
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
	value, numValueTokens, err := parseExpression(tokens[idx:], line)
	if err != nil {
		return GlobalDefinition{}, 0, err
	}
	idx += numValueTokens
	if tokens[idx].Type != Newline {
		return GlobalDefinition{}, 0, msg(line, column, "Global not terminated with newline.")
	}
	idx++
	return GlobalDefinition{line, column, target.Content, value, pkg}, idx, nil
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
	case OpenParen:
		var err error
		expr, idx, err = parseOpenParen(tokens)
		if err != nil {
			return nil, 0, err
		}
	default:
		return nil, 0, msg(line, column, "Improper expression: "+fmt.Sprintf("%#v", token))
	}
	return expr, idx, nil
}

func parseOpenParen(tokens []Token) (Expression, int, error) {
	line := tokens[0].LineNumber
	column := tokens[0].Column
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
	if functionCall {
		if leadingCall == nil {
			expr = FunctionCall{line, column, op, arguments}
		} else {
			expr = FunctionCall{line, column, leadingCall, arguments}
		}
	} else {
		expr = Operation{line, column, op.Content, arguments}
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
	var params []string
	for true {
		if tokens[idx].Type == Space && tokens[idx+1].Type == IdentifierWord {
			params = append(params, tokens[idx+1].Content)
			idx += 2
		} else if tokens[idx].Type == Space && tokens[idx+1].Type == Newline {
			idx += 2
			break
		} else if tokens[idx].Type == Newline {
			idx++
			break
		} else {
			return FunctionDefinition{}, 0, msg(line, column, "Expecting parameter name or end of line.")
		}
	}
	body, nTokens, err := parseBody(tokens[idx:], indentationSpaces)
	if err != nil {
		return FunctionDefinition{}, 0, err
	}
	idx += nTokens
	return FunctionDefinition{
		line, column,
		name.Content,
		params,
		body,
		pkg,
	}, idx, nil
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
			tokens[idx+1].Content == "elif" {
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
		return ElseifClause{}, 0, msg(line, column, "Elif clause condition not followed by newline.")
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
	if tokens[idx].Type != IdentifierWord {
		return ForeachStatement{}, 0, msg(line, column, "Expecting identifier for the values in foreach.")
	}
	valName := tokens[idx].Content
	idx++
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
		indexName, valName,
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
			"forinc/fordec end expression not followed by newline.")
	}
	idx++
	body, numTokens, err := parseBody(tokens[idx:], indentation+indentationSpaces)
	if err != nil {
		return ForincStatement{}, 0, err
	}
	idx += numTokens
	return ForincStatement{
		line, column,
		indexName,
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
	// return statement defaults to returning nil
	if tokens[idx].Type == Newline {
		return ReturnStatement{line, column, Token{NilLiteral, "nil", line, tokens[idx].Column}}, 2, nil
	}
	if tokens[idx].Type == Space && tokens[idx+1].Type == Newline {
		return ReturnStatement{line, column, Token{NilLiteral, "nil", line, tokens[idx].Column}}, 3, nil
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
	if tokens[idx].Type != Newline {
		return ReturnStatement{}, 0, msg(line, column, "Missing newline.")
	}
	idx++
	return ReturnStatement{line, column, value}, idx, nil
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

func parseAssignment(tokens []Token) (AssignmentStatement, int, error) {
	line := tokens[0].LineNumber
	column := tokens[0].Column
	idx := 1
	if tokens[idx].Type != Space {
		return AssignmentStatement{}, 0, msg(line, column, "Missing space.")
	}
	idx++
	if tokens[idx].Type != IdentifierWord {
		return AssignmentStatement{}, 0, msg(line, column, "Expecting target of assignment.")
	}
	target := tokens[idx]
	idx++
	if tokens[idx].Type != Space {
		return AssignmentStatement{}, 0, msg(line, column, "Missing space.")
	}
	idx++
	expr, nTokens, err := parseExpression(tokens[idx:], line)
	if err != nil {
		return AssignmentStatement{}, 0, err
	}
	idx += nTokens
	if tokens[idx].Type == Space {
		idx++
	}
	if tokens[idx].Type != Newline {
		return AssignmentStatement{}, 0, msg(line, tokens[0].Column, "Missing newline at end of assignment.")
	}
	idx++
	return AssignmentStatement{line, column, target.Content, expr}, idx, nil
}

func parseLocals(tokens []Token) (LocalsStatement, int, error) {
	line := tokens[0].LineNumber
	column := tokens[0].Column
	idx := 1
	var locals []string
	for true {
		if tokens[idx].Type == Space && tokens[idx+1].Type == IdentifierWord {
			locals = append(locals, tokens[idx+1].Content)
			idx += 2
		} else if tokens[idx].Type == Space && tokens[idx+1].Type == Newline {
			idx += 2
			break
		} else if tokens[idx].Type == Newline {
			idx++
			break
		} else {
			return LocalsStatement{}, 0, msg(line, column, "Expecting local variable name or end of line.")
		}
	}
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
					case "return":
						statement, numTokens, err = parseReturn(tokens[i:])
					case "forinc":
						statement, numTokens, err = parseForinc(tokens[i:], indentation, false)
					case "fordec":
						statement, numTokens, err = parseForinc(tokens[i:], indentation, true)
					case "break":
						statement, numTokens, err = parseBreak(tokens[i:])
					case "continue":
						statement, numTokens, err = parseContinue(tokens[i:])
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
					numTokens++ // include the newline
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
