// Package syntaxhighlight provides syntax highlighting for code. It currently
// uses a language-independent lexer and performs decently on JavaScript, Java,
// Ruby, Python, Go, and C.
package highlight

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/scanner"
	"text/template"
	"unicode"
	"unicode/utf8"

	"github.com/sourcegraph/annotate"
)

var keywords = map[string]bool{
	"global": true,
	"func":   true,
	"if":     true,
	"elseif": true,
	"else":   true,
	"while":  true,
	"as":     true,
	"asadd":  true,
	"asinc":  true,
	"asdec":  true,
	"assub":  true,
	"asmul":  true,
	"asdiv":  true,
	"locals": true,
	"return": true,
}

var operators = map[string]bool{
	"add":    true,
	"sub":    true,
	"mul":    true,
	"div":    true,
	"mod":    true,
	"eq":     true,
	"neq":    true,
	"id":     true,
	"not":    true,
	"lt":     true,
	"gt":     true,
	"lte":    true,
	"gte":    true,
	"get":    true,
	"set":    true,
	"append": true,
	"or":     true,
	"and":    true,
	"print":  true,
	"prompt": true,
	"concat": true,
	"list":   true,
	"map":    true,
	"len":    true,
}

var valueWords = map[string]bool{
	"true":  true,
	"false": true,
	"nil":   true,
}

// var keywords = map[string]struct{}{
// 	"BEGIN":            {},
// 	"END":              {},
// 	"False":            {},
// 	"Infinity":         {},
// 	"NaN":              {},
// 	"None":             {},
// 	"True":             {},
// 	"abstract":         {},
// 	"alias":            {},
// 	"align_union":      {},
// 	"alignof":          {},
// 	"and":              {},
// 	"append":           {},
// 	"as":               {},
// 	"asm":              {},
// 	"assert":           {},
// 	"auto":             {},
// 	"axiom":            {},
// 	"begin":            {},
// 	"bool":             {},
// 	"boolean":          {},
// 	"break":            {},
// 	"byte":             {},
// 	"caller":           {},
// 	"case":             {},
// 	"catch":            {},
// 	"char":             {},
// 	"class":            {},
// 	"concept":          {},
// 	"concept_map":      {},
// 	"const":            {},
// 	"const_cast":       {},
// 	"constexpr":        {},
// 	"continue":         {},
// 	"debugger":         {},
// 	"decltype":         {},
// 	"def":              {},
// 	"default":          {},
// 	"defined":          {},
// 	"del":              {},
// 	"delegate":         {},
// 	"delete":           {},
// 	"die":              {},
// 	"do":               {},
// 	"double":           {},
// 	"dump":             {},
// 	"dynamic_cast":     {},
// 	"elif":             {},
// 	"else":             {},
// 	"elsif":            {},
// 	"end":              {},
// 	"ensure":           {},
// 	"enum":             {},
// 	"eval":             {},
// 	"except":           {},
// 	"exec":             {},
// 	"exit":             {},
// 	"explicit":         {},
// 	"export":           {},
// 	"extends":          {},
// 	"extern":           {},
// 	"false":            {},
// 	"final":            {},
// 	"finally":          {},
// 	"float":            {},
// 	"float32":          {},
// 	"float64":          {},
// 	"for":              {},
// 	"foreach":          {},
// 	"friend":           {},
// 	"from":             {},
// 	"func":             {},
// 	"function":         {},
// 	"generic":          {},
// 	"get":              {},
// 	"global":           {},
// 	"goto":             {},
// 	"if":               {},
// 	"implements":       {},
// 	"import":           {},
// 	"in":               {},
// 	"inline":           {},
// 	"instanceof":       {},
// 	"int":              {},
// 	"int8":             {},
// 	"int16":            {},
// 	"int32":            {},
// 	"int64":            {},
// 	"interface":        {},
// 	"is":               {},
// 	"lambda":           {},
// 	"last":             {},
// 	"late_check":       {},
// 	"local":            {},
// 	"long":             {},
// 	"make":             {},
// 	"map":              {},
// 	"module":           {},
// 	"mutable":          {},
// 	"my":               {},
// 	"namespace":        {},
// 	"native":           {},
// 	"new":              {},
// 	"next":             {},
// 	"nil":              {},
// 	"no":               {},
// 	"nonlocal":         {},
// 	"not":              {},
// 	"null":             {},
// 	"nullptr":          {},
// 	"operator":         {},
// 	"or":               {},
// 	"our":              {},
// 	"package":          {},
// 	"pass":             {},
// 	"print":            {},
// 	"private":          {},
// 	"property":         {},
// 	"protected":        {},
// 	"public":           {},
// 	"raise":            {},
// 	"redo":             {},
// 	"register":         {},
// 	"reinterpret_cast": {},
// 	"require":          {},
// 	"rescue":           {},
// 	"retry":            {},
// 	"return":           {},
// 	"self":             {},
// 	"set":              {},
// 	"short":            {},
// 	"signed":           {},
// 	"sizeof":           {},
// 	"static":           {},
// 	"static_assert":    {},
// 	"static_cast":      {},
// 	"strictfp":         {},
// 	"struct":           {},
// 	"sub":              {},
// 	"super":            {},
// 	"switch":           {},
// 	"synchronized":     {},
// 	"template":         {},
// 	"then":             {},
// 	"this":             {},
// 	"throw":            {},
// 	"throws":           {},
// 	"transient":        {},
// 	"true":             {},
// 	"try":              {},
// 	"type":             {},
// 	"typedef":          {},
// 	"typeid":           {},
// 	"typename":         {},
// 	"typeof":           {},
// 	"undef":            {},
// 	"undefined":        {},
// 	"union":            {},
// 	"unless":           {},
// 	"unsigned":         {},
// 	"until":            {},
// 	"use":              {},
// 	"using":            {},
// 	"var":              {},
// 	"virtual":          {},
// 	"void":             {},
// 	"volatile":         {},
// 	"wantarray":        {},
// 	"when":             {},
// 	"where":            {},
// 	"while":            {},
// 	"with":             {},
// 	"yield":            {},
// }

