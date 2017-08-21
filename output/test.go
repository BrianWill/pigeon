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

type Foo struct {
	bar string
	cat float64
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
func _main() {
	var a Foo
	var b _List
	var c map[float64]string
	debug := func(line int) {
		var globals = map[string]interface{}{}
		var locals = map[string]interface{}{
			"a": a,
			"b": b,
			"c": c,
		}
		//_p.PollContinue(line, globals, locals)
	}
	if _breakpoints[16] {
		debug(16)
	}
	a = Foo{"hi", float64(3)}
	if _breakpoints[17] {
		debug(17)
	}
	b = (func() (_list _List) {
		(*_list) = make([]interface{}, 3)
		(*_list)[0] = float64(5)
		(*_list)[1] = float64(2)
		(*_list)[2] = float64(9)
		return
	})()
	if _breakpoints[18] {
		debug(18)
	}
	c = map[float64]string{float64(5): "hi", float64(9): "yo"}
	if _breakpoints[19] {
		debug(19)
	}
	b[float64(0)] = float64(3)
	if _breakpoints[20] {
		debug(20)
	}
	c[float64(3)] = "hi"
	if _breakpoints[21] {
		debug(21)
	}
	(_fmt.Print("bla"))
	return nil
}

func main() {
	go _p.PollBreakpoints(&_breakpoints)
	_main()
}
