package staticPigeon

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
)

/* All identifiers get prefixed with _ to avoid collisions with Go reserved words and predefined identifiers */
// returns map of valid breakpoints
func compile(definitions []Definition) (string, map[string]bool, error) {
	structs := map[string]StructDefinition{}
	funcs := map[string]FunctionDefinition{}
	methods := map[string]MethodDefinition{}
	interfaces := map[string]InterfaceDefinition{}
	imports := []ImportDefinition{}
	types := map[string]DataType{}
	packageNames := map[string]bool{}
	var ctx CodeContext
	ctx.Locals = map[string]Variable{}
	ctx.Globals = map[string]GlobalDefinition{}
	ctx.FuncTypes = map[string]FunctionType{}
	ctx.ValidBreakpoints = map[string]bool{}
	ctx.Types = types
	for _, def := range definitions {
		switch d := def.(type) {
		case GlobalDefinition:
			if packageNames[d.Name] {
				return "", nil, errors.New("Duplicate top-level name: " + d.Name)
			}
			ctx.Globals[d.Name] = d
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

import _fmt "fmt"

var _breakpoints = make(map[int]bool)

type _List []interface{}

func _newList(items ...interface{}) *_List {
	return &items
}

func (l *_List) append(item interface{}) {
	*l = append(*l, item)
}

func (l *_List) set(idx float64, item interface{}) {
	(*l)[int64(idx)] = item
}

func (l *_List) len() float64 {
	return float64(len(*l))
}

func _Prompt(args ...interface{}) {
	if len(args) > 1 {
		_fmt.Print(args...)
	}
	
}

`
	err := processStructs(structs, methods, types)
	if err != nil {
		return "", nil, err
	}
	c, err := compileInterfaces(interfaces, types)
	if err != nil {
		return "", nil, err
	}
	code += c
	for _, st := range structs {
		st := types[st.Name].(Struct)
		err := findImplementors(&st, interfaces, types)
		if err != nil {
			return "", nil, err
		}
		c, err := compileStruct(st, types)
		if err != nil {
			return "", nil, err
		}
		code += c
	}
	for _, fn := range funcs {
		fnType, err := getFunctionType(fn, types)
		if err != nil {
			return "", nil, err
		}
		ctx.FuncTypes[fn.Name] = fnType
	}
	c, err = compileGlobals(ctx)
	if err != nil {
		return "", nil, err
	}
	code += c
	for _, fn := range methods {
		c, err := compileMethod(fn, ctx)
		if err != nil {
			return "", nil, err
		}
		code += c
	}
	for _, fn := range funcs {
		c, err := compileFunc(fn, ctx)
		if err != nil {
			return "", nil, err
		}
		code += c
	}
	code += `

func main() {
	go _p.PollBreakpoints(&_breakpoints)
	_main()
}
`
	return code, ctx.ValidBreakpoints, nil
}

func findImplementors(st *Struct, interfaces map[string]InterfaceDefinition, types map[string]DataType) error {
Outer:
	for _, iface := range interfaces {
		for _, sig := range iface.Methods {
			ft, ok := st.Methods[sig.Name]
			if !ok {
				continue Outer
			}
			mt, err := sig.getFunctionType(types)
			if err != nil {
				return err
			}
			if !reflect.DeepEqual(ft, mt) {
				continue Outer
			}
		}
		st.Implements[iface.Name] = true
	}
	return nil
}

func compileStruct(st Struct, types map[string]DataType) (string, error) {
	code := "type " + st.Name + " struct {\n"
	for i, n := range st.MemberNames {
		t, err := compileType(st.MemberTypes[i])
		if err != nil {
			return "", err
		}
		code += n + " " + t + "\n"
	}
	return code + "}\n", nil
}

func compileInterfaces(interfaces map[string]InterfaceDefinition, types map[string]DataType) (string, error) {
	code := "\n"
	for _, inter := range interfaces {
		code += "type " + inter.Name + " interface {\n"
		for _, sig := range inter.Methods {
			// validate each method
			_, err := sig.getFunctionType(types)
			if err != nil {
				return "", err
			}
			code += sig.Name + "("
			for _, pt := range sig.ParamTypes {
				t, err := canonicalType(pt, types)
				if err != nil {
					return "", err
				}
				c, err := compileType(t)
				if err != nil {
					return "", err
				}
				code += c + ", "
			}
			code += ") ("
			for _, rt := range sig.ReturnTypes {
				t, err := canonicalType(rt, types)
				if err != nil {
					return "", err
				}
				c, err := compileType(t)
				if err != nil {
					return "", err
				}
				code += c + ", "
			}
			code += ")\n"
		}
		code += "}\n"
	}
	return code, nil
}

// replace all StructDefinition in types with Structs. Verifies that no Struct is illegally recursive
func processStructs(structs map[string]StructDefinition, methods map[string]MethodDefinition, types map[string]DataType) error {
	var processStruct func(StructDefinition, map[string]DataType, []string) (Struct, error)
	processStruct = func(st StructDefinition, types map[string]DataType, containingStructs []string) (Struct, error) {
		if _, ok := types[st.Name]; !ok {
			return Struct{}, fmt.Errorf("Invalid struct type on line %d.", st.LineNumber)
		}
		newStruct := Struct{
			LineNumber:  st.LineNumber,
			Column:      st.Column,
			Name:        st.Name,
			MemberNames: make([]string, len(st.Members)),
			MemberTypes: make([]DataType, len(st.Members)),
			Implements:  map[string]bool{},
			Methods:     map[string]FunctionType{},
		}
		for i, m := range st.Members {
			dt, err := canonicalType(m.Type, types)
			if err != nil {
				return Struct{}, err
			}
			newStruct.MemberNames[i] = m.Name
			newStruct.MemberTypes[i] = dt
			switch t := dt.(type) {
			case StructDefinition:
				for _, name := range containingStructs {
					if st.Name == name {
						return Struct{}, fmt.Errorf("Struct cannot recursively contain itself on line %d", st.LineNumber)
					}
				}
				st, err := processStruct(t, types, append(containingStructs, st.Name))
				if err != nil {
					return Struct{}, err
				}
				newStruct.MemberTypes[i] = st
			}
		}
		return newStruct, nil
	}
	for _, st := range structs {
		s, err := processStruct(st, types, []string{})
		if err != nil {
			return err
		}
		types[s.Name] = s
	}
	for _, meth := range methods {
		dt, err := canonicalType(meth.Receiver.Type, types)
		if err != nil {
			return err
		}
		if st, ok := dt.(Struct); ok {
			funcType, err := meth.getFunctionType(types)
			if err != nil {
				return err
			}
			st.Methods[meth.Name] = funcType
		} else {
			return errors.New("Method has non-struct receiver. Line " + strconv.Itoa(meth.LineNumber))
		}
	}
	return nil
}

func (m MethodDefinition) getFunctionType(types map[string]DataType) (FunctionType, error) {
	fd := FunctionDefinition{
		Parameters:  m.Parameters,
		ReturnTypes: m.ReturnTypes,
	}
	return getFunctionType(fd, types)
}

func getFunctionType(fn FunctionDefinition, types map[string]DataType) (FunctionType, error) {
	params := make([]DataType, len(fn.Parameters))
	for i, p := range fn.Parameters {
		dt, err := canonicalType(p.Type, types)
		if err != nil {
			return FunctionType{}, err
		}
		params[i] = dt
	}
	returnTypes := make([]DataType, len(fn.ReturnTypes))
	for i, rt := range fn.ReturnTypes {
		dt, err := canonicalType(rt, types)
		if err != nil {
			return FunctionType{}, err
		}
		returnTypes[i] = dt
	}
	return FunctionType{params, returnTypes}, nil
}

func (s Signature) getFunctionType(types map[string]DataType) (FunctionType, error) {
	params := make([]DataType, len(s.ParamTypes))
	for i, p := range s.ParamTypes {
		dt, err := canonicalType(p, types)
		if err != nil {
			return FunctionType{}, err
		}
		params[i] = dt
	}
	returnTypes := make([]DataType, len(s.ReturnTypes))
	for i, rt := range s.ReturnTypes {
		dt, err := canonicalType(rt, types)
		if err != nil {
			return FunctionType{}, err
		}
		returnTypes[i] = dt
	}
	return FunctionType{params, returnTypes}, nil
}

// assumes both are valid types and that all type names are unique
func isType(child DataType, parent DataType, exact bool) bool {
	if reflect.DeepEqual(child, parent) {
		return true
	}
	switch c := child.(type) {
	case InterfaceDefinition:
		switch p := parent.(type) {
		case InterfaceDefinition:
			return c.Name == p.Name
		}
	case Struct:
		switch p := parent.(type) {
		case InterfaceDefinition:
			if exact {
				return false
			}
			// return true if the child implements the parent interface
			for name, ok := range c.Implements {
				if ok && name == p.Name {
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
	case "P":
		if len(params) != 1 {
			return nil, errors.New("Pointer type has wrong number of type parameters.")
		}
		return BuiltinType{"P", params}, nil
	case "N", "Str", "Bool", "E", "Any":
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

// not the same as a 'type assertion'
// 'type expression' is parens starting with a type to create a value of that type
func compileTypeExpression(te TypeExpression, ctx CodeContext) (string, []DataType, error) {
	lineStr := strconv.Itoa(te.LineNumber)
	dt, err := canonicalType(te.Type, ctx.Types)
	if err != nil {
		return "", nil, err
	}
	switch t := dt.(type) {
	case BuiltinType:
		switch t.Name {
		case "M":
			if len(t.Params) != 2 {
				return "", nil, errors.New("Invalid type expression. Map must have two type parameters. Line " + lineStr)
			}
			mapType, err := compileType(t)
			if err != nil {
				return "", nil, err
			}
			mapType += "{"
			if len(te.Operands)%2 != 0 {
				return "", nil, errors.New("Invalid type expression. Map must have even number of operands. Line " + lineStr)
			}
			for i := 0; i < len(te.Operands); i += 2 {
				key, returnedTypes, err := compileExpression(te.Operands[i], ctx)
				if err != nil {
					return "", nil, err
				}
				if len(returnedTypes) != 1 || !isType(returnedTypes[0], t.Params[0], false) {
					return "", nil, errors.New("Invalid type expression. Map key of wrong type. Line " + lineStr)
				}
				val, returnedTypes, err := compileExpression(te.Operands[i+1], ctx)
				if err != nil {
					return "", nil, err
				}
				if len(returnedTypes) != 1 || !isType(returnedTypes[0], t.Params[1], false) {
					return "", nil, errors.New("Invalid type expression. Map val of wrong type. Line " + lineStr)
				}
				mapType += key + ": " + val + ", "
			}
			mapType += "}"
			return mapType, []DataType{t}, nil
		case "L":
			if len(t.Params) != 1 {
				return "", nil, errors.New("Invalid type expression. List must have one type parameter. Line " + lineStr)
			}
			expr := "(func () (_list _List) {\n"
			expr += "(*_list) = make([]interface{}, " + strconv.Itoa(len(te.Operands)) + ")\n"
			if err != nil {
				return "", nil, err
			}
			for i := 0; i < len(te.Operands); i++ {
				val, returnedTypes, err := compileExpression(te.Operands[i], ctx)
				if err != nil {
					return "", nil, err
				}
				if len(returnedTypes) != 1 || !isType(returnedTypes[0], t.Params[0], false) {
					return "", nil, errors.New("Invalid type expression. List val of wrong type. Line " + lineStr)
				}
				expr += "(*_list)[" + strconv.Itoa(i) + "] = " + val + "\n"
			}
			expr += `return
		    })()`
			return expr, []DataType{t}, nil
		case "N", "P", "Str", "Bool":
			return "", nil, errors.New("Invalid type expression. Cannot create an N, P, Str, or Bool. Line " + lineStr)
		}
	case FunctionType:
		return "", nil, errors.New("Invalid type expression. Cannot create a function with a type expression. Line " + lineStr)
	case Struct:
		if len(t.MemberNames) != len(te.Operands) {
			return "", nil, errors.New("Invalid type expression. Wrong number of args for creating struct. Line " + lineStr)
		}
		code := t.Name + "{"
		for i, argType := range t.MemberTypes {
			expr, returnTypes, err := compileExpression(te.Operands[i], ctx)
			if err != nil {
				return "", nil, err
			}
			if len(returnTypes) != 1 || !isType(returnTypes[0], argType, false) {
				return "", nil, errors.New("Invalid type expression. Wrong type of arg for creating struct. Line " + lineStr)
			}
			code += expr + ", "
		}
		code += "}"
		return code, []DataType{t}, nil
	case InterfaceDefinition:
		return "", nil, errors.New("Invalid type expression. Cannot create interface value. Line " + lineStr)
	}
	// should be unreachable
	return "", nil, errors.New("Invalid type expression. Line " + lineStr)
}

func compileExpression(e Expression, ctx CodeContext) (string, []DataType, error) {
	var code string
	var returnedTypes []DataType
	var err error
	switch e := e.(type) {
	case Operation:
		code, returnedTypes, err = compileOperation(e, ctx)
		if err != nil {
			return "", nil, err
		}
	case FunctionCall:
		code, returnedTypes, err = compileFunctionCall(e, ctx)
		if err != nil {
			return "", nil, err
		}
	case MethodCall:
		fmt.Println("compiling method ", e)
		code, returnedTypes, err = compileMethodCall(e, ctx)
		if err != nil {
			return "", nil, err
		}
	case TypeExpression:
		code, returnedTypes, err = compileTypeExpression(e, ctx)
		if err != nil {
			return "", nil, err
		}
	case ParsedDataType:
		dt, err := canonicalType(e, ctx.Types)
		if err != nil {
			return "", nil, err
		}
		code, err = compileType(dt)
		if err != nil {
			return "", nil, err
		}
		returnedTypes = []DataType{BuiltinType{"Type", nil}}
	case Token:
		switch e.Type {
		case IdentifierWord:
			name := e.Content
			if v, ok := ctx.Locals[name]; ok {
				code = name
				rt, err := canonicalType(v.Type, ctx.Types)
				if err != nil {
					return "", nil, err
				}
				returnedTypes = []DataType{rt}
			} else if v, ok := ctx.Globals[name]; ok {
				code = "g_" + name
				rt, err := canonicalType(v.Type, ctx.Types)
				if err != nil {
					return "", nil, err
				}
				returnedTypes = []DataType{rt}
			} else {
				return "", nil, fmt.Errorf("Name %s on line %d is undefined.", name, e.LineNumber)
			}
		case NumberLiteral:
			code = "float64(" + e.Content + ")"
			returnedTypes = []DataType{BuiltinType{"N", nil}}
		case StringLiteral:
			code = e.Content
			returnedTypes = []DataType{BuiltinType{"Str", nil}}
		case BooleanLiteral:
			code = e.Content
			returnedTypes = []DataType{BuiltinType{"Bool", nil}}
		case NilLiteral:
			code = "_p.Nil(0)"
		}
	}
	return code, returnedTypes, nil
}

func compileGlobals(ctx CodeContext) (string, error) {
	code := ""
	for _, g := range ctx.Globals {
		code += "var " + g.Name + " "
		t, err := canonicalType(g.Type, ctx.Types)
		if err != nil {
			return "", err
		}
		c, err := compileType(t)
		if err != nil {
			return "", err
		}
		code += c
		c, returnedTypes, err := compileExpression(g.Value, ctx)
		if err != nil {
			return "", err
		}
		if len(returnedTypes) != 1 {
			return "", errors.New("Initial value of global does not match the declared type.")
		}
		if t != returnedTypes[0] {
			return "", errors.New("Initial value of global does not match the declared type.")
		}
		code += c + "\n"
		ctx.ValidBreakpoints[strconv.Itoa(g.LineNumber)] = true
	}
	return code, nil
}

func isList(dt DataType) (DataType, bool) {
	t, ok := dt.(BuiltinType)
	if !ok || t.Name != "L" {
		return nil, false
	}
	return t.Params[0], true
}

func isMap(dt DataType) (DataType, DataType, bool) {
	t, ok := dt.(BuiltinType)
	if !ok || t.Name != "M" {
		return nil, nil, false
	}
	return t.Params[0], t.Params[1], true
}

// assumes a valid data type. Accepts Struct but not a StructDefinition
func compileType(dt DataType) (string, error) {
	switch t := dt.(type) {
	case BuiltinType:
		switch t.Name {
		case "N":
			return "float64", nil
		case "Bool":
			return "bool", nil
		case "Str":
			return "string", nil
		case "E":
			return "error", nil
		case "Any":
			return "interface{}", nil
		case "L":
			return "*_List", nil
		case "M":
			keyType, err := compileType(t.Params[0])
			if err != nil {
				return "", err
			}
			valType, err := compileType(t.Params[1])
			if err != nil {
				return "", err
			}
			return "map[" + keyType + "]" + valType, nil
		case "P":
			pointerType, err := compileType(t.Params[0])
			if err != nil {
				return "", err
			}
			return "*" + pointerType, nil
		}
	case InterfaceDefinition:
		return t.Name, nil
	case Struct:
		return t.Name, nil
	case FunctionType:
		typeStr := "func( "
		for _, paramType := range t.Params {
			s, err := compileType(paramType)
			if err != nil {
				return "", err
			}
			typeStr += s + ", "
		}
		typeStr += ") "
		if len(t.ReturnTypes) > 0 {
			typeStr += "("
			for _, returnType := range t.ReturnTypes {
				s, err := compileType(returnType)
				if err != nil {
					return "", err
				}
				typeStr += s + ", "
			}
			typeStr += ")"
		}
		return typeStr, nil
	case StructDefinition:
		panic("Invalid type")
	}

	return "", nil
}

// returns code snippet ending with '\n\n'
func compileFunc(fn FunctionDefinition, ctx CodeContext) (string, error) {
	ctx.Locals = map[string]Variable{}
	header := "func " + fn.Name + "("
	for i, param := range fn.Parameters {
		dt, err := canonicalType(param.Type, ctx.Types)
		if err != nil {
			return "", err
		}
		typeCode, err := compileType(dt)
		if err != nil {
			return "", err
		}
		header += param.Name + " " + typeCode
		if i < len(fn.Parameters)-1 {
			header += ", "
		}
		ctx.Locals[param.Name] = param
	}
	header += ") ("
	returnTypes := make([]DataType, len(fn.ReturnTypes))
	for i, rt := range fn.ReturnTypes {
		dt, err := canonicalType(rt, ctx.Types)
		if err != nil {
			return "", err
		}
		returnTypes[i] = dt
		typeCode, err := compileType(dt)
		if err != nil {
			return "", err
		}
		header += typeCode
		if i < len(fn.ReturnTypes)-1 {
			header += ", "
		}
	}
	header += ") {\n"
	if len(fn.Body) < 1 {
		return "", errors.New("Function should contain at least one statement.")
	}

	bodyStatements := fn.Body
	if localsStatement, ok := bodyStatements[0].(LocalsStatement); ok {
		for _, v := range localsStatement.Vars {
			header += "var "
			if _, ok := ctx.Locals[v.Name]; ok {
				return "", fmt.Errorf("Local variable %s on line %d is already defined as a parameter.",
					v.Name, v.LineNumber)
			}
			ctx.Locals[v.Name] = v
			dt, err := canonicalType(v.Type, ctx.Types)
			if err != nil {
				return "", err
			}
			typeCode, err := compileType(dt)
			if err != nil {
				return "", err
			}
			header += v.Name + " " + typeCode + "\n"
		}
		bodyStatements = bodyStatements[1:]
	}
	header += genDebugFn(ctx.Locals, ctx.Globals)
	body, err := compileBody(bodyStatements, returnTypes, ctx)
	if err != nil {
		return "", err
	}
	return header + body + "\n}\n", nil
}

// returns code snippet ending with '\n\n'
func compileMethod(meth MethodDefinition, ctx CodeContext) (string, error) {
	ctx.Locals = map[string]Variable{}
	dt, err := canonicalType(meth.Receiver.Type, ctx.Types)
	if err != nil {
		return "", err
	}
	receiverType, err := compileType(dt)
	if err != nil {
		return "", err
	}
	header := "func (" + meth.Receiver.Name + " " + receiverType + ") " + meth.Name + "("
	for i, param := range meth.Parameters {
		dt, err := canonicalType(param.Type, ctx.Types)
		if err != nil {
			return "", err
		}
		typeCode, err := compileType(dt)
		if err != nil {
			return "", err
		}
		header += param.Name + " " + typeCode
		if i < len(meth.Parameters)-1 {
			header += ", "
		}
		ctx.Locals[param.Name] = param
	}
	header += ") ("
	returnTypes := make([]DataType, len(meth.ReturnTypes))
	for i, rt := range meth.ReturnTypes {
		dt, err := canonicalType(rt, ctx.Types)
		if err != nil {
			return "", err
		}
		returnTypes[i] = dt
		typeCode, err := compileType(dt)
		if err != nil {
			return "", err
		}
		header += typeCode
		if i < len(meth.ReturnTypes)-1 {
			header += ", "
		}
	}
	header += ") {\n"
	if len(meth.Body) < 1 {
		return "", errors.New("Function should contain at least one statement.")
	}

	bodyStatements := meth.Body
	if localsStatement, ok := bodyStatements[0].(LocalsStatement); ok {
		for _, v := range localsStatement.Vars {
			header += "var "
			if _, ok := ctx.Locals[v.Name]; ok {
				return "", fmt.Errorf("Local variable %s on line %d is already defined as a parameter.",
					v.Name, v.LineNumber)
			}
			ctx.Locals[v.Name] = v
			dt, err := canonicalType(v.Type, ctx.Types)
			if err != nil {
				return "", err
			}
			typeCode, err := compileType(dt)
			if err != nil {
				return "", err
			}
			header += v.Name + " " + typeCode + "\n"
		}
		bodyStatements = bodyStatements[1:]
	}
	header += genDebugFn(ctx.Locals, ctx.Globals)
	body, err := compileBody(bodyStatements, returnTypes, ctx)
	if err != nil {
		return "", err
	}
	return header + body + "\n}\n", nil
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
	//_p.PollContinue(line, globals, locals)
}
`
	return s
}

