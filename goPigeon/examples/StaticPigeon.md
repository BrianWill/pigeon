static pigeon
    member import only
    no channels, no goroutines
    only one number type: N (a float64)
    lists instead of arrays and slices
    maps
    structs and pointers
    interfaces
    forall loop
    Any type
    methods
    typeswitch
    function variables, but no anon functions
    no return parameters
    no variadic functions (only a few special built-in functions are variadic)
    no embedding
    no const, iota
    break, continue, no goto
    bad ops abort the program, no defer or recover




# local variables and assignment

Every local variable declared with `var` must be followed by the type of the variable. For example:

```
var apple S            // declare a string (denoted `S`) variable `apple`
var orange N           // declare a number (denoted `N`) variable `orange`
var banana Bool        // declare a boolean (denoted `Bool`) variable `banana`
```

Only types can begin with uppercase letters, so variables and functions must all begin with lowercase letters.

A variable can only be assigned values of its declared type:

```
var apple S
as apple "hi"          // OK
as apple 5             // compile error: cannot assign a number to a string variable
as apple nil           // compile error: cannot assign nil to a string variable
```

When declared, a variable starts out with the default value for its type:

- the default number value is 0
- the default string value is an empty string, ""
- the default boolean value is false

We can declare a variable in an `as` statement by suffixing the target variable with `'`. The type of the variable is implicit from the type of the value being assigned to it:

```
as x' "hi"        // because "hi" is a string, new variable `x` is a string
as y' x           // because `x` is a string, new variable `y` is a string
as y' 4           // because `4` is a number, new variable `y` is a number
```

## string equality and relational comparisons

We can use `eq` to compare strings for equality:

```
as s' "hi"
as s2' "hi"
as s3' "yo"
as bool' (eq s s2)       // true
as bool (eq s s3)        // false
```

The relational operators, `lt` and `gt`, compare the first non-equal character of the two strings by their Unicode values:

```
as s' "hit"
as s2' "hiss"
as bool' (lt s s2)         // false ("t" is not less than "s")
as bool (gt s s2)          // true ("t" greater than "s")
```

# global variables

A global variable is declared and initialized with a `global` statement:

```
global x S "hi"               // create global string variable 'x' initialized to "hi"
```

# function return types

A function must specify the type of value it returns:

```
// function `foo` returns a boolean
func foo : Bool
    // ...
```

A function must return the type of value it claims to return:

```
func apple : Bool
    return true      // OK

func banana : Bool
    return 5         // compile error: the function says it will return a boolean, not a number

func orange : Bool
    return           // compile error: the function says it will return a boolean, so cannot return nothing
```

By declaring the return type, the compiler knows what type of value any function call returns:

```
var x Bool
as x (apple)          // OK because `apple` returns a boolean
var y N
as y (apple)          // compile error: cannot assign a boolean to number variable `y`
```

# function parameter types

A function must specify the types of its parameters:

```
// function `mango` returns a boolean and has two parameters: number parameter `x` and string parameter `y`
func mango x N y S : Bool
    return true
```

When calling a function, we must pass arguments of the types it expects:

```
(mango 3 "hi")         // OK
(mango 3)              // compile error: too few arguments
(mango 3 "hi" 4)       // compile error: too many arguments
(mango false "hi")     // compile error: argument types do not match the parameters
(mango "hi" 3)         // compile error: argument types do not match the parameters
```

# multi-return functions

A function can return multiple values:

```
// function `kiwi` returns a number and a string
func kiwi : N S
    return 3 "hi"
```

A multi-return function cannot be called where only a single value is expected:

```
(kiwi)                   // OK: the returned values get discarded
as x' (add 3 (kiwi))     // compile error: 'kiwi' returns two values, not just a single number
```

An `as` statement can have multiple assignment targets to receive the multiple values returned by a function:

```
var x N
var y S
as x y (kiwi)          // OK: 'x' is assigned the number and 'y' is assigned the string
as x (kiwi)            // compile error: too few assignment targets
```

If we don't care about one or more of the values returned, we can discard them by using `_`, the 'blank identifier', as the target of assignment:

```
as a' b' c' (foo)       // (foo) returns 3 values
as _ x' _ (foo)         // discard first and third values returned by (foo)
```

# lists

A list in StaticPigeon must be homogenous, meaning it must be composed of values of all the same type. The value type is declared in angle brackets suffixing `L`:

```
var x L<N>       // variable 'x' stores a list of numbers
var y L<S>       // variable 'y' stores a list of strings
var z L<L<N>>    // variable 'z' stores a list of lists of numbers
```

