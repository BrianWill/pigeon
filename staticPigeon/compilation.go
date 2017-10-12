package staticPigeon

import (
	"errors"
	"fmt"
	"go/build"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

/* All identifiers get prefixed with _ to avoid collisions with Go reserved words and predefined identifiers */
// returns map of valid breakpoints
func compile(pkg *Package, packages map[string]*Package, outputDir string, isMain bool) error {
	code := ""
	if isMain {
		code += "package main\n"
	} else {
		code += "package " + pkg.Prefix + "\n"
	}

	code += `import _fmt "fmt"
import _std "github.com/BrianWill/pigeon/staticPigeon/stdlib"
`
	c, err := compileImports(pkg.ImportDefs, packages, outputDir)
	if err != nil {
		return err
	}
	code += c

	code += compileNativeImports(pkg.NativeImports)

	err = processStructs(pkg)
	if err != nil {
		return err
	}
	c, err = compileInterfaces(pkg)
	if err != nil {
		return err
	}
	code += c
	for _, st := range pkg.Structs {
		if st.Pkg != pkg {
			continue
		}
		err := findImplementors(&st, pkg)
		if err != nil {
			return err
		}
		c, err := compileStruct(&st, pkg.Types)
		if err != nil {
			return err
		}
		code += c
	}
	// check that all function parameter and return types are valid
	for _, fn := range pkg.Funcs {
		_, err := getFunctionType(fn)
		if err != nil {
			return err
		}
	}
	c, err = compileGlobals(pkg)
	if err != nil {
		return err
	}
	code += c
	for _, methByStruct := range pkg.Methods {
		for _, m := range methByStruct {
			c, err := compileMethod(m)
			if err != nil {
				return err
			}
			code += c
		}
	}
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
	if isMain {
		code += `
		
		func main() {
			_main()
		}
		`
	}

	pkg.Code = code
	return nil
}

func findImplementors(st *Struct, pkg *Package) error {
Outer:
	for _, iface := range pkg.Interfaces {
		for _, sig := range iface.Methods {
			ft, ok := st.Methods[sig.Name]
			if !ok {
				continue Outer
			}
			mt, err := sig.getFunctionType(pkg)
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

func compileImports(imports map[string]ImportDefinition, packages map[string]*Package, outputDir string) (string, error) {
	code := ""
	for _, imp := range imports {
		path, err := filepath.Abs(imp.Path)
		if err != nil {
			return "", err
		}
		code += "import _" + packages[path].Prefix + `"` + outputDir + packages[path].Prefix + `"` + "\n"
	}
	return code, nil
}

func compileNativeImports(imports map[string]string) string {
	code := ""
	for prefix, path := range imports {
		code += "import " + prefix + `"` + path + `"` + "\n"
	}
	return code
}

func compileStruct(st *Struct, types map[string]DataType) (string, error) {
	code := "type " + st.Name + " struct {\n"
	for i, n := range st.MemberNames {
		t, err := compileType(st.MemberTypes[i], st.Pkg)
		if err != nil {
			return "", err
		}
		code += strings.Title(n) + " " + t + "\n"
	}
	code += st.NativeCode
	return code + "}\n", nil
}

func compileInterfaces(pkg *Package) (string, error) {
	code := "\n"
	for _, inter := range pkg.Interfaces {
		if inter.Pkg != pkg {
			continue
		}
		code += "type " + inter.Name + " interface {\n"
		for _, sig := range inter.Methods {
			// validate each method
			_, err := sig.getFunctionType(pkg)
			if err != nil {
				return "", err
			}
			code += sig.Name + "("
			for _, pt := range sig.ParamTypes {
				t, err := getDataType(pt, pkg)
				if err != nil {
					return "", err
				}
				c, err := compileType(t, pkg)
				if err != nil {
					return "", err
				}
				code += c + ", "
			}
			code += ") ("
			for _, rt := range sig.ReturnTypes {
				t, err := getDataType(rt, pkg)
				if err != nil {
					return "", err
				}
				c, err := compileType(t, pkg)
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

// populates pkg.Structs and verifies that no Struct is illegally recursive
func processStructs(pkg *Package) error {
	var processStruct func(Struct, *Package, []Struct) error
	processStruct = func(s Struct, pkg *Package, containingStructs []Struct) error {
		st := pkg.StructDefs[s.Name]
		for i, m := range st.Members {
			dt, err := getDataType(m.Type, pkg)
			if err != nil {
				return err
			}
			s.MemberNames[i] = m.Name
			s.MemberTypes[i] = dt
			switch t := dt.(type) {
			case Struct:
				for _, cst := range containingStructs {
					if t.Name == cst.Name {
						return msg(st.LineNumber, st.Column, "Struct cannot recursively contain itself.")
					}
				}
				err := processStruct(t, pkg, append(containingStructs, t))
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
	for _, st := range pkg.StructDefs {
		s := Struct{
			LineNumber:  st.LineNumber,
			Column:      st.Column,
			Name:        st.Name,
			MemberNames: make([]string, len(st.Members)),
			MemberTypes: make([]DataType, len(st.Members)),
			Implements:  map[string]bool{},
			Methods:     map[string]FunctionType{},
			NativeCode:  st.NativeCode,
			Pkg:         st.Pkg,
		}
		pkg.Structs[s.Name] = s
		pkg.Types[s.Name] = s
	}
	for _, st := range pkg.Structs {
		err := processStruct(st, pkg, []Struct{st})
		if err != nil {
			return err
		}
	}
	for _, methByStruct := range pkg.Methods {
		for _, meth := range methByStruct {
			dt, err := getDataType(meth.Receiver.Type, pkg)
			if err != nil {
				return err
			}
			if st, ok := dt.(Struct); ok {
				funcType, err := meth.getFunctionType()
				if err != nil {
					return err
				}
				st.Methods[meth.Name] = funcType
			} else {
				return msg(meth.LineNumber, meth.Column, "Method has non-struct receiver.")
			}
		}
	}
	return nil
}

func (m MethodDefinition) getFunctionType() (FunctionType, error) {
	fd := FunctionDefinition{
		Parameters:  m.Parameters,
		ReturnTypes: m.ReturnTypes,
		Pkg:         m.Pkg,
	}
	return getFunctionType(fd)
}

func getFunctionType(fn FunctionDefinition) (FunctionType, error) {
	params := make([]DataType, len(fn.Parameters))
	for i, p := range fn.Parameters {
		dt, err := getDataType(p.Type, fn.Pkg)
		if err != nil {
			return FunctionType{}, err
		}
		params[i] = dt
	}
	returnTypes := make([]DataType, len(fn.ReturnTypes))
	for i, rt := range fn.ReturnTypes {
		dt, err := getDataType(rt, fn.Pkg)
		if err != nil {
			return FunctionType{}, err
		}
		returnTypes[i] = dt
	}
	return FunctionType{params, returnTypes}, nil
}

func (s Signature) getFunctionType(pkg *Package) (FunctionType, error) {
	params := make([]DataType, len(s.ParamTypes))
	for i, p := range s.ParamTypes {
		dt, err := getDataType(p, pkg)
		if err != nil {
			return FunctionType{}, err
		}
		params[i] = dt
	}
	returnTypes := make([]DataType, len(s.ReturnTypes))
	for i, rt := range s.ReturnTypes {
		dt, err := getDataType(rt, pkg)
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

func getDataType(parsed ParsedDataType, pkg *Package) (DataType, error) {
	if parsed.Type == "A" {
		if len(parsed.Params) != 2 {
			return nil, errors.New("Array type must have two type parameters.")
		}
		t, err := getDataType(parsed.Params[0], pkg)
		if err != nil {
			return nil, err
		}
		size, err := strconv.Atoi(parsed.Params[1].Type)
		if err != nil {
			return nil, errors.New("Array type must have integer as second type parameter.")
		}
		return ArrayType{size, t}, nil
	}
	params := make([]DataType, len(parsed.Params))
	for i, v := range parsed.Params {
		t, err := getDataType(v, pkg)
		if err != nil {
			return nil, err
		}
		params[i] = t
	}
	returnTypes := make([]DataType, len(parsed.ReturnTypes))
	for i, v := range parsed.ReturnTypes {
		t, err := getDataType(v, pkg)
		if err != nil {
			return nil, err
		}
		returnTypes[i] = t
	}
	switch parsed.Type {
	case "Fn":
		return FunctionType{params, returnTypes}, nil
	case "L":
		if len(params) != 1 {
			return nil, errors.New("List type has wrong number of type parameters.")
		}
		return BuiltinType{"L", params}, nil
	case "S":
		if len(params) != 1 {
			return nil, errors.New("List type has wrong number of type parameters.")
		}
		return BuiltinType{"S", params}, nil
	case "Ch":
		if len(params) != 1 {
			return nil, errors.New("Channel type has wrong number of type parameters.")
		}
		return BuiltinType{"Ch", params}, nil
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
	case "I", "F", "Byte", "Str", "Bool", "Err", "Any":
		if len(params) != 0 {
			return nil, errors.New("Type " + parsed.Type + " should not have any type parameters.")
		}
		return BuiltinType{parsed.Type, params}, nil
	default:
		t, ok := pkg.Types[parsed.Type]
		if !ok {
			return nil, errors.New("Unknown type. " + fmt.Sprint(parsed.Type))
		}
		if len(parsed.Params) > 0 || len(parsed.ReturnTypes) > 0 {
			return nil, errors.New("Type " + parsed.Type + " should not have any type parameters.")
		}
		return t, nil
	}
}

// not the same as a 'type assertion'
// 'type expression' is parens starting with a type to create a value of that type
func compileTypeExpression(te TypeExpression, pkg *Package, locals map[string]Variable) (string, []DataType, error) {
	line := te.LineNumber
	column := te.Column
	dt, err := getDataType(te.Type, pkg)
	if err != nil {
		return "", nil, err
	}
	switch t := dt.(type) {
	case BuiltinType:
		switch t.Name {
		case "I", "F", "Byte":
			if len(t.Params) != 0 {
				return "", nil, msg(line, column, "Invalid type expression: "+t.Name+" cannot have type parameters.")
			}
			if len(te.Operands) != 1 {
				return "", nil, msg(line, column, "Invalid type expression: "+t.Name+" must have one (and just one) operand.")
			}
			expr, returnedTypes, err := compileExpression(te.Operands[0], pkg, locals)
			if err != nil {
				return "", nil, err
			}
			if len(returnedTypes) != 1 {
				return "", nil, msg(line, column, "Invalid type expression: "+t.Name+" must have one (and just one) operand.")
			}
			if !isNumber(returnedTypes[0]) {
				return "", nil, msg(line, column, "Invalid type expression: "+t.Name+" must have number operand.")
			}
			var numberTypes = map[string]string{
				"I":    "int64",
				"F":    "float64",
				"Byte": "byte",
			}
			code := numberTypes[t.Name] + "(" + expr + ")"
			return code, []DataType{t}, nil
		case "M":
			if len(t.Params) != 2 {
				return "", nil, msg(line, column, "Invalid type expression. Map must have two type parameters.")
			}
			mapType, err := compileType(t, pkg)
			if err != nil {
				return "", nil, err
			}
			mapType += "{"
			if len(te.Operands)%2 != 0 {
				return "", nil, msg(line, column, "Invalid type expression. Map must have even number of operands.")
			}
			for i := 0; i < len(te.Operands); i += 2 {
				key, returnedTypes, err := compileExpression(te.Operands[i], pkg, locals)
				if err != nil {
					return "", nil, err
				}
				if len(returnedTypes) != 1 || !isType(returnedTypes[0], t.Params[0], false) {
					return "", nil, msg(line, column, "Invalid type expression. Map key of wrong type.")
				}
				val, returnedTypes, err := compileExpression(te.Operands[i+1], pkg, locals)
				if err != nil {
					return "", nil, err
				}
				if len(returnedTypes) != 1 || !isType(returnedTypes[0], t.Params[1], false) {
					return "", nil, msg(line, column, "Invalid type expression. Map val of wrong type.")
				}
				mapType += key + ": " + val + ", "
			}
			mapType += "}"
			return mapType, []DataType{t}, nil
		case "L":
			if len(t.Params) != 1 {
				return "", nil, msg(line, column, "Invalid type expression. List must have one type parameter.")
			}
			expr := "(func () *_std.List {\n"
			expr += "var _list _std.List = make([]interface{}, " + strconv.Itoa(len(te.Operands)) + ")\n"
			if err != nil {
				return "", nil, err
			}
			for i := 0; i < len(te.Operands); i++ {
				val, returnedTypes, err := compileExpression(te.Operands[i], pkg, locals)
				if err != nil {
					return "", nil, err
				}
				if len(returnedTypes) != 1 || !isType(returnedTypes[0], t.Params[0], false) {
					return "", nil, msg(line, column, "Invalid type expression. List val of wrong type.")
				}
				expr += "_list[" + strconv.Itoa(i) + "] = " + val + "\n"
			}
			expr += `return &_list
		    })()`
			return expr, []DataType{t}, nil
		case "Ch":
			if len(t.Params) != 1 {
				return "", nil, msg(line, column, "Invalid type expression. Channel must have one type parameter.")
			}
			if len(te.Operands) != 1 {
				return "", nil, msg(line, column, "Invalid type expression. Must specify size for channel.")
			}
			size, returnedTypes, err := compileExpression(te.Operands[0], pkg, locals)
			if err != nil {
				return "", nil, err
			}
			if len(returnedTypes) != 1 {
				return "", nil, msg(line, column, "Invalid type expression. Channel must have one (and just one) operand.")
			}
			if !isInteger(returnedTypes[0]) {
				return "", nil, msg(line, column, "Invalid type expression. Channel must have an integer operand.")
			}
			channelType, err := compileType(t.Params[0], pkg)
			if err != nil {
				return "", nil, err
			}
			code := "make(chan " + channelType + ", " + size + ")"
			return code, []DataType{t}, nil
		default:
			return "", nil, msg(line, column, "Invalid type expression. Cannot create type "+t.Name+".")
		}
	case ArrayType:
		if len(te.Operands) != t.Size {
			return "", nil, msg(line, column, "Array expression must have number of operands that matches the length.")
		}
		param, err := compileType(t.Type, pkg)
		if err != nil {
			return "", nil, err
		}
		code := "[" + strconv.Itoa(t.Size) + "]" + param + "{"
		for i := 0; i < len(te.Operands); i++ {
			val, returnedTypes, err := compileExpression(te.Operands[i], pkg, locals)
			if err != nil {
				return "", nil, err
			}
			if len(returnedTypes) != 1 || !isType(returnedTypes[0], t.Type, false) {
				return "", nil, msg(line, column, "Invalid type expression. List val of wrong type.")
			}
			code += val + ", "
		}
		code += "}"
		return code, nil, nil
	case FunctionType:
		return "", nil, msg(line, column, "Invalid type expression. Cannot create a function with a type expression.")
	case Struct:
		if len(t.MemberNames) != len(te.Operands) {
			return "", nil, msg(line, column, "Invalid type expression. Wrong number of args for creating struct.")
		}
		code := t.Name + "{"
		for i, argType := range t.MemberTypes {
			expr, returnTypes, err := compileExpression(te.Operands[i], pkg, locals)
			if err != nil {
				return "", nil, err
			}
			if len(returnTypes) != 1 || !isType(returnTypes[0], argType, false) {
				return "", nil, msg(line, column, "Invalid type expression. Wrong type of arg for creating struct.")
			}
			code += expr + ", "
		}
		code += "}"
		return code, []DataType{t}, nil
	case InterfaceDefinition:
		return "", nil, msg(line, column, "Invalid type expression. Cannot create interface value.")
	}
	// should be unreachable
	return "", nil, msg(line, column, "Invalid type expression.")
}

func compileExpression(e Expression, pkg *Package, locals map[string]Variable) (string, []DataType, error) {
	var code string
	var returnedTypes []DataType
	var err error
	switch e := e.(type) {
	case Operation:
		code, returnedTypes, err = compileOperation(e, pkg, locals)
		if err != nil {
			return "", nil, err
		}
	case FunctionCall:
		code, returnedTypes, err = compileFunctionCall(e, pkg, locals)
		if err != nil {
			return "", nil, err
		}
	case MethodCall:
		fmt.Println("compiling method ", e)
		code, returnedTypes, err = compileMethodCall(e, pkg, locals)
		if err != nil {
			return "", nil, err
		}
	case TypeExpression:
		code, returnedTypes, err = compileTypeExpression(e, pkg, locals)
		if err != nil {
			return "", nil, err
		}
	case ParsedDataType:
		dt, err := getDataType(e, pkg)
		if err != nil {
			return "", nil, err
		}
		code, err = compileType(dt, pkg)
		if err != nil {
			return "", nil, err
		}
		returnedTypes = []DataType{BuiltinType{"Type", nil}}
	case Token:
		switch e.Type {
		case IdentifierWord:
			name := e.Content
			if v, ok := locals[name]; ok {
				code = name
				rt, err := getDataType(v.Type, pkg)
				if err != nil {
					return "", nil, err
				}
				returnedTypes = []DataType{rt}
			} else if v, ok := pkg.Globals[name]; ok {
				if v.Pkg == pkg {
					code = "G_" + name
				} else {
					code = "_" + v.Pkg.Prefix + ".G_" + name
				}
				rt, err := getDataType(v.Type, pkg)
				if err != nil {
					return "", nil, err
				}
				returnedTypes = []DataType{rt}
			} else if v, ok := pkg.Funcs[name]; ok {
				if v.Pkg == pkg {
					code = name
				} else {
					code = "_" + v.Pkg.Prefix + "." + name
				}
				code = name
				rt, err := getFunctionType(v)
				if err != nil {
					return "", nil, err
				}
				returnedTypes = []DataType{rt}
			} else {
				return "", nil, msg(e.LineNumber, e.Column, "Name is undefined.")
			}
		case NumberLiteral:
			if strings.Index(e.Content, ".") == -1 {
				code = "int64(" + e.Content + ")"
				returnedTypes = []DataType{BuiltinType{"I", nil}}
			} else {
				code = "float64(" + e.Content + ")"
				returnedTypes = []DataType{BuiltinType{"F", nil}}
			}
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

func compileGlobals(pkg *Package) (string, error) {
	code := ""
	for _, g := range pkg.Globals {
		if g.Pkg != pkg {
			continue
		}
		code += "var G_" + g.Name + " "
		t, err := getDataType(g.Type, pkg)
		if err != nil {
			return "", err
		}
		c, err := compileType(t, pkg)
		if err != nil {
			return "", err
		}
		code += c + " = "
		c, returnedTypes, err := compileExpression(g.Value, pkg, map[string]Variable{})
		if err != nil {
			return "", err
		}
		if len(returnedTypes) != 1 {
			return "", msg(g.LineNumber, g.Column, "Initial value of global does not match the declared type.")
		}
		if !isType(returnedTypes[0], t, false) {
			return "", msg(g.LineNumber, g.Column, "Initial value of global does not match the declared type.")
		}
		code += c + "\n"
		pkg.ValidBreakpoints[strconv.Itoa(g.LineNumber)] = true
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

func isNumber(dt DataType) bool {
	t, ok := dt.(BuiltinType)
	if !ok {
		return false
	}
	return t.Name == "I" || t.Name == "F" || t.Name == "Byte"
}

func isChannel(dt DataType) (bool, DataType) {
	t, ok := dt.(BuiltinType)
	if !ok || t.Name != "Ch" {
		return false, nil
	}
	return true, t.Params[0]
}

func isInteger(dt DataType) bool {
	t, ok := dt.(BuiltinType)
	if !ok {
		return false
	}
	return t.Name == "I" || t.Name == "Byte"
}

func isMap(dt DataType) (DataType, DataType, bool) {
	t, ok := dt.(BuiltinType)
	if !ok || t.Name != "M" {
		return nil, nil, false
	}
	return t.Params[0], t.Params[1], true
}

// assumes a valid data type. Accepts Struct but not a StructDefinition
func compileType(dt DataType, pkg *Package) (string, error) {
	switch t := dt.(type) {
	case BuiltinType:
		switch t.Name {
		case "I":
			return "int64", nil
		case "F":
			return "float64", nil
		case "Byte":
			return "byte", nil
		case "Bool":
			return "bool", nil
		case "Str":
			return "string", nil
		case "Err":
			return "error", nil
		case "Any":
			return "interface{}", nil
		case "L":
			return "*_std.List", nil
		case "S":
			param, err := compileType(t.Params[0], pkg)
			if err != nil {
				return "", err
			}
			return "[]" + param, nil
		case "Ch":
			param, err := compileType(t.Params[0], pkg)
			if err != nil {
				return "", err
			}
			return "chan " + param, nil
		case "M":
			keyType, err := compileType(t.Params[0], pkg)
			if err != nil {
				return "", err
			}
			valType, err := compileType(t.Params[1], pkg)
			if err != nil {
				return "", err
			}
			return "map[" + keyType + "]" + valType, nil
		case "P":
			pointerType, err := compileType(t.Params[0], pkg)
			if err != nil {
				return "", err
			}
			return "*" + pointerType, nil
		}
	case ArrayType:
		param, err := compileType(t.Type, pkg)
		if err != nil {
			return "", err
		}
		return "[" + strconv.Itoa(t.Size) + "]" + param, nil
	case InterfaceDefinition:
		if t.Pkg != pkg {
			return "_" + t.Pkg.Prefix + "." + t.Name, nil
		}
		return t.Name, nil
	case Struct:
		if t.Pkg != pkg {
			return "_" + t.Pkg.Prefix + "." + t.Name, nil
		}
		return t.Name, nil
	case FunctionType:
		typeStr := "func( "
		for _, paramType := range t.Params {
			s, err := compileType(paramType, pkg)
			if err != nil {
				return "", err
			}
			typeStr += s + ", "
		}
		typeStr += ") "
		if len(t.ReturnTypes) > 0 {
			typeStr += "("
			for _, returnType := range t.ReturnTypes {
				s, err := compileType(returnType, pkg)
				if err != nil {
					return "", err
				}
				typeStr += s + ", "
			}
			typeStr += ")"
		}
		return typeStr, nil
	case StructDefinition:
		return "", errors.New("Invalid type")
	}

	return "", nil
}

// returns code snippet ending with '\n\n'
func compileFunc(fn FunctionDefinition) (string, error) {
	locals := map[string]Variable{}
	header := "func " + strings.Title(fn.Name) + "("
	for i, param := range fn.Parameters {
		dt, err := getDataType(param.Type, fn.Pkg)
		if err != nil {
			return "", err
		}
		typeCode, err := compileType(dt, fn.Pkg)
		if err != nil {
			return "", err
		}
		header += param.Name + " " + typeCode
		if i < len(fn.Parameters)-1 {
			header += ", "
		}
		locals[param.Name] = param
	}
	header += ") ("
	returnTypes := make([]DataType, len(fn.ReturnTypes))
	for i, rt := range fn.ReturnTypes {
		dt, err := getDataType(rt, fn.Pkg)
		if err != nil {
			return "", err
		}
		returnTypes[i] = dt
		typeCode, err := compileType(dt, fn.Pkg)
		if err != nil {
			return "", err
		}
		header += typeCode
		if i < len(fn.ReturnTypes)-1 {
			header += ", "
		}
	}
	header += ") {\n"
	if fn.NativeCode != "" {
		return header + fn.NativeCode + "\n}\n", nil
	}
	if len(fn.Body) < 1 {
		return "", msg(fn.LineNumber, fn.Column, "Function should contain at least one statement.")
	}
	bodyStatements := fn.Body
	// account for locals statement
	if localsStatement, ok := bodyStatements[0].(LocalsStatement); ok {
		for _, v := range localsStatement.Vars {
			header += "var "
			if _, ok := locals[v.Name]; ok {
				return "", msg(v.LineNumber, v.Column, "Local variable "+v.Name+" is already defined as a parameter.")
			}
			locals[v.Name] = v
			dt, err := getDataType(v.Type, fn.Pkg)
			if err != nil {
				return "", err
			}
			typeCode, err := compileType(dt, fn.Pkg)
			if err != nil {
				return "", err
			}
			header += v.Name + " " + typeCode
			if t, ok := dt.(BuiltinType); ok {
				switch t.Name {
				case "L":
					header += " = new(_std.List)"
				case "M", "Ch":
					header += " = make(" + typeCode + ")"
				}
			}
			header += "\n"
		}
		bodyStatements = bodyStatements[1:]
	}
	// account for local funcs
	for {
		if len(bodyStatements) == 0 {
			break
		}
		localFunc, ok := bodyStatements[0].(LocalFuncStatement)
		if !ok {
			break
		}
		code, err := compileLocalFunc(localFunc, fn.Pkg, locals)
		if err != nil {
			return "", err
		}
		header += code + "\n"
		bodyStatements = bodyStatements[1:]
	}
	header += genDebugFn(locals, fn.Pkg.Globals, fn.Pkg)
	body, err := compileBody(bodyStatements, returnTypes, fn.Pkg, locals)
	if err != nil {
		return "", err
	}
	return header + body + "\n}\n", nil
}

// returns code snippet ending with '\n\n'
func compileMethod(meth MethodDefinition) (string, error) {
	locals := map[string]Variable{}
	dt, err := getDataType(meth.Receiver.Type, meth.Pkg)
	if err != nil {
		return "", err
	}
	receiverType, err := compileType(dt, meth.Pkg)
	if err != nil {
		return "", err
	}
	header := "func (" + meth.Receiver.Name + " " + receiverType + ") " + meth.Name + "("
	for i, param := range meth.Parameters {
		dt, err := getDataType(param.Type, meth.Pkg)
		if err != nil {
			return "", err
		}
		typeCode, err := compileType(dt, meth.Pkg)
		if err != nil {
			return "", err
		}
		header += param.Name + " " + typeCode
		if i < len(meth.Parameters)-1 {
			header += ", "
		}
		locals[param.Name] = param
	}
	header += ") ("
	returnTypes := make([]DataType, len(meth.ReturnTypes))
	for i, rt := range meth.ReturnTypes {
		dt, err := getDataType(rt, meth.Pkg)
		if err != nil {
			return "", err
		}
		returnTypes[i] = dt
		typeCode, err := compileType(dt, meth.Pkg)
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
		return "", msg(meth.LineNumber, meth.Column, "FMethod should contain at least one statement.")
	}
	bodyStatements := meth.Body
	if localsStatement, ok := bodyStatements[0].(LocalsStatement); ok {
		for _, v := range localsStatement.Vars {
			header += "var "
			if _, ok := locals[v.Name]; ok {
				return "", msg(v.LineNumber, v.Column, "Local variable "+v.Name+" is already defined as a parameter.")
			}
			locals[v.Name] = v
			dt, err := getDataType(v.Type, meth.Pkg)
			if err != nil {
				return "", err
			}
			typeCode, err := compileType(dt, meth.Pkg)
			if err != nil {
				return "", err
			}
			header += v.Name + " " + typeCode
			if t, ok := dt.(BuiltinType); ok {
				switch t.Name {
				case "L":
					header += " = new(_std.List)"
				case "M", "Ch":
					header += " = make(" + typeCode + ")"
				}
			}
			header += v.Name + " " + typeCode + "\n"
		}
		bodyStatements = bodyStatements[1:]
	}
	for {
		if len(bodyStatements) == 0 {
			break
		}
		localFunc, ok := bodyStatements[0].(LocalFuncStatement)
		if !ok {
			break
		}
		code, err := compileLocalFunc(localFunc, meth.Pkg, locals)
		if err != nil {
			return "", err
		}
		header += code + "\n"
		bodyStatements = bodyStatements[1:]
	}
	header += genDebugFn(locals, meth.Pkg.Globals, meth.Pkg)
	body, err := compileBody(bodyStatements, returnTypes, meth.Pkg, locals)
	if err != nil {
		return "", err
	}
	return header + body + "\n}\n", nil
}

func compileLocalFunc(fn LocalFuncStatement, pkg *Package, outerLocals map[string]Variable) (string, error) {
	if _, ok := outerLocals[fn.Name]; ok {
		return "", msg(fn.LineNumber, fn.Column, "Local func cannot have same name as another local.")
	}
	paramTypes := make([]ParsedDataType, len(fn.Parameters))
	for i, v := range fn.Parameters {
		paramTypes[i] = v.Type
	}
	outerLocals[fn.Name] = Variable{fn.LineNumber, fn.Column, fn.Name,
		ParsedDataType{fn.LineNumber, "Fn", paramTypes, fn.ReturnTypes},
	}
	locals := map[string]Variable{}
	header := fn.Name + " := func("
	for i, param := range fn.Parameters {
		dt, err := getDataType(param.Type, pkg)
		if err != nil {
			return "", err
		}
		typeCode, err := compileType(dt, pkg)
		if err != nil {
			return "", err
		}
		header += param.Name + " " + typeCode
		if i < len(fn.Parameters)-1 {
			header += ", "
		}
		locals[param.Name] = param
	}
	header += ") ("
	returnTypes := make([]DataType, len(fn.ReturnTypes))
	for i, rt := range fn.ReturnTypes {
		dt, err := getDataType(rt, pkg)
		if err != nil {
			return "", err
		}
		returnTypes[i] = dt
		typeCode, err := compileType(dt, pkg)
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
		return "", msg(fn.LineNumber, fn.Column, "Function should contain at least one statement.")
	}
	bodyStatements := fn.Body
	if localsStatement, ok := bodyStatements[0].(LocalsStatement); ok {
		for _, v := range localsStatement.Vars {
			header += "var "
			if _, ok := locals[v.Name]; ok {
				return "", msg(v.LineNumber, v.Column, "Local variable "+v.Name+"is already defined as a parameter.")
			}
			locals[v.Name] = v
			dt, err := getDataType(v.Type, pkg)
			if err != nil {
				return "", err
			}
			typeCode, err := compileType(dt, pkg)
			if err != nil {
				return "", err
			}
			header += v.Name + " " + typeCode
			if t, ok := dt.(BuiltinType); ok {
				switch t.Name {
				case "L":
					header += " = new(_std.List)"
				case "M", "Ch":
					header += " = make(" + typeCode + ")"
				}
			}
			header += "\n"
		}
		bodyStatements = bodyStatements[1:]
	}
	header += genDebugFn(locals, pkg.Globals, pkg)
	for name, v := range outerLocals {
		if _, ok := locals[name]; !ok {
			locals[name] = v
		}
	}
	body, err := compileBody(bodyStatements, returnTypes, pkg, locals)
	if err != nil {
		return "", err
	}
	return header + body + "\n}\n", nil
}

func genDebugFn(locals map[string]Variable, globals map[string]GlobalDefinition, pkg *Package) string {
	s := `debug := func(line int) {
	var globals = map[string]interface{}{
`
	for k, g := range globals {
		if g.Pkg == pkg {
			s += fmt.Sprintf("\"%s\": G_%s,\n", k, k)
		} else {
			s += fmt.Sprintf("\"%s\": _%s.G_%s,\n", k, g.Pkg.Prefix, k)
		}
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

func compileGoStatement(s GoStatement, pkg *Package, locals map[string]Variable) (string, error) {
	expr, _, err := compileExpression(s.Call, pkg, locals)
	if err != nil {
		return "", err
	}
	return "go " + expr + "\n", nil
}

func compileIfStatement(s IfStatement, expectedReturnTypes []DataType, pkg *Package, locals map[string]Variable) (string, error) {
	c, returnedTypes, err := compileExpression(s.Condition, pkg, locals)
	if err != nil {
		return "", err
	}
	if len(returnedTypes) != 1 || !isType(returnedTypes[0], BuiltinType{"Bool", nil}, true) {
		return "", msg(s.LineNumber, s.Column, "if condition does not return one value or returns non-bool.")
	}
	code := "if " + c
	c, err = compileBody(s.Body, expectedReturnTypes, pkg, locals)
	if err != nil {
		return "", nil
	}
	code += " {\n" + c + "}"
	for _, elif := range s.Elifs {
		c, returnedTypes, err := compileExpression(elif.Condition, pkg, locals)
		if err != nil {
			return "", err
		}
		if !isType(returnedTypes[0], BuiltinType{"Bool", nil}, true) {
			return "", msg(elif.LineNumber, elif.Column, "Elif condition expression does not return a boolean.")
		}
		code += " else if " + c + ".(bool) {\n"
		c, err = compileBody(elif.Body, expectedReturnTypes, pkg, locals)
		if err != nil {
			return "", err
		}
		code += c + "}"
	}
	if len(s.Else.Body) > 0 {
		c, err := compileBody(s.Else.Body, expectedReturnTypes, pkg, locals)
		if err != nil {
			return "", err
		}
		code += " else {\n" + c + "}"
	}
	return code + "\n", nil
}

func compileTypeswitchStatement(s TypeswitchStatement, expectedReturnTypes []DataType, pkg *Package,
	locals map[string]Variable) (string, error) {
	expr, rts, err := compileExpression(s.Value, pkg, locals)
	if err != nil {
		return "", err
	}
	if len(rts) != 1 {
		return "", msg(s.LineNumber, s.Column, "typeswitch expression does not return one value.")
	}
	inter, ok := rts[0].(InterfaceDefinition)
	if !ok {
		return "", msg(s.LineNumber, s.Column, "typeswitch expression does not an interface value.")
	}
	code := "{\n _inter := " + expr + "\n"
	for i, c := range s.Cases {
		caseType, err := getDataType(c.Variable.Type, pkg)
		if err != nil {
			return "", err
		}
		if !isType(caseType, inter, false) {
			return "", msg(s.LineNumber, s.Column, "typeswitch case type is not an implementor of the interface.")
		}
		t, err := compileType(caseType, pkg)
		if err != nil {
			return "", err
		}
		name := c.Variable.Name
		newLocals := map[string]Variable{}
		for k, v := range locals {
			newLocals[k] = v
		}
		newLocals[name] = c.Variable
		body, err := compileBody(c.Body, expectedReturnTypes, pkg, newLocals)
		if err != nil {
			return "", nil
		}
		if i > 0 {
			code += " else "
		}
		code += "if " + name + ", _ok := _inter.(" + t + "); _ok { \n"
		code += genDebugFn(newLocals, pkg.Globals, pkg)
		code += body + "}"
	}
	if s.Default != nil {
		body, err := compileBody(s.Default, expectedReturnTypes, pkg, locals)
		if err != nil {
			return "", nil
		}
		code += " else { \n" + body + "}"
	}
	return code + "\n}\n", nil
}

func compileWhileStatement(s WhileStatement, expectedReturnTypes []DataType, pkg *Package, locals map[string]Variable) (string, error) {
	c, returnedTypes, err := compileExpression(s.Condition, pkg, locals)
	if err != nil {
		return "", err
	}
	if len(returnedTypes) != 1 {
		return "", msg(s.LineNumber, s.Column, "while condition expression must one value (a boolean).")
	}
	if !isType(returnedTypes[0], BuiltinType{"Bool", nil}, true) {
		return "", msg(s.LineNumber, s.Column, "while condition expression does not return a boolean.")
	}
	code := "for " + c + " {\n"
	c, err = compileBody(s.Body, expectedReturnTypes, pkg, locals)
	if err != nil {
		return "", err
	}
	return code + c + "}\n", nil
}

func compileForeachStatement(s ForeachStatement, expectedReturnTypes []DataType, pkg *Package, locals map[string]Variable) (string, error) {
	if _, ok := locals[s.IndexName]; ok {
		return "", msg(s.LineNumber, s.Column, "foreach index name conflicts with an existing local variable.")
	}
	if _, ok := locals[s.ValName]; ok {
		return "", msg(s.LineNumber, s.Column, "foreach val name conflicts with an existing local variable.")
	}
	newLocals := map[string]Variable{}
	for k, v := range locals {
		newLocals[k] = v
	}
	newLocals[s.IndexName] = Variable{s.LineNumber, s.Column, s.IndexName, s.IndexType}
	newLocals[s.ValName] = Variable{s.LineNumber, s.Column, s.ValName, s.ValType}
	collExpr, returnedTypes, err := compileExpression(s.Collection, pkg, newLocals)
	if err != nil {
		return "", err
	}
	if len(returnedTypes) != 1 {
		return "", msg(s.LineNumber, s.Column, "foreach collection expression improperly returns more than one value.")
	}
	indexType, err := getDataType(s.IndexType, pkg)
	if err != nil {
		return "", err
	}
	valType, err := getDataType(s.ValType, pkg)
	if err != nil {
		return "", err
	}
	code := "for _i, _v := range "
	switch t := returnedTypes[0].(type) {
	case BuiltinType:
		if t.Name != "L" && t.Name != "M" && t.Name != "S" {
			return "", msg(s.LineNumber, s.Column, "foreach collection type must be a list or map.")
		}
		if t.Name == "L" {
			if !isNumber(indexType) {
				return "", msg(s.LineNumber, s.Column, "Expected foreach index variable to be a number.")
			}
			if !isType(t.Params[0], valType, false) {
				return "", msg(s.LineNumber, s.Column, "Improper foreach val type for list.")
			}
			code += "*"
		} else if t.Name == "M" {
			if !isType(t.Params[0], indexType, false) {
				return "", msg(s.LineNumber, s.Column, "Improper foreach index type for map.")
			}
			if !isType(t.Params[1], valType, false) {
				return "", msg(s.LineNumber, s.Column, "Improper foreach val type for map.")
			}
		}
	case ArrayType:
		if !isNumber(indexType) {
			return "", msg(s.LineNumber, s.Column, "Expected foreach index variable to be a number.")
		}
		if !isType(t.Type, valType, false) {
			return "", msg(s.LineNumber, s.Column, "Improper foreach val type for array.")
		}
	default:
		return "", msg(s.LineNumber, s.Column, "foreach collection type must be a list, map, slice, or array.")
	}
	code += collExpr + " { \n"
	code += s.IndexName + " := _i \n"
	code += s.ValName + " := _v \n"
	code += genDebugFn(newLocals, pkg.Globals, pkg)
	body, err := compileBody(s.Body, expectedReturnTypes, pkg, newLocals)
	if err != nil {
		return "", err
	}
	code += body + "}\n"
	return code, nil
}

func compileBody(statements []Statement, expectedReturnTypes []DataType, pkg *Package, locals map[string]Variable) (string, error) {
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
			c, err = compileIfStatement(s, expectedReturnTypes, pkg, locals)
		case WhileStatement:
			c, err = compileWhileStatement(s, expectedReturnTypes, pkg, locals)
		case ForeachStatement:
			c, err = compileForeachStatement(s, expectedReturnTypes, pkg, locals)
		case GoStatement:
			c, err = compileGoStatement(s, pkg, locals)
		case AssignmentStatement:
			c, err = compileAssignmentStatement(s, pkg, locals)
		case SelectStatement:
			c, err = compileSelectStatement(s, expectedReturnTypes, pkg, locals)
		case TypeswitchStatement:
			c, err = compileTypeswitchStatement(s, expectedReturnTypes, pkg, locals)
		case ReturnStatement:
			c, err = compileReturnStatement(s, expectedReturnTypes, pkg, locals)
		case FunctionCall:
			c, _, err = compileFunctionCall(s, pkg, locals)
			c += "\n"
		case MethodCall:
			c, _, err = compileMethodCall(s, pkg, locals)
			c += "\n"
		case Operation:
			if s.Operator != "set" && s.Operator != "print" && s.Operator != "println" &&
				s.Operator != "prompt" && s.Operator != "push" {
				return "", msg(s.LineNumber, s.Column, "Improper operation as statement. Only set, push, print, println, "+
					"and prompt can be standalone statements.")
			}
			c, _, err = compileOperation(s, pkg, locals)
			c += "\n"
		}
		if err != nil {
			return "", err
		}
		code += c
	}
	return code, nil
}

func compileSelectStatement(s SelectStatement, expectedReturnTypes []DataType, pkg *Package, locals map[string]Variable) (string, error) {
	code := "select {\n"
	for _, clause := range s.Clauses {
		switch clause := clause.(type) {
		case SelectSendClause:
			ch, channelTypes, err := compileExpression(clause.Channel, pkg, locals)
			if err != nil {
				return "", err
			}
			if len(channelTypes) != 1 {
				return "", msg(clause.LineNumber, clause.Column, "Select send clause expecting channel expression.")
			}
			ok, param := isChannel(channelTypes[0])
			if !ok {
				return "", msg(clause.LineNumber, clause.Column, "Select send clause expecting channel expression.")
			}
			val, valTypes, err := compileExpression(clause.Value, pkg, locals)
			if err != nil {
				return "", err
			}
			if len(valTypes) != 1 {
				return "", msg(clause.LineNumber, clause.Column, "Select send clause expecting expression returning one value.")
			}
			if !isType(valTypes[0], param, false) {
				return "", msg(clause.LineNumber, clause.Column, "Select send clause expecting expression with type matching its channel.")
			}
			code += "case " + ch + " <- " + val + ":\n"
			c, err := compileBody(clause.Body, expectedReturnTypes, pkg, locals)
			if err != nil {
				return "", err
			}
			code += c + "\n"
		case SelectRcvClause:
			ch, channelTypes, err := compileExpression(clause.Channel, pkg, locals)
			if err != nil {
				return "", err
			}
			if len(channelTypes) != 1 {
				return "", msg(clause.LineNumber, clause.Column, "Select send clause expecting channel expression.")
			}
			ok, param := isChannel(channelTypes[0])
			if !ok {
				return "", msg(clause.LineNumber, clause.Column, "Select send clause expecting channel expression.")
			}
			dt, err := getDataType(clause.Target.Type, pkg)
			if err != nil {
				return "", err
			}
			// because we must use := syntax, must be exact match
			if !isType(param, dt, true) {
				return "", msg(clause.LineNumber, clause.Column, "Select rcv clause has improper type target variable.")
			}
			code += "case " + clause.Target.Name + " := <- " + ch + ":\n"
			newLocals := map[string]Variable{}
			for name, v := range locals {
				newLocals[name] = v
			}
			newLocals[clause.Target.Name] = clause.Target
			code += genDebugFn(newLocals, pkg.Globals, pkg)
			c, err := compileBody(clause.Body, expectedReturnTypes, pkg, newLocals)
			if err != nil {
				return "", err
			}
			code += c + "\n"
		}
	}
	if s.Default.Body != nil {
		c, err := compileBody(s.Default.Body, expectedReturnTypes, pkg, locals)
		if err != nil {
			return "", err
		}
		code += "default:\n" + c
	}

	return code + "\n}\n", nil
}

func compileAssignmentStatement(s AssignmentStatement, pkg *Package, locals map[string]Variable) (string, error) {
	valCode, valueTypes, err := compileExpression(s.Value, pkg, locals)
	if err != nil {
		return "", err
	}
	if len(valueTypes) != len(s.Targets) {
		return "", msg(s.LineNumber, s.Column, "Wrong number of targets in assignment.")
	}
	code := ""
	for i, target := range s.Targets {
		switch t := target.(type) {
		case Token:
			if t.Type != IdentifierWord {
				return "", msg(s.LineNumber, s.Column, "Assignment to non-identifier.")
			}
		case Operation:
			if t.Operator != "dr" && t.Operator != "get" && t.Operator != "ref" {
				return "", msg(s.LineNumber, s.Column, "Improper target of assignment.")
			}
			if t.Operator == "get" {
				t.Operator = "asget"
				target = t
			}
		default:
			return "", msg(s.LineNumber, s.Column, "Improper target of assignment.")
		}
		expr, rts, err := compileExpression(target, pkg, locals)
		if err != nil {
			return "", err
		}
		// shouldn't be the case that any target expression returns more than one value
		if len(rts) != 1 {
			return "", errors.New("Improper target of assignment on line")
		}
		if !isType(valueTypes[i], rts[0], false) {
			return "", msg(s.LineNumber, s.Column, "Value in assignment does not match expected type.")
		}
		code += expr
		if i < len(s.Targets)-1 {
			code += ", "
		}
	}
	return code + " = " + valCode + "\n", nil
}

func compileReturnStatement(s ReturnStatement, expectedReturnTypes []DataType, pkg *Package, locals map[string]Variable) (string, error) {
	if len(s.Values) != len(expectedReturnTypes) {
		return "", msg(s.LineNumber, s.Column, "Return statement has wrong number of values.")
	}
	code := "return "
	for i, v := range s.Values {
		c, returnedTypes, err := compileExpression(v, pkg, locals)
		if err != nil {
			return "", err
		}
		if len(returnedTypes) != 1 {
			return "", msg(s.LineNumber, s.Column, "Expression in return statement returns more than one value.")
		}
		if !isType(returnedTypes[0], expectedReturnTypes[i], false) {
			return "", msg(s.LineNumber, s.Column, "Wrong type in return statement.")
		}
		code += c
		if i < len(s.Values)-1 {
			code += ", "
		}
	}
	return code + "\n", nil
}

func compileMethodCall(s MethodCall, pkg *Package, locals map[string]Variable) (string, []DataType, error) {
	if len(s.Arguments) < 1 {
		return "", nil, msg(s.LineNumber, s.Column, "Method call has no receiver.")
	}
	receiver, receiverTypes, err := compileExpression(s.Receiver, pkg, locals)
	if err != nil {
		return "", nil, err
	}
	if len(receiverTypes) != 1 {
		return "", nil, msg(s.LineNumber, s.Column, "Method call receiver expression does not return one value.")
	}
	var ft FunctionType
Outer:
	switch receiverType := receiverTypes[0].(type) {
	case Struct:
		var ok bool
		ft, ok = receiverType.Methods[s.MethodName]
		if !ok {
			return "", nil, msg(s.LineNumber, s.Column, "Method call struct receiver does not have such a method.")
		}
	case InterfaceDefinition:
		for _, sig := range receiverType.Methods {
			if sig.Name == s.MethodName {
				var err error
				ft, err = sig.getFunctionType(pkg)
				if err != nil {
					return "", nil, err
				}
				break Outer
			}
		}
		return "", nil, msg(s.LineNumber, s.Column, "Method call receiver does not have a method of that name.")
	default:
		return "", nil, msg(s.LineNumber, s.Column, "Method call receiver must be a struct or interface value.")
	}

	code := receiver + "." + s.MethodName + "("
	for i, exp := range s.Arguments {
		c, returnedTypes, err := compileExpression(exp, pkg, locals)
		if err != nil {
			return "", nil, err
		}
		if len(returnedTypes) != 1 {
			return "", nil, msg(s.LineNumber, s.Column, "Method call argument does not return one value.")
		}
		if !isType(returnedTypes[0], ft.Params[i], false) {
			return "", nil, msg(s.LineNumber, s.Column, "Method call argument is wrong type.")
		}
		code += c + ", " // Go is OK with comma after last arg, so don't need special case for last arg
	}
	return code + ")", ft.ReturnTypes, nil
}

func compileFunctionCall(s FunctionCall, pkg *Package, locals map[string]Variable) (string, []DataType, error) {
	code := ""
	var ft FunctionType
	var ok bool
	switch s := s.Function.(type) {
	case Operation:
		c, returnedTypes, err := compileOperation(s, pkg, locals)
		if err != nil {
			return "", nil, err
		}
		if len(returnedTypes) != 1 {
			return "", nil, msg(s.LineNumber, s.Column, "operation at start of parens must return a function to call.")
		}
		ft, ok = returnedTypes[0].(FunctionType)
		if !ok {
			return "", nil, msg(s.LineNumber, s.Column, "operation at start of parens returned something other than a function.")
		}
		code += c
	case FunctionCall:
		c, returnedTypes, err := compileFunctionCall(s, pkg, locals)
		if err != nil {
			return "", nil, err
		}
		if len(returnedTypes) != 1 {
			return "", nil, msg(s.LineNumber, s.Column, "function call at start of parens must return a function to call.")
		}
		ft, ok = returnedTypes[0].(FunctionType)
		if !ok {
			return "", nil, msg(s.LineNumber, s.Column, "function call at start of parens returned something other than a function.")
		}
		code += c
	case Token: // will always be an identifier
		v, ok := locals[s.Content]
		if ok {
			var err error
			dt, err := getDataType(v.Type, pkg)
			if err != nil {
				return "", nil, err
			}
			ft, ok = dt.(FunctionType)
			if !ok {
				return "", nil, msg(s.LineNumber, s.Column, "calling non-function.")
			}
			code += s.Content
		} else {
			fnDef, ok := pkg.Funcs[s.Content] // previous check means we don't have to check for zero val
			if !ok {
				return "", nil, msg(s.LineNumber, s.Column, "calling non-existent function.")
			}
			var err error
			ft, err = getFunctionType(fnDef)
			if err != nil {
				return "", nil, err
			}
			if fnDef.Pkg == pkg {
				code += strings.Title(s.Content)
			} else {
				code += "_" + fnDef.Pkg.Prefix + "." + strings.Title(s.Content) // calling imported function
			}
		}
	}
	code += "(" // start of arguments
	for i, exp := range s.Arguments {
		c, returnedTypes, err := compileExpression(exp, pkg, locals)
		if err != nil {
			return "", nil, err
		}
		if len(returnedTypes) != 1 {
			return "", nil, msg(s.LineNumber, s.Column, "argument expression in function call doesn't return one value.")
		}
		if !isType(returnedTypes[0], ft.Params[i], false) {
			return "", nil, msg(s.LineNumber, s.Column, "argument of wrong type in function call.")
		}
		code += c + ", " // Go is OK with comma after last arg, so don't need special case for last arg
	}
	if len(s.Arguments) > 0 {
		code = code[:len(code)-2] // drop last comma and space
	}
	return code + ")", ft.ReturnTypes, nil
}

func (s Struct) getMemberType(name string) (DataType, error) {
	for i, n := range s.MemberNames {
		if n == name {
			return s.MemberTypes[i], nil
		}
	}
	return nil, errors.New("Struct does not contain member '" + name + "'")
}

func compileOperation(o Operation, pkg *Package, locals map[string]Variable) (string, []DataType, error) {
	operandCode := make([]string, len(o.Operands))
	operandTypes := make([]DataType, len(o.Operands))
	for i, expr := range o.Operands {
		if i == 1 && (o.Operator == "get" || o.Operator == "asget") {
			if st, ok := operandTypes[i].(Struct); ok {
				var returnType DataType
				if token, ok := expr.(Token); ok {
					if token.Type == IdentifierWord {
						var err error
						returnType, err = st.getMemberType(token.Content)
						if err != nil {
							return "", nil, err
						}
						return operandCode[0] + "." + strings.Title(token.Content),
							[]DataType{returnType}, nil
					}
				}
			}
		}
		c, returnTypes, err := compileExpression(expr, pkg, locals)
		if err != nil {
			return "", nil, err
		}
		if len(returnTypes) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "operand expression returns more than one value.")
		}
		operandCode[i] = c
		operandTypes[i] = returnTypes[0]
	}
	code := "("
	var returnType DataType
	switch o.Operator {
	case "add":
		if len(o.Operands) < 2 {
			return "", nil, msg(o.LineNumber, o.Column, "add operations requires at least two operands.")
		}
		t := operandTypes[0]
		if !isNumber(t) {
			return "", nil, msg(o.LineNumber, o.Column, "add operation has non-number operand")
		}
		for i := range o.Operands {
			if !isType(operandTypes[i], t, true) {
				return "", nil, msg(o.LineNumber, o.Column, "add operation has non-number operand")
			}
			code += operandCode[i]
			if i < len(o.Operands)-1 {
				code += " + "
			}
		}
		returnType = t
	case "sub":
		if len(o.Operands) < 2 {
			return "", nil, msg(o.LineNumber, o.Column, "sub operation requires at least two operands")
		}
		t := operandTypes[0]
		if !isNumber(t) {
			return "", nil, msg(o.LineNumber, o.Column, "sub operation has non-number operand")
		}
		for i := range o.Operands {
			if !isType(operandTypes[i], t, true) {
				return "", nil, msg(o.LineNumber, o.Column, "sub operation has non-number operand")
			}
			code += operandCode[i]
			if i < len(o.Operands)-1 {
				code += " - "
			}
		}
		returnType = t
	case "mul":
		if len(o.Operands) < 2 {
			return "", nil, msg(o.LineNumber, o.Column, "mul operation requires at least two operands")
		}
		t := operandTypes[0]
		if !isNumber(t) {
			return "", nil, msg(o.LineNumber, o.Column, "mul operation has non-number operand")
		}
		for i := range o.Operands {
			if !isType(operandTypes[i], t, true) {
				return "", nil, msg(o.LineNumber, o.Column, "mul operation has non-number operand")
			}
			code += operandCode[i]
			if i < len(o.Operands)-1 {
				code += " * "
			}
		}
		returnType = t
	case "div":
		if len(o.Operands) < 2 {
			return "", nil, msg(o.LineNumber, o.Column, "div operation requires at least two operands")
		}
		t := operandTypes[0]
		if !isNumber(t) {
			return "", nil, msg(o.LineNumber, o.Column, "div operation has non-number operand")
		}
		for i := range o.Operands {
			if !isType(operandTypes[i], t, true) {
				return "", nil, msg(o.LineNumber, o.Column, "div operation has non-number operand")
			}
			code += operandCode[i]
			if i < len(o.Operands)-1 {
				code += " / "
			}
		}
		returnType = t
	case "mod":
		if len(o.Operands) != 2 {
			return "", nil, msg(o.LineNumber, o.Column, "mod operation requires two operands")
		}
		t := operandTypes[0]
		if !isNumber(t) {
			return "", nil, msg(o.LineNumber, o.Column, "mod operation has non-number operand")
		}
		for i := range o.Operands {
			if !isType(operandTypes[i], t, true) {
				return "", nil, msg(o.LineNumber, o.Column, "mod operation has non-number operand")
			}
			code += "int64(" + operandCode[i] + ")"
			if i < len(o.Operands)-1 {
				code += " % "
			}
		}
		returnType = t
	case "eq":
		if len(o.Operands) < 2 {
			return "", nil, msg(o.LineNumber, o.Column, "eq operation requires at least two operands")
		}
		returnType = BuiltinType{"Bool", nil}
		for i := 0; i < len(o.Operands)-1; i++ {
			if !isType(operandTypes[i], operandTypes[0], true) ||
				!isType(operandTypes[i+1], operandTypes[0], true) {
				return "", nil, msg(o.LineNumber, o.Column, "eq operation has mismatched operand types")
			}
			if i > 0 {
				code += " && "
			}
			code += operandCode[i] + " == " + operandCode[i+1]
		}
	case "neq":
		if len(o.Operands) < 2 {
			return "", nil, msg(o.LineNumber, o.Column, "neq operation requires at least two operands")
		}
		returnType = BuiltinType{"Bool", nil}
		for i := 0; i < len(o.Operands)-1; i++ {
			if !isType(operandTypes[i], operandTypes[0], true) ||
				!isType(operandTypes[i+1], operandTypes[0], true) {
				return "", nil, msg(o.LineNumber, o.Column, "neq operation has mismatched operand types")
			}
			if i > 0 {
				code += " && "
			}
			code += operandCode[i] + " != " + operandCode[i+1]
		}
	case "not":
		if len(o.Operands) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "not operation requires one operand")
		}
		returnType = BuiltinType{"Bool", nil}
		if !isType(operandTypes[0], returnType, true) {
			return "", nil, msg(o.LineNumber, o.Column, "not operation has a non-bool operand")
		}
		code += "!" + operandCode[0]
	case "lt":
		if len(o.Operands) < 2 {
			return "", nil, msg(o.LineNumber, o.Column, "lt operation requires at least two operands")
		}
		returnType = BuiltinType{"Bool", nil}
		t := operandTypes[0]
		if !isNumber(t) {
			return "", nil, msg(o.LineNumber, o.Column, "lt operation has non-number operand")
		}
		for i := 0; i < len(o.Operands)-1; i++ {
			if !isType(operandTypes[i], t, true) ||
				!isType(operandTypes[i+1], t, true) {
				return "", nil, msg(o.LineNumber, o.Column, "lt operation has non-number operand")
			}
			if i > 0 {
				code += " && "
			}
			code += operandCode[i] + " < " + operandCode[i+1]
		}
	case "gt":
		if len(o.Operands) < 2 {
			return "", nil, msg(o.LineNumber, o.Column, "gt operation requires at least two operands")
		}
		returnType = BuiltinType{"Bool", nil}
		t := operandTypes[0]
		if !isNumber(t) {
			return "", nil, msg(o.LineNumber, o.Column, "lt operation has non-number operand")
		}
		for i := 0; i < len(o.Operands)-1; i++ {
			if !isType(operandTypes[i], t, true) ||
				!isType(operandTypes[i+1], t, true) {
				return "", nil, msg(o.LineNumber, o.Column, "gt operation has non-number operand")
			}
			if i > 0 {
				code += " && "
			}
			code += operandCode[i] + " > " + operandCode[i+1]
		}
	case "lte":
		if len(o.Operands) < 2 {
			return "", nil, msg(o.LineNumber, o.Column, "lte operation requires at least two operands")
		}
		returnType = BuiltinType{"Bool", nil}
		t := operandTypes[0]
		if !isNumber(t) {
			return "", nil, msg(o.LineNumber, o.Column, "lt operation has non-number operand")
		}
		for i := 0; i < len(o.Operands)-1; i++ {
			if !isType(operandTypes[i], t, true) ||
				!isType(operandTypes[i+1], t, true) {
				return "", nil, msg(o.LineNumber, o.Column, "lte operation has non-number operand")
			}
			if i > 0 {
				code += " && "
			}
			code += operandCode[i] + " <= " + operandCode[i+1]

		}
	case "gte":
		if len(o.Operands) < 2 {
			return "", nil, msg(o.LineNumber, o.Column, "gte operation requires at least two operands")
		}
		returnType = BuiltinType{"Bool", nil}
		t := operandTypes[0]
		if !isNumber(t) {
			return "", nil, msg(o.LineNumber, o.Column, "lt operation has non-number operand")
		}
		for i := 0; i < len(o.Operands)-1; i++ {
			if !isType(operandTypes[i], t, true) ||
				!isType(operandTypes[i+1], t, true) {
				return "", nil, msg(o.LineNumber, o.Column, "gte operation has non-number operand")
			}
			if i > 0 {
				code += " && "
			}
			code += operandCode[i] + " >= " + operandCode[i+1]
		}
	case "send":
		if len(o.Operands) != 2 {
			return "", nil, msg(o.LineNumber, o.Column, "send operation requires two operands")
		}
		ok, param := isChannel(operandTypes[0])
		if !ok {
			return "", nil, msg(o.LineNumber, o.Column, "send operation's first operand must be a channel")
		}
		if !isType(operandTypes[0], param, false) {
			return "", nil, msg(o.LineNumber, o.Column, "send operation's second operand is wrong type for the channel")
		}
		code += operandCode[0] + " <- " + operandCode[1]
		returnType = nil
	case "rcv":
		if len(o.Operands) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "rcv operation requires one operand")
		}
		ok, param := isChannel(operandTypes[0])
		if !ok {
			return "", nil, msg(o.LineNumber, o.Column, "rcv operation's first operand must be a channel")
		}
		code += " <- " + operandCode[0]
		returnType = param
	case "get", "asget":
		if len(o.Operands) < 2 {
			return "", nil, msg(o.LineNumber, o.Column, "get operation has too few operands")
		}
		switch t := operandTypes[0].(type) {
		case BuiltinType:
			switch t.Name {
			case "M":
				returnType = t.Params[1]
				if !isType(operandTypes[1], t.Params[0], true) {
					return "", nil, msg(o.LineNumber, o.Column, "get operation on map has wrong type as second operand")
				}
				code += operandCode[0] + "[" + operandCode[1] + "]"
			case "L", "S":
				returnType = t.Params[0]
				dt, err := compileType(returnType, pkg)
				if err != nil {
					return "", nil, err
				}
				if !isNumber(operandTypes[1]) {
					return "", nil, msg(o.LineNumber, o.Column, "get operation on list or slice requires a number as second operand")
				}
				if t.Name == "L" {
					code += "(*"
				} else if t.Name == "S" {
					code += "("
				}
				code += operandCode[0] + ")[int64(" + operandCode[1] + ")]"
				if t.Name == "L" && o.Operator == "get" {
					code += ".(" + dt + ")"
				}
			default:
				return "", nil, msg(o.LineNumber, o.Column, "get operation requires a list or map as first operand.")
			}
		case ArrayType:
			returnType = t.Type
			if !isNumber(operandTypes[1]) {
				return "", nil, errors.New("get operation on an array requires a number as second operand")
			}
			code += operandCode[0] + "[int64(" + operandCode[1] + ")]"
		default:
			return "", nil, msg(o.LineNumber, o.Column, "get operation requires a list or map as first operand.")
		}
	case "set":
		if len(o.Operands) != 3 {
			return "", nil, msg(o.LineNumber, o.Column, "set operation requires three operands")
		}
		switch t := operandTypes[0].(type) {
		case BuiltinType:
			switch t.Name {
			case "M":
				if !isType(operandTypes[1], t.Params[0], true) {
					return "", nil, msg(o.LineNumber, o.Column, "set operation on map has wrong type as second operand")
				}
				if !isType(operandTypes[2], t.Params[1], false) {
					return "", nil, msg(o.LineNumber, o.Column, "set operation on map has wrong type as third operand")
				}
				code += "func () {" + operandCode[0] + "[" + operandCode[1] + "] = " + operandCode[2] + "}()"
			case "L":
				if !isNumber(operandTypes[1]) {
					return "", nil, msg(o.LineNumber, o.Column, "set operation requires a number as second operand")
				}
				if !isType(operandTypes[2], t.Params[0], false) {
					return "", nil, msg(o.LineNumber, o.Column, "set operation on list has wrong type as third operand")
				}
				code += "func () {(*" + operandCode[0] + ")[" + operandCode[1] + "] = " + operandCode[2] + "}()"
			case "S":
				if !isNumber(operandTypes[1]) {
					return "", nil, msg(o.LineNumber, o.Column, "set operation requires a number as second operand")
				}
				if !isType(operandTypes[2], t.Params[0], false) {
					return "", nil, msg(o.LineNumber, o.Column, "set operation on list has wrong type as third operand")
				}
				code += "func () {" + operandCode[0] + "[" + operandCode[1] + "] = " + operandCode[2] + "}()"
			}
		case ArrayType:
			if !isNumber(operandTypes[1]) {
				return "", nil, msg(o.LineNumber, o.Column, "set operation requires a number as second operand")
			}
			if !isType(operandTypes[2], t.Type, false) {
				return "", nil, msg(o.LineNumber, o.Column, "set operation on list has wrong type as third operand")
			}
			code += "func () {" + operandCode[0] + "[" + operandCode[1] + "] = " + operandCode[2] + "}()"
		default:
			return "", nil, msg(o.LineNumber, o.Column, "set operation requires a list, map, slice, or array as first operand")
		}
		returnType = nil
	case "push":
		if len(o.Operands) != 2 {
			return "", nil, msg(o.LineNumber, o.Column, "push operation requires two operands")
		}
		switch t := operandTypes[0].(type) {
		case BuiltinType:
			if t.Name != "L" {
				return "", nil, msg(o.LineNumber, o.Column, "push operation's first operand must be a list.")
			}
			if !isType(operandTypes[1], t.Params[0], false) {
				return "", nil, msg(o.LineNumber, o.Column, "push operation's second operand is not valid for the list.")
			}
			code += operandCode[0] + ".append(" + operandCode[1] + ")"
		default:
			return "", nil, msg(o.LineNumber, o.Column, "push operation requires first operand to be a list.")
		}
		returnType = nil
	case "append":
		if len(o.Operands) != 2 {
			return "", nil, msg(o.LineNumber, o.Column, "append operation requires two operands")
		}
		switch t := operandTypes[0].(type) {
		case BuiltinType:
			if t.Name != "S" {
				return "", nil, msg(o.LineNumber, o.Column, "append operation's first operand must be a slice.")
			}
			if !isType(operandTypes[1], t.Params[0], false) {
				return "", nil, msg(o.LineNumber, o.Column, "append operation's second operand is not valid for the slice.")
			}
			code += "append(" + operandCode[0] + ", " + operandCode[1] + ")"
		default:
			return "", nil, msg(o.LineNumber, o.Column, "append operation requires first operand to be a slice.")
		}
		returnType = operandTypes[0]
	case "slice":
		if len(o.Operands) != 3 {
			return "", nil, errors.New("slice operation requires three operands")
		}
		if !isNumber(operandTypes[1]) || !isNumber(operandTypes[2]) {
			return "", nil, msg(o.LineNumber, o.Column, "slice operation's second and third operands must be numbers.")
		}
		switch t := operandTypes[0].(type) {
		case BuiltinType:
			if t.Name != "S" {
				return "", nil, msg(o.LineNumber, o.Column, "append operation's first operand must be a slice.")
			}
			returnType = operandTypes[0]
		case ArrayType:
			returnType = BuiltinType{"S", []DataType{t.Type}}
		default:
			return "", nil, msg(o.LineNumber, o.Column, "append operation requires first operand to be a slice.")
		}
		code += operandCode[0] + "[int64(" + operandCode[1] + "):int64(" + operandCode[2] + ")]"
	case "or":
		if len(o.Operands) < 2 {
			return "", nil, msg(o.LineNumber, o.Column, "or operation requires at least two operands")
		}
		returnType = BuiltinType{"Bool", nil}
		for i := range o.Operands {
			if !isType(operandTypes[i], returnType, true) {
				return "", nil, msg(o.LineNumber, o.Column, "or operation has non-boolean operand")
			}
			code += operandCode[i]
			if i < len(o.Operands)-1 {
				code += " || "
			}
		}
	case "and":
		if len(o.Operands) < 2 {
			return "", nil, msg(o.LineNumber, o.Column, "and operation requires at least two operands")
		}
		returnType = BuiltinType{"Bool", nil}
		for i := range o.Operands {
			if !isType(operandTypes[i], returnType, true) {
				return "", nil, msg(o.LineNumber, o.Column, "and operation has non-boolean operand")
			}
			code += operandCode[i]
			if i < len(o.Operands)-1 {
				code += " && "
			}
		}
	case "ref":
		if len(o.Operands) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "ref operation requires a single operand.")
		}
		switch e := o.Operands[0].(type) {
		case Token:
			switch e.Type {
			case IdentifierWord:
				name := e.Content
				if v, ok := locals[name]; ok {
					rt, err := getDataType(v.Type, pkg)
					if err != nil {
						return "", nil, err
					}
					returnType = BuiltinType{"P", []DataType{rt}}
					code += "&" + name
				} else if v, ok := pkg.Globals[name]; ok {
					code += "&G_" + name
					rt, err := getDataType(v.Type, pkg)
					if err != nil {
						return "", nil, err
					}
					returnType = BuiltinType{"P", []DataType{rt}}
				} else {
					return "", nil, msg(e.LineNumber, e.Column, "Name is undefined.")
				}
			default:
				return "", nil, msg(o.LineNumber, o.Column, "ref operation has improper operand.")
			}
		case Operation:
			if e.Operator != "get" {
				return "", nil, msg(o.LineNumber, o.Column, "ref operation has improper operand.")
			}
			if len(o.Operands) != 2 {
				return "", nil, msg(o.LineNumber, o.Column, "get operation requires two operands")
			}
			t, ok := operandTypes[0].(BuiltinType)
			if !ok || (t.Name != "L" && t.Name != "M") {
				return "", nil, msg(o.LineNumber, o.Column, "get operation requires a list or map as first operand")
			}
			switch t.Name {
			case "M":
				returnType = t.Params[1]
				if !isType(operandTypes[1], t.Params[0], true) {
					return "", nil, msg(o.LineNumber, o.Column, "get operation on map has wrong type as second operand")
				}
				code += "&" + operandCode[0] + "[" + operandCode[1] + "]"
			case "L":
				returnType = t.Params[0]
				if !isNumber(operandTypes[1]) {
					return "", nil, msg(o.LineNumber, o.Column, "get operation requires a number as second operand")
				}
				code += "&(*" + operandCode[0] + ")[int64(" + operandCode[1] + ")]"
			}
		default:
			return "", nil, msg(o.LineNumber, o.Column, "ref operation requires a single operand.")
		}
	case "dr":
		if len(o.Operands) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "dr operation requires a single operand.")
		}
		dt, ok := operandTypes[0].(BuiltinType)
		if !ok && dt.Name != "P" {
			return "", nil, msg(o.LineNumber, o.Column, "dr operation requires a pointer operand.")
		}
		returnType = dt.Params[0]
		code += "*" + operandCode[0]
	case "band":
		if len(o.Operands) != 2 {
			return "", nil, msg(o.LineNumber, o.Column, "'band' operation requires two operands")
		}
		if !isNumber(operandTypes[0]) || !isNumber(operandTypes[1]) {
			return "", nil, msg(o.LineNumber, o.Column, "'band' operation requires two number operands")
		}
		code += operandCode[0] + " & " + operandCode[1]
	case "bor":
		if len(o.Operands) != 2 {
			return "", nil, msg(o.LineNumber, o.Column, "'bor' operation requires two operands")
		}
		if !isNumber(operandTypes[0]) || !isNumber(operandTypes[1]) {
			return "", nil, msg(o.LineNumber, o.Column, "'bor' operation requires two number operands")
		}
		code += operandCode[0] + " | " + operandCode[1]
	case "bxor":
		if len(o.Operands) != 2 {
			return "", nil, msg(o.LineNumber, o.Column, "'bxor' operation requires two operands")
		}
		if !isNumber(operandTypes[0]) || !isNumber(operandTypes[1]) {
			return "", nil, msg(o.LineNumber, o.Column, "'bxor' operation requires two number operands")
		}
		code += operandCode[0] + " ^ " + operandCode[1]
	case "bnot":
		if len(o.Operands) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "'bnot' operation requires one operand")
		}
		if !isNumber(operandTypes[0]) {
			return "", nil, msg(o.LineNumber, o.Column, "'bnot' operation requires one number operand")
		}
		code += "^" + operandCode[1]
	case "print":
		if len(o.Operands) < 1 {
			return "", nil, msg(o.LineNumber, o.Column, "'print' operation requires at least one operand")
		}
		code += "_fmt.Print("
		for i := range o.Operands {
			code += operandCode[i] + ", "
		}
		code += ")"
	case "println":
		if len(o.Operands) < 1 {
			return "", nil, msg(o.LineNumber, o.Column, "'println' operation requires at least one operand")
		}
		code += "_fmt.Println("
		for i := range o.Operands {
			code += operandCode[i] + ", "
		}
		code += ")"
	case "prompt":
		code += "_std.Prompt("
		for i := range o.Operands {
			code += operandCode[i] + ", "
		}
		code += ")"
	case "concat":
		if len(o.Operands) < 2 {
			return "", nil, msg(o.LineNumber, o.Column, "concat operation requires at least two operands")
		}
		returnType = BuiltinType{"Str", nil}
		for i := range o.Operands {
			if !isType(operandTypes[i], returnType, true) {
				return "", nil, msg(o.LineNumber, o.Column, "concat operation has non-string operand")
			}
			code += operandCode[i]
			if i < len(o.Operands)-1 {
				code += " + "
			}
		}
	case "len":
		if len(o.Operands) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "len operation requires one operand")
		}
		returnType = BuiltinType{"I", nil}
		switch t := operandTypes[0].(type) {
		case BuiltinType:
			switch t.Name {
			case "L":
				code += "len(*" + operandCode[0] + ")"
			case "M", "S":
				code += "len(" + operandCode[0] + ")"
			default:
				return "", nil, msg(o.LineNumber, o.Column, "len operand must be a list or map")
			}
		case ArrayType:
			code += "len(" + operandCode[0] + ")"
		default:
			return "", nil, msg(o.LineNumber, o.Column, "len operation requires a list, map, array, or slice as operand")
		}
	case "istype":
		if len(o.Operands) != 2 {
			return "", nil, msg(o.LineNumber, o.Column, "istype operation requires two operands")
		}
		parsedType, ok := o.Operands[0].(ParsedDataType)
		if !ok {
			return "", nil, msg(o.LineNumber, o.Column, "istype first operand must be a data type")
		}
		dt, err := getDataType(parsedType, pkg)
		if err != nil {
			return "", nil, err
		}
		if !isType(dt, operandTypes[1], false) {
			return "", nil, msg(o.LineNumber, o.Column, "istype first operand must be a type implementing "+
				"interface type of the second operand")
		}
		code += operandCode[1] + ".(" + operandCode[0] + "))"
		return code, []DataType{dt, BuiltinType{"Bool", nil}}, nil
	case "randInt":
		if len(o.Operands) > 0 {
			return "", nil, msg(o.LineNumber, o.Column, "randInt operation takes no operands")
		}
		returnType = BuiltinType{"I", nil}
		code += "_std.RandInt()"
	case "randIntN":
		if len(o.Operands) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "randIntN operation takes one integer operand")
		}
		returnType = BuiltinType{"I", nil}
		if !isType(operandTypes[0], returnType, true) {
			return "", nil, msg(o.LineNumber, o.Column, "randIntN operation has non-integer operand")
		}
		code += "_std.RandIntN(" + operandCode[0] + ")"
	case "randFloat":
		if len(o.Operands) > 0 {
			return "", nil, msg(o.LineNumber, o.Column, "randFloat operation takes no operands")
		}
		returnType = BuiltinType{"F", nil}
		code += "_std.RandFloat()"
	case "parseInt":
		if len(o.Operands) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "parseInt operation takes one string operand")
		}
		if !isType(operandTypes[0], BuiltinType{"Str", nil}, true) {
			return "", nil, msg(o.LineNumber, o.Column, "parseInt operation has non-string operand")
		}
		code += "_std.ParseInt(" + operandCode[0] + ")"
		return code, []DataType{BuiltinType{"I", nil}, BuiltinType{"Err", nil}}, nil
	case "parseFloat":
		if len(o.Operands) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "parseFloat operation takes one string operand")
		}
		returnType = BuiltinType{"F", nil}
		if !isType(operandTypes[0], BuiltinType{"Str", nil}, true) {
			return "", nil, msg(o.LineNumber, o.Column, "parseFloat operation has non-string operand")
		}
		code += "_std.ParseFloat(" + operandCode[0] + ")"
		return code, []DataType{BuiltinType{"F", nil}, BuiltinType{"Err", nil}}, nil
	case "formatInt":
		if len(o.Operands) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "formatInt operation takes one integer operand")
		}
		returnType = BuiltinType{"Str", nil}
		if !isType(operandTypes[0], BuiltinType{"I", nil}, true) {
			return "", nil, msg(o.LineNumber, o.Column, "formatInt operation has non-integer operand")
		}
		code += "_std.FormatInt(" + operandCode[0] + ")"
	case "formatFloat":
		if len(o.Operands) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "formatFloat operation takes one float operand")
		}
		returnType = BuiltinType{"Str", nil}
		if !isType(operandTypes[0], BuiltinType{"F", nil}, true) {
			return "", nil, msg(o.LineNumber, o.Column, "formatFloat operation has non-float operand")
		}
		code += "_std.FormatFloat(" + operandCode[0] + ")"
	case "parseTime":
		if len(o.Operands) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "parseTime operation takes one string operand")
		}
		if !isType(operandTypes[0], BuiltinType{"Str", nil}, true) {
			return "", nil, msg(o.LineNumber, o.Column, "parseTime operation has non-string operand")
		}
		code += "_std.ParseTime(" + operandCode[0] + ")"
		return code, []DataType{BuiltinType{"I", nil}, BuiltinType{"Err", nil}}, nil
	case "timeNow":
		if len(o.Operands) != 0 {
			return "", nil, msg(o.LineNumber, o.Column, "TimeNow operation takes no operands")
		}
		returnType = BuiltinType{"I", nil}
		code += "_std.TimeNow()"
	case "formatTime":
		if len(o.Operands) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "formatTime operation takes one integer operand")
		}
		returnType = BuiltinType{"Str", nil}
		if !isType(operandTypes[0], BuiltinType{"I", nil}, true) {
			return "", nil, msg(o.LineNumber, o.Column, "formatTime operation has non-integer operand")
		}
		code += "_std.FormatTime(" + operandCode[0] + ")"
	}

	code += ")"
	if returnType == nil {
		return code, []DataType{}, nil
	}
	return code, []DataType{returnType}, nil
}