const _Kind_name = "WhitespaceStringKeywordCommentTypeLiteralPunctuationPlaintextTagHTMLTagHTMLAttrNameHTMLAttrValueDecimal"

var _Kind_index = [...]uint8{0, 10, 16, 23, 30, 34, 41, 52, 61, 64, 71, 83, 96, 103}

func (i Kind) GoString() string {
	if i+1 >= Kind(len(_Kind_index)) {
		return fmt.Sprintf("syntaxhighlight.Kind(%d)", i)
	}
	return "syntaxhighlight." + _Kind_name[_Kind_index[i]:_Kind_index[i+1]]
}

// Kind represents a syntax highlighting kind (class) which will be assigned to tokens.
// A syntax highlighting scheme (style) maps text style properties to each token kind.
type Kind uint8

// A set of supported highlighting kinds
const (
	Whitespace Kind = iota
	String
	Keyword
	Comment
	Operator
	ValueWord
	Type
	Literal
	Punctuation
	Plaintext
	Tag
	HTMLTag
	HTMLAttrName
	HTMLAttrValue
	Decimal
)

//go:generate gostringer -type=Kind

// Printer implements an interface to render highlighted output
// (see HTMLPrinter for the implementation of this interface)
type Printer interface {
	Print(w io.Writer, kind Kind, tokText string) error
}

// HTMLConfig holds the HTML class configuration to be used by annotators when
// highlighting code.
type HTMLConfig struct {
	String        string
	Keyword       string
	Comment       string
	Type          string
	Literal       string
	Punctuation   string
	Plaintext     string
	Tag           string
	HTMLTag       string
	HTMLAttrName  string
	HTMLAttrValue string
	Decimal       string
	Whitespace    string
	Operator      string
	ValueWord     string

	AsOrderedList bool
}

// HTMLPrinter implements Printer interface and is used to produce
// HTML-based highligher
type HTMLPrinter HTMLConfig

// Class returns the set class for a given token Kind.
func (c HTMLConfig) Class(kind Kind) string {
	switch kind {
	case String:
		return c.String
	case Keyword:
		return c.Keyword
	case Operator:
		return c.Operator
	case ValueWord:
		return c.ValueWord
	case Comment:
		return c.Comment
	case Type:
		return c.Type
	case Literal:
		return c.Literal
	case Punctuation:
		return c.Punctuation
	case Plaintext:
		return c.Plaintext
	case Tag:
		return c.Tag
	case HTMLTag:
		return c.HTMLTag
	case HTMLAttrName:
		return c.HTMLAttrName
	case HTMLAttrValue:
		return c.HTMLAttrValue
	case Decimal:
		return c.Decimal
	}
	return ""
}

var lineNum = 2

// Print is the function that emits highlighted source code using
// <span class="...">...</span> wrapper tags
func (p HTMLPrinter) Print(w io.Writer, kind Kind, tokText string) error {
	if p.AsOrderedList {
		if i := strings.Index(tokText, "\n"); i > -1 {
			if err := p.Print(w, kind, tokText[:i]); err != nil {
				return err
			}
			w.Write([]byte("</li>\n<li lineNum=\"" + strconv.Itoa(lineNum) + "\">"))
			lineNum++
			if err := p.Print(w, kind, tokText[i+1:]); err != nil {
				return err
			}
			return nil
		}
	}

	class := ((HTMLConfig)(p)).Class(kind)
	if class != "" {
		_, err := w.Write([]byte(`<span class="`))
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, class)
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(`">`))
		if err != nil {
			return err
		}
	}
	template.HTMLEscape(w, []byte(tokText))
	if class != "" {
		_, err := w.Write([]byte(`</span>`))
		if err != nil {
			return err
		}
	}
	return nil
}