To create a list, we use the list type like an operator, providing zero or more values as operands:

```
var x L<N>
as x (L<N> 5 -12 90)       // create a list of numbers with three values: 5, -12, 90
var y L<S>
as y (L<S> "hi" "bye")     // create a list of strings with two values: "hi", "bye"
var z L<N>
as z x                     // OK
as x y                     // compile error: can only assign lists of numbers to 'x'
as y (L<S> 7 false)        // compile error: the list can only have string values
```

To access the values of a list, we have the `get` and `set` operators:

```
as a' (L<N> 5 -12 90)
as b' (get a 0)             // assign 5 to 'b'
as c' (get a 2)             // assign 90 to 'c'
as d' (get a 3)             // runtime error: index 3 is out of bounds

(set a 0 -8)                // set index 0 of the list to -8
(set a 3 70)                // runtime error: index 3 is out of bounds
```

The `append` operator appends a value to the end of a list, increasing its size by one:

```
as a' (L<N> 5 -12 90)
(append a 75)               // increase length of list by one, setting index 3 of the list to 75
```

The `len` ('length') operator returns the number of elements in the list:

```
as a' (L<N> 5 -12 90)
as b' (len a)               // assign 3 to 'b'
```

# maps

A map in StaticPigeon must be homogenous, meaning it must be composed of keys of all the same type and values of all the same type. The keys can only be numbers or strings. The key and value types are declared in angle brackets suffixing `M`:

```
var x M<N S>         // variable 'x' stores a map of numbers to strings
var y M<S Bool>      // variable 'y' stores a map of strings to booleans
var z M<S M<N N>>    // variable 'z' stores a map of strings to maps of numbers to numbers
```

To create a map, we use the map type like an operator, providing zero or more key-value pairs as operands:

```
as x' (M<S N> "hi" 5 "bye" -12)     // create a map of strings to numbers with two key-value 
                                    // pairs: "hi" with the value 5, "bye" with the value -12
as y' (M<S B>)                      // create a map of strings to booleans with no key-value pairs
as x y                              // compile error: can only assign maps of strings to numbers to 'x'
```

To access the key-value pairs, we use the `get` and `set` operators:

```
as a' (M<S N> "hi" 5 "bye" -12)
as b' (get a "hi")                   // assign 5 to 'b'
as c' (get a "bye")                  // assign 90 to 'c'
as d' (get a "yo")                   // runtime error: the map has no key "yo"

(set a "hi" -8)                      // set key "hi" to have the value -8
(set a "yo" 70)                      // create new pair: key "yo" with the value 70
```

The `hasKey` operator returns true if a map contains a specified key:

```
as a' (M<S N> "hi" 5 "bye" -12)
as b' (hasKey a "hi")                // assign true to 'b'
as c' (hasKey a "yo")                // assign false to 'c'
```

The `len` ('length') operator returns the number of elements in the map:

```
as a' (M<S N> "hi" 5 "bye" -12)
as b' (len a)                        // assign 2 to 'b'
```

# `get` and `set` shorthand

We can use square brackets as shorthand for `get`:

```
as a' (L<N> 5 -12 90)
as b' a[0]                    // as b' (get a 0)
```

Similarly, we can use square brackets in the target of assignment as shorthand for `set`:

```
as a' (M<S N> "hi" 5)
as a["yo"] -8                 // (set a "yo" -8)
```

If no value is given, the variable starts out with the default value for its type.

# foreach loops

Because looping through all the elements of a list or map is such a common thing to do, we have a convenience for it, the `forall` loop. Whereas a conventional loop is written like so:

```
as list' (L<N> 70 80 90 100 110)

// prints 0 70 1 80 2 90 3 100 4 110
as i' 0
while (lt i (len list))
    (print i)
    (print list[i])                
    asinc i
```

...we can write the same loop with `foreach`:

```
as list' (L<N> 70 80 90 100 110)
// prints 0 70 1 80 2 90 3 100 4 110
foreach i v arr
    (print i)                  // print the index
    (print v)                  // print the value   
```

A `foreach` declares two variables and takes an array or slice. In each iteration of the loop, the first variable is assigned an index, and the second variable is assigned the value at that index in the slice.

For a map, the first variable is the key while the second is the value:

```
as map' (M<S N> "hi" 5 "bye" -12)
// prints "hi", 5, "bye", -12
foreach k v map
   (print k)
   (print v) 
```

The key-value pairs of a map have no sense of order, so no guarantees are made about the order in which `foreach` iterates through the key-value pairs: the order is effectively random!

