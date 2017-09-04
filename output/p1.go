package p1

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

var G_bar float64 = float64(4)

func Foo() float64 {
	debug := func(line int) {
		var globals = map[string]interface{}{
			"bar": g_bar,
		}
		var locals = map[string]interface{}{}
		//_p.PollContinue(line, globals, locals)
	}
	if _breakpoints[2] {
		debug(2)
	}
	return float64(5)

}