func compileIfStatement(s IfStatement, expectedReturnTypes []DataType, ctx CodeContext) (string, error) {
	c, returnedTypes, err := compileExpression(s.Condition, ctx)
	if err != nil {
		return "", err
	}
	if len(returnedTypes) != 1 || !isType(returnedTypes[0], BuiltinType{"Bool", nil}, true) {
		return "", fmt.Errorf("if condition does not return one value or returns non-bool on line %d", s.LineNumber)
	}
	code := "if " + c
	c, err = compileBody(s.Body, expectedReturnTypes, ctx)
	if err != nil {
		return "", nil
	}
	code += " {\n" + c + "}"
	for _, elif := range s.Elifs {
		c, returnedTypes, err := compileExpression(elif.Condition, ctx)
		if err != nil {
			return "", err
		}
		if !isType(returnedTypes[0], BuiltinType{"Bool", nil}, true) {
			return "", errors.New("Elif condition expression does not return a boolean on line " + strconv.Itoa(elif.LineNumber))
		}
		code += " else if " + c + ".(bool) {\n"
		c, err = compileBody(elif.Body, expectedReturnTypes, ctx)
		if err != nil {
			return "", err
		}
		code += c + "}"
	}
	if len(s.Else.Body) > 0 {
		c, err := compileBody(s.Else.Body, expectedReturnTypes, ctx)
		if err != nil {
			return "", err
		}
		code += " else {\n" + c + "}"
	}
	return code + "\n", nil
}

