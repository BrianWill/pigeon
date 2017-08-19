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
	var x float64
	var y float64
	debug := func(line int) {
		var globals = map[string]interface{}{}
		var locals = map[string]interface{}{
			"x": x,
			"y": y,
		}
		//_p.PollContinue(line, globals, locals)
	}
	if _breakpoints[12] {
		debug(12)
	}
	x = sum(y, float64(3))
	if _breakpoints[13] {
		debug(13)
	}
	(_fmt.Print(x))
	if _breakpoints[14] {
		debug(14)
	}
	if float64(3) == float64(6) {
		if _breakpoints[15] {
			debug(15)
		}
		(_Prompt("yo"))
	}
	if _breakpoints[16] {
		debug(16)
	}
	(_fmt.Print("yo"))
	if _breakpoints[17] {
		debug(17)
	}
	(_fmt.Print("bla"))
	return nil
}

func main() {
	go _p.PollBreakpoints(&_breakpoints)
	_main()
}
