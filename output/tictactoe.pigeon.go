package main

import _p "github.com/BrianWill/pigeon/stdlib"

var _breakpoints = make(map[int]bool)

var g_topRow interface{} = _p.List("_", "_", "_")
var g_middleRow interface{} = _p.List("_", "_", "_")
var g_bottomRow interface{} = _p.List("_", "_", "_")

func playerMove(currentPlayer interface{}) interface{} {
	var move, row, col, slot interface{}
	_p.NullOp(move, row, col, slot)
	debug := func(line int) {
		var globals = map[string]interface{}{
			"topRow":    g_topRow,
			"middleRow": g_middleRow,
			"bottomRow": g_bottomRow,
		}
		var locals = map[string]interface{}{
			"currentPlayer": currentPlayer,
			"move":          move,
			"row":           row,
			"col":           col,
			"slot":          slot,
		}
		_p.PollContinue(line, globals, locals)
	}
	if _breakpoints[7] {
		debug(7)
	}
	move = _p.Nil(0)
	if _breakpoints[8] {
		debug(8)
	}
	for _p.Eq(move, _p.Nil(0)).(bool) {
		if _breakpoints[9] {
			debug(9)
		}
		row = _p.Nil(0)
		if _breakpoints[10] {
			debug(10)
		}
		for _p.Eq(row, _p.Nil(0)).(bool) {
			if _breakpoints[11] {
				debug(11)
			}
			row = _p.Prompt("Select [t]op, [m]iddle, or [b]ottom row, player", currentPlayer)
			if _breakpoints[12] {
				debug(12)
			}
			if _p.Eq(row, "t").(bool) {
				if _breakpoints[13] {
					debug(13)
				}
				row = g_topRow
			} else if _p.Eq(row, "m").(bool) {
				if _breakpoints[15] {
					debug(15)
				}
				row = g_middleRow
			} else if _p.Eq(row, "b").(bool) {
				if _breakpoints[17] {
					debug(17)
				}
				row = g_bottomRow
			} else {
				if _breakpoints[19] {
					debug(19)
				}
				_p.Print("Invalid input. Try again.")
				if _breakpoints[20] {
					debug(20)
				}
				row = _p.Nil(0)
			}
		}
		if _breakpoints[21] {
			debug(21)
		}
		col = _p.Nil(0)
		if _breakpoints[22] {
			debug(22)
		}
		for _p.Eq(col, _p.Nil(0)).(bool) {
			if _breakpoints[23] {
				debug(23)
			}
			col = _p.Prompt("Select [l]eft, [m]iddle, or [r]ight column, player", currentPlayer)
			if _breakpoints[24] {
				debug(24)
			}
			if _p.Eq(col, "l").(bool) {
				if _breakpoints[25] {
					debug(25)
				}
				col = float64(0)
			} else if _p.Eq(col, "m").(bool) {
				if _breakpoints[27] {
					debug(27)
				}
				col = float64(1)
			} else if _p.Eq(col, "r").(bool) {
				if _breakpoints[29] {
					debug(29)
				}
				col = float64(2)
			} else {
				if _breakpoints[31] {
					debug(31)
				}
				_p.Print("Invalid input. Try again.")
				if _breakpoints[32] {
					debug(32)
				}
				col = _p.Nil(0)
			}
		}
		if _breakpoints[33] {
			debug(33)
		}
		slot = _p.Get(row, col)
		if _breakpoints[34] {
			debug(34)
		}
		if _p.Eq(slot, "_").(bool) {
			if _breakpoints[35] {
				debug(35)
			}
			_p.Set(row, col, currentPlayer)
			if _breakpoints[36] {
				debug(36)
			}
			move = true
		} else {
			if _breakpoints[38] {
				debug(38)
			}
			_p.Print("That slot is occupied! Try again.")
		}
	}
	return nil
}
func winner() interface{} {
	var topRowFull, middleRowFull, bottomRowFull interface{}
	_p.NullOp(topRowFull, middleRowFull, bottomRowFull)
	debug := func(line int) {
		var globals = map[string]interface{}{
			"topRow":    g_topRow,
			"middleRow": g_middleRow,
			"bottomRow": g_bottomRow,
		}
		var locals = map[string]interface{}{
			"topRowFull":    topRowFull,
			"middleRowFull": middleRowFull,
			"bottomRowFull": bottomRowFull,
		}
		_p.PollContinue(line, globals, locals)
	}
	if _breakpoints[45] {
		debug(45)
	}
	if _p.And(_p.Neq(_p.Get(g_topRow, float64(0)), "_"), _p.Eq(_p.Get(g_topRow, float64(0)), _p.Get(g_topRow, float64(1)), _p.Get(g_topRow, float64(2)))).(bool) {
		if _breakpoints[46] {
			debug(46)
		}
		return _p.Get(g_topRow, float64(0))
	}
	if _breakpoints[48] {
		debug(48)
	}
	if _p.And(_p.Neq(_p.Get(g_middleRow, float64(0)), "_"), _p.Eq(_p.Get(g_middleRow, float64(0)), _p.Get(g_middleRow, float64(1)), _p.Get(g_middleRow, float64(2)))).(bool) {
		if _breakpoints[49] {
			debug(49)
		}
		return _p.Get(g_middleRow, float64(0))
	}
	if _breakpoints[51] {
		debug(51)
	}
	if _p.And(_p.Neq(_p.Get(g_bottomRow, float64(0)), "_"), _p.Eq(_p.Get(g_bottomRow, float64(0)), _p.Get(g_bottomRow, float64(1)), _p.Get(g_bottomRow, float64(2)))).(bool) {
		if _breakpoints[52] {
			debug(52)
		}
		return _p.Get(g_bottomRow, float64(0))
	}
	if _breakpoints[54] {
		debug(54)
	}
	if _p.And(_p.Neq(_p.Get(g_topRow, float64(0)), "_"), _p.Eq(_p.Get(g_topRow, float64(0)), _p.Get(g_middleRow, float64(0)), _p.Get(g_bottomRow, float64(0)))).(bool) {
		if _breakpoints[55] {
			debug(55)
		}
		return _p.Get(g_topRow, float64(0))
	}
	if _breakpoints[57] {
		debug(57)
	}
	if _p.And(_p.Neq(_p.Get(g_topRow, float64(1)), "_"), _p.Eq(_p.Get(g_topRow, float64(1)), _p.Get(g_middleRow, float64(1)), _p.Get(g_bottomRow, float64(1)))).(bool) {
		if _breakpoints[58] {
			debug(58)
		}
		return _p.Get(g_topRow, float64(1))
	}
	if _breakpoints[60] {
		debug(60)
	}
	if _p.And(_p.Neq(_p.Get(g_topRow, float64(2)), "_"), _p.Eq(_p.Get(g_topRow, float64(2)), _p.Get(g_middleRow, float64(2)), _p.Get(g_bottomRow, float64(2)))).(bool) {
		if _breakpoints[61] {
			debug(61)
		}
		return _p.Get(g_topRow, float64(2))
	}
	if _breakpoints[63] {
		debug(63)
	}
	if _p.And(_p.Neq(_p.Get(g_topRow, float64(0)), "_"), _p.Eq(_p.Get(g_topRow, float64(0)), _p.Get(g_middleRow, float64(1)), _p.Get(g_bottomRow, float64(2)))).(bool) {
		if _breakpoints[64] {
			debug(64)
		}
		return _p.Get(g_topRow, float64(0))
	}
	if _breakpoints[66] {
		debug(66)
	}
	if _p.And(_p.Neq(_p.Get(g_bottomRow, float64(0)), "_"), _p.Eq(_p.Get(g_bottomRow, float64(0)), _p.Get(g_middleRow, float64(1)), _p.Get(g_topRow, float64(2)))).(bool) {
		if _breakpoints[67] {
			debug(67)
		}
		return _p.Get(g_bottomRow, float64(0))
	}
	if _breakpoints[69] {
		debug(69)
	}
	topRowFull = _p.And(_p.Neq(_p.Get(g_topRow, float64(0)), "_"), _p.Neq(_p.Get(g_topRow, float64(1)), "_"), _p.Neq(_p.Get(g_topRow, float64(2)), "_"))
	if _breakpoints[70] {
		debug(70)
	}
	middleRowFull = _p.And(_p.Neq(_p.Get(g_middleRow, float64(0)), "_"), _p.Neq(_p.Get(g_middleRow, float64(1)), "_"), _p.Neq(_p.Get(g_middleRow, float64(2)), "_"))
	if _breakpoints[71] {
		debug(71)
	}
	bottomRowFull = _p.And(_p.Neq(_p.Get(g_bottomRow, float64(0)), "_"), _p.Neq(_p.Get(g_bottomRow, float64(1)), "_"), _p.Neq(_p.Get(g_bottomRow, float64(2)), "_"))
	if _breakpoints[72] {
		debug(72)
	}
	if _p.And(topRowFull, middleRowFull, bottomRowFull).(bool) {
		if _breakpoints[73] {
			debug(73)
		}
		return "tie"
	}
	if _breakpoints[74] {
		debug(74)
	}
	return "_"
}
func _main() interface{} {
	var w, done, currentPlayer interface{}
	_p.NullOp(w, done, currentPlayer)
	debug := func(line int) {
		var globals = map[string]interface{}{
			"topRow":    g_topRow,
			"middleRow": g_middleRow,
			"bottomRow": g_bottomRow,
		}
		var locals = map[string]interface{}{
			"currentPlayer": currentPlayer,
			"w":             w,
			"done":          done,
		}
		_p.PollContinue(line, globals, locals)
	}
	if _breakpoints[78] {
		debug(78)
	}
	currentPlayer = "X"
	if _breakpoints[79] {
		debug(79)
	}
	done = false
	if _breakpoints[80] {
		debug(80)
	}
	for _p.Not(done).(bool) {
		if _breakpoints[81] {
			debug(81)
		}
		_p.Print(_p.Concat(g_topRow, "\n", g_middleRow, "\n", g_bottomRow, "\n"))
		if _breakpoints[82] {
			debug(82)
		}
		w = winner()
		if _breakpoints[83] {
			debug(83)
		}
		if _p.Eq(w, "X").(bool) {
			if _breakpoints[84] {
				debug(84)
			}
			_p.Print("X's win!")
			if _breakpoints[85] {
				debug(85)
			}
			done = true
		} else if _p.Eq(w, "O").(bool) {
			if _breakpoints[87] {
				debug(87)
			}
			_p.Print("O's win!")
		} else if _p.Eq(w, "tie").(bool) {
			if _breakpoints[89] {
				debug(89)
			}
			_p.Print("Tie!")
			if _breakpoints[90] {
				debug(90)
			}
			done = true
		} else {
			if _breakpoints[92] {
				debug(92)
			}
			playerMove(currentPlayer)
			if _breakpoints[94] {
				debug(94)
			}
			if _p.Eq(currentPlayer, "X").(bool) {
				if _breakpoints[95] {
					debug(95)
				}
				currentPlayer = "O"
			} else {
				if _breakpoints[97] {
					debug(97)
				}
				currentPlayer = "X"
			}
		}
	}
	return nil
}

func main() {
	go _p.PollBreakpoints(&_breakpoints)
	_main()
}
