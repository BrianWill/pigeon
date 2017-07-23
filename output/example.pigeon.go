package main

import _pigeon "github.com/BrianWill/pigeon/stdlib"

var _topRow interface{} = _pigeon.List("_", "_", "_")
var _middleRow interface{} = _pigeon.List("_", "_", "_")
var _bottomRow interface{} = _pigeon.List("_", "_", "_")

func _playerMove(_currentPlayer interface{}) interface{} {
	var _move, _row, _col, _slot interface{}
	_move = _pigeon.Null(0)
	for _pigeon.Eq(_move, _pigeon.Null(0)).(bool) {
		_row = _pigeon.Null(0)
		for _pigeon.Eq(_row, _pigeon.Null(0)).(bool) {
			_row = _pigeon.Prompt("Select [t]op, [m]iddle, or [b]ottom row, player ", _currentPlayer)
			if _pigeon.Eq(_row, "t").(bool) {
				_row = _topRow
			} else if _pigeon.Eq(_row, "m").(bool) {
				_row = _middleRow
			} else if _pigeon.Eq(_row, "b").(bool) {
				_row = _bottomRow
			} else {
				_pigeon.Print("Invalid input. Try again.")
				_row = _pigeon.Null(0)
			}
		}
		_col = _pigeon.Null(0)
		for _pigeon.Eq(_col, _pigeon.Null(0)).(bool) {
			_col = _pigeon.Prompt("Select [l]eft, [m]iddle, or [r]ight column, player ", _currentPlayer)
			if _pigeon.Eq(_col, "l").(bool) {
				_row = float64(0)
			} else if _pigeon.Eq(_row, "m").(bool) {
				_row = float64(1)
			} else if _pigeon.Eq(_row, "r").(bool) {
				_row = float64(2)
			} else {
				_pigeon.Print("Invalid input. Try again.")
				_row = _pigeon.Null(0)
			}
		}
		_slot = _pigeon.Get(_row, _col)
		if _pigeon.Eq(_slot, "_").(bool) {
			_pigeon.Set(_row, _col, _currentPlayer)
			_move = true
		} else {
			_pigeon.Print("That slot is occupied! Try again.")
		}
	}
	return nil
}
func _winner() interface{} {
	var _topRowFull, _middleRowFull, _bottomRowFull interface{}
	if _pigeon.And(_pigeon.Neq(_pigeon.Get(_topRow, float64(0)), "_"), _pigeon.Eq(_pigeon.Get(_topRow, float64(0)), _pigeon.Get(_topRow, float64(1)), _pigeon.Get(_topRow, float64(2)))).(bool) {
		return _pigeon.Get(_topRow, float64(0))
	}
	if _pigeon.And(_pigeon.Neq(_pigeon.Get(_middleRow, float64(0)), "_"), _pigeon.Eq(_pigeon.Get(_middleRow, float64(0)), _pigeon.Get(_middleRow, float64(1)), _pigeon.Get(_middleRow, float64(2)))).(bool) {
		return _pigeon.Get(_middleRow, float64(0))
	}
	if _pigeon.And(_pigeon.Neq(_pigeon.Get(_bottomRow, float64(0)), "_"), _pigeon.Eq(_pigeon.Get(_bottomRow, float64(0)), _pigeon.Get(_bottomRow, float64(1)), _pigeon.Get(_bottomRow, float64(2)))).(bool) {
		return _pigeon.Get(_bottomRow, float64(0))
	}
	if _pigeon.And(_pigeon.Neq(_pigeon.Get(_topRow, float64(0)), "_"), _pigeon.Eq(_pigeon.Get(_topRow, float64(0)), _pigeon.Get(_middleRow, float64(0)), _pigeon.Get(_bottomRow, float64(0)))).(bool) {
		return _pigeon.Get(_topRow, float64(0))
	}
	if _pigeon.And(_pigeon.Neq(_pigeon.Get(_topRow, float64(1)), "_"), _pigeon.Eq(_pigeon.Get(_topRow, float64(1)), _pigeon.Get(_middleRow, float64(1)), _pigeon.Get(_bottomRow, float64(1)))).(bool) {
		return _pigeon.Get(_topRow, float64(1))
	}
	if _pigeon.And(_pigeon.Neq(_pigeon.Get(_topRow, float64(2)), "_"), _pigeon.Eq(_pigeon.Get(_topRow, float64(2)), _pigeon.Get(_middleRow, float64(2)), _pigeon.Get(_middleRow, float64(2)))).(bool) {
		return _pigeon.Get(_topRow, float64(2))
	}
	if _pigeon.And(_pigeon.Neq(_pigeon.Get(_topRow, float64(0)), "_"), _pigeon.Eq(_pigeon.Get(_topRow, float64(0)), _pigeon.Get(_middleRow, float64(1)), _pigeon.Get(_bottomRow, float64(2)))).(bool) {
		return _pigeon.Get(_topRow, float64(0))
	}
	if _pigeon.And(_pigeon.Neq(_pigeon.Get(_bottomRow, float64(0)), "_"), _pigeon.Eq(_pigeon.Get(_bottomRow, float64(0)), _pigeon.Get(_middleRow, float64(1)), _pigeon.Get(_topRow, float64(2)))).(bool) {
		return _pigeon.Get(_bottomRow, float64(0))
	}
	_topRowFull = _pigeon.And(_pigeon.Neq(_pigeon.Get(_topRow, float64(0)), "_"), _pigeon.Neq(_pigeon.Get(_topRow, float64(1)), "_"), _pigeon.Neq(_pigeon.Get(_topRow, float64(2)), "_"))
	_middleRowFull = _pigeon.And(_pigeon.Neq(_pigeon.Get(_middleRow, float64(0)), "_"), _pigeon.Neq(_pigeon.Get(_middleRow, float64(1)), "_"), _pigeon.Neq(_pigeon.Get(_middleRow, float64(2)), "_"))
	_bottomRowFull = _pigeon.And(_pigeon.Neq(_pigeon.Get(_bottomRow, float64(0)), "_"), _pigeon.Neq(_pigeon.Get(_bottomRow, float64(1)), "_"), _pigeon.Neq(_pigeon.Get(_bottomRow, float64(2)), "_"))
	if _pigeon.And(_topRowFull, _middleRowFull, _bottomRowFull).(bool) {
		return "tie"
	}
	return "_"
}
func _main() interface{} {
	var _w, _done, _currentPlayer interface{}
	_currentPlayer = "X"
	_done = false
	for _pigeon.Not(_done).(bool) {
		_pigeon.Print(_pigeon.Concat(_topRow, "\n", _middleRow, "\n", _bottomRow, "\n"))
		_w = _winner()
		if _pigeon.Eq(_w, "X").(bool) {
			_pigeon.Print("X's win!")
			_done = true
		} else if _pigeon.Eq(_w, "O").(bool) {
			_pigeon.Print("O's win!")
		} else if _pigeon.Eq(_w, "tie").(bool) {
			_pigeon.Print("Tie!")
			_done = true
		} else {
			_playerMove(_currentPlayer)
			if _pigeon.Eq(_currentPlayer, "X").(bool) {
				_currentPlayer = "O"
			} else {
				_currentPlayer = "X"
			}
		}
	}
	return nil
}

func main() {
	_main()
}
