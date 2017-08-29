package main

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
			"a": a,
			"b": b,
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
func testing() {
	var n float64
	var p *float64
	debug := func(line int) {
		var globals = map[string]interface{}{}
		var locals = map[string]interface{}{
			"n": n,
			"p": p,
		}
		//_p.PollContinue(line, globals, locals)
	}
	if _breakpoints[33] {
		debug(33)
	}
	p = (&n)
	if _breakpoints[34] {
		debug(34)
	}
	n = (*p)

}
func _main() {
	var a Foo
	var b *_List
	var c map[float64]string
	var d Roger
	var e bool
	debug := func(line int) {
		var globals = map[string]interface{}{}
		var locals = map[string]interface{}{
			"a": a,
			"b": b,
			"c": c,
			"d": d,
			"e": e,
		}
		//_p.PollContinue(line, globals, locals)
	}
	if _breakpoints[38] {
		debug(38)
	}
	(b.append("hi"))
	if _breakpoints[39] {
		debug(39)
	}
	a = Foo{"hi", float64(3)}
	if _breakpoints[40] {
		debug(40)
	}
	for _i, _v := range b {
		i = _i
		v = _v
		if _breakpoints[41] {
			debug(41)
		}
		(_fmt.Print(i, v))
	}
	if _breakpoints[42] {
		debug(42)
	}
	d = a
	if _breakpoints[43] {
		debug(43)
	}
	d.foo(float64(5), "hi")
	if _breakpoints[44] {
		debug(44)
	}
	a.doStuff(float64(4))
	if _breakpoints[45] {
		debug(45)
	}
	{
		_inter := d
		if a, _ok := _inter.(Foo); _ok {
			if _breakpoints[47] {
				debug(47)
			}
			(_fmt.Println("d is Foo"))
		} else {
			if _breakpoints[49] {
				debug(49)
			}
			(_fmt.Println("d is not Foo"))
		}
	}
	if _breakpoints[50] {
		debug(50)
	}
	b = (func() (_list _List) {
		(*_list) = make([]interface{}, 2)
		(*_list)[0] = "yo"
		(*_list)[1] = "byte"
		return
	})()
	if _breakpoints[51] {
		debug(51)
	}
	c = map[float64]string{float64(5): "hi", float64(9): "yo"}
	if _breakpoints[52] {
		debug(52)
	}
	(*b[int64(float64(0))].(string))
	if _breakpoints[53] {
		debug(53)
	}
	(*b[int64(float64(0))].(string)) = "asdf"
	if _breakpoints[54] {
		debug(54)
	}
	(c[float64(3)]) = "hi"
	if _breakpoints[55] {
		debug(55)
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
