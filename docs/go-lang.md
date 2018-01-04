## Learning Go after GoPigeon

Once you've learned GoPigeon, learning Go is primarily a matter of learning the different syntax and the essential parts of the Go standard library (which is much, much larger than what GoPigeon offers). Beyond syntax, there are a few few additional features (nested functions, closures, go routines, channels, const).

## infix operators, function calls, and assignment

The operators in Go are symbols rather than words:

```
+       addition
-       subtraction, negation
*       multiplication
/       division
==      equality
!=      not equality
!       not
<       less than
>       greater than
<=      less than or equal
>=      greater than or equal
&&      logical and
||      logical or
&       bitwise and
|       bitwise or
^       bitwise not
```

Most of these are binary infix operators, meaning they are always written in between their two operands. A few, like ! and ^, are unary operators written before their single operand:


```go
(((-x) + y) * z)             // (mul (add (neg x) y) z)
((!true) && false)           // (and (not true) false)
```

The operators have an *order of precedence*, such that parentheses are only needed when we wish to subvert the order of precedence. For example, `*` has higher precedence than `+`, so these are equivalent:

```go
((x * y) + z)                // (add (mul x y) z)
x * y + z                    // (add (mul x y) z)
```

...but here the parentheses cause the addition to be done first:

```go
x * (y + z)                  // (mul x (add y z))
```

Unlike in Pigeon, we can surround any expression in extra parentheses. These Go expressions are all equivalent:

```go
x
(x)
((x))
(((x)))
```

To call a function, we put parens *after* the function name and put the arguments inside the parens, separated by commas:

```go
foo(a, b, c)                 // (foo a b c)
x + foo(a, b, c)             // (add x (foo a b c))
```

An assignment statement is denoted by `=`, with the target of assignment on the left and the value on the right:

```go
x = 5                        // as x 5
```

As a shorthand, some operators can be combined with `=`:

```
x += 3            // x = x + 3
x /= 3            // x = x / 3
x -= 3            // x = x - 3
x *= 3            // x = x * 3
```

## basic data types

```
int                 32- or 64-bit signed integer
int8                8-bit signed integer
int16               16-bit signed integer
int32               32-bit signed integer
int64               64-bit signed integer

uint                32- or 64-bit unsigned integer
uint8               8-bit unsigned integer
uint16              16-bit unsigned integer
uint32              32-bit unsigned integer
uint64              64-bit unsigned integer
byte                alias for uint8

float32             32-bit floating-point
float64             64-bit floating-point

string              UTF-8 enoded string
bool                boolean
```

Whether `int` and `uint` are 32- or 64-bit depends upon the platform we compile for.

## function definitions

Functions are written with the parameters and return types in parens, separated by commas, followed by the body inside curly braces:

```go
// foo takes an int and a string and returns a byte and a float32
func foo(a int, b string) (byte, float32) {
    // ... body goes here
}
```

If a function has just one return type, we can omit the parens around the return type.

The `main` function has no parameters or return type:

```go
func main() {
    // ...
}
```

Note the parens for the parameters are never omited, even for a function with no parameters.

## increment and decrement statements

Because adding `1` or subtracting `1` from an integer variable is so common, Go allows shorthand statements with `++` (increment) and `--` (decrement):

```go
i := 4
i++       // i = i + 1
i--       // i = i - 1
```

## `if` and `for` syntax