// returns map of valid breakpoints
func Compile(inputFilename string, outputDir string) (map[string]*Package, error) {
	packages := map[string]*Package{}
	_, err := ProcessPackage(inputFilename, packages, []string{}, outputDir)
	if err != nil {
		return nil, err
	}
	return packages, nil
}

var packagePrefixNum = 0

// 'packages' = all previously processed packages
// 'ancestorPaths' = all package full paths up the chain from this one we're processing (needed for detecting recursive dependencies)
func ProcessPackage(filename string, packages map[string]*Package, ancestorPaths []string, outputDir string) (*Package, error) {
	// if std lib
	if filename == "file" || filename == "http" {
		filename = build.Default.GOPATH + "/src/github.com/BrianWill/pigeon/staticPigeon/stdlib/" +
			filename + ".spigeon"
	}
	path, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}
	isMain := len(ancestorPaths) == 0
	for _, p := range ancestorPaths {
		if p == path {
			return nil, errors.New("Recurssive package import: " + p)
		}
	}
	ancestorPaths = append(ancestorPaths, path)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	tokens, err := lex(string(data) + "\n")
	if err != nil {
		return nil, err
	}

	pkg := &Package{
		Globals:          map[string]GlobalDefinition{},
		Types:            map[string]DataType{},
		ImportedTypes:    map[DataType]bool{},
		ValidBreakpoints: map[string]bool{},
		StructDefs:       map[string]StructDefinition{},
		Structs:          map[string]Struct{},
		Funcs:            map[string]FunctionDefinition{},
		Methods:          map[string]map[string]MethodDefinition{},
		Interfaces:       map[string]InterfaceDefinition{},
		ImportDefs:       map[string]ImportDefinition{},
		ImportedPackages: map[string]*Package{},
		NativeImports:    map[string]string{},
		FullPath:         path,
		Prefix:           "p" + strconv.Itoa(packagePrefixNum),
	}
	packagePrefixNum++
	packages[path] = pkg
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
				return nil, errors.New("Duplicate top-level name: " + d.Name)
			}
			pkg.Globals[d.Name] = d
			packageNames[un] = true
		case FunctionDefinition:
			un := strings.ToUpper(d.Name)
			if packageNames[un] {
				return nil, errors.New("Duplicate top-level name: " + d.Name)
			}
			pkg.Funcs[d.Name] = d
			packageNames[un] = true
		case StructDefinition:
			un := strings.ToUpper(d.Name)
			if packageNames[un] {
				return nil, errors.New("Duplicate top-level name: " + d.Name)
			}
			pkg.StructDefs[d.Name] = d
			packageNames[un] = true
		case InterfaceDefinition:
			un := strings.ToUpper(d.Name)
			if packageNames[un] {
				return nil, errors.New("Duplicate top-level name: " + d.Name)
			}
			pkg.Interfaces[d.Name] = d
			pkg.Types[d.Name] = d
			packageNames[un] = true
		case MethodDefinition:
			st, ok := pkg.Methods[d.Name]
			if !ok {
				st = map[string]MethodDefinition{}
				pkg.Methods[d.Name] = st
			}
			_, ok = st[d.Receiver.Name]
			if ok {
				return nil, errors.New("Duplicate method " + d.Name + " defined for type " + d.Receiver.Name)
			}
			st[d.Receiver.Name] = d
		case NativeImportDefinition:
			pkg.NativeImports[d.Alias] = d.Path
		case ImportDefinition:
			pkg.ImportDefs[d.Path] = d
			var otherPkg *Package
			otherPkg = packages[d.Path]
			if otherPkg == nil {
				var err error
				otherPkg, err = ProcessPackage(d.Path, packages, ancestorPaths, outputDir)
				if err != nil {
					return nil, err
				}
			}
			for _, foreignType := range otherPkg.Types {
				pkg.ImportedTypes[foreignType] = true
			}
			for i, v := range d.Names {
				localName := v
				foreignName := localName
				if d.Aliases[i] != "" {
					localName = d.Aliases[i]
				}
				un := strings.ToUpper(localName)
				if packageNames[un] {
					return nil, errors.New("Duplicate top-level name: " + localName)
				}
				packageNames[un] = true
				g, ok := otherPkg.Globals[foreignName]
				if ok {
					pkg.Globals[localName] = g
					continue
				}
				st, ok := otherPkg.Structs[foreignName]
				if ok {
					pkg.Structs[localName] = st
					pkg.Types[localName] = st
					continue
				}
				f, ok := otherPkg.Funcs[foreignName]
				if ok {
					pkg.Funcs[localName] = f
					continue
				}
				inter, ok := otherPkg.Interfaces[foreignName]
				if ok {
					pkg.Interfaces[localName] = inter
					pkg.Types[localName] = inter
					continue
				}
				return nil, errors.New("Imported name is not an exported name in foreign package: " + foreignName)
			}
		default:
			return nil, errors.New("Unrecognized definition")
		}
	}
	err = compile(pkg, packages, outputDir, isMain)
	if err != nil {
		return nil, err
	}
	return pkg, nil
}
