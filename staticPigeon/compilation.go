package staticPigeon

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/davecgh/go-spew/spew"
)

/* All identifiers get prefixed with _ to avoid collisions with Go reserved words and predefined identifiers */
// returns map of valid breakpoints
func compile(definitions []Definition) (string, map[string]bool, error) {
	globals := make(map[string]GlobalDefinition)
	structs := make(map[string]StructDefinition)
	funcs := make(map[string]FunctionDefinition)
	methods := make(map[string]MethodDefinition)
	interfaces := make(map[string]InterfaceDefinition)
	imports := []ImportDefinition{}
	types := map[string]DataType{}
	packageNames := map[string]bool{}
	for _, def := range definitions {
		switch d := def.(type) {
		case GlobalDefinition:
			if packageNames[d.Name] {
				return "", nil, errors.New("Duplicate top-level name: " + d.Name)
			}
			globals[d.Name] = d
			packageNames[d.Name] = true
		case FunctionDefinition:
			if packageNames[d.Name] {
				return "", nil, errors.New("Duplicate top-level name: " + d.Name)
			}
			funcs[d.Name] = d
			packageNames[d.Name] = true
		case StructDefinition:
			if packageNames[d.Name] {
				return "", nil, errors.New("Duplicate top-level name: " + d.Name)
			}
			structs[d.Name] = d
			types[d.Name] = d
			packageNames[d.Name] = true
		case MethodDefinition:
			if packageNames[d.Name] {
				return "", nil, errors.New("Duplicate top-level name: " + d.Name)
			}
			methods[d.Name] = d
			packageNames[d.Name] = true
		case InterfaceDefinition:
			if packageNames[d.Name] {
				return "", nil, errors.New("Duplicate top-level name: " + d.Name)
			}
			interfaces[d.Name] = d
			types[d.Name] = d
			packageNames[d.Name] = true
		case ImportDefinition:
			name := ""
			for i, v := range d.Names {
				name = v
				if d.Aliases[i] != "" {
					name = d.Aliases[i]
				}
			}
			if packageNames[name] {
				return "", nil, errors.New("Duplicate top-level name: " + name)
			}
			imports = append(imports, d)
			packageNames[name] = true
		default:
			return "", nil, errors.New("Unrecognized definition")
		}
	}

	code := `package main

import _p "github.com/BrianWill/pigeon/stdlib"

var _breakpoints = make(map[int]bool)

`
	validBreakpoints := map[string]bool{}

	c, err := compileGlobals(globals, types, validBreakpoints)
	if err != nil {
		return "", nil, err
	}
	code += c

	code += `

func main() {
	go _p.PollBreakpoints(&_breakpoints)
	_main()
}
`

	return code, validBreakpoints, nil
}

func compileType(dataType DataType) (string, error) {
	return "", nil
}

// assumes both are valid types and that all type names are unique
func isType(child DataType, parent DataType, exact bool) bool {
	if child == parent {
		return true
	}
	switch c := child.(type) {
	case InterfaceDefinition:
		switch p := parent.(type) {
		case InterfaceDefinition:
			return c.Name == p.Name
		}
	case StructDefinition:
		switch p := parent.(type) {
		case InterfaceDefinition:
			if exact {
				return false
			}
			// return true if the child implements the parent interface
			for _, v := range c.Implements {
				if v == p.Name {
					return true
				}
			}
			return false
		case StructDefinition:
			return c.Name == p.Name
		}
	case BuiltinType:
		switch p := parent.(type) {
		case BuiltinType:
			if c.Name != p.Name || len(c.Params) != len(p.Params) {
				return false
			}
			for i := range c.Params {
				if !isType(c.Params[i], p.Params[i], true) {
					return false
				}
			}
			return true
		}
	}
	return false
}

