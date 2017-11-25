package dynamicPigeon

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

/* All identifiers get prefixed with _ to avoid collisions with Go reserved words and predefined identifiers */
// returns map of valid breakpoints
func compile(pkg *Package, outputDir string) error {
	code := "package main\n"

	code += `import _fmt "fmt"
import _log "log"
import _std "github.com/BrianWill/pigeon/dynamicPigeon/stdlib"
`

	c, err := compileGlobals(pkg)
	if err != nil {
		return err
	}
	code += c
	for _, fn := range pkg.Funcs {
		if fn.Pkg != pkg {
			continue
		}
		c, err := compileFunc(fn)
		if err != nil {
			return err
		}
		code += c
	}

	code += `
		
		func main() {
			_fmt.Println()
			_main()
		}
		`

	pkg.Code = code
	return nil
}

func compileExpression(e Expression, pkg *Package, locals map[string]string) (string, error) {
	var code string
	var err error
	switch e := e.(type) {
	case Operation:
		code, err = compileOperation(e, pkg, locals)
		if err != nil {
			return "", err
		}
	case FunctionCall:
		code, err = compileFunctionCall(e, pkg, locals)
		if err != nil {
			return "", err
		}
	case Token:
		switch e.Type {
		case IdentifierWord:
			name := e.Content
			if _, ok := locals[name]; ok {
				code = name
			} else if v, ok := pkg.Globals[name]; ok {
				if v.Pkg == pkg {
					code = "G_" + name
				}
			} else if v, ok := pkg.Funcs[name]; ok {
				if v.Pkg == pkg {
					code = name
				}
				code = name
			} else {
				return "", msg(e.LineNumber, e.Column, "Name is undefined.")
			}
		case NumberLiteral:
			code = "float64(" + e.Content + ")"
		case StringLiteral:
			code = e.Content
		case BooleanLiteral:
			code = e.Content
		case NilLiteral:
			code = "_std.Nil(0)"
		}
	}
	return code, nil
}

func compileGlobals(pkg *Package) (string, error) {
	code := ""
	for _, g := range pkg.Globals {
		if g.Pkg != pkg {
			continue
		}
		code += "var G_" + g.Name + " interface{} = "
		c, err := compileExpression(g.Value, pkg, map[string]string{})
		if err != nil {
			return "", err
		}
		code += c + "\n"
		pkg.ValidBreakpoints[strconv.Itoa(g.LineNumber)] = true
	}
	return code, nil
}

// returns code snippet ending with '\n\n'
func compileFunc(fn FunctionDefinition) (string, error) {
	locals := map[string]string{}
	header := "func " + strings.Title(fn.Name) + "(_params ...interface{}) interface{} { \n"
	header += "if len(_params) != " + strconv.Itoa(len(fn.Parameters)) + ` {
	_log.Fatalln("Call to function ` + fn.Name + ` has the wrong number of arguments.")
	}
	`
	for i, param := range fn.Parameters {
		header += "var " + param + " interface{} = _params[" + strconv.Itoa(i) + "]\n"
		locals[param] = param
	}
	if len(fn.Body) < 1 {
		return "", msg(fn.LineNumber, fn.Column, "Function should contain at least one statement.")
	}
	bodyStatements := fn.Body
	// account for locals statement
	if localsStatement, ok := bodyStatements[0].(LocalsStatement); ok {
		for _, v := range localsStatement.Vars {
			header += "var "
			if _, ok := locals[v]; ok {
				return "", msg(localsStatement.LineNumber, localsStatement.Column, "Local variable "+v+" is already defined as a parameter.")
			}
			locals[v] = v
			header += v + " interface{}\n"
			header += "_std.NullOp(" + v + ")\n"
		}
		bodyStatements = bodyStatements[1:]
	}
	header += genDebugFn(locals, fn.Pkg.Globals, fn.Pkg)
	body, err := compileBody(bodyStatements, fn.Pkg, locals, false)
	if err != nil {
		return "", err
	}
	return header + body + "\nreturn nil\n}\n", nil
}

func genDebugFn(locals map[string]string, globals map[string]GlobalDefinition, pkg *Package) string {
	s := `debug := func(line int) {
	var globals = map[string]interface{}{
`
	for k := range globals {
		s += fmt.Sprintf("\"%s\": G_%s,\n", k, k)
	}
	s += `}
	var locals = map[string]interface{}{
`
	for k := range locals {
		s += fmt.Sprintf("\"%s\": %s,\n", k, k)
	}
	s += `}
	_fmt.Println(globals, locals)
	//_p.PollContinue(line, globals, locals)
}
`
	return s
}

