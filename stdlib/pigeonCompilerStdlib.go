package stdlib

import "fmt"

type Null int

func Add(numbers ...interface{}) interface{} {
	var sum float64
	for _, n := range numbers {
		switch n := n.(type) {
		case float64:
			sum += n
		default:
			panic("Attempted to add a non-number.")
		}
	}
	return sum
}

func Sub(numbers ...interface{}) interface{} {
	var sum float64
	for _, n := range numbers {
		switch n := n.(type) {
		case float64:
			sum -= n
		default:
			panic("Attempted to subtract a non-number.")
		}
	}
	return sum
}

func Mul(numbers ...interface{}) interface{} {
	if len(numbers) < 2 {
		panic("Multiplication operation has too few operands.")
	}
	product, ok := numbers[0].(float64)
	if !ok {
		panic("Attempted to multiply a non-number.")
	}
	for _, n := range numbers[1:] {
		switch n := n.(type) {
		case float64:
			product *= n
		default:
			panic("Attempted to multiply a non-number.")
		}
	}
	return product
}

func Div(numbers ...interface{}) interface{} {
	if len(numbers) < 2 {
		panic("Division operation has too few operands.")
	}
	quotient, ok := numbers[0].(float64)
	if !ok {
		panic("Attempted to divide a non-number.")
	}
	for _, n := range numbers[1:] {
		switch n := n.(type) {
		case float64:
			quotient /= n
		default:
			panic("Attempted to divide a non-number.")
		}
	}
	return quotient
}

func Mod(numbers ...interface{}) interface{} {
	if len(numbers) != 2 {
		panic("Modulus operation does not have two operands.")
	}
	a, ok1 := numbers[0].(float64)
	b, ok2 := numbers[1].(float64)
	if !ok1 || !ok2 {
		panic("Attempted modulus with a non-number.")
	}
	return float64(int(a) % int(b))
}

func Eq(values ...interface{}) interface{} {
	if len(values) < 2 {
		panic("Attempted equality test with fewer than 2 operands.")
	}

	for _, val := range values {
		switch val.(type) {
		case float64, bool, string, Null:
		default:
			panic("Attempted quality test with type other than a number, boolean, string, or null.")
		}
	}

	switch val := values[0].(type) {
	case float64:
		for _, v := range values[1:] {
			f, ok := v.(float64)
			if !ok || f != val {
				return false
			}
		}
	case bool:
		for _, v := range values[1:] {
			b, ok := v.(bool)
			if !ok || b != val {
				return false
			}
		}
	case string:
		for _, v := range values[1:] {
			s, ok := v.(string)
			if !ok || s != val {
				return false
			}
		}
	case Null:
		for _, v := range values[1:] {
			_, ok := v.(Null)
			if !ok {
				return false
			}
		}
	}
	return true
}

func Neq(values ...interface{}) interface{} {
	return !Eq(values).(bool)
}

func Id(vals ...interface{}) interface{} {
	if len(vals) < 2 {
		panic("Too few operands for 'id' operation.")
	}
	first := vals[0]
	for _, v := range vals[1:] {
		if first != v {
			return false
		}
	}
	return true
}

func Not(vals ...interface{}) interface{} {
	if len(vals) != 1 {
		panic("Incorrect number of operands for get operation.")
	}
	b, ok := vals[0].(bool)
	if !ok {
		panic("Attempted logical not operation on a non-boolean value.")
	}
	return !b
}

func Lt(numbers ...interface{}) interface{} {
	if len(numbers) < 2 {
		panic("Too few operands for 'lt' operation.")
	}
	prev, ok := numbers[0].(float64)
	if !ok {
		panic("Attempted 'lt' operation on a non-number.")
	}
	for _, n := range numbers[1:] {
		f, ok := n.(float64)
		if !ok {
			panic("Attempted 'lt' operation on a non-number.")
		}
		if prev >= f {
			return false
		}
		prev = f
	}
	return true
}

func Gt(numbers ...interface{}) interface{} {
	if len(numbers) < 2 {
		panic("Too few operands for 'gt' operation.")
	}
	prev, ok := numbers[0].(float64)
	if !ok {
		panic("Attempted 'gt' operation on a non-number.")
	}
	for _, n := range numbers[1:] {
		f, ok := n.(float64)
		if !ok {
			panic("Attempted 'gt' operation on a non-number.")
		}
		if prev <= f {
			return false
		}
		prev = f
	}
	return true
}

