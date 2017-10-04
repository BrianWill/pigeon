package std

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

type List []interface{}

func NewList(items ...interface{}) *List {
	var l List = List(items)
	return &l
}

var Breakpoints = make(map[int]bool)

func (l *List) append(item interface{}) {
	*l = append(*l, item)
}

func (l *List) set(idx float64, item interface{}) {
	(*l)[int64(idx)] = item
}

func (l *List) len() float64 {
	return float64(len(*l))
}

func Prompt(args ...interface{}) {
	if len(args) > 1 {
		fmt.Print(args...)
	}
}

func init() {
	t := time.Now()
	rand.Seed(t.UnixNano())
}

var RandFloat = rand.Float64

var RandInt = rand.Int63

var RandIntN = rand.Int63n

func ParseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func ParseInt(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func FormatFloat(f float64) string {
	return strconv.FormatFloat(f, 'E', -1, 64)
}

func FormatInt(i int64) string {
	return strconv.FormatInt(i, 10)
}

func TimeNow() int64 {
	return time.Now().Unix()
}

func FormatTime(i int64) string {
	return time.Unix(i, 0).Format(time.UnixDate)
}

func ParseTime(s string) (int64, error) {
	t, err := time.Parse(time.UnixDate, s)
	if err != nil {
		return t.Unix(), nil
	}
	return 0, err
}