func compileIfStatement(s IfStatement, pkg *Package, locals map[string]string, insideLoop bool) (string, error) {
	c, err := compileExpression(s.Condition, pkg, locals)
	if err != nil {
		return "", err
	}

	code := `
	{
		_cond, _ok := (` + c + `).(bool)
		if !_ok {
			_log.Fatalln("If condition must be a boolean.")
		}
		if _cond {
	`
	c, err = compileBody(s.Body, pkg, locals, insideLoop)
	if err != nil {
		return "", nil
	}
	code += c
	for _, elif := range s.Elifs {
		c, err := compileExpression(elif.Condition, pkg, locals)
		if err != nil {
			return "", err
		}

		code += `
		} else {
			_cond, _ok := (` + c + `).(bool)
			if !_ok {
				_log.Fatalln("Elif condition must be a boolean.")
			}
			if _cond {
		`
		c, err = compileBody(elif.Body, pkg, locals, insideLoop)
		if err != nil {
			return "", err
		}
		code += c
	}
	if len(s.Else.Body) > 0 {
		code += `
		} else {	
		`
		c, err := compileBody(s.Else.Body, pkg, locals, insideLoop)
		if err != nil {
			return "", err
		}
		code += c
	}
	for i := 0; i < 2+len(s.Elifs); i++ {
		code += `
			}
		`
	}
	return code, nil
}

func compileWhileStatement(s WhileStatement, pkg *Package, locals map[string]string) (string, error) {
	c, err := compileExpression(s.Condition, pkg, locals)
	if err != nil {
		return "", err
	}
	code := `
	for {
		_cond, _ok := (` + c + `).(bool)
		if !_ok {
			_log.Fatalln("While loop condition must be a boolean.")
		}
		if !_cond {
			break
		}
	`
	c, err = compileBody(s.Body, pkg, locals, true)
	if err != nil {
		return "", err
	}
	return code + c + "\n}\n", nil
}

func compileForincStatement(s ForincStatement, pkg *Package, locals map[string]string) (string, error) {
	if _, ok := locals[s.IndexName]; ok {
		return "", msg(s.LineNumber, s.Column, "forinc index name conflicts with an existing local variable.")
	}
	newLocals := map[string]string{}
	for k, v := range locals {
		newLocals[k] = v
	}
	newLocals[s.IndexName] = s.IndexName
	startExpr, err := compileExpression(s.StartVal, pkg, newLocals)
	if err != nil {
		return "", err
	}
	endExpr, err := compileExpression(s.EndVal, pkg, newLocals)
	if err != nil {
		return "", err
	}
	code := `{
	_start, _ok := (interface{}(` + startExpr + `)).(float64)
	if !_ok {
		panic("Forinc/fordec start value is not a number.")
	}
	_end, _ok := (interface{}(` + endExpr + `)).(float64)
	if !_ok {
		panic("Forinc/fordec end value is not a number.")
	}
	`
	if s.Dec {
		code += "_start-- \n"
	}

	code += `for ` + s.IndexName + ` := _start; ` + s.IndexName
	if s.Dec {
		code += " >= "
	} else {
		code += " < "
	}
	code += " _end; " + s.IndexName
	if s.Dec {
		code += "--"
	} else {
		code += "++"
	}
	code += " { \n"

	//code += genDebugFn(newLocals, pkg.Globals, pkg)
	body, err := compileBody(s.Body, pkg, newLocals, true)
	if err != nil {
		return "", err
	}
	code += body + "}\n}\n"
	return code, nil
}

func compileForeachStatement(s ForeachStatement, pkg *Package, locals map[string]string) (string, error) {
	if _, ok := locals[s.IndexName]; ok {
		return "", msg(s.LineNumber, s.Column, "foreach index name conflicts with an existing local variable.")
	}
	if _, ok := locals[s.ValName]; ok {
		return "", msg(s.LineNumber, s.Column, "foreach val name conflicts with an existing local variable.")
	}
	if s.IndexName == s.ValName {
		return "", msg(s.LineNumber, s.Column, "foreach index name conflicts with val name.")
	}
	collExpr, err := compileExpression(s.Collection, pkg, locals)
	if err != nil {
		return "", err
	}
	newLocals := map[string]string{}
	for k, v := range locals {
		newLocals[k] = v
	}
	newLocals[s.IndexName] = s.IndexName
	newLocals[s.ValName] = s.ValName

	body, err := compileBody(s.Body, pkg, newLocals, true)
	if err != nil {
		return "", err
	}

	code := `
		switch _c := ` + collExpr + `.(type) {
		case _std.ListType:
			for _i, _v := range *_c.List {
`
	code += s.IndexName + " := interface{}(float64(_i))\n"
	code += s.ValName + " := interface{}(_v)\n"
	code += "_std.NullOp(" + s.IndexName + ")\n"
	code += "_std.NullOp(" + s.ValName + ")\n"
	code += body
	code += `
			}
		case _std.MapType:
			for _k, _v := range _c {
`
	code += s.IndexName + " := interface{}(_k)\n"
	code += s.ValName + " := interface{}(_v)\n"
	code += "_std.NullOp(" + s.IndexName + ")\n"
	code += "_std.NullOp(" + s.ValName + ")\n"
	code += body
	code += `
			}
		default:
			_log.Fatalln("Foreach collection must be a list or map.")
		}
	
	`

	return code, nil
}

