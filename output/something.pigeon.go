package main

import _p "github.com/BrianWill/pigeon/stdlib"

var _breakpoints = make(map[int]bool)

func _main() interface{} {
	debug := func(line int) {
		var globals = map[string]interface{}{}
		var locals = map[string]interface{}{}
		_p.PollContinue(line, globals, locals)
	}
	if _breakpoints[2] {
		debug(2)
	}
	_p.Print("hi")
	if _breakpoints[3] {
		debug(3)
	}
	_p.Print("yo")
	if _breakpoints[4] {
		debug(4)
	}
	_p.Print("bla")
	return nil
}

func main() {
	go _p.PollBreakpoints(&_breakpoints)
	_main()
}
