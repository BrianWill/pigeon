package std

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type List []interface{}

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

func Prompt(args ...interface{}) string {
	if len(args) >= 1 {
		fmt.Println(args...)
	}
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	if scanner.Err() != nil {
		log.Fatalln(scanner.Err())
	}
	s := scanner.Text()
	return s
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

func Floor(i float64) float64 {
	return math.Floor(i)
}

func Ceil(i float64) float64 {
	return math.Ceil(i)
}

func NoOp(discardMe ...interface{}) {
	// nada
}

func StrLen(s string) int64 {
	return int64(utf8.RuneCountInString(s))
}

func CreateFile(name string) (int64, string) {
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return 0, err.Error()
	}
	id := int64(f.Fd())
	openFilesById[id] = f
	return id, ""
}

func OpenFile(name string) (int64, string) {
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return 0, err.Error()
	}
	id := int64(f.Fd())
	openFilesById[id] = f
	return id, ""
}

var openFilesById = map[int64]*os.File{}

func CloseFile(id int64) string {
	if f, ok := openFilesById[id]; ok {
		if err := f.Close(); err != nil {
			log.Fatal("Error closing file.")
		}
		delete(openFilesById, id)
		return ""
	} else {
		return "Error closing file: no open file has id '" + strconv.FormatInt(id, 10) + "'"
	}
}

func ReadFile(id int64, bytes []byte) (int64, string) {
	if f, ok := openFilesById[id]; ok {
		n, err := f.Read(bytes)
		if err != nil {
			return int64(n), err.Error()
		} else {
			return int64(n), ""
		}
	} else {
		return 0, "Error reading file: no open file has id '" + strconv.FormatInt(id, 10) + "'"
	}
}

func WriteFile(id int64, bytes []byte) (int64, string) {
	if f, ok := openFilesById[id]; ok {
		n, err := f.Write(bytes)
		if err != nil {
			return int64(n), err.Error()
		} else {
			return int64(n), ""
		}
	} else {
		return 0, "Error writing file: no open file has id '" + strconv.FormatInt(id, 10) + "'"
	}
}

func SeekFile(id int64, offset int64) (int64, string) {
	if f, ok := openFilesById[id]; ok {
		n, err := f.Seek(offset, 1)
		if err != nil {
			return int64(n), err.Error()
		} else {
			return int64(n), ""
		}
	} else {
		return 0, "Error seeking file: no open file has id '" + strconv.FormatInt(id, 10) + "'"
	}
}

func SeekFileStart(id int64, offset int64) (int64, string) {
	if f, ok := openFilesById[id]; ok {
		n, err := f.Seek(offset, 0)
		if err != nil {
			return int64(n), err.Error()
		} else {
			return int64(n), ""
		}
	} else {
		return 0, "Error seeking file: no open file has id '" + strconv.FormatInt(id, 10) + "'"
	}
}

func SeekFileEnd(id int64, offset int64) (int64, string) {
	if f, ok := openFilesById[id]; ok {
		n, err := f.Seek(offset, 2)
		if err != nil {
			return int64(n), err.Error()
		} else {
			return int64(n), ""
		}
	} else {
		return 0, "Error seeking file: no open file has id '" + strconv.FormatInt(id, 10) + "'"
	}
}