func compileTypeswitchStatement(s TypeswitchStatement, expectedReturnTypes []DataType, ctx CodeContext) (string, error) {
	expr, rts, err := compileExpression(s.Value, ctx)
	if err != nil {
		return "", err
	}

	if len(rts) != 1 {
		return "", fmt.Errorf("typeswitch expression does not return one value on line %d", s.LineNumber)
	}
	inter, ok := rts[0].(InterfaceDefinition)
	if !ok {
		return "", fmt.Errorf("typeswitch expression does not an interface value on line %d", s.LineNumber)
	}
	code := "{\n _inter := " + expr + "\n"
	for i, c := range s.Cases {
		caseType, err := canonicalType(c.Variable.Type, ctx.Types)
		if err != nil {
			return "", err
		}
		if !isType(caseType, inter, false) {
			return "", fmt.Errorf("typeswitch case type is not an implementor of the interface. Line %d", s.LineNumber)
		}
		t, err := compileType(caseType)
		if err != nil {
			return "", err
		}
		name := c.Variable.Name
		locals := map[string]Variable{}
		for k, v := range ctx.Locals {
			locals[k] = v
		}
		locals[name] = c.Variable
		ctx.Locals = locals
		body, err := compileBody(c.Body, expectedReturnTypes, ctx)
		if err != nil {
			return "", nil
		}
		if i > 0 {
			code += " else "
		}
		code += "if " + name + ", _ok := _inter.(" + t + "); _ok { \n" + body + "}"
	}
	if s.Default != nil {
		body, err := compileBody(s.Default, expectedReturnTypes, ctx)
		if err != nil {
			return "", nil
		}
		code += " else { \n" + body + "}"
	}
	return code + "\n}\n", nil
}

