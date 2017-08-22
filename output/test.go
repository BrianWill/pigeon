package main

import _fmt "fmt"

var _breakpoints = make(map[int]bool)

type _List *[]interface{}

func _newList(items ...interface{}) _List {
	return &items
}

func (l _List) append(item interface{}) {
	*l = append(*l, item)
}

func (l _List) set(idx float64, item interface{}) {
	(*l)[int64(idx)] = item
}

func (l _List) get(idx float64, item interface{}) interface{} {
	return (*l)[int64(idx)]
}

func (l _List) len() float64 {
	return float64(len(*l))
}

func _Prompt(args ...interface{}) {
	if len(args) > 1 {
		_fmt.Print(args...)
	}

}

type Roger interface {
	foo(float64, string) (float64, Foo)
}
type Foo struct {
	bar string
	cat float64
}

func (f Foo) foo(a float64, b string) (float64, Foo) {
	debug := func(line int) {
		var globals = map[string]interface{}{}
		var locals = map[string]interface{}{
			"b": b,
			"a": a,
		}
		//_p.PollContinue(line, globals, locals)
	}
	if _breakpoints[21] {
		debug(21)
	}
	(_fmt.Print("hi"))
	if _breakpoints[22] {
		debug(22)
	}
	return float64(3), Foo{"hi", float64(5)}

}
func (f Foo) doStuff(apple float64) string {
	debug := func(line int) {
		var globals = map[string]interface{}{}
		var locals = map[string]interface{}{
			"apple": apple,
		}
		//_p.PollContinue(line, globals, locals)
	}
	if _breakpoints[25] {
		debug(25)
	}
	return "hi"

}
func _main() {
	var a Foo
	var b _List
	var c map[float64]string
	var d Roger
	debug := func(line int) {
		var globals = map[string]interface{}{}
		var locals = map[string]interface{}{
			"a": a,
			"b": b,
			"c": c,
			"d": d,
		}
		//_p.PollContinue(line, globals, locals)
	}
	if _breakpoints[29] {
		debug(29)
	}
	a = Foo{"hi", float64(3)}
	if _breakpoints[30] {
		debug(30)
	}
	d = a
	if _breakpoints[31] {
		debug(31)
	}
	d.foo(float64(5), "hi")
	if _breakpoints[32] {
		debug(32)
	}
	a.doStuff(float64(4))
	if _breakpoints[33] {
		debug(33)
	}
	b = (func() (_list _List) {
		(*_list) = make([]interface{}, 3)
		(*_list)[0] = float64(5)
		(*_list)[1] = float64(2)
		(*_list)[2] = float64(9)
		return
	})()
	if _breakpoints[34] {
		debug(34)
	}
	c = map[float64]string{float64(5): "hi", float64(9): "yo"}
	if _breakpoints[35] {
		debug(35)
	}
	b[float64(0)] = float64(3)
	if _breakpoints[36] {
		debug(36)
	}
	c[float64(3)] = "hi"
	if _breakpoints[37] {
		debug(37)
	}
	(_fmt.Print("bla"))

}
func sum(a float64, b float64) float64 {
	debug := func(line int) {
		var globals = map[string]interface{}{}
		var locals = map[string]interface{}{
			"a": a,
			"b": b,
		}
		//_p.PollContinue(line, globals, locals)
	}
	if _breakpoints[3] {
		debug(3)
	}
	return (a + b)

}
func giveNums() (float64, float64) {
	var a float64
	debug := func(line int) {
		var globals = map[string]interface{}{}
		var locals = map[string]interface{}{
			"a": a,
		}
		//_p.PollContinue(line, globals, locals)
	}
	if _breakpoints[7] {
		debug(7)
	}
	a = float64(5)
	if _breakpoints[8] {
		debug(8)
	}
	return a, float64(7)

}
func doNothing() {
	debug := func(line int) {
		var globals = map[string]interface{}{}
		var locals = map[string]interface{}{}
		//_p.PollContinue(line, globals, locals)
	}
	if _breakpoints[11] {
		debug(11)
	}
	(_fmt.Print("hi"))

}

func main() {
	go _p.PollBreakpoints(&_breakpoints)
	_main()
}
