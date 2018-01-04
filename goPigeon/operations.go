package goPigeon

import (
	"strings"
)

func compileMakeOp(o Operation, pkg *Package, locals map[string]Variable) (string, []DataType, error) {
	// make op doesn't parse unless it has 2 operands, so we needn't check
	numStr, rts, err := compileExpression(o.Operands[0], pkg, locals)
	if err != nil {
		return "", nil, err
	}
	if len(rts) != 1 {
		return "", nil, msg(o.LineNumber, o.Column, "'make' operation has improper second operand.")
	}
	if !isType(rts[0], BuiltinType{"I", nil}, true) {
		return "", nil, msg(o.LineNumber, o.Column, "'make' operation requires at least two operands.")
	}
	if len(o.MakeType.Params) != 1 {
		return "", nil, msg(o.LineNumber, o.Column, "'make' operation requires a slice or list type")
	}
	dt, err := getDataType(o.MakeType.Params[0], pkg)
	if err != nil {
		return "", nil, err
	}
	typeStr, err := compileType(dt, pkg)
	if err != nil {
		return "", nil, err
	}
	switch o.MakeType.Type {
	case "S":
		code := "(make([]" + typeStr + "," + numStr + "))"
		return code, []DataType{BuiltinType{"S", []DataType{dt}}}, nil
	case "L":
		code := "(func () *_std.List {\n"
		code += "var _list _std.List = make([]interface{}, " + numStr + ")\n"
		code += `return &_list
		})()`
		return code, []DataType{BuiltinType{"S", []DataType{dt}}}, nil
	default:
		return "", nil, msg(o.LineNumber, o.Column, "'make' operation requires a slice or list type")
	}
}