func compileWhileStatement(s WhileStatement, expectedReturnTypes []DataType, ctx CodeContext) (string, error) {
	c, returnedTypes, err := compileExpression(s.Condition, ctx)
	if err != nil {
		return "", err
	}
	if len(returnedTypes) != 1 {
		return "", errors.New("while condition expression must one value (a boolean) on line " + strconv.Itoa(s.LineNumber))
	}
	if !isType(returnedTypes[0], BuiltinType{"Bool", nil}, true) {
		return "", errors.New("while condition expression does not return a boolean on line " + strconv.Itoa(s.LineNumber))
	}
	code := "for " + c + " {\n"
	c, err = compileBody(s.Body, expectedReturnTypes, ctx)
	if err != nil {
		return "", err
	}
	return code + c + "}\n", nil
}

func compileForeachStatement(s ForeachStatement, expectedReturnTypes []DataType, ctx CodeContext) (string, error) {
	lineStr := strconv.Itoa(s.LineNumber)
	if _, ok := ctx.Locals[s.IndexName]; ok {
		return "", errors.New("foreach index name conflicts with an existing local variable on line " + lineStr)
	}
	if _, ok := ctx.Locals[s.ValName]; ok {
		return "", errors.New("foreach val name conflicts with an existing local variable on line " + lineStr)
	}
	locals := map[string]Variable{}
	for k, v := range ctx.Locals {
		locals[k] = v
	}
	locals[s.IndexName] = Variable{s.LineNumber, s.Column, s.IndexName, s.IndexType}
	locals[s.ValName] = Variable{s.LineNumber, s.Column, s.ValName, s.ValType}
	ctx.Locals = locals
	collExpr, returnedTypes, err := compileExpression(s.Collection, ctx)
	if err != nil {
		return "", err
	}
	if len(returnedTypes) != 1 {
		return "", errors.New("foreach collection expression improperly returns more than one value on line " + lineStr)
	}
	collType, ok := returnedTypes[0].(BuiltinType)
	if !ok || (collType.Name != "L" && collType.Name != "M") {
		return "", errors.New("foreach collection type must be a list or map on line " + lineStr)
	}
	indexType, err := canonicalType(s.IndexType, ctx.Types)
	if err != nil {
		return "", err
	}
	valType, err := canonicalType(s.ValType, ctx.Types)
	if err != nil {
		return "", err
	}
	if collType.Name == "L" {
		if !isType(indexType, BuiltinType{"N", nil}, true) {
			return "", errors.New("Expected foreach index variable to be a number on line " + lineStr)
		}
		if !isType(collType.Params[0], valType, false) {
			return "", errors.New("Improper foreach val type for list on line " + lineStr)
		}
	} else if collType.Name == "M" {
		if !isType(collType.Params[0], indexType, false) {
			return "", errors.New("Improper foreach index type for map on line " + lineStr)
		}
		if !isType(collType.Params[1], valType, false) {
			return "", errors.New("Improper foreach val type for map on line " + lineStr)
		}
	}
	code := "for _i, _v := range " + collExpr + " { \n"
	code += s.IndexName + " = _i \n"
	code += s.ValName + " = _v \n"
	body, err := compileBody(s.Body, expectedReturnTypes, ctx)
	if err != nil {
		return "", err
	}
	code += body + "}\n"
	return code, nil
}

