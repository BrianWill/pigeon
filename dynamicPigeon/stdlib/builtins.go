package stdlib

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"reflect"
	"time"
)

type Nil int

type ListType struct {
	List *[]interface{}
}

var Breakpoints = make(map[int]bool)

type MapType map[interface{}]interface{}

func Add(numbers ...interface{}) interface{} {
	if len(numbers) < 2 {
		log.Fatalln("Add operation has too few operands.")
	}
	var sum float64
	for _, n := range numbers {
		switch n := n.(type) {
		case float64:
			sum += n
		default:
			log.Fatalln("Attempted to add a non-number.")
		}
	}
	return sum
}

func Inc(numbers ...interface{}) interface{} {
	if len(numbers) != 1 {
		log.Fatalln("Inc operation has too few operands.")
	}
	val, ok := numbers[0].(float64)
	if !ok {
		log.Fatalln("Attempted to inc a non-number.")
	}
	return val + 1
}

func Dec(numbers ...interface{}) interface{} {
	if len(numbers) != 1 {
		log.Fatalln("Dec operation has too few operands.")
	}
	val, ok := numbers[0].(float64)
	if !ok {
		log.Fatalln("Attempted to inc a non-number.")
	}
	return val - 1
}

func Sub(numbers ...interface{}) interface{} {
	if len(numbers) < 2 {
		log.Fatalln("Sub operation has too few operands.")
	}
	val, ok := numbers[0].(float64)
	if !ok {
		log.Fatalln("Attempted to subtract a non-number.")
	}
	for _, n := range numbers[1:] {
		switch n := n.(type) {
		case float64:
			val -= n
		default:
			log.Fatalln("Attempted to subtract a non-number.")
		}
	}
	return val
}

func Mul(numbers ...interface{}) interface{} {
	if len(numbers) < 2 {
		log.Fatalln("Mul operation has too few operands.")
	}
	product, ok := numbers[0].(float64)
	if !ok {
		log.Fatalln("Attempted to multiply a non-number.")
	}
	for _, n := range numbers[1:] {
		switch n := n.(type) {
		case float64:
			product *= n
		default:
			log.Fatalln("Attempted to multiply a non-number.")
		}
	}
	return product
}

func Div(numbers ...interface{}) interface{} {
	if len(numbers) < 2 {
		log.Fatalln("Div operation has too few operands.")
	}
	quotient, ok := numbers[0].(float64)
	if !ok {
		log.Fatalln("Attempted to divide a non-number.")
	}
	for _, n := range numbers[1:] {
		switch n := n.(type) {
		case float64:
			quotient /= n
		default:
			log.Fatalln("Attempted to divide a non-number.")
		}
	}
	return quotient
}

func Mod(numbers ...interface{}) interface{} {
	if len(numbers) != 2 {
		log.Fatalln("Modulus operation does not have two operands.")
	}
	a, ok1 := numbers[0].(float64)
	b, ok2 := numbers[1].(float64)
	if !ok1 || !ok2 {
		log.Fatalln("Attempted modulus with a non-number.")
	}
	return float64(int(a) % int(b))
}

