package main

import _pigeon "github.com/BrianWill/pigeon/stdlib"

var _bruce interface{}
var _y interface{}
var _ted interface{}
var _x interface{}
var _by5 interface{}
var _thom interface{}
var _z interface{}
var _by3 interface{}
var _leo interface{}
var _ed interface{}

func _eric() interface{} {
	_pigeon.Print("hello")
	return nil
}
func _ryan(_bat interface{}, _goat interface{}) interface{} {
	var _ted2 interface{}
	_ted2 = float64(4)
	_pigeon.Print(_pigeon.Sub(_goat, _ted2))
	return nil
}
func main() {
	_eric()
	_ryan(float64(4), float64(-9))
	_ryan(float64(3), float64(5))
	_ted = _pigeon.Dict(float64(5), float64(2), "yo", float64(-1))
	_pigeon.Print(_pigeon.Len(_ted))
	_pigeon.Print(_pigeon.Get(_ted, "yo"))
	_pigeon.Set(_ted, "yo", float64(8))
	_pigeon.Print(_pigeon.Get(_ted, "yo"))
	_pigeon.Set(_ted, float64(21), false)
	_pigeon.Print(_pigeon.Get(_ted, float64(21)))
	_pigeon.Print("\n\"\tyup!\"\n")
	_x = float64(1)
	for _pigeon.Lte(_x, float64(15)).(bool) {
		_by3 = _pigeon.Eq(_pigeon.Mod(_x, float64(3)), float64(0))
		_by5 = _pigeon.Eq(_pigeon.Mod(_x, float64(5)), float64(0))
		if _pigeon.And(_by3, _by5).(bool) {
			_pigeon.Print("fizzbuzz")
		} else if _by3.(bool) {
			_pigeon.Print("fizz")
		} else if _by5.(bool) {
			_pigeon.Print("buzz")
		} else {
			_pigeon.Print(_x)
		}
		_x = _pigeon.Add(_x, float64(1))
	}
	_leo = _pigeon.List()
	_bruce = _leo
	_pigeon.Print(_pigeon.Len(_leo))
	_pigeon.Append(_bruce, "salut")
	_pigeon.Print(_pigeon.Len(_leo))
	_ed = _pigeon.List(float64(88), float64(8))
	_thom = _pigeon.List(float64(88), float64(8))
	_pigeon.Print(_pigeon.Id(_ed, _thom))
	_thom = _ed
	_pigeon.Print(_pigeon.Id(_ed, _thom))
	_x = float64(1)
	_y = _x
	_z = _x
	_pigeon.Print(_pigeon.Id(_x, _z))
	_x = float64(-7)
	_y = float64(-7)
	if _pigeon.Eq(_x, float64(3)).(bool) {
		_pigeon.Print("hi")
	} else if _pigeon.Eq(_x, float64(5)).(bool) {
		_pigeon.Print("bye")
	} else if _pigeon.Eq(_x, float64(-7), _y).(bool) {
		_pigeon.Print("moo")
	} else if _pigeon.Eq(_x, float64(14)).(bool) {
		_pigeon.Print("woof")
	} else {
		_pigeon.Print("grunt")
	}
	_x = float64(8)
	if _pigeon.Eq(_pigeon.Mod(_x, float64(2)), float64(0)).(bool) {
		_pigeon.Print("even")
	} else {
		_pigeon.Print("odd")
	}
	_pigeon.Print(_pigeon.And(true, false))
	_pigeon.Print(_pigeon.Or(true, false, true))
	_pigeon.Print(_pigeon.Or(false, false))
	_z = float64(6)
	for _pigeon.Gt(_z, float64(0)).(bool) {
		_pigeon.Print(_z)
		_z = _pigeon.Sub(_z, float64(1))
	}
}