func compileBody(statements []Statement, expectedReturnTypes []DataType, ctx CodeContext) (string, error) {
	var code string
	for _, s := range statements {
		line := s.Line()
		ctx.ValidBreakpoints[strconv.Itoa(line)] = true
		code += fmt.Sprintf("if _breakpoints[%d] {debug(%d)}\n", line, line)
		var c string
		var err error
		switch s := s.(type) {
		case IfStatement:
			c, err = compileIfStatement(s, expectedReturnTypes, ctx)
		case WhileStatement:
			c, err = compileWhileStatement(s, expectedReturnTypes, ctx)
		case ForeachStatement:
			c, err = compileForeachStatement(s, expectedReturnTypes, ctx)
		case AssignmentStatement:
			c, err = compileAssignmentStatement(s, ctx)
		case TypeswitchStatement:
			c, err = compileTypeswitchStatement(s, expectedReturnTypes, ctx)
		case ReturnStatement:
			c, err = compileReturnStatement(s, expectedReturnTypes, ctx)
		case FunctionCall:
			c, _, err = compileFunctionCall(s, ctx)
			c += "\n"
		case MethodCall:
			c, _, err = compileMethodCall(s, ctx)
			c += "\n"
		case Operation:
			c, _, err = compileOperation(s, ctx)
			c += "\n"
		}
		if err != nil {
			return "", err
		}
		code += c
	}
	return code, nil
}

