package main

import _fmt "fmt"
import _p1 "p1"

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

func _main() {
	debug := func(line int) {
		var globals = map[string]interface{}{
			"bar": g_bar,
		}
		var locals = map[string]interface{}{}
		//_p.PollContinue(line, globals, locals)
	}
	if _breakpoints[7] {
		debug(7)
	}
	(_fmt.Print(_p1.Foo()))
	if _breakpoints[8] {
		debug(8)
	}
	(_fmt.Print(_p1.G_bar))

}

func main() {
	go _p.PollBreakpoints(&_breakpoints)
	_main()
}
