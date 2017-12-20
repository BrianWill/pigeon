package std

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type List []interface{}

func NewList(items ...interface{}) *List {
	l := List(items)
	return &l
}

func (s *List) String() string {
	strs := make([]string, len(*s))
	for i, v := range *s {
		strs[i] = fmt.Sprint(v)
	}
	return "[" + strings.Join(strs, ", ") + "]"
}

var Breakpoints = make(map[int]bool)

func (l *List) Append(item interface{}) {
	*l = append(*l, item)
}

func (l *List) Set(idx int64, item interface{}) {
	(*l)[idx] = item
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

func Charlist(s string) *List {
	l := List(make([]interface{}, len(s)))
	for i, v := range s {
		l[i] = string(v)
	}
	return &l
}

func Runelist(s string) *List {
	l := List(make([]interface{}, len(s)))
	for i, v := range s {
		l[i] = int64(v)
	}
	return &l
}

func Charslice(s string) []string {
	strs := make([]string, len(s))
	for i, v := range s {
		strs[i] = string(v)
	}
	return strs
}

func Runeslice(s string) []int64 {
	strs := make([]int64, len(s))
	for i, v := range s {
		strs[i] = int64(v)
	}
	return strs
}

func Runelist2string(r *List) string {
	runes := make([]rune, len(*r))
	for i, v := range *r {
		runes[i] = rune(v.(int64))
	}
	return string(runes)
}

func Charlist2string(s *List) string {
	strs := make([]string, len(*s))
	for i, v := range *s {
		strs[i] = v.(string)
	}
	return strings.Join(strs, "")
}

func Runeslice2string(r []int64) string {
	runes := make([]rune, len(r))
	for i, v := range r {
		runes[i] = rune(v)
	}
	return string(runes)
}

func Charslice2string(s []string) string {
	return strings.Join(s, "")
}

func NoOp(discardMe ...interface{}) {
	// nada
}