func canonicalType(parsed ParsedDataType, types map[string]DataType) (DataType, error) {
	params := make([]DataType, len(parsed.Params))
	for i, v := range parsed.Params {
		t, err := canonicalType(v, types)
		if err != nil {
			return nil, err
		}
		params[i] = t
	}
	returnTypes := make([]DataType, len(parsed.ReturnTypes))
	for i, v := range parsed.ReturnTypes {
		t, err := canonicalType(v, types)
		if err != nil {
			return nil, err
		}
		returnTypes[i] = t
	}
	switch parsed.Type {
	case "F":
		return FunctionType{params, returnTypes}, nil
	case "L":
		if len(params) != 1 {
			return nil, errors.New("List type has wrong number of type parameters.")
		}
		return BuiltinType{"L", params}, nil
	case "M":
		if len(params) != 2 {
			return nil, errors.New("Map type has wrong number of type parameters.")
		}
		return BuiltinType{"M", params}, nil
	case "N", "Str", "Bool":
		if len(params) != 0 {
			return nil, errors.New("Type " + parsed.Type + " should not have any type parameters.")
		}
		return BuiltinType{parsed.Type, params}, nil
	default:
		t, ok := types[parsed.Type]
		if !ok {
			return nil, errors.New("Unknown type.")
		}
		if len(parsed.Params) > 0 || len(parsed.ReturnTypes) > 0 {
			return nil, errors.New("Type " + parsed.Type + " should not have any type parameters.")
		}
		return t, nil
	}
}

func compileExpression(e Expression, locals map[string]Variable, globals map[string]GlobalDefinition,
	types map[string]DataType) (string, []DataType, error) {
	var code string
	var returnedType DataType
	var err error
	switch e := e.(type) {
	case Operation:
		code, returnedType, err = compileOperation(e, locals, globals, types)
		if err != nil {
			return "", nil, err
		}
	case FunctionCall:
		code, returnedType, err = compileFunctionCall(e, locals, globals, types)
		if err != nil {
			return "", nil, err
		}
	case Token:
		switch e.Type {
		case IdentifierWord:
			name := e.Content
			if v, ok := locals[name]; ok {
				code = name
				returnedType, err = canonicalType(v.Type, types)
				if err != nil {
					return "", nil, err
				}
			} else if v, ok := globals[name]; ok {
				code = "g_" + name
				returnedType, err = canonicalType(v.Type, types)
				if err != nil {
					return "", nil, err
				}
			} else {
				return "", nil, fmt.Errorf("Name %s on line %d is undefined.", name, e.LineNumber)
			}
		case NumberLiteral:
			code = "float64(" + e.Content + ")"
			returnedType = BuiltinType{"N", nil}
		case StringLiteral:
			code = e.Content
			returnedType = BuiltinType{"Str", nil}
		case BooleanLiteral:
			code = e.Content
			returnedType = BuiltinType{"Bool", nil}
		case NilLiteral:
			code = "_p.Nil(0)"
		}
	}
	return code, returnedType, nil
}

func compileGlobals(globals map[string]GlobalDefinition, types map[string]DataType,
	validBreakpoints map[string]bool) (string, error) {
	code := ""
	for _, g := range globals {
		code += "var " + g.Name + " "
		t, err := canonicalType(g.Type, types)
		if err != nil {
			return "", err
		}
		c, err := compileType(t)
		if err != nil {
			return "", err
		}
		code += c
		c, returnedType, err := compileExpression(g.Value, map[string]Variable{}, globals, types)
		if err != nil {
			return "", err
		}
		if t != returnedType {
			return "", errors.New("Initial value of global does not match the declared type.")
		}
		code += c + "\n"
		validBreakpoints[strconv.Itoa(g.LineNumber)] = true
	}
	return "", nil
}