func Eq(values ...interface{}) interface{} {
	if len(values) < 2 {
		log.Fatalln("Attempted equality test with fewer than 2 operands.")
	}

	for _, val := range values {
		switch val.(type) {
		case float64, bool, string, Nil, ListType, MapType:
		default:
			log.Fatalln("Attempted equality test with type other than a number, boolean, string, or null.")
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
	case Nil:
		for _, v := range values[1:] {
			_, ok := v.(Nil)
			if !ok {
				return false
			}
		}
	case ListType:
		for _, v := range values[1:] {
			_, ok := v.(ListType)
			if !ok || reflect.DeepEqual(v, val) {
				return false
			}
		}
	case MapType:
		for _, v := range values[1:] {
			_, ok := v.(MapType)
			if !ok || reflect.DeepEqual(v, val) {
				return false
			}
		}
	}
	return true
}

func Neq(values ...interface{}) interface{} {
	return !Eq(values...).(bool)
}

func Id(vals ...interface{}) interface{} {
	if len(vals) < 2 {
		log.Fatalln("Too few operands for 'id' operation.")
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
		log.Fatalln("Incorrect number of operands for get operation.")
	}
	b, ok := vals[0].(bool)
	if !ok {
		log.Fatalln("Attempted logical not operation on a non-boolean value.")
	}
	return !b
}

func Lt(numbers ...interface{}) interface{} {
	if len(numbers) < 2 {
		log.Fatalln("Too few operands for 'lt' operation.")
	}
	prev, ok := numbers[0].(float64)
	if !ok {
		log.Fatalln("Attempted 'lt' operation on a non-number.")
	}
	for _, n := range numbers[1:] {
		f, ok := n.(float64)
		if !ok {
			log.Fatalln("Attempted 'lt' operation on a non-number.")
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
		log.Fatalln("Too few operands for 'gt' operation.")
	}
	prev, ok := numbers[0].(float64)
	if !ok {
		log.Fatalln("Attempted 'gt' operation on a non-number.")
	}
	for _, n := range numbers[1:] {
		f, ok := n.(float64)
		if !ok {
			log.Fatalln("Attempted 'gt' operation on a non-number.")
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
		log.Fatalln("Too few operands for 'lte' operation.")
	}
	prev, ok := numbers[0].(float64)
	if !ok {
		log.Fatalln("Attempted 'lte' operation on a non-number.")
	}
	for _, n := range numbers[1:] {
		f, ok := n.(float64)
		if !ok {
			log.Fatalln("Attempted 'lte' operation on a non-number.")
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
		log.Fatalln("Too few operands for 'gte' operation.")
	}
	prev, ok := numbers[0].(float64)
	if !ok {
		log.Fatalln("Attempted 'gte' operation on a non-number.")
	}
	for _, n := range numbers[1:] {
		f, ok := n.(float64)
		if !ok {
			log.Fatalln("Attempted 'gte' operation on a non-number.")
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
		log.Fatalln("Incorrect number of operands for 'get' operation.")
	}
	switch v := args[0].(type) {
	case ListType:
		f, ok := args[1].(float64)
		if !ok {
			log.Fatalln("Second operand to 'get' on a list should be a number.")
		}
		if int(f) >= len(*v.List) {
			log.Fatalln("Index of 'get' exceeds bounds of the list.")
		}
		return (*v.List)[int(f)]
	case MapType:
		switch key := args[1].(type) {
		case float64, string:
			return v[key]
		default:
			log.Fatalln("Second operand to 'get' on a map should be a string or number.")
		}
	default:
		log.Fatalln("First operand to 'get' must be a map or a list.")
	}
	return nil
}

func Set(args ...interface{}) interface{} {
	if len(args) != 3 {
		log.Fatalln("Incorrect number of operands for 'set' operation.")
	}
	switch v := args[0].(type) {
	case ListType:
		f, ok := args[1].(float64)
		if !ok {
			log.Fatalln("Second operand to 'set' on a list should be a number.")
		}
		(*v.List)[int(f)] = args[2]
	case MapType:
		switch key := args[1].(type) {
		case float64, string:
			v[key] = args[2]
		default:
			log.Fatalln("Second operand to 'set' on a map should be a string or number.")
		}
	default:
		log.Fatalln("First operand to 'set' must be a map or a list.")
	}
	return Nil(0)
}

func Push(args ...interface{}) interface{} {
	if len(args) < 2 {
		log.Fatalln("Too few operands for 'push' operation.")
	}
	list, ok := args[0].(ListType)
	if !ok {
		log.Fatalln("First operand to 'push' must be a list.")
	}
	for _, v := range args[1:] {
		*list.List = append(*list.List, v)
	}
	return Nil(0)
}

func Or(args ...interface{}) interface{} {
	if len(args) < 2 {
		log.Fatalln("Too few operands for 'or' operation.")
	}
	for _, a := range args {
		b, ok := a.(bool)
		if !ok {
			log.Fatalln("Operands of 'or' must be booleans.")
		}
		if b {
			return true
		}
	}
	return false
}

func And(args ...interface{}) interface{} {
	if len(args) < 2 {
		log.Fatalln("Too few operands for 'or' operation.")
	}
	for _, a := range args {
		b, ok := a.(bool)
		if !ok {
			log.Fatalln("Operands of 'or' must be booleans.")
		}
		if !b {
			return false
		}
	}
	return true
}

func Print(args ...interface{}) interface{} {
	if len(args) == 0 {
		log.Fatalln("Print operation needs at least one operand.")
	}
	fmt.Print(args...)
	return Nil(0)
}

func Println(args ...interface{}) interface{} {
	if len(args) == 0 {
		log.Fatalln("Println operation needs at least one operand.")
	}
	fmt.Println(args...)
	return Nil(0)
}

func Prompt(args ...interface{}) interface{} {
	if len(args) >= 1 {
		fmt.Println(args...)
	}
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	if scanner.Err() != nil {
		log.Fatalln(scanner.Err())
	}
	s := scanner.Text()
	return s
}

func List(args ...interface{}) interface{} {
	list := make([]interface{}, len(args))
	for i, a := range args {
		list[i] = a
	}
	return ListType{&list}
}

func Concat(args ...interface{}) interface{} {
	if len(args) < 1 {
		log.Fatalln("Concat operation needs two or more operands.")
	}
	return fmt.Sprint(args...)
}

func Floor(args ...interface{}) interface{} {
	if len(args) != 1 {
		log.Fatalln("'floor' operation needs one operand.")
	}
	switch v := args[0].(type) {
	case float64:
		return math.Floor(v)
	default:
		log.Fatalln("'floor' operation operand must be a number")
		return nil
	}
}

func RandNum(args ...interface{}) interface{} {
	if len(args) != 0 {
		log.Fatalln("'randNum' operation should have no operands.")
	}
	return rand.Float64()
}

func Map(args ...interface{}) interface{} {
	if len(args) == 0 {
		log.Fatalln("'Map' operation needs at least one operand.")
	}
	if len(args)%2 != 0 {
		log.Fatalln("'Map' operations needs an even number of operands.")
	}
	_map := make(MapType)
	for i := 0; i < len(args); {
		_map[args[i]] = args[i+1]
		i += 2
	}
	return _map
}

func Len(args ...interface{}) interface{} {
	if len(args) != 1 {
		log.Fatalln("'len' operator must have just one operand.")
	}
	switch a := args[0].(type) {
	case ListType:
		return float64(len(*a.List))
	case MapType:
		return float64(len(a))
	case string:
		return float64(len(a))
	default:
		log.Fatalln("'len' operator operand must be a map or list.")
		return nil
	}
}

func Charlist(args ...interface{}) interface{} {
	if len(args) != 1 {
		log.Fatalln("'charlist' operation needs one operand.")
	}
	s, ok := args[0].(string)
	if !ok {
		log.Fatalln("'charlist' operation operand must be a string.")
	}
	list := make([]interface{}, len(s))
	for i, a := range []rune(s) {
		list[i] = string(a)
	}
	return ListType{&list}
}

func Getchar(args ...interface{}) interface{} {
	if len(args) != 2 {
		log.Fatalln("'getchar' operation needs two operands.")
	}
	s, ok := args[0].(string)
	if !ok {
		log.Fatalln("'getchar' operation's first operand must be a string.")
	}
	idx, ok := args[1].(float64)
	if !ok {
		log.Fatalln("'getchar' operation's second operand must be a number.")
	}
	for i, a := range []rune(s) {
		if i == int(idx) {
			return string(a)
		}
	}
	log.Fatalln("index for 'getchar' operation is out of bounds")
	return ListType{}
}

func (l ListType) String() string {
	list := l.List
	if len(*l.List) == 0 {
		return "[]"
	}
	s := "["
	for _, v := range *list {
		s += fmt.Sprintf("%v ", v)
	}
	return s[:len(s)-1] + "]"
}

type DebugVar struct {
	name   string
	val    interface{}
	global bool
}

// do nothing (used to supress unused variable compile errors)
func NullOp(args ...interface{}) {
	// do nothing
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