func compileOperation(o Operation, pkg *Package, locals map[string]Variable) (string, []DataType, error) {
	if o.Operator == "make" {
		return compileMakeOp(o, pkg, locals)
	}
	operandCode := make([]string, len(o.Operands))
	operandTypes := make([]DataType, len(o.Operands))
	for i, expr := range o.Operands {
		if i == 1 {
			if o.Operator == "get" {
				switch st := operandTypes[0].(type) {
				case Struct:
					switch token := expr.(type) {
					case Token:
						if token.Type == IdentifierWord {
							returnType, err := st.getMemberType(token.Content)
							if err != nil {
								return "", nil, err
							}
							return operandCode[0] + "." + strings.Title(token.Content),
								[]DataType{returnType}, nil
						}
					}
				}
			} else if o.Operator == "set" {
				if len(o.Operands) != 3 {
					return "", nil, msg(o.LineNumber, o.Column, "'set' operation requires 3 operands")
				}
				switch st := operandTypes[0].(type) {
				case Struct:
					switch token := expr.(type) {
					case Token:
						if token.Type == IdentifierWord {
							returnType, err := st.getMemberType(token.Content)
							if err != nil {
								return "", nil, err
							}
							val, valTypes, err := compileExpression(o.Operands[2], pkg, locals)
							if err != nil {
								return "", nil, err
							}
							if len(valTypes) != 1 {
								return "", nil, msg(o.LineNumber, o.Column, "'set' operation value expression should return just one value")
							}
							if !isType(valTypes[0], returnType, false) {
								return "", nil, msg(o.LineNumber, o.Column, "'set' operation value expression has wrong type for the target struct field")
							}
							rt, err := compileType(returnType, pkg)
							if err != nil {
								return "", nil, err
							}
							return "(func () " + rt + " { _t := " + val + "; " + operandCode[0] +
									"." + strings.Title(token.Content) + " = _t; return _t }())",
								[]DataType{returnType}, nil
						}
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
			return "", nil, msg(o.LineNumber, o.Column, "'add' operations requires at least two operands.")
		}
		t := operandTypes[0]
		if !isNumber(t) {
			return "", nil, msg(o.LineNumber, o.Column, "'add' operation has non-number operand")
		}
		for i := range o.Operands {
			if !isType(operandTypes[i], t, true) {
				return "", nil, msg(o.LineNumber, o.Column, "'add' operation has operand whose type differs from the others")
			}
			code += operandCode[i]
			if i < len(o.Operands)-1 {
				code += " + "
			}
		}
		returnType = t
	case "sub":
		if len(o.Operands) < 2 {
			return "", nil, msg(o.LineNumber, o.Column, "'sub' operation requires at least two operands")
		}
		t := operandTypes[0]
		if !isNumber(t) {
			return "", nil, msg(o.LineNumber, o.Column, "'sub' operation has non-number operand")
		}
		for i := range o.Operands {
			if !isType(operandTypes[i], t, true) {
				return "", nil, msg(o.LineNumber, o.Column, "'sub' operation has operand whose type differs from the others")
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
	case "inc":
		if len(o.Operands) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "inc operation requires one operand.")
		}
		t := operandTypes[0]
		if !isNumber(t) {
			return "", nil, msg(o.LineNumber, o.Column, "inc operation has non-number operand")
		}
		code += operandCode[0] + " + 1"
		returnType = t
	case "dec":
		if len(o.Operands) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "dec operation requires one operand.")
		}
		t := operandTypes[0]
		if !isNumber(t) {
			return "", nil, msg(o.LineNumber, o.Column, "dec operation has non-number operand")
		}
		code += operandCode[0] + " - 1"
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
	case "get":
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
				return "", nil, msg(o.LineNumber, o.Column, "get operation on an array requires a number as second operand")
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
				code += operandCode[0] + ".Set(int64(" + operandCode[1] + "), " + operandCode[2] + ")"
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
			code += operandCode[0] + ".Append(" + operandCode[1] + ")"
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
			return "", nil, msg(o.LineNumber, o.Column, "'slice' operation requires three operands")
		}
		if !isNumber(operandTypes[1]) || !isNumber(operandTypes[2]) {
			return "", nil, msg(o.LineNumber, o.Column, "'slice' operation's second and third operands must be numbers.")
		}
		switch t := operandTypes[0].(type) {
		case BuiltinType:
			switch t.Name {
			case "Str":
				returnType = BuiltinType{"S", []DataType{BuiltinType{"Str", nil}}}
			case "S":
				returnType = operandTypes[0]
			default:
				return "", nil, msg(o.LineNumber, o.Column, "'slice' operation's first operand must be a slice or string.")
			}
		case ArrayType:
			returnType = BuiltinType{"S", []DataType{t.Type}}
		default:
			return "", nil, msg(o.LineNumber, o.Column, "'slice' operation requires first operand to be a slice.")
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
					return "", nil, msg(e.LineNumber, e.Column, "Name is undefined: "+name)
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
		returnType = BuiltinType{"Str", nil}
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
	case "getchar":
		if len(o.Operands) != 2 {
			return "", nil, msg(o.LineNumber, o.Column, "getchar operation requires two operands")
		}
		returnType = BuiltinType{"Str", nil}
		if !isType(operandTypes[0], returnType, true) {
			return "", nil, msg(o.LineNumber, o.Column, "getchar's first operand must be a string")
		}
		if !isInteger(operandTypes[1]) {
			return "", nil, msg(o.LineNumber, o.Column, "getchar's second operand must be an integer or byte")
		}
		code += "string(" + operandCode[0] + "[" + operandCode[1] + "])"
	case "getrune":
		if len(o.Operands) != 2 {
			return "", nil, msg(o.LineNumber, o.Column, "getchar operation requires two operands")
		}
		returnType = BuiltinType{"I", nil}
		if !isType(operandTypes[0], returnType, true) {
			return "", nil, msg(o.LineNumber, o.Column, "getchar's first operand must be a string")
		}
		if !isInteger(operandTypes[1]) {
			return "", nil, msg(o.LineNumber, o.Column, "getchar's second operand must be an integer or byte")
		}
		code += operandCode[0] + "[" + operandCode[1] + "]"
	case "charlist":
		if len(o.Operands) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "charlist operation requires one operand")
		}
		returnType = BuiltinType{"L", []DataType{BuiltinType{"Str", nil}}}
		if !isType(operandTypes[0], BuiltinType{"Str", nil}, true) {
			return "", nil, msg(o.LineNumber, o.Column, "charlist operand must be a string")
		}
		code += "_std.Charlist(" + operandCode[0] + ")"
	case "runelist":
		if len(o.Operands) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "runelist operation requires one operand")
		}
		returnType = BuiltinType{"L", []DataType{BuiltinType{"I", nil}}}
		if !isType(operandTypes[0], BuiltinType{"Str", nil}, true) {
			return "", nil, msg(o.LineNumber, o.Column, "runelist operand must be a string")
		}
		code += "_std.Runelist(" + operandCode[0] + ")"
	case "charslice":
		if len(o.Operands) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "charslice operation requires one operand")
		}
		returnType = BuiltinType{"S", []DataType{BuiltinType{"Str", nil}}}
		if !isType(operandTypes[0], BuiltinType{"Str", nil}, true) {
			return "", nil, msg(o.LineNumber, o.Column, "charslice operand must be a string")
		}
		code += "_std.Charslice(" + operandCode[0] + ")"
	case "runeslice":
		if len(o.Operands) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "runeslice operation requires one operand")
		}
		returnType = BuiltinType{"S", []DataType{BuiltinType{"I", nil}}}
		if !isType(operandTypes[0], BuiltinType{"Str", nil}, true) {
			return "", nil, msg(o.LineNumber, o.Column, "runeslice operand must be a string")
		}
		code += "_std.Runeslice(" + operandCode[0] + ")"
	case "byteslice":
		if len(o.Operands) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "'byteslice' operation requires one string operand")
		}
		returnType = BuiltinType{"S", []DataType{BuiltinType{"Byte", nil}}}
		if !isType(operandTypes[0], BuiltinType{"Str", nil}, true) {
			return "", nil, msg(o.LineNumber, o.Column, "'byteslice' operand must be a string")
		}
		code += "[]byte(" + operandCode[0] + ")"
	case "len":
		if len(o.Operands) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "len operation requires one operand")
		}
		returnType = BuiltinType{"I", nil}
		switch t := operandTypes[0].(type) {
		case BuiltinType:
			switch t.Name {
			case "Str":
				code += "_std.StrLen(" + operandCode[0] + ")"
			case "L":
				code += "int64(len(*" + operandCode[0] + "))"
			case "M", "S":
				code += "int64(len(" + operandCode[0] + "))"
			default:
				return "", nil, msg(o.LineNumber, o.Column, "len operand must be a list or map")
			}
		case ArrayType:
			code += "int64(len(" + operandCode[0] + "))"
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
	case "floor":
		if len(o.Operands) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "'floor' operation takes one float operand")
		}
		returnType = BuiltinType{"F", nil}
		if !isType(operandTypes[0], BuiltinType{"F", nil}, true) {
			return "", nil, msg(o.LineNumber, o.Column, "'floor' operation has non-float operand")
		}
		code += "_std.Floor(" + operandCode[0] + ")"
	case "ceil":
		if len(o.Operands) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "'ceil' operation takes one float operand")
		}
		returnType = BuiltinType{"F", nil}
		if !isType(operandTypes[0], BuiltinType{"F", nil}, true) {
			return "", nil, msg(o.LineNumber, o.Column, "'ceil' operation has non-float operand")
		}
		code += "_std.Ceil(" + operandCode[0] + ")"
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
	case "createFile":
		if len(o.Operands) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "'createFile' operation takes one string operand")
		}
		if !isType(operandTypes[0], BuiltinType{"Str", nil}, true) {
			return "", nil, msg(o.LineNumber, o.Column, "'cerateFile' operation has non-string operand")
		}
		code += "_std.CreateFile(" + operandCode[0] + ")"
		code += ")"
		return code, []DataType{BuiltinType{"I", nil}, BuiltinType{"Str", nil}}, nil
	case "openFile":
		if len(o.Operands) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "'openFile' operation takes one string operand")
		}
		if !isType(operandTypes[0], BuiltinType{"Str", nil}, true) {
			return "", nil, msg(o.LineNumber, o.Column, "'openFile' operation has non-string operand")
		}
		code += "_std.OpenFile(" + operandCode[0] + ")"
		code += ")"
		return code, []DataType{BuiltinType{"I", nil}, BuiltinType{"Str", nil}}, nil
	case "closeFile":
		if len(o.Operands) != 1 {
			return "", nil, msg(o.LineNumber, o.Column, "'closeFile' operation takes one integer operand")
		}
		if !isType(operandTypes[0], BuiltinType{"I", nil}, true) {
			return "", nil, msg(o.LineNumber, o.Column, "'closeFile' operation has non-string operand")
		}
		code += "_std.CloseFile(" + operandCode[0] + ")"
		code += ")"
		return code, []DataType{BuiltinType{"Str", nil}}, nil
	case "readFile":
		if len(o.Operands) != 2 {
			return "", nil, msg(o.LineNumber, o.Column, "'readFile' operation takes one integer and one slice of bytes")
		}
		if !isType(operandTypes[0], BuiltinType{"I", nil}, true) {
			return "", nil, msg(o.LineNumber, o.Column, "'readFile' first operator should be an integer (a file id)")
		}
		if !isType(operandTypes[1], BuiltinType{"S", []DataType{BuiltinType{"Byte", nil}}}, true) {
			return "", nil, msg(o.LineNumber, o.Column, "'readFile' second operator should be a slice of bytes")
		}
		code += "_std.ReadFile(" + operandCode[0] + "," + operandCode[1] + ")"
		code += ")"
		return code, []DataType{BuiltinType{"I", nil}, BuiltinType{"Str", nil}}, nil
	case "writeFile":
		if len(o.Operands) != 2 {
			return "", nil, msg(o.LineNumber, o.Column, "'writeFile' operation takes one integer and one slice of bytes")
		}
		if !isType(operandTypes[0], BuiltinType{"I", nil}, true) {
			return "", nil, msg(o.LineNumber, o.Column, "'writeFile' first operator should be an integer (a file id)")
		}
		if !isType(operandTypes[1], BuiltinType{"S", []DataType{BuiltinType{"Byte", nil}}}, true) {
			return "", nil, msg(o.LineNumber, o.Column, "'writeFile' second operator should be a slice of bytes")
		}
		code += "_std.WriteFile(" + operandCode[0] + "," + operandCode[1] + ")"
		code += ")"
		return code, []DataType{BuiltinType{"I", nil}, BuiltinType{"Str", nil}}, nil
	case "seekFile":
		if len(o.Operands) != 2 {
			return "", nil, msg(o.LineNumber, o.Column, "'seekFile' operation takes one integer and one slice of bytes")
		}
		if !isType(operandTypes[0], BuiltinType{"I", nil}, true) {
			return "", nil, msg(o.LineNumber, o.Column, "'seekFile' first operator should be an integer (a file id)")
		}
		if !isType(operandTypes[1], BuiltinType{"S", []DataType{BuiltinType{"Byte", nil}}}, true) {
			return "", nil, msg(o.LineNumber, o.Column, "'seekFile' second operator should be a slice of bytes")
		}
		code += "_std.SeekFile(" + operandCode[0] + "," + operandCode[1] + ")"
		code += ")"
		return code, []DataType{BuiltinType{"I", nil}, BuiltinType{"Str", nil}}, nil
	case "seekFileStart":
		if len(o.Operands) != 2 {
			return "", nil, msg(o.LineNumber, o.Column, "'seekFileStart' operation takes one integer and one slice of bytes")
		}
		if !isType(operandTypes[0], BuiltinType{"I", nil}, true) {
			return "", nil, msg(o.LineNumber, o.Column, "'seekFileStart' first operator should be an integer (a file id)")
		}
		if !isType(operandTypes[1], BuiltinType{"S", []DataType{BuiltinType{"Byte", nil}}}, true) {
			return "", nil, msg(o.LineNumber, o.Column, "'seekFileStart' second operator should be a slice of bytes")
		}
		code += "_std.SeekFileStart(" + operandCode[0] + "," + operandCode[1] + ")"
		code += ")"
		return code, []DataType{BuiltinType{"I", nil}, BuiltinType{"Str", nil}}, nil
	case "seekFileEnd":
		if len(o.Operands) != 2 {
			return "", nil, msg(o.LineNumber, o.Column, "'seekFileStart' operation takes one integer and one slice of bytes")
		}
		if !isType(operandTypes[0], BuiltinType{"I", nil}, true) {
			return "", nil, msg(o.LineNumber, o.Column, "'seekFileStart' first operator should be an integer (a file id)")
		}
		if !isType(operandTypes[1], BuiltinType{"S", []DataType{BuiltinType{"Byte", nil}}}, true) {
			return "", nil, msg(o.LineNumber, o.Column, "'seekFileStart' second operator should be a slice of bytes")
		}
		code += "_std.SeekFileEnd(" + operandCode[0] + "," + operandCode[1] + ")"
		code += ")"
		return code, []DataType{BuiltinType{"I", nil}, BuiltinType{"Str", nil}}, nil
	}

	code += ")"
	if returnType == nil {
		return code, []DataType{}, nil
	}
	return code, []DataType{returnType}, nil
}