// // returns code snippet ending with '\n\n'
func compileFunc(fn FunctionDefinition, globals map[string]GlobalDefinition,
	types map[string]DataType, validBreakpoints map[string]bool) (string, error) {
	locals := map[string]Variable{}

	header := "func " + fn.Name + "("
	for _, param := range fn.Parameters {
		header += param.Name + " interface{}, "
		locals[param.Name] = param
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
	body, err := compileBody(bodyStatements, fn, locals, globals, types, validBreakpoints)
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

func genDebugFn(locals map[string]Variable, globals map[string]GlobalDefinition) string {
	s := `debug := func(line int) {
	var globals = map[string]interface{}{
`
	for k := range globals {
		s += fmt.Sprintf("\"%s\": g_%s,\n", k, k)
	}
	s += `}
	var locals = map[string]interface{}{
`
	for k := range locals {
		s += fmt.Sprintf("\"%s\": %s,\n", k, k)
	}
	s += `}
	_p.PollContinue(line, globals, locals)
}
`
	return s
}

func compileIfStatement(s IfStatement, expectedReturnTypes []DataType, locals map[string]Variable,
	globals map[string]GlobalDefinition, types map[string]DataType,
	validBreakpoints map[string]bool) (string, error) {
	c, returnedType, err := compileExpression(s.Condition, locals, globals, types)
	if err != nil {
		return "", err
	}
	code := "if " + c + ".(bool)"
	c, err = compileBody(s.Body, expectedReturnTypes, locals, globals, types, validBreakpoints)
	if err != nil {
		return "", nil
	}
	code += " {\n" + c + "}"
	for _, elif := range s.Elifs {
		c, returnedType, err := compileExpression(elif.Condition, locals, globals, types)
		if err != nil {
			return "", err
		}
		if !isType(returnedType, BuiltinType{"Bool", nil}, true) {
			return "", errors.New("Elif condition expression does not return a boolean on line " + strconv.Itoa(elif.LineNumber))
		}
		code += " else if " + c + ".(bool) {\n"
		c, err = compileBody(elif.Body, expectedReturnTypes, locals, globals, types, validBreakpoints)
		if err != nil {
			return "", err
		}
		code += c + "}"
	}
	if len(s.Else.Body) > 0 {
		c, err := compileBody(s.Else.Body, expectedReturnTypes, locals, globals, types, validBreakpoints)
		if err != nil {
			return "", err
		}
		code += " else {\n" + c + "}"
	}
	return code + "\n", nil
}

func compileWhileStatement(s WhileStatement, expectedReturnTypes []DataType, locals map[string]Variable,
	globals map[string]GlobalDefinition, types map[string]DataType, validBreakpoints map[string]bool) (string, error) {
	c, returnedType, err := compileExpression(s.Condition, locals, globals, types)
	if err != nil {
		return "", err
	}
	if !isType(returnedType, BuiltinType{"Bool", nil}, true) {
		return "", errors.New("while condition expression does not return a boolean on line " + strconv.Itoa(s.LineNumber))
	}
	code := "for " + c + ".(bool) {\n"
	c, err = compileBody(s.Body, expectedReturnTypes, locals, globals, types, validBreakpoints)
	if err != nil {
		return "", err
	}
	return code + c + "}\n", nil
}

func compileBody(statements []Statement, expectedReturnTypes []DataType, locals map[string]Variable,
	globals map[string]GlobalDefinition, types map[string]DataType,
	validBreakpoints map[string]bool) (string, error) {
	var code string
	for _, s := range statements {
		line := s.Line()
		validBreakpoints[strconv.Itoa(line)] = true
		code += fmt.Sprintf("if _breakpoints[%d] {debug(%d)}\n", line, line)
		var c string
		var err error
		var returnedType DataType
		switch s := s.(type) {
		case IfStatement:
			c, err = compileIfStatement(s, expectedReturnTypes, locals, globals, types, validBreakpoints)
		case WhileStatement:
			c, err = compileWhileStatement(s, expectedReturnTypes, locals, globals, types, validBreakpoints)
		case AssignmentStatement:
			c, err = compileAssignmentStatement(s, locals, globals, types)
		case ReturnStatement:
			c, err = compileReturnStatement(s, expectedReturnTypes, locals, globals, types)
		case FunctionCall:
			c, returnedType, err = compileFunctionCall(s, locals, globals, types)
			c += "\n"
		case Operation:
			c, returnedType, err = compileOperation(s, locals, globals, types)
			c += "\n"
		}
		if err != nil {
			return "", err
		}
		code += c
	}
	return code, nil
}

func compileAssignmentStatement(s AssignmentStatement, locals map[string]Variable, globals map[string]GlobalDefinition,
	types map[string]DataType) (string, error) {
	switch target := s.Target.(type) {
	case Token:
		lineStr := strconv.Itoa(target.LineNumber)
		if target.Type != IdentifierWord {
			return "", errors.New("Assignment to non-identifier on line " + lineStr)
		}
		name := target.Content
		local, isLocal := locals[name]
		global, isGlobal := globals[name]
		if !isLocal && !isGlobal {
			return "", errors.New("Assignment to non-existent variable on line " + lineStr)
		}
		c, returnedType, err := compileExpression(s.Value, locals, globals, types)
		if err != nil {
			return "", err
		}
		// TODO mutiple assignment
		var parsedType ParsedDataType
		if isLocal {
			parsedType = local.Type
		} else if isGlobal {
			parsedType = global.Type

		}
		dataType, err := canonicalType(parsedType, types)
		if err != nil {
			return "", err
		}
		if !isType(dataType, returnedType, false) {
			return "", errors.New("Value in assignment does not match expected type on line " + lineStr)
		}
		return target.Content + " = " + c + "\n", nil
	case Operation:
		if target.Operator != "get" {
			return "", errors.New("Improper target of assignment on line " + strconv.Itoa(target.LineNumber))
		}
		// turn the get op into a set op
		target.Operator = "set"
		target.Operands = append(target.Operands, s.Value)
		c, returnedType, err := compileExpression(target, locals, globals, types)
		if err != nil {
			return "", err
		}
		return c + "\n", nil
	default:
		// TODO give Expression LineNumber() method so we can get a line number here
		return "", errors.New("Invalid target of assignment.")
	}
}

func compileReturnStatement(s ReturnStatement, expectedReturnTypes []DataType, locals map[string]Variable,
	globals map[string]GlobalDefinition, types map[string]DataType) (string, error) {
	c, dataType, err := compileExpression(s.Value, locals, globals, types)
	if err != nil {
		return "", err
	}
	if !isType(dataType, expectedReturnTypes[0], false) {
		return "", errors.New("Wrong data type for return on " + strconv.Itoa(s.LineNumber))
	}
	return "return " + c + "\n", nil
}

func compileFunctionCall(s FunctionCall, locals map[string]Variable, globals map[string]GlobalDefinition,
	types map[string]DataType) (string, DataType, error) {
	var code string
	var returnedType DataType
	var c string
	var err error
	switch s := s.Function.(type) {
	case Operation:
		c, returnedType, err = compileOperation(s, locals, globals, types)
		if err != nil {
			return "", nil, err
		}
		code += c
	case FunctionCall:
		c, returnedType, err = compileFunctionCall(s, locals, globals, types)
		if err != nil {
			return "", nil, err
		}
		code += c
		// TODO have to assert type of function
	case Token: // will always be an identifier
		code += s.Content
	}
	code += "(" // start of arguments
	for _, exp := range s.Arguments {
		c, returnedType, err = compileExpression(exp, locals, globals, types)
		if err != nil {
			return "", nil, err
		}
		code += c + ", " // Go is OK with comma after last arg, so don't need special case for last arg
	}
	if len(s.Arguments) > 0 {
		code = code[:len(code)-2] // drop last comma and space
	}
	return code + ")", returnedType, nil
}

func compileOperation(o Operation, locals map[string]Variable, globals map[string]GlobalDefinition,
	types map[string]DataType) (string, DataType, error) {
	operandCode := make([]string, len(o.Operands))
	operandTypes := make([]DataType, len(o.Operands))
	for i, expr := range o.Operands {
		c, returnType, err := compileExpression(expr, locals, globals, types)
		if err != nil {
			return "", nil, err
		}
		operandCode[i] = c
		operandTypes[i] = returnType
	}
	code := "("
	var returnType DataType
	switch o.Operator {
	case "add":
		if len(o.Operands) < 2 {
			return "", nil, errors.New("add operations requires at least two operands")
		}
		returnType = BuiltinType{"N", nil}
		for i := range o.Operands {
			if !isType(operandTypes[i], returnType, true) {
				return "", nil, errors.New("add operation has non-number operand")
			}
			code += operandCode[i]
			if i < len(o.Operands)-1 {
				code += " + "
			}
		}
	case "sub":
		if len(o.Operands) < 2 {
			return "", nil, errors.New("sub operation requires at least two operands")
		}
		returnType = BuiltinType{"N", nil}
		for i := range o.Operands {
			if !isType(operandTypes[i], returnType, true) {
				return "", nil, errors.New("sub operation has non-number operand")
			}
			code += operandCode[i]
			if i < len(o.Operands)-1 {
				code += " - "
			}
		}
	case "mul":
		if len(o.Operands) < 2 {
			return "", nil, errors.New("mul operation requires at least two operands")
		}
		returnType = BuiltinType{"N", nil}
		for i := range o.Operands {
			if !isType(operandTypes[i], returnType, true) {
				return "", nil, errors.New("mul operation has non-number operand")
			}
			code += operandCode[i]
			if i < len(o.Operands)-1 {
				code += " * "
			}
		}
	case "div":
		if len(o.Operands) < 2 {
			return "", nil, errors.New("div operation requires at least two operands")
		}
		returnType = BuiltinType{"N", nil}
		for i := range o.Operands {
			if !isType(operandTypes[i], returnType, true) {
				return "", nil, errors.New("div operation has non-number operand")
			}
			code += operandCode[i]
			if i < len(o.Operands)-1 {
				code += " / "
			}
		}
	case "mod":
		if len(o.Operands) != 2 {
			return "", nil, errors.New("mod operation requires two operands")
		}
		returnType = BuiltinType{"N", nil}
		for i := range o.Operands {
			if !isType(operandTypes[i], returnType, true) {
				return "", nil, errors.New("mod operation has non-number operand")
			}
			code += "int64(" + operandCode[i] + ")"
			if i < len(o.Operands)-1 {
				code += " % "
			}
		}
	case "eq":
		if len(o.Operands) < 2 {
			return "", nil, errors.New("eq operation requires at least two operands")
		}
		returnType = BuiltinType{"Bool", nil}
		for i := 0; i < len(o.Operands)-1; i++ {
			if !isType(operandTypes[i], operandTypes[0], true) ||
				!isType(operandTypes[i+1], operandTypes[0], true) {
				return "", nil, errors.New("eq operation has mismatched operand types")
			}
			if i > 0 {
				code += " && "
			}
			code += operandCode[i] + " == " + operandCode[i+1]
		}
	case "neq":
		if len(o.Operands) < 2 {
			return "", nil, errors.New("neq operation requires at least two operands")
		}
		returnType = BuiltinType{"Bool", nil}
		for i := 0; i < len(o.Operands)-1; i++ {
			if !isType(operandTypes[i], operandTypes[0], true) ||
				!isType(operandTypes[i+1], operandTypes[0], true) {
				return "", nil, errors.New("neq operation has mismatched operand types")
			}
			if i > 0 {
				code += " && "
			}
			code += operandCode[i] + " != " + operandCode[i+1]
		}
	case "not":
		if len(o.Operands) != 1 {
			return "", nil, errors.New("not operation requires one operand")
		}
		returnType = BuiltinType{"Bool", nil}
		if !isType(operandTypes[0], returnType, true) {
			return "", nil, errors.New("not operation has a non-bool operand")
		}
		code += "!" + operandCode[0]
	case "lt":
		if len(o.Operands) < 2 {
			return "", nil, errors.New("lt operation requires at least two operands")
		}
		returnType = BuiltinType{"Bool", nil}
		for i := 0; i < len(o.Operands)-1; i++ {
			if !isType(operandTypes[i], BuiltinType{"N", nil}, true) ||
				!isType(operandTypes[i+1], BuiltinType{"N", nil}, true) {
				return "", nil, errors.New("lt operation has non-number operand")
			}
			if i > 0 {
				code += " && "
			}
			code += operandCode[i] + " < " + operandCode[i+1]
		}
	case "gt":
		if len(o.Operands) < 2 {
			return "", nil, errors.New("gt operation requires at least two operands")
		}
		returnType = BuiltinType{"Bool", nil}
		for i := 0; i < len(o.Operands)-1; i++ {
			if !isType(operandTypes[i], BuiltinType{"N", nil}, true) ||
				!isType(operandTypes[i+1], BuiltinType{"N", nil}, true) {
				return "", nil, errors.New("gt operation has non-number operand")
			}
			if i > 0 {
				code += " && "
			}
			code += operandCode[i] + " > " + operandCode[i+1]
		}
	case "lte":
		if len(o.Operands) < 2 {
			return "", nil, errors.New("lte operation requires at least two operands")
		}
		returnType = BuiltinType{"Bool", nil}
		for i := 0; i < len(o.Operands)-1; i++ {
			if !isType(operandTypes[i], BuiltinType{"N", nil}, true) ||
				!isType(operandTypes[i+1], BuiltinType{"N", nil}, true) {
				return "", nil, errors.New("lte operation has non-number operand")
			}
			if i > 0 {
				code += " && "
			}
			code += operandCode[i] + " <= " + operandCode[i+1]

		}
	case "gte":
		if len(o.Operands) < 2 {
			return "", nil, errors.New("gte operation requires at least two operands")
		}
		returnType = BuiltinType{"Bool", nil}
		for i := 0; i < len(o.Operands)-1; i++ {
			if !isType(operandTypes[i], BuiltinType{"N", nil}, true) ||
				!isType(operandTypes[i+1], BuiltinType{"N", nil}, true) {
				return "", nil, errors.New("gte operation has non-number operand")
			}
			if i > 0 {
				code += " && "
			}
			code += operandCode[i] + " >= " + operandCode[i+1]
		}
	case "get":
		if len(o.Operands) != 2 {
			return "", nil, errors.New("get operation requires two operands")
		}
		t, ok := operandTypes[0].(BuiltinType)
		if !ok || (t.Name != "N" && t.Name != "M") {
			return "", nil, errors.New("get operation requires a list or map as first operand")
		}
		switch t.Name {
		case "M":
			returnType = t.Params[1]
			if !isType(operandTypes[1], t.Params[0], true) {
				return "", nil, errors.New("get operation on map has wrong type as second operand")
			}
		case "L":
			returnType = t.Params[0]
			if !isType(operandTypes[1], BuiltinType{"N", nil}, true) {
				return "", nil, errors.New("get operation requires a number as second operand")
			}
		}
		code += operandCode[0] + "[" + operandCode[1] + "]"
	case "set":
		if len(o.Operands) != 3 {
			return "", nil, errors.New("set operation requires three operands")
		}
		t, ok := operandTypes[0].(BuiltinType)
		if !ok || (t.Name != "N" && t.Name != "M") {
			return "", nil, errors.New("set operation requires a list or map as first operand")
		}
		switch t.Name {
		case "M":
			if !isType(operandTypes[1], t.Params[0], true) {
				return "", nil, errors.New("set operation on map has wrong type as second operand")
			}
			if !isType(operandTypes[2], t.Params[1], false) {
				return "", nil, errors.New("set operation on map has wrong type as third operand")
			}
		case "L":
			if !isType(operandTypes[1], BuiltinType{"N", nil}, true) {
				return "", nil, errors.New("set operation requires a number as second operand")
			}
			if !isType(operandTypes[2], t.Params[0], false) {
				return "", nil, errors.New("set operation on list has wrong type as third operand")
			}
		}
		returnType = nil
		code += "func () {" + operandCode[0] + "[" + operandCode[1] + "] = " + operandCode[2] + "}()"
	case "append":
	case "or":
		if len(o.Operands) < 2 {
			return "", nil, errors.New("or operation requires at least two operands")
		}
		returnType = BuiltinType{"Bool", nil}
		for i := range o.Operands {
			if !isType(operandTypes[i], returnType, true) {
				return "", nil, errors.New("or operation has non-boolean operand")
			}
			code += operandCode[i]
			if i < len(o.Operands)-1 {
				code += " || "
			}
		}
	case "and":
		if len(o.Operands) < 2 {
			return "", nil, errors.New("and operation requires at least two operands")
		}
		returnType = BuiltinType{"Bool", nil}
		for i := range o.Operands {
			if !isType(operandTypes[i], returnType, true) {
				return "", nil, errors.New("and operation has non-boolean operand")
			}
			code += operandCode[i]
			if i < len(o.Operands)-1 {
				code += " && "
			}
		}
	case "print":
	case "prompt":
	case "concat":
		if len(o.Operands) < 2 {
			return "", nil, errors.New("concat operation requires at least two operands")
		}
		returnType = BuiltinType{"Str", nil}
		for i := range o.Operands {
			if !isType(operandTypes[i], returnType, true) {
				return "", nil, errors.New("concat operation has non-string operand")
			}
			code += operandCode[i]
			if i < len(o.Operands)-1 {
				code += " + "
			}
		}
	case "list":
	case "map":
	case "len":
	}

	code += ")"
	return code, returnType, nil
}

// func Highlight(code []byte) ([]byte, error) {
// 	return highlight.AsHTML(code, highlight.OrderedList())
// }

// func CompileAndRun(filename string) (*exec.Cmd, error) {
// 	filename, _, err := Compile(filename)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return Run(filename)
// }

// func Run(filename string) (*exec.Cmd, error) {
// 	cmd := exec.Command("go", "run", filename)
// 	cmd.Stdin = os.Stdin
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr
// 	err := cmd.Start()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return cmd, nil
// }

// returns map of valid breakpoints
func Compile(inputFilename string) (string, map[string]bool, error) {
	data, err := ioutil.ReadFile(inputFilename)
	if err != nil {
		return "", nil, err
	}
	tokens, err := lex(string(data) + "\n")
	if err != nil {
		return "", nil, err
	}
	fmt.Println("tokens: ", tokens)
	definitions, err := parse(tokens)
	if err != nil {
		return "", nil, err
	}
	code, validBreakpoints, err := compile(definitions)
	if err != nil {
		return "", nil, err
	}
	spew.Dump("breakpoints", validBreakpoints)
	spew.Dump("compiled", code)
	return "", nil, nil
	// outputFilename := outputDir + "/" + inputFilename + ".go"
	// err = ioutil.WriteFile(outputFilename, []byte(code), os.ModePerm)
	// if err != nil {
	// 	return "", nil, err
	// }
	// err = exec.Command("go", "fmt", outputFilename).Run()
	// if err != nil {
	// 	return "", nil, err
	// }
	// return outputFilename, validBreakpoints, nil
}