The bodies of `if`'s, `for`'s (Go's equivalent of `while`), and other such constructs are surrounded in curly braces:

```go
if x < 3 {
    // body
} else if y == 2 {            // note that we write 'else if' instead of 'elif'
    // body
}

for i > 0 {
    // body
}
```

Line indentation is not siginificant in Go, but standard style is to indent code as we do in Pigeon.

As a convenience, `for` loops can be written in this form:

```
for precondition; condition; postcondition {
    body
}
```

The pre-condition is a declaration and assignment using the `:=` syntax. It is executed once, at the start of the loop before the condition is first tested.

The post-condition is an assignment or increment/decrement operation. It is executed every time after the body is executed but before the condition is tested again. (If the condition is false the first time, the post-condition is skipped over entirely like the rest of the body.)

This variant of `for` is especially handy for looping through a range of integers:

```go
// this loop calls 'foo' with the values: 0 1 2 3 4
for i := 0; i < 5; i++ {
    foo(i)
}
```

With a normal `for`, the same thing would be written:

```go
i := 0
for i < 5 {
    foo(i)
    i++   
}
```

The above code is exactly the same except for one subtle difference: when the variable is declared in the pre-condition, it belongs to the scope of the for body:

```go
for i := 0; i < 5; i++ {
    foo(i)
}
bar(i)        // compile error: 'i' does not exist in this scope
```

The equivalent of Pigeon's `foreach` is written with `:= range`:

```go
foo := []int{8, 4, 7}
for i, v := range foo {
    // ... i is the index, v is the value
}
```

(Notice we don't specify the types of the two variables because it is inferred from the type of 'foo'. GoPigeon could give us this same convenience but choses not to for the sake of explicitness.)

## variable declarations

A `var` statement creates a local variable. They can be put anywhere in a function, not just at the top, but a variable is only considered to exist after its `var` statement:

```go
foo(x)                 // compile error: x does not exist here
var x int              // x exists starting here
x = 3
```

A declared variable starts out with the 'zero value' (default value) for its type. For number types, this is 0.

For concision, we can assign to a variable in its `var` statement:

```go
var x int = 3          // create int varible x with initial value 3
```

...is equivalent to:

```go
var x int
x = 3
```

If we initialize the variable in a `var` this way, we can leave the type to be inferred from the value assigned:

```go
var foo = "hi"         // var foo string = "hi"
var bar = true         // var bar bool = true
var ack = foo()        // ack will have the type of whatever foo returns
var x = 5              // var x int = 5
var y = 5.2            // var x float64 = 5.2
```

(Note that an integer is assumed to be an `int` and a floating-point value is assume to be a `float64`.)

As further convenience, we can write the above with `:=` instead of `var`:

```go
foo := "hi"         // var foo string = "hi"
bar := true         // var bar bool = true
ack := foo()        // ack will have the type of whatever foo returns
x := 5              // var x int = 5
y := 5.2            // var x float64 = 5.2
```

If a variable `var` or `:=` statement is inside the body of some construct like an `if` or loop, the variable it creates only exists within that body:

```go
if x < 3 {
    var y int   // this y variable only exists within the if
    // ...
}
```

Inside a scope (a body), we can create a variable with the same name as a variable from the outer scope. In the subscope, that name will refer to the inner variable, not the outer one:

```go
var x int
// ...
if x < 3 {
    var x int   // this 'x' variable is different from the 'x' of the outer scope
    x = 6       // assign to 'x' of the inner scope
    foo(x)      // pass 'x' of the inner scope
    // ...
}
```

If these name conflicts are ever a problem, simply rename one or both of the conflicting variables.

The authors of Go decided that unused local variables are generally unintentional, and so the compiler complains about them to help us catch our mistakes:

```go
// the compiler doesn't care that parameter 'c' is never used
func foo(a int, b int, c int) int {
    var s string        // the compiler will give us an error because 's' is never used
    return a + b
}
```

If a function returns multiple values but we don't want to use them all, we can assign the values to the 'blank identifier', effectively discarding those values:

```go
// assume foo returns three values
a, b, c := foo()                        // ok
d, e := foo()                           // compile error: must have three assignment targets
f, g, _ := foo()                        // ok: discard last value
_, h, _ := foo()                        // ok: discard first and last values
_, j, k := foo()                        // ok: discard first values
_, _, l := foo()                        // ok: discard first and second values
```

The blank identifier is also useful in `for-range` loops when we want only the index/key or only the value:

```go
for _, v := range foo {
    // ... we only want the values of foo, not the indexes
}

for i, _ := range foo {
    // ... we only want the indexes of foo, not the values
}

// this is shorthand for the previous loop
for i := range foo {
    // ... we only want the indexes of foo, not the values
}
```

As a shorthand, we can combine multiple successive `var` statements into one using parens:

```go
var (
    a = 5
    b = "hi"
    c = 2
)
```

...is equivalent to...

```go
var a = 5
var b = "hi"
var c = 2
```

## globals

Global variables are created with `var` statements outside of any function:

```go
// outside any function
var foo string = "hi"            // global string variable foo with initial value "hi"
var bar = true                   // global bool variable bar with initial value true (type is left inferred)
ack := 3                         // compile error: cannot use := to create a global
var ack int                      // global int variable ack with initial value 0 (the default int value)
```

The initial value of a global can be any expression. These expressions are evaluated before the initial call to 'main'. If a global initialization expression uses the value of another global, the compiler will figure out the necessary initialization order:

```go
// outside any function
// the compiler will initialize bar, then foo, then ack
var foo = bar * 2
var bar = 4
var ack = 7 - foo                
```

The compiler will give you an error if your global initializations depend upon each other in a loop:

```go
// outside any function
// compile error: dependency loop: bar depends upon ack, which depends upon foo, which depends upon bar
var foo = bar * 2
var bar = ack
var ack = 7 - foo
```

## constants

Number literals in Go are called *constants* and have no particular type, meaning, say, `52` is neither a `uint8`, an `int64`, or any other type of integer. When assigning a number constant to a variable, the compiler simply requires that the value be valid for the variable's type:

```go
var x float32 = 53.8       // OK
var y int = 53.8           // compile error: 53.8 is not a valid int value  
var z byte = 9000          // compile error: 9000 is not a valid byte value (the max byte value is 255)
```

A `const` statement creates a named constant. These are not variables: they are just constant values represented by a name. The value 'assigned' in `const` must be a compile-time expression:

```go
const x = 3.5                 // x is now a name for the constant 3.5
x = 7                         // compile error: cannot assign to a constant (except at creation)
const y = 9 * 10              // OK
const z = foo()               // compile error: functions can only be called at runtime
```

A `const` statement at the top-level of code is global. A `const` statement in a function is local to the scope.

If we specify a type for a constant, the compiler considers it to be a value of that type and only that type:

```go
const x uint16 = 500
var y int = x                // compile error: cannot assign a uint16 value to an int variable
```

We can give a constant a type:

```go
const foo float32 = 53.8       // constant foo has type float32
var bar float32 = foo          // OK
var ack int = foo              // compile error: foo is not an int
foo = 4.89                     // compile error: cannot assign a new value to a constant
```

## iota

In the parentheses form of `const`, we can use the reserved word `iota` as the value for the first constant and leave the value of the other constants implied. The first constant will be 0, the second will be 1, the third 2, *etc.*:

```go
const (
    foo = iota           // 0
    bar                  // 1
    ack                  // 2
)
```

If we specify a type for the first constant, all other constants will have the same type:

```go
const (
    foo int64 = iota     // int64(0)
    bar                  // int64(1)
    ack                  // int64(2)
)
```

The word `iota` can be used in an expression. The same expression is used to generate all the constant values, with `iota` as 0 for the first constant and incrementing by 1 for each additional constant:

```go
const (
    a = 3 * iota         // 0     (3 * 0)
    b                    // 3     (3 * 1)
    c                    // 6     (3 * 2)
    d                    // 9     (3 * 3)
    e                    // 12    (3 * 4)
    f                    // 15    (3 * 5)
)
```

## semi-colon insertion

Several kinds of statements (including `var`, assignments, and function call statements) require semi-colons at the end:

```go
var x int;
x = 3;
foo(x);
```

However, before it reads your code, the Go compiler will insert semi-colons at the end of any line ending with:

- an identifier (a name defined in code, such as a variable or function name)
- a constant (*e.g.* `35.7`, `“hi”`, `true`, `false`)
- `break`
- `continue`
- `fallthrough`
- `return`
- `++`
- `--`
- `)`
- `}`

Common practice is to not write these semi-colons explicitly.

## pointer syntax

A pointer type is denoted by prefixing * on the type:

```go
var x *int                    // P<I>
var y **string                // P<P<Str>>
```

The equivalent of Pigeon's `ref` and `dr` are `&` and `*`. To assign to the location referenced by a pointer, we don't have `set` but instead assign to a dereference of the pointer:

```go
var x *int
var y int
x = &y                       // as x (ref y)
*x = 5                       // (set x 5)
var z int = *x               // assign 5 to z
```

Be clear that `*` has three meanings:

- x * y (binary operator for multiplication)
- *x (unary operator for dereferencing)
- var x *int (type modifier to make a pointer)

## array and slice syntax

An array type is denoted by prefixing the size of the array in `[]`, and a slice type is denoted by prefixing empty `[]`:

```go
var x [4]int          // x is an array of 4 ints
var y []int           // y is a slice of ints
```

An array or slice value can be created by suffixing the type with values listed inside `{}`:

```go
var x [4]int = [4]int{7, 3, 5, 2}        // (A<I 4> 7 3 5 2)
var y []int = []int{3, 8, 9, 1, -11}     // (S<I> 3 8 9 1 -11)
```

Instead of `get`, we suffix an array or slice with the index inside `[]`:

```go
var x [4]int = [4]int{7, 3, 5, 2}        // (A<I 4> 7 3 5 2)
var z int = x[2]                         // (get x 2)
```

Instead of `set`, we assign to an array or slice suffixed with the index inside `[]`:

```go
var x [4]int = [4]int{7, 3, 5, 2}        // (A<I 4> 7 3 5 2)
x[2] = 100                               // (set x 2 100)
```

Instead of `slice`, we suffix an array or slice with start and end indexes separated by a colon inside `[]`:

```go
var x [4]int = [4]int{7, 3, 5, 2}        // (A<I 4> 7 3 5 2)
var y []int = x[1:3]                     // (slice x 1 3)
```

The `make`, `append`, `len`, and `cap` operators are considered built-in functions (because they syntatically look like function calls):

```go
var x []string = make([]string, 10, 20)              // as x (make S<Str> 10 20)
x = append(x, "hi")
len(x)                                               // 11
cap(x)                                               // 20
```

## anonymous functions

A function type is denoted by `func` followed by the parameter types and return types in parens, separated by commas:

```go
var x func (int, string) (bool, int)     // Fn<I, Str : Bool, I>
```

Unlike GoPigeon, Go has 'anonymous functions', functions written as expressions:

```go
var x func (int) int         // variable x is a function taking an int and returning an int
x = func (a int) {           // assign a new function to x
    return a + 3
}
var y int = x(11)            // assign 14 to y
```

An anonymous function has access to the variables in the containing scope that precede it:

```go
var b int = 9
x :=  func (a int) {
    return a + b             // b of the containing scope
}
```

## closures

## variadic functions

A variadic function is a function in which the last parameter is a slice denoted by `...` instead of the normal `[]`. A variadic function is called not by passing a slice to this last parameter but rather zero or more elements that get automatically bundled into a new slice:

```go
// 'foo' is variadic
// 'b' is a []int but gets its argument in a special way
func foo(a string, b ...int) {
    // ...
}

func main() {
    foo("hi", 3, 2, 7)         // passes []int{3, 2, 7} to parameter 'b'
    foo("hi", 3)               // passes []int{3} to parameter 'b'
    foo("hi")                  // passes []int{} to parameter 'b'
}
```

This minor syntax allowance simply spares us from creating these new slices explicitly in each call:

```go
// what we would have to write instead if 'foo' were not variadic
func main() {
    foo("hi", []int{3, 2, 7})
    foo("hi", []int{3})
    foo("hi", []int{})
}
```

If we want to pass an already existing slice to a variadic function, we can do so using `...` as a suffix on the last argument:

```go
func main() {
    x := []int{3, 2, 7}
    foo("hi", x...)            // passes the slice to parameter 'b'
}
```

## return variables

The return types of a function can be given associated variables. A return statement with no explict values returns the value(s) of the return variable(s). The return variables have their default values at the start of the call:

```go
// 'bar' has a return variable 'a' of type int 
func bar(x int) (a int, b string) {
    // 'a' starts out 0, 'b' starts out ""
    a = 3
    b = "hi"
    if x > 7 {
        return        // implicitly returns 'a' and 'b'
    }
    return x, b     
}

func main() {
    i, s := bar(10)    // 3, "hi"
    i, s = bar(5)      // 5, "hi"
}
```

Return variables can occasionally make a function look a bit cleaner in some cases where the function has many return statements. There are also some scenarios involving `defer` statements (discussed later) where return variables are needed.

## maps

A map type is denoted `map[x]y`, where 'x' is the key type and 'y' is the value type. We use `[]` to get and set on a map:

```go
var x map[int]string                  // variable x stores a reference to a map (but the defualt value is nil)
x = make(map[int]string)              // we can use make to create a new empty map of the specified type
len(x)                                // 0
x[5] = "hi"                           // set key 5 to have the value "hi"
len(x)                                // 1
var s string = x[5]                   // assign "hi" to s
```

For comparison, the GoPigeon equivalent:

```
locals x M<I Str> s Str
as x (make M<I Str>)
(len x)
(set x 5 "hi")
(len x)
as s (get x 5)
```

We can use `{}` after a map type to create a new map with zero or more key-value pairs:

```go
var x map[int]string
x = map[int]string{}                     // assign to a new empty map to x
len(x)                                   // 0
x = map[int]string{7: "hi", 84: "yo"}
len(x)                                   // 2
var s string = x[84]                     // 84
```

For comparison, the GoPigeon equivalent:

```
locals x M<I Str> s Str
as x (M<I Str>)
(len x)
as x (M<I Str> 7 "hi" 84 "yo")
(len x)
as s (get x 84)
```

## named types

Using `type`, we can define a *named type*. The new type is not an alias:

```go
type fred int          // define named type fred to be an int

func main() {
    var x fred = 5     // OK (a fred is really just an int)
    var y int = 4
    x = y              // compile error: cannot assign an int to a fred variable
}
```

It may seem strange to create a new type that is just like an existing type, but it is sometimes useful to make distinctions between different uses for the same underlying data representation. For example, floating-point numbers can represent all kinds of things, like quantities of money, mass, or time. If we create three distinct types for money, mass, and time, the compiler can then catch cases where we misuse values of these types:

```go
type dollars float64     
type seconds float64

func makeItRain(d dollars) dollars {
    // because dollars are really just float64's, we can perform arithmetic upon them
    return d * 100                           
}

func main() {
    var myMoney dollars = 3.50          // OK (number constants have no specific type)
    var myTime seconds = 60
    myMoney = makeItRain(myMoney)       // OK
    myMoney = makeItRain(myTime)        // compile error: argument must be of type dollars
}
```

Above, the compiler catches when we mistakenly try passing a seconds value to a function defined to take a dollars value.

## methods

A method is written like a function but with the receiver in parens before the function name:

```go
func (c Cat) sleep(hours float64) float64 {
    // ... body
}
```

The equivalent in GoPigeon is written:

```
method sleep c Cat hours F : F
    // ... body
```

There is no method call operator. Instead we invoke methods with the dot operator:

```go
c := cat{}
c.sleep(4.3)          // call the sleep method with receiver 'c' and argument 4.3
```

As a convenience, we can invoke methods of non-pointer types *via* pointers without having to explicitly derefernece:

```go
c := &cat{}            // c is a cat pointer
c.sleep(4.3)           // (*c).sleep(4.3)
```

We can also call methods of pointer types on non-pointer values without having to explicitly use `&`:

```go
// a method of pointer-to-Cat
func (c *Cat) sleep(hours float64) float64 {
    // ... body
}

func main() {
    c := cat{}             // c is a regular cat, not a pointer-to-cat
    c.sleep(4.3)           // (&c).sleep(4.3)
}
```

## method values

A *method value* creates a function that represents a method but also has a bound value for the receiver. A method value is written like a method call without parens or arguments:

```
type kim int

func (k kim) foo(a int) int {
    return int(k) + a
}

func main() {
    v := kim(5)
    a := v.foo(2)                   // 7
    var f func(int) int
    f = v.foo                       // creates a function with bound value kim(5)
    b := f(2)                       // 7
}
```

(A struct type cannot have a field and method of the same name because then this syntax for method values would be ambiguous.)

## method expressions

A *method expression* creates a function that represents a method but replaces the receiver with an ordinary parameter. A method expression is written with the receiver type in parens, followed by dot and the method name:

```
type kim int

func (k kim) foo(a int) int {
    return int(k) + a
}

func main() {
    v := kim(5)
    a := v.foo(2)                   // 7
    var f func(kim, int) int
    f = (kim).foo                   // creates a function taking a kim and an int and returning an int
    b := f(v, 2)                    // 7
}
```

## structs

To define a struct:

```go
type cat struct {
    name string
    age int
    weight float32
}
```

To create values of the struct type, we use `{}` with values in the order we defined the fields:

```go
var c cat = cat{"Mittens", 10, 12.0}     // a cat with a name, age, and weight
c = cat{}                                // a cat with default values for its fields: "", 0, 0.0
c = cat{"Mittens", 10}                   // compile error: must provide values for all fields or no fields
```

If we specify the field names, we can write them in any order, and omitted fields have their default value:

```go
c = cat{weight: 12.0, name: "Mittens"}        // c = cat{"Mittens", 0, 12.0}
```

As a convenience, the dot operator used on a pointer to a struct implicitly dereferences the struct:

```go
c := &cat{"Mittens", 10, 12.0}     // c is a pointer to a cat
s := c.name                        // (*c).name
```

When declaring a struct type, if we omit the name of a struct field, the field is *embedded*, and the field’s name is implicitly the same as its type:

```go
type foo struct {
    a int
    b string
    c float32
}

type bar struct {
    a float32
    foo                      // embed 'foo' inside 'bar'
    x string
}

func main() {
    var b bar = bar{}
    var f foo = b.foo        // assign 'foo' field to variable 'f'
    b.foo.a = 3              // assign to field 'a' of the 'foo' field
}
```

As a convenience, the fields of an embedded struct can be accessed as if they are directly part of the embedding struct (even though they really aren’t!). However, if the embedding struct has a field of the same name, the embedded struct’s field can only be accessed *via* the embedded struct:

```go
func main() {
    var b bar = bar{}
    b.c = "hi"               // b.foo.c = "hi"
    b.a = 35.2               // assign to the float32 field of 'Bar'
    b.foo.a = 12             // assign to the int field of 'Foo'
}
```

If an embedded type has methods, we can call them as if they are directly methods of the embedding struct, but the embedded struct is what's actually passed as the receiver:

```go
func (f foo) roger() int {
    return f.a
}

func main() {
    var b bar = bar{}
    b.foo.a = 9
    x := b.roger()           // 9  (b.foo passed as receiver)
    y := b.foo.roger()       // 9
}
```

Methods of embedded structs count towards the embedding struct implementing interfaces:

```go
type alice interface {
    bob()
    carol()
}

type foo struct {
    // ...
}

type bar struct {
    // ...
    foo                      // embed 'foo' inside 'bar'
}

func (f foo) bob() {
    // ...
}

func (b bar) carol() {
    // ...
}

func main() {
    var a alice = bar{}      // OK: 'bar' implements 'alice'
}
```

A struct can embed pointers to structs:

```go
type foo struct {
    a int
    b string
    c float32
}

type bar struct {
    a float32
    *foo                     // embed pointer to 'foo' inside 'bar'
    x string
}

func main() {
    var b bar = bar{}
    b.foo = &foo{}           // the 'foo' pointer needs an actual foo to point to
    b.foo.a = 3              // (*b.foo).A = 3
    var f *foo = b.foo       // assign 'foo' field to variable 'f'
}
```

## interfaces

To define an interface:

```go
type Sleeper interface {
    // implementors of sleeper must have a method sleep with a single float64 parameter and returning nothing
    sleep(float64)
    // implementors of sleeper must have a method wake with no parameters and returning a float64
    wake() float64 
}
```

The above would be written in GoPigeon as:

```
interface Sleeper
    sleep F
    wake : F
```

The equivalent of GoPigeon's Any is confusingly called the 'empty interface' and written `interface{}`:

```go
var x interface{}          // an empty interface variable
x = 5                      // OK
x = "hi"                   // OK
c = cat{}                  // OK
switch x.(type) {
case int:
    // ... x references an int
case string:
    // ... x references a string
case cat:
    // ... x references a cat
default:
    // ... x references something other than an integer, string, or Cat
}
```

The GoPigeon equivalent of above is:

```
locals x Any
as x 5
as x "hi"
as c (Cat)
typeswitch x
case I
    // ... x references an integer
case Str
    // ... x references a string
case Cat
    // ... x references a Cat
default
    // ... x references something other than an integer, string, or Cat
```

We can also use a type assertion operation to get the value referenced by an interface value:

```go
var x interface{}         // an empty interface variable
x = 5                     // OK
i := x.(int)              // get the referenced int value from 'x'
s := x.(string)           // runtime error because 'x' is not referencing a string
```

A type assertion in a multi-value context will not panic and returns a boolean indicating if the interface value references the specified type:

```go
var x interface{}
x = cat{"Mittens", 10, 12.0}
c, ok := x.(cat)      // assign cat{"Mittens", 10, 12.0} to 'c' and true to 'ok'
x = 5
c, ok = x.(cat)       // assign cat{} to 'c' and false to 'ok'
```

Only use the single-value form of type assertion in cases where it's certain that the interface value references the specified type. In all other cases, use type switches or the multi-value form of type assertion.

## defer statements

A *defer statement* defers execution of a function or method call. Every `defer` adds another call to a list belonging to the containing function or method call; when the call ends, its list of defered calls are executed in reverse order (*i.e.* the last defered call runs first).

```go
// prints: "1", then "2", then "3", then "4"
func foo() {
    fmt.Println("1")
    defer fmt.Println("4")
    defer fmt.Println("3")
    if false {
        defer fmt.Println("never")   // this defer statement is never executed, so this call is never defered
    }
    fmt.Println("2")
}
```

Defering calls can be useful for doing clean-up business, such as making sure a file is closed when execution leaves a call.

## panics

A runtime error in Go is called a ***panic***. A few things which trigger panics:

- accessing an array or slice index that is out of bounds
- invoking a method via a nil interface value
- sending to a closed channel
- asserting the wrong type using the single-return form of type assertion

When a panic occurs in a goroutine, execution backs out of the call chain, executing all deferred calls as it goes. For example, say a goroutine executes A, which calls B, which calls C, which panics. If A, B, and C have deferred calls before the panic, the deferred calls will run in reverse order: C, then B, then A.

Once a panic backs execution out of a goroutine, the whole program aborts regardless whether other goroutines are still executing.

Calling the built-in function panic triggers a panic in the current goroutine. Deliberately triggering panics is sometimes appropriate, such as when the caller passed bad arguments. (Passing bad arguments is a bug, not an error: we should fix the code to stop passing bad arguments.)

```go
func foo(a int, b int) int {
    // ...
    if badInput {
        panic()
    }
    // ...
}
```

We can stop a panic and resume a goroutine’s normal execution using the built-in function `recover`. When called directly from a defered call, `recover` stops the panic from propagating up to the next call:

```go
func foo() {
    defer func() {
        fmt.Println("still recovering")
    }()

    defer func() {
        recover()
        fmt.Println("recovering")
    }()

    panic()
}

func main() {
    foo()           // prints: "recovering", then "still recovering"
    // ... execution continues normally
}
```

Above, we recover in a defered call of foo, so execution resumes normally where foo was called. But what if foo returned a value?

```go
func foo() int {
    defer func() {
        recover()
    }()

    panic()
    return 3
}

func main() {
    z := foo()         // 0
}
```

Here, the recovered call returns a default value. Using return variables, defered calls can set the return value to something else:

```go
func foo() (a int) {
    defer func() {
        recover()
        a = 5
    }()

    panic()
    return 3
}

func main() {
    z := foo()         // 5
}
```

We can pass a single value of any type to panic. This value is then returned by recover (as an empty interface value):

```go
func foo() (a int) {
    defer func() {
        a = recover().(int)   // type assert into an int
    }()

    panic(7)
    return 3
}

func main() {
    z := foo()         // 7
}
```

If no value is passed to panic, recover returns nil.

A call to recover only works inside a defered call during a panic. All other calls to recover do nothign and return nil:

```go
func main() {
    z := recover()     // does nothing and returns empty interface value nil
}
```

If a panic is triggered while a panic is already in progress, the defered call where the second panic occurs aborts, but otherwise the panic continues as normal.

## channels

Channels types are denoted with the reserved word `chan` followed by the type of element in the channel. We create actual channel values with the `make` builtin:

```
var ch chan int            // create variable 'ch' to reference a channel of ints
ch = make(chan int, 10)    // assign to 'ch' a new channel of ints with a capacity of 10
```

## send- and receive-only channels

A normal channel reference is *bidirectional*. The compiler lets us send and receive *via* a bidirectional reference.

Send- and receive-only channel references are *unidirectional*: the compiler lets us only send or receive *via* these references, respectively. A send-only type reference is denoted `chan<-`. A receive-only type reference is denoted `<-chan`.

We can cast from a bidirectional reference to a unidirectional reference, but not the other way around. We cannot cast between send-only and receive-only references.

Be clear that a channel itself is always bidirectional: we create an actual channel value with `make`, but our channel expressions are just references. A single channel may be referenced by any number of bidirectional and unidirectional channel references.

```go
var bi chan int = make(chan int, 10)              // create a channel

// notice we must surround the chan types in parens to cast
var s chan<- int = (chan<- int)(bi)               // cast to a send-only reference
var r <-chan int = (<-chan int)(bi)               // cast to a receive-only reference

// all three variables reference the same channel
bi <- 3                 // ok
<-bi                    // ok
s <- 3                  // ok
<-s                     // compiler error: cannot receive via send-only reference
r <- 3                  // compile error: cannot send via receive-only reference
<-r                     // ok
```

We can leave casts from bidirectional to unidirectional references implicit:

```go
var bi chan int = make(chan int, 10)              
var send chan<- int = bi                   // cast is implicit
var receive <-chan int = bi                // cast is implicit
```

So why use unidirectional references? Very commonly, we intend for a particular goroutine to only send to a particular channel or only receive from that channel. Unidirectional references help us enforce that intention. When we spawn a new goroutine, we can pass it only a unidirectional reference, thereby ensuring the goroutine will only read or write the channel, not do both.

## closing channels

The built-in function `close` closes a channel. We can still receive from a closed channel, but sending to a closed channel triggers a panic. Once a closed channel has no more values to receive, any subsequent receive operations will return the zero value of the type without ever blocking:

```go
ch := make(chan int, 3)
ch <- 1
ch <- 2
ch <- 3
close(ch)
a := <-ch   // 1
b := <-ch   // 2
c := <-ch   // 3
d := <-ch   // 0
e := <-ch   // 0
```

To distinguish between a zero value sent through a channel and a zero value indicating the channel has closed, the receive operator can return two values. The first returned value is the value read from the channel, and the second is a boolean (true indicating the value was sent):

```go
ch := make(chan int, 3)
// ...
val, ok := <-ch      // 'ok' will be true if the value was sent
```

Closing a channel *via* a receive-only reference triggers a panic. (This makes sense because generally only senders know when there is no more stuff to send.)

Closing a channel which has already been closed triggers a panic.

## for-range with channels

A for-range loop is a convenient way to read from a channel until it closes. Each iteration receives a value from the channel (and will block accordingly, like any normal receive operation). Once the channel is closed and empty, the loop ends.

```go
ch := make(chan int, 10)
ch <- 6
ch <- 4
close(ch)

// after two iterations, loop ends because the channel is closed
for v := range ch {
    fmt.Println(v)
}
```

The loop above is simply a more compact way to write the below:

```go
for v, ok := <-ch; ok; v, ok = <-ch {
    fmt.Println(v)
}
```

## goto statements and labels

We can prefix statements with *labels*, names suffixed with colons:

```go
george: foo()           // a statement with the label 'george'
maria: if x < 3 {       // an 'if' statement with the label 'maria'
    // ...
}
```

The name of a label must be unique among other labels within the same function/method.

Having labeled a statement, we can jump execution to that statement with a goto statement in the same function/method:

```go
if x < 3 {
    goto george         // jump execution to the statement labeled 'george'
}
bar()
george: foo()
```

A goto statement may not jump to a position where variables should exist but their declarations have been skipped over:

```go
if x < 3 {
    goto george         // compile error: cannot jump over declaration of 'y'
}
y := 3
george: foo()      
```

When we nest loops within other loops, we can break or continue an outer loop from an inner loop using labels:

```go
var arr [30][10]int
// ... assign values to the array
sarah: 
for i := 0; i < 30; i++ {
    for j := 0; j < 10; j++ {
        v := arr[i][j]
        if v == 99 {
            break sarah             // break out of loop with the label 'sarah'
        }
        foo(v)
    }
}
```

(For visual clarity, it’s often best to write a label on the line preceding the statement which it labels.)

## packages and imports



## reflection

With type assertions, we can test if an interface value references a value of a specific type, but what if we want to know if the type is something more general, like an array, or a slice, or a number type? The special package “reflect” gives us the means to query the types of values referenced in interface values at run time, *a.k.a.* to do reflection. With reflection, we can write functions that take in interface values but then branch to handle different types of input differently. The fmt.Println function, for example, is a variadic function taking a slice of empty interface values, and it uses reflection to discover the types of these inputs and then create an appropriate text representation for any kind of input.

Reflection is not a commonly used feature, so we won’t cover it here, but it’s worth mentioning because some parts of the standard library rely upon it.