func compileBody(statements []Statement, pkg *Package, locals map[string]string, insideLoop bool) (string, error) {
	var code string
	for _, s := range statements {
		line := s.Line()
		lineStr := strconv.Itoa(line)
		pkg.ValidBreakpoints[lineStr] = true
		code += fmt.Sprintf("if _std.Breakpoints[%d] {debug(%d)}\n", line, line)
		var c string
		var err error
		switch s := s.(type) {
		case IfStatement:
			c, err = compileIfStatement(s, pkg, locals, insideLoop)
		case WhileStatement:
			c, err = compileWhileStatement(s, pkg, locals)
		case ForeachStatement:
			c, err = compileForeachStatement(s, pkg, locals)
		case ForincStatement:
			c, err = compileForincStatement(s, pkg, locals)
		case AssignmentStatement:
			c, err = compileAssignmentStatement(s, pkg, locals)
		case ReturnStatement:
			c, err = compileReturnStatement(s, pkg, locals)
		case BreakStatement:
			if insideLoop {
				c += "break \n"
			} else {
				err = msg(s.LineNumber, s.Column, "cannot have break statement outside a loop.")
			}
		case ContinueStatement:
			if insideLoop {
				c += "continue \n"
			} else {
				err = msg(s.LineNumber, s.Column, "cannot have continue statement outside a loop.")
			}
		case FunctionCall:
			c, err = compileFunctionCall(s, pkg, locals)
			c += "\n"
		case Operation:
			if s.Operator != "set" && s.Operator != "print" && s.Operator != "println" &&
				s.Operator != "prompt" && s.Operator != "push" {
				return "", msg(s.LineNumber, s.Column, "Improper operation as statement. Only set, push, print, println, "+
					"and prompt can be standalone statements.")
			}
			c, err = compileOperation(s, pkg, locals)
			c += "\n"
		}
		if err != nil {
			return "", err
		}
		code += c
	}
	return code, nil
}

func compileAssignmentStatement(s AssignmentStatement, pkg *Package, locals map[string]string) (string, error) {
	valCode, err := compileExpression(s.Value, pkg, locals)
	if err != nil {
		return "", err
	}
	return s.Target + " = " + valCode + "\n", nil
}

func compileReturnStatement(s ReturnStatement, pkg *Package, locals map[string]string) (string, error) {
	c, err := compileExpression(s.Value, pkg, locals)
	if err != nil {
		return "", err
	}
	return "return " + c + "\n", nil
}

func compileFunctionCall(s FunctionCall, pkg *Package, locals map[string]string) (string, error) {
	code := ""
	switch s := s.Function.(type) {
	case Operation:
		c, err := compileOperation(s, pkg, locals)
		if err != nil {
			return "", err
		}
		code += c
	case FunctionCall:
		c, err := compileFunctionCall(s, pkg, locals)
		if err != nil {
			return "", err
		}
		code += c
	case Token: // will always be an identifier
		if _, ok := locals[s.Content]; ok {
			code += s.Content
		} else {
			// previous check means we don't have to check for zero val
			if _, ok := pkg.Funcs[s.Content]; !ok {
				return "", msg(s.LineNumber, s.Column, "calling non-existent function.")
			}
			code += strings.Title(s.Content)
		}
	}
	code += "(" // start of arguments
	for _, exp := range s.Arguments {
		c, err := compileExpression(exp, pkg, locals)
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

func compileOperation(o Operation, pkg *Package, locals map[string]string) (string, error) {
	operandCode := make([]string, len(o.Operands))
	for i, expr := range o.Operands {
		c, err := compileExpression(expr, pkg, locals)
		if err != nil {
			return "", err
		}
		operandCode[i] = c
	}
	code := "(_std." + strings.Title(o.Operator) + "("
	for _, expr := range o.Operands {
		c, err := compileExpression(expr, pkg, locals)
		if err != nil {
			return "", err
		}
		code += c + ", "
	}
	return code + "))", nil
}

func Compile(filename string, outputDir string) (*Package, error) {
	path, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	tokens, err := lex(string(data) + "\r\n")
	if err != nil {
		return nil, err
	}

	pkg := &Package{
		Globals:          map[string]GlobalDefinition{},
		ValidBreakpoints: map[string]bool{},
		Funcs:            map[string]FunctionDefinition{},
	}
	definitions, err := parse(tokens, pkg)
	if err != nil {
		return nil, err
	}

	packageNames := map[string]bool{}
	for _, def := range definitions {
		switch d := def.(type) {
		case GlobalDefinition:
			un := strings.ToUpper(d.Name)
			if packageNames[un] {
				return nil, msg(d.LineNumber, d.Column, "Duplicate top-level name: "+d.Name)
			}
			pkg.Globals[d.Name] = d
			packageNames[un] = true
		case FunctionDefinition:
			un := strings.ToUpper(d.Name)
			if packageNames[un] {
				return nil, msg(d.LineNumber, d.Column, "Duplicate top-level name: "+d.Name)
			}
			pkg.Funcs[d.Name] = d
			packageNames[un] = true
		default:
			return nil, errors.New("Unrecognized definition")
		}
	}
	err = compile(pkg, outputDir)
	if err != nil {
		return nil, err
	}
	return pkg, nil
}