func compileAssignmentStatement(s AssignmentStatement, ctx CodeContext) (string, error) {
	lineStr := strconv.Itoa(s.LineNumber)
	valCode, valueTypes, err := compileExpression(s.Value, ctx)
	if err != nil {
		return "", err
	}
	if len(valueTypes) != len(s.Targets) {
		return "", errors.New("Wrong number of targets in assignment on line " + lineStr)
	}
	code := ""
	for i, target := range s.Targets {
		switch t := target.(type) {
		case Token:
			if t.Type != IdentifierWord {
				return "", errors.New("Assignment to non-identifier on line " + lineStr)
			}
		case Operation:
			if t.Operator != "dr" && t.Operator != "get" && t.Operator != "ref" {
				return "", errors.New("Improper target of assignment on line " + lineStr)
			}
		default:
			return "", errors.New("Improper target of assignment on line " + lineStr)
		}
		expr, rts, err := compileExpression(target, ctx)
		if err != nil {
			return "", err
		}
		// shouldn't be the case that any target expression returns more than one value
		if len(rts) != 1 {
			return "", errors.New("Improper target of assignment on line")
		}
		if !isType(valueTypes[i], rts[0], false) {
			return "", errors.New("Value in assignment does not match expected type on line " + lineStr)
		}
		code += expr
		if i < len(s.Targets)-1 {
			code += ", "
		}
	}
	return code + " = " + valCode + "\n", nil
}