type Annotator interface {
	Annotate(start int, kind Kind, tokText string) (*annotate.Annotation, error)
}

type HTMLAnnotator HTMLConfig

func (a HTMLAnnotator) Annotate(start int, kind Kind, tokText string) (*annotate.Annotation, error) {
	class := ((HTMLConfig)(a)).Class(kind)
	if class != "" {
		left := []byte(`<span class="`)
		left = append(left, []byte(class)...)
		left = append(left, []byte(`">`)...)
		return &annotate.Annotation{
			Start: start, End: start + len(tokText),
			Left: left, Right: []byte("</span>"),
		}, nil
	}
	return nil, nil
}

// Option is a type of the function that can modify
// one or more of the options in the HTMLConfig structure.
type Option func(options *HTMLConfig)

// OrderedList allows you to format the output as an ordered list
// to have line numbers in the output.
//
// Example:
// AsHTML(input, OrderedList())
func OrderedList() Option {
	return func(o *HTMLConfig) {
		o.AsOrderedList = true
	}
}

// DefaultHTMLConfig provides class names that match those of google-code-prettify
// (https://code.google.com/p/google-code-prettify/).
var DefaultHTMLConfig = HTMLConfig{
	String:        "str",
	Keyword:       "kwd",
	Comment:       "com",
	Type:          "typ",
	Literal:       "lit",
	Punctuation:   "pun",
	Plaintext:     "pln",
	Tag:           "tag",
	HTMLTag:       "htm",
	HTMLAttrName:  "atn",
	HTMLAttrValue: "atv",
	Decimal:       "dec",
	Operator:      "op",
	ValueWord:     "val",
	Whitespace:    "",
}

func Print(s *scanner.Scanner, w io.Writer, p Printer) error {
	tok := s.Scan()
	for tok != scanner.EOF {
		tokText := s.TokenText()
		err := p.Print(w, tokenKind(tok, tokText), tokText)
		if err != nil {
			return err
		}

		tok = s.Scan()
	}

	return nil
}

func Annotate(src []byte, a Annotator) (annotate.Annotations, error) {
	s := NewScanner(src)

	var anns annotate.Annotations
	read := 0

	tok := s.Scan()
	for tok != scanner.EOF {
		tokText := s.TokenText()

		ann, err := a.Annotate(read, tokenKind(tok, tokText), tokText)
		if err != nil {
			return nil, err
		}
		read += len(tokText)
		if ann != nil {
			anns = append(anns, ann)
		}

		tok = s.Scan()
	}

	return anns, nil
}

// AsHTML converts source code into an HTML-highlighted version;
// It accepts optional configuration parameters to control rendering
// (see OrderedList as one example)
func AsHTML(src []byte, options ...Option) ([]byte, error) {
	opt := DefaultHTMLConfig
	for _, f := range options {
		f(&opt)
	}

	var buf bytes.Buffer
	if opt.AsOrderedList {
		lineNum = 2 // reset global
		buf.Write([]byte("<ol>\n<li lineNum=\"1\">"))
	}
	err := Print(NewScanner(src), &buf, HTMLPrinter(opt))
	if opt.AsOrderedList {
		buf.Write([]byte("</li>\n</ol>"))
	}
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// NewScanner is a helper that takes a []byte src, wraps it in a reader and creates a Scanner.
func NewScanner(src []byte) *scanner.Scanner {
	return NewScannerReader(bytes.NewReader(src))
}

// NewScannerReader takes a reader src and creates a Scanner.
func NewScannerReader(src io.Reader) *scanner.Scanner {
	var s scanner.Scanner
	s.Init(src)
	s.Error = func(_ *scanner.Scanner, _ string) {}
	s.Whitespace = 0
	s.Mode = s.Mode ^ scanner.SkipComments
	return &s
}

func tokenKind(tok rune, tokText string) Kind {
	switch tok {
	case scanner.Ident:
		if _, ok := keywords[tokText]; ok {
			return Keyword
		}
		if _, ok := operators[tokText]; ok {
			return Operator
		}
		if _, ok := valueWords[tokText]; ok {
			return ValueWord
		}
		if r, _ := utf8.DecodeRuneInString(tokText); unicode.IsUpper(r) {
			return Type
		}
		return Plaintext
	case scanner.Float, scanner.Int:
		return Decimal
	case scanner.Char, scanner.String, scanner.RawString:
		return String
	case scanner.Comment:
		return Comment
	}
	if unicode.IsSpace(tok) {
		return Whitespace
	}
	return Punctuation
}