# structs

A struct (short for 'structure') is a data type we define, composed of one or more elements, called 'fields'. Structs are defined at the top-level of code, along side the globals and functions. After the reserved word `struct`, we specify the name of the struct (which must begin with a capital letter), and then list the fields indented on successive lines, each field having a name and type:

```
// define a struct type Cat with three fields
struct Cat
    name S          // string field 'name'
    age N           // number field 'age'
    weight N        // number field 'weight'
```

Only types can begin with uppercase letters, so fields must all begin with lowercase letters.

Having defined a struct, we can then create variables and values of its type:

```
var x Cat
as x (Cat "Mittens" 10 12.3)     // assign to 'x' a new Cat value
```

(When creating a struct value, the field values are expected in the same order they were declared in the struct. The fields of an uninitialized struct variable all have the default values of their types.)

To access the fields of a struct value, we use the `get` sand `set` operators:

```
as y' (get x age)          // assign 10 to 'y'
(set x name 11)            // set the age field of 'x' to 11
```

(Notice that we specify the field names directly by their names, not as strings.)

As shorthand for `get` and `set` to access fields, we can use the dot operator:

```
as y' x.age               // assign 10 to 'y'
as x.age 11               // set the age field of 'x' to 11
```

# methods

A *method* is like a function, but the first parameter called the 'receiver'. The receiver's type can only be a struct, not a number or string:s

```
// a method with Cat receiver 'x', number parameter 'y', and which returns a string
method pineapple x Cat y N : S
    return "yo"
```

When called, a method's name is prefixed with a dot:

```
as c' (Cat "Oscar" 13 15.5)
(.pineapple c 404)            // call the pineapple method with 'c' passed to the Cat receiver
```

A method belongs to the namespace of its receiver type, so we can have as many methods of the same name as we like as long as they all have different receiver types. For example, we can define a `pineapple` method with a Dog receiver which is separate from the one defined with a Cat receiver:

```
// a method with Dog receiver 'a', string parameter 'b', boolean parameter 'c', and which returns a number
method pineapple a Dog b S c Bool: N
    return 4
```

When calling a method, the compiler knows which method is being called based on the type of the first argument:

```
as d' (Dog)
(.pineapple d "bark" true)   // call the pineapple method with a Dog receiver
```

Method names do not conflict with the names of functions or variables. If we define a `pineapple` function, it is separate from any method(s) named `pineapple`.

# interfaces

An interface is a data type that simply lists method signatures (method names, parameter types, and return types).

```
// an interface named Justin with two method signatures
interface Justin
    foo N : S    // a method named foo with a number parameter and returning a string
    bar          // a method named bar with no parameters
```

When a struct has methods matching all of the signatures of an interface, that struct is said to 'implement' the interface.

```
// struct Dog implements the Justin interface:

method foo d Dog n N : S
    return "yo"

method bar d Dog
    return
```

We cannot create values of an interface type itself, but any values of a struct implementing an interface are valid values of that interface:

```
var x Justin
as x (Dog)                        // OK if struct Dog implements Justin
as x (Cat "Fluffy" 13 15.5)       // OK if struct Cat implements Justin
```

Above, though the variable is assigned a Dog and a Cat, the compiler doesn't ever presume to know what type of value is currently assigned to 'x': as far as the compiler is concerned, `x` is always a Justin value, not any struct type in particular. Therefore, the only thing the compiler lets us do with `x` is invoke the methods of Justin:

```
method meow c Cat : S
    return "meow"

var x Justin
as x (Cat "Fluffy" 13 15.5)        // OK: assume Cat implements Justin
as x.name "Mittens"                // compile error: the Justin interface has no fields
(.meow x)                          // compile error: the Justin interface has no method 'meow'
(.bar x)                           // OK: the Justin interface has a method bar
```

Which actual method gets called *via* an interface value depends on what actual struct the interface value holds at that moment:

```
var x Justin
as x (Cat "Fluffy" 13 15.5)        
(.bar x)                           // call method bar of Cat
as x (Dog)
(.bar x)                           // call method bar of Dog
as 
```

A struct can implement any number of interfaces.

# the `isActual` operator

The `isActual` operator returns true if an interface value holds a particular struct type:

```
var x Justin
as x (Dog)        
as y' (isActual Dog x)               // return true (because 'x' holds a Dog value)
as z' (isActual Cat x)               // return false (because 'x' does not hold a Cat value)
```

# the `actual` operator

The `actual` operator asserts the struct type held in an interface value and returns that struct value as its own struct type:

```
var j Justin
as j (Dog)  

var d Dog
as d (actual Dog j)                   // returns the Dog value held in 'j'

as j (Cat "Mittens" 7 10.2)
as d (actual Dog j)                   // runtime error: 'j' does not currently hold a Dog
```

# the Any type

The special type Any is like an interface with no methods. All other types are considered to implement the Any type, so all values are valid Any values:

```
var a Any
as a 3              // OK
as a "hi"           // OK
as a (L<N>)         // OK
as a (Dog)          // OK
```

To get a value held as an Any back as its own type, use the `actual` operator:

```
var a Any
as a 3
var n N
as n (actual N a)        // assign 3 to 'n'
```

# switchType

```
switchType x
case N S
    // ...x here is original type
case L<N>
    // ...x here is L<N>
case Foo
    // ...x here is Foo
default
    // ...x here is original type
```


# pointers

A *pointer* is a value that represents a memory address of some variable or field. Like with lists and maps, there is no single pointer type but rather a pointer type for every other type in the language, *e.g.* the `P<N>` type is a pointer to a number variable or field.

```
var x P<N>            // declare variable 'x' as a pointer to a number
var y P<S>            // declare variable 'y' as a pointer to a string
var z P<L<S>>         // declare variable 'z' as a pointer to a list of strings
```

To get a pointer value that represents the address of a variable, we use the `ref` ('reference') operator:

```
var message S
var p P<S>
as p (ref message)    // assign address of 'message' to 'p'

var num N
as p (ref num)        // compile error: cannot assign a pointer to a number to 'p'
```

We can also use the `ref` operator on fields of a struct value:

```
var c Cat
var p P<S>
as p (ref c.name)     // assign address of 'c.name' to 'p'
```

To retrieve the value at the location represented by a pointer value, we use the `der` ('dereference') operator:

```
var x N
as x 5

var p P<N>
as p (ref x)        

var y N
as y (der x)           // assign 5 to 'y'
```

The special value `nil` is a pointer value that points to nothing. A pointer of any type can be `nil`. An uninitialized pointer variable defaults to `nil`. Dereference `nil` is an error that aborts the program.

(The default value for a list variable is an empty list. The default value for a map variable is an empty map.)

# packages

Each file of code in StaticPigeon is its own *package*, its own module of code. To use the globals, functions, structs, and interfaces defined in another package, we must import them. An `import` statement specifies a package with a file path string and then, on successive lines, the name(s) of elements to import from that package:

```
// import 'Cat', 'Dog', 'foo', and 'bar' from 
// package "somedir/otherpackage.sp" (a relative path)
import "somedir/otherpackage.sp"
    Cat
    Dog
    foo
    bar
```

Imported names cannot conflict with other global, function, struct, or interface names in the package. To resolve name conflicts, we can alias our imports:

```
import "other.sp"
    apple
    banana mango   // what "other.sp" calls 'banana' will be known as 'mango' in this package
    orange
```

The import statements must proceded everything else in the package.

Methods do not need to be imported: if you import a struct, you can call all of its methods. A method can only be defined in the same package as its receiver struct type.

# functions and methods as values

We can create variables that store functions:

```
// 'potato' is a variable for storing a function with a number parameter 
// and a string parameter and returning a string
var potato F<N S : S>

// 'turnip' is a variable for storing a function with no parameters and returning nothing
var turnip F
```

Methods are not values, but we can use `meth2func` to create a function from a method:

```
method pumpkin c Cat n N : S
    // ...


// somewhere else in code...
var carrot F<Cat N : S>
as carrot (meth2func pumpkin Cat)
```

Note that, because a method could be define for multiple receivers, we must specify a particular struct type when using `meth2func`.

# dot operator with pointers

If we use the dot operator on a pointer to a struct, the pointer is implicitly derefereneced:

```
as c' (Cat)
as p' (ref c)           // 'p' is a pointer to Cat
as s' (der c).name        

// the dereference above can be left implicit
as s' c.name             
```

Similarly, if we pass a pointer to a struct as the first argument to a method, the pointer is implicitly dereferenced:

```
as c' (Cat)
as p' (ref c)           // 'p' is a pointer to Cat
(.sleep (der c)) 

// the dereference above can be left implicit
(.sleep c)              
```

# the E (error) interface



# nil

Types that can be nil: pointer, map, list, interface.




# read and write files


# time

# os

# strings
# strconv

# encoding

The `newE` operator creates an error wrapping a string 





===========================================

DynamicPigeon
StaticPigeon
GoPigeon