func compileReturnStatement(s ReturnStatement, expectedReturnTypes []DataType, ctx CodeContext) (string, error) {
	lineStr := strconv.Itoa(s.LineNumber)
	if len(s.Values) != len(expectedReturnTypes) {
		return "", errors.New("Return statement has wrong number of values on line " + lineStr)
	}
	code := "return "
	for i, v := range s.Values {
		c, returnedTypes, err := compileExpression(v, ctx)
		if err != nil {
			return "", err
		}
		if len(returnedTypes) != 1 {
			return "", errors.New("Expression in return statement returns more than one value on line " + lineStr)
		}
		if !isType(returnedTypes[0], expectedReturnTypes[i], false) {
			return "", errors.New("Wrong type in return statement on line " + lineStr)
		}
		code += c
		if i < len(s.Values)-1 {
			code += ", "
		}
	}
	return code + "\n", nil
}

func compileMethodCall(s MethodCall, ctx CodeContext) (string, []DataType, error) {
	lineStr := strconv.Itoa(s.LineNumber)
	if len(s.Arguments) < 1 {
		return "", nil, errors.New("Method call has not receiver on line " + lineStr)
	}
	receiver, receiverTypes, err := compileExpression(s.Receiver, ctx)
	if err != nil {
		return "", nil, err
	}
	if len(receiverTypes) != 1 {
		return "", nil, errors.New("Method call receiver expression does not return one value on line " + lineStr)
	}
	var ft FunctionType
Outer:
	switch receiverType := receiverTypes[0].(type) {
	case Struct:
		var ok bool
		ft, ok = receiverType.Methods[s.MethodName]
		if !ok {
			return "", nil, errors.New("Method call struct receiver does not have such a method on line " + lineStr)
		}
	case InterfaceDefinition:
		for _, sig := range receiverType.Methods {
			if sig.Name == s.MethodName {
				var err error
				ft, err = sig.getFunctionType(ctx.Types)
				if err != nil {
					return "", nil, err
				}
				break Outer
			}
		}
		return "", nil, errors.New("Method call receiver does not have a method of that name on " + lineStr)
	default:
		return "", nil, errors.New("Method call receiver must be a struct or interface value on line " + lineStr)
	}

	code := receiver + "." + s.MethodName + "("
	for i, exp := range s.Arguments {
		c, returnedTypes, err := compileExpression(exp, ctx)
		if err != nil {
			return "", nil, err
		}
		if len(returnedTypes) != 1 {
			return "", nil, errors.New("Method call argument does not return one value on line " + lineStr)
		}
		if !isType(returnedTypes[0], ft.Params[i], false) {
			return "", nil, errors.New("Method call argument is wrong type on line " + lineStr)
		}
		code += c + ", " // Go is OK with comma after last arg, so don't need special case for last arg
	}
	return code + ")", ft.ReturnTypes, nil
}

func compileFunctionCall(s FunctionCall, ctx CodeContext) (string, []DataType, error) {
	lineStr := strconv.Itoa(s.LineNumber)
	code := ""
	var ft FunctionType
	var ok bool
	switch s := s.Function.(type) {
	case Operation:
		c, returnedTypes, err := compileOperation(s, ctx)
		if err != nil {
			return "", nil, err
		}
		if len(returnedTypes) != 1 {
			return "", nil, errors.New("operation at start of parens must return a function to call on line " + lineStr)
		}
		ft, ok = returnedTypes[0].(FunctionType)
		if !ok {
			return "", nil, errors.New("operation at start of parens returned something other than a function on line " + lineStr)
		}
		code += c
	case FunctionCall:
		c, returnedTypes, err := compileFunctionCall(s, ctx)
		if err != nil {
			return "", nil, err
		}
		if len(returnedTypes) != 1 {
			return "", nil, errors.New("function call at start of parens must return a function to call on line " + lineStr)
		}
		ft, ok = returnedTypes[0].(FunctionType)
		if !ok {
			return "", nil, errors.New("function call at start of parens returned something other than a function on line " + lineStr)
		}
		code += c
	case Token: // will always be an identifier
		code += s.Content
		ft, ok = ctx.FuncTypes[s.Content]
		if !ok {
			return "", nil, errors.New("calling nonexistent function on line " + lineStr)
		}
	}
	code += "(" // start of arguments
	for i, exp := range s.Arguments {
		c, returnedTypes, err := compileExpression(exp, ctx)
		if err != nil {
			return "", nil, err
		}
		if len(returnedTypes) != 1 {
			return "", nil, errors.New("argument expression in function call doesn't return one value on line " + lineStr)
		}
		if !isType(returnedTypes[0], ft.Params[i], false) {
			return "", nil, errors.New("argument of wrong type in function call on line " + lineStr)
		}
		code += c + ", " // Go is OK with comma after last arg, so don't need special case for last arg
	}
	if len(s.Arguments) > 0 {
		code = code[:len(code)-2] // drop last comma and space
	}
	return code + ")", ft.ReturnTypes, nil
}