func Lte(numbers ...interface{}) interface{} {
	if len(numbers) < 2 {
		panic("Too few operands for 'lte' operation.")
	}
	prev, ok := numbers[0].(float64)
	if !ok {
		panic("Attempted 'lte' operation on a non-number.")
	}
	for _, n := range numbers[1:] {
		f, ok := n.(float64)
		if !ok {
			panic("Attempted 'lte' operation on a non-number.")
		}
		if prev > f {
			return false
		}
		prev = f
	}
	return true
}

func Gte(numbers ...interface{}) interface{} {
	if len(numbers) < 2 {
		panic("Too few operands for 'gte' operation.")
	}
	prev, ok := numbers[0].(float64)
	if !ok {
		panic("Attempted 'gte' operation on a non-number.")
	}
	for _, n := range numbers[1:] {
		f, ok := n.(float64)
		if !ok {
			panic("Attempted 'gte' operation on a non-number.")
		}
		if prev < f {
			return false
		}
		prev = f
	}
	return true
}

func Get(args ...interface{}) interface{} {
	if len(args) != 2 {
		panic("Incorrect number of operands for 'get' operation.")
	}
	switch v := args[0].(type) {
	case *[]interface{}:
		f, ok := args[1].(float64)
		if !ok {
			panic("Second operand to 'get' on a list should be a number.")
		}
		return (*v)[int(f)]
	case map[interface{}]interface{}:
		switch key := args[1].(type) {
		case float64, string:
			return v[key]
		default:
			panic("Second operand to 'get' on a dict should be a string or number.")
		}		
	default:
		panic("First operand to 'get' must be a dict or a list.")
	}
	return nil
}

func Set(args ...interface{}) interface{} {
	if len(args) != 3 {
		panic("Incorrect number of operands for 'set' operation.")
	}
	switch v := args[0].(type) {
	case *[]interface{}:
		f, ok := args[1].(float64)
		if !ok {
			panic("Second operand to 'set' on a list should be a number.")
		}
		(*v)[int(f)] = args[2]
	case map[interface{}]interface{}:
		switch key := args[1].(type) {
		case float64, string:
			v[key] = args[2]
		default:
			panic("Second operand to 'set' on a dict should be a string or number.")
		}
	default:
		panic("First operand to 'set' must be a dict or a list.")
	}
	return Null(0)
}

func Append(args ...interface{}) interface{} {
	if len(args) < 2 {
		panic("Too few operands for 'append' operation.")
	}
	list, ok := args[0].(*[]interface{})
	if !ok {
		panic("First operand to 'append' must be a list.")
	}
	for _, v := range args[1:] {
		*list = append(*list, v)
	}
	return Null(0)
}

func Or(args ...interface{}) interface{} {
	if len(args) < 2 {
		panic("Too few operands for 'or' operation.")
	}
	for _, a := range args {
		b, ok := a.(bool)
		if !ok {
			panic("Operands of 'or' must be booleans.")
		}
		if b {
			return true
		}
	}
	return false
}

func And(args ...interface{}) interface{} {
	if len(args) < 2 {
		panic("Too few operands for 'or' operation.")
	}
	for _, a := range args {
		b, ok := a.(bool)
		if !ok {
			panic("Operands of 'or' must be booleans.")
		}
		if !b {
			return false
		}
	}
	return true
}

func Print(args ...interface{}) interface{} {
	if len(args) == 0 {
		panic("Print operation needs at least one operand.")
	}
	fmt.Println(args[0])
	// TODO may need to customize printing for some types
	return Null(0)
}

func List(args ...interface{}) interface{} {
	list := make([]interface{}, len(args))
	for i, a := range args {
		list[i] = a
	}
	return &list
}

func Dict(args ...interface{}) interface{} {
	if len(args) == 0 {
		panic("'Dict' operation needs at least one operand.")
	}
	if len(args) % 2 != 0 {
		panic("'Dict' operations needs an even number of operands.")
	}
	dict := make(map[interface{}]interface{})
	for i := 0; i < len(args); {
		dict[args[i]] = args[i + 1] 
		i += 2
	}
	return dict
}

func Len(args ...interface{}) interface{} {
	if len(args) != 1 {
		panic("'len' operator must have just one operand.")
	}
	switch a := args[0].(type) {
	case *[]interface{}:
		return len(*a)
	case map[interface{}]interface{}:
		return len(a)
	default:
		panic("'len' operator operand must be a dict or list.")
	}
	return nil
}