func compileOperation(o Operation, ctx CodeContext) (string, []DataType, error) {
	lineStr := strconv.Itoa(o.LineNumber)
	operandCode := make([]string, len(o.Operands))
	operandTypes := make([]DataType, len(o.Operands))
	for i, expr := range o.Operands {
		c, returnTypes, err := compileExpression(expr, ctx)
		if err != nil {
			return "", nil, err
		}
		if len(returnTypes) != 1 {
			return "", nil, errors.New("operand expression returns more than one value on line " + strconv.Itoa(o.LineNumber))
		}
		operandCode[i] = c
		operandTypes[i] = returnTypes[0]
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
		if !ok || (t.Name != "L" && t.Name != "M") {
			return "", nil, errors.New("get operation requires a list or map as first operand. Line " + lineStr)
		}
		switch t.Name {
		case "M":
			returnType = t.Params[1]
			if !isType(operandTypes[1], t.Params[0], true) {
				return "", nil, errors.New("get operation on map has wrong type as second operand")
			}
			code += operandCode[0] + "[" + operandCode[1] + "]"
		case "L":
			returnType = t.Params[0]
			dt, err := compileType(returnType)
			if err != nil {
				return "", nil, err
			}
			if !isType(operandTypes[1], BuiltinType{"N", nil}, true) {
				return "", nil, errors.New("get operation requires a number as second operand")
			}
			code += "*" + operandCode[0] + "[int64(" + operandCode[1] + ")].(" + dt + ")"
		}
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
		if len(o.Operands) != 2 {
			return "", nil, errors.New("append operation requires two operands")
		}
		dt, ok := isList(operandTypes[0])
		if !ok {
			return "", nil, errors.New("append operation requires first operand to be a list on line " + lineStr)
		}
		if !isType(operandTypes[1], dt, false) {
			return "", nil, errors.New("append operation's second operand is not valid for the list. Line " + lineStr)
		}
		code += operandCode[0] + ".append(" + operandCode[1] + ")"
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
	case "ref":
		if len(o.Operands) != 1 {
			return "", nil, errors.New("ref operation requires a single operand. Line " + lineStr)
		}
		switch e := o.Operands[0].(type) {
		case Token:
			switch e.Type {
			case IdentifierWord:
				name := e.Content
				if v, ok := ctx.Locals[name]; ok {
					rt, err := canonicalType(v.Type, ctx.Types)
					if err != nil {
						return "", nil, err
					}
					returnType = BuiltinType{"P", []DataType{rt}}
					code += "&" + name
				} else if v, ok := ctx.Globals[name]; ok {
					code += "&g_" + name
					rt, err := canonicalType(v.Type, ctx.Types)
					if err != nil {
						return "", nil, err
					}
					returnType = BuiltinType{"P", []DataType{rt}}
				} else {
					return "", nil, fmt.Errorf("Name %s on line %d is undefined.", name, e.LineNumber)
				}
			default:
				return "", nil, errors.New("ref operation has improper operand. Line " + lineStr)
			}
		case Operation:
			if e.Operator != "get" {
				return "", nil, errors.New("ref operation has improper operand. Line " + lineStr)
			}
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
				code += "&" + operandCode[0] + "[" + operandCode[1] + "]"
			case "L":
				returnType = t.Params[0]
				if !isType(operandTypes[1], BuiltinType{"N", nil}, true) {
					return "", nil, errors.New("get operation requires a number as second operand")
				}
				code += "&(*" + operandCode[0] + ")[int64(" + operandCode[1] + ")]"
			}
		default:
			return "", nil, errors.New("ref operation requires a single operand. Line " + lineStr)
		}
	case "dr":
		if len(o.Operands) != 1 {
			return "", nil, errors.New("dr operation requires a single operand. Line " + lineStr)
		}
		dt, ok := operandTypes[0].(BuiltinType)
		if !ok && dt.Name != "P" {
			return "", nil, errors.New("dr operation requires a pointer operand. Line " + lineStr)
		}
		returnType = dt.Params[0]
		code += "*" + operandCode[0]
	case "print":
		if len(o.Operands) < 1 {
			return "", nil, errors.New("'print' operation requires at least one operand")
		}
		code += "_fmt.Print("
		for i := range o.Operands {
			code += operandCode[i] + ", "
		}
		code += ")"
	case "println":
		if len(o.Operands) < 1 {
			return "", nil, errors.New("'println' operation requires at least one operand")
		}
		code += "_fmt.Println("
		for i := range o.Operands {
			code += operandCode[i] + ", "
		}
		code += ")"
	case "prompt":
		code += "_Prompt("
		for i := range o.Operands {
			code += operandCode[i] + ", "
		}
		code += ")"
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
	case "len":
		if len(o.Operands) != 1 {
			return "", nil, errors.New("len operation requires one operand")
		}
		dt, ok := operandTypes[0].(BuiltinType)
		if !ok {
			return "", nil, errors.New("")
		}
		if dt.Name != "L" && dt.Name != "M" {
			return "", nil, errors.New("len operand must be a list or map")
		}
		returnType = BuiltinType{"N", nil}
		code += "len(" + operandCode[0] + ")"
	case "istype":
		if len(o.Operands) != 2 {
			return "", nil, errors.New("istype operation requires two operands")
		}
		parsedType, ok := o.Operands[0].(ParsedDataType)
		if !ok {
			return "", nil, errors.New("istype first operand must be a data type")
		}
		dt, err := canonicalType(parsedType, ctx.Types)
		if err != nil {
			return "", nil, err
		}
		if !isType(dt, operandTypes[1], false) {
			return "", nil, errors.New("istype first operand must be a type implementing interface type of the second operand")
		}
		code += operandCode[1] + ".(" + operandCode[0] + "))"
		return code, []DataType{dt, BuiltinType{"Bool", nil}}, nil
	}

	code += ")"
	if returnType == nil {
		return code, []DataType{}, nil
	}
	return code, []DataType{returnType}, nil
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
	definitions, err := parse(tokens)
	if err != nil {
		return "", nil, err
	}
	code, validBreakpoints, err := compile(definitions)
	if err != nil {
		return "", nil, err
	}

	return code, validBreakpoints, nil

}
