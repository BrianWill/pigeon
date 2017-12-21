# StaticPigeon reference

Experienced programmers can probably glean everything they need to know from this reference, but beginners should watch the [tutorial videos](http://youtube.com/).

## definitions

A StaticPigeon program consists of definitions at the top-level of code starting with these reserved words:

### `func`

```
// a function named 'foo' that returns an integer and a string and expects two parameters:
// 'x' (a boolean)
// 'y' (a float)
func foo x Bool y F : I Str
    // the body has two statements
    (println x y)
    return 3 "yo"
```

### `global`

```
global x I 0          // a global variable named 'x' of type integer with initial value 0
global s Str "hi"     // a global variable named 's' of type string with initial value "hi"
```

### `struct`

```
// a struct type named 'Dog' with two fields
struct Dog
    name Str         // a field named 'name' of type string
    weight F         // a field named 'weight' of type float
```

### `interface`

```
// an interface named 'Mover' with three method signatures
interface Mover
    teleport I I      // a method named 'teleport' returning nothing and expecting two integers
    swim F : Str      // a method named 'swim' returning a string and expecting a float
    walk F            // a method named 'walk' returning nothing and expecting a float
```

### `method`

```
// a method of struct 'Dog' named 'foo' that returns a string
// and expects a single parameter 's' (a string)
method foo d Dog s Str : Str
    // the body has two statements
    (println d s)
    return "bark"
```

## data types

Booleans (`Bool`) remain unchanged from DynamicPigeon. The default boolean value is `false`.

Strings (`Str`) have additional operators related to slices. The default string value is `""` (an empty string). 

Whereas DynamicPigeon has just one type of number, StaticPigeon has three:

- 64-bit signed integers (`I`)
- 8-bit unsigned integers (`Byte`)
- 64-bit floating-point (`F`)

A number literal with a decimal point is a floating-point number; a number literal *without* a decimal point is an integer. Byte values are created by using `Byte` as an operator with an integer operand:

```
25              // an integer
25.0            // a float
(Byte 25)       // a byte
```

The inputs of an arithmetic operation must all be the same. The type returned is the type of the inputs:

```
(add 5 2 8.2)                // compile error: mixing integers and floats
(mul 5 2 8)                  // returns an integer
(mul 5.0 2.0 8.0)            // returns a float
(add (Byte 5) (Byte 2))      // returns a byte
```

The default integer value is `0`, and the default float value is `0.0`.

### lists

Lists in StaticPigeon must be 'homogenous', meaning a single list can store only one type of thing, *e.g.* a list of integers can only store integers, a list of booleans can only store booleans, a list of strings can only store strings, *etc.*

For every list, we must denote the type of the list. This type is then enforced by the compiler in `get` and `set` operations on the list:

```
func main
    locals a L<I> b L<Str> c Str d I
    as a (L<I> 7 -3 2)     // assign to 'a' a new list of integers with values 7, -3, 2
    as c (get a 0)         // compile error: cannot assign integer to 'c'
    as d (get a 0)         // OK
    (set a 0 98.6)         // compile error: cannot set float in list of integers
    (set a 0 98)           // OK
```

### maps

Every map must specify the type of its keys and the type of its values. Keys can be integers, floats, or strings; the values can be any type:

```
func main
    locals a M<I Str> b Str
    as a (M<I Str> 3 "hi" 9 "yo")    // assign to 'a' a new map of intgers to strings with two key-value pairs
    as b (get a 3)                   // "hi"
    as b (get a 4)                   // "" (default string value because the map has no key 4)
    (set a 8 "aloha")                  // add new key-value pair: key 8 with value "aloha"
    (set a 8.0 "aloha")                // compile error: key must be an integer, not a float
```

The default map value is nil.

### functions

Functions can be treated like values, but unlike in DynamicPigeon, functions with different signatures are considered different types, *e.g. a function returning an integer is a different type from a function returning a string.

```
func foo a I b Str : F
    // ... do something

func bar
    // ... do something

func main
    locals x Fn<I Str : F> y Fn<>
    as x foo
    as y bar
    as x bar         // compile error: wrong type of function
    (x 3 "hi")       // calls 'foo' (because 'x' currently references 'foo') 
    (y)              // calls 'bar' (because 'y' currently references 'bar')
```

### arrays

An array is like a list but has a fixed size upon creation. The size of an array is considered to be integral to its type, *e.g.* an array of 5 strings is a different type from an array of 4 strings.

An array variable stores an array directly, not by reference. Passing an array as input to a function copies the array to the parameter. Likewise, an array returned from a function is copied to the caller.

A default array value has default values for all of its elements.

### slices

A slice represents a subportion of an array. Each slice value has three components:

- a reference to some index within some array
- a 'length': an integer representing the logical length of the slice from that index within the array
- a 'capacity': an integer representing the number of indexes starting from that index within the array

A default slice value has length and capacity of 0.

### structs

A struct ('structure') is a composite data type defined by the programmer. A struct type has one or more fields, each with a name and type. A value of a struct type has its own set of the fields.

A default struct value has default values for all of its fields.

### pointers

A pointer is a value representing a storage location, *i.e.* a memory address. Pointers are distinguised at compile time by the type they point to, *e.g.* a pointer to an integer is a different type from a pointer to a string.

Using the `ref` ('reference') operator, we can get a pointer representing the address of a variable (or an index of an array variable, or a field of a struct variable).

The `dr` ('dereference') operator returns the value at the address represented by a pointer.

The `sr` ('set at reference') operator stores a value at the address represented by a pointer.

The default value for a pointer is `nil`.

### interfaces

An interface is a data type defined by the programmer which represents a set of method signatures. Any struct with methods matching all the signatures of an interface is said to 'implement' that interface.

An interface value is composed of two references: a reference to a value of an implementing struct type, along with a a reference to that struct's type itself.

The methods of an interface can be called on an interface value. (The actual method called depends at runtime upon the struct type which the interface value references.)

A `typeswitch` statement lets us branch on the type referenced by an interface value and to use the referenced value as its own type.

The default interface value is `nil`.

The special built-in interface type `Any` has no method signatures, and every type (even non-structs) is considered to implement `Any`.

## statements

There are several kinds of statements:

### `locals`

```
func main
    locals x I y Str           // 'main' has two local variables: 'x' (an integer) and 'y' (a string)
    locals z Bool              // compile error: only the first statement of a function can be a 'locals' statement
```

### `as`

```
func foo : I Str
    return 9 "yo"

func main
    locals x I y Str
    as x 3                    // assign 3 to 'x'
    as x "hi"                 // compile error: cannot assign a string to an integer variable
    as x y (foo)              // assign 9 to 'x', "yo" to 'y'
```

### `return`

```
func foo a F : I
    if (gt a 10)
        return "hi"           // compilation error: this function says it returns an integer, not a string         
    return 3                                  
```

### `if`

```
func foo x F
    if (eq x 4.3)
        (println "x equals 4.3")
    elif (eq x 1.689)
        (println "x equals 1.689")
    elif (eq x 7.9)
        (println "x equals 7.9")
    else
        (println "x does not equal 4.3, 1.689, or 7.9")
```

### `while`

```
func main
    locals x I               // initial value of 'x' is 0
    // this loop prints: 0 1 2 3  
    while (lt x 4)
        (println x)
        as x (inc x)         // increase value of 'x' by one
```

### `forinc`, `fordec`

```
func main
    // this loop prints: 0 1 2 3
    forinc x I 0 4
        (println x)
```

```
func main
    // this loop prints: 3 2 1 0
    fordec x I 4 0
        (println x)
```

### `foreach`

```
func main
    locals fruits L<Str>
    as fruits (L<Str> "banana" "apple" "grape" "orange")
    // this loop prints: 0 banana, 1 apple, 2 grape, 3 orange
    foreach i I s Str fruits
        (println i s)
```

### `break`, `continue`

```
func main
    locals fruits L<Str>
    as fruits (L<Str> "banana" "apple" "grape" "orange")
    // this loop prints: banana apple
    foreach i I s Str fruits
        if (eq s "grape")
            break              // jumps execution out of the loop
        (println s)
```

```
func main
    // this loop prints 1 3 5 7 9
    forinc i I 0 10
        if (eq 0 (mod i 2))        // if 'i' is an even number
            continue
        (println i)
```

### `typeswitch`

```
// assume Fruit is an interface implemented by structs Banana, Orange, and others
func foo f Fruit 
    // the order of the cases does not matter
    typeswitch f
    case b Banana
        // ... executed if 'f' references an Banana
    case o Orange
        // ... executed if 'f' references an Orange
    // the default clause (if present) must come last
    default
        // ... executed if 'f' references something other than a Banana or Orange
```

## arithmetic operators

`add` ('addition')

```
func main
    (add 3 5)            // 8
    (add 3 5 -14)        // -6
    (add 3.0 -5.0)       // -2.0
    (add 3.0 5)          // compile error: operands must be all integers or all floats
```

`sub` ('subtraction')

```
func main
    (sub 3 5)            // -2
    (sub 3 5 -14)        // 12
    (sub 3.0 -5.0)       // 8.0
    (sub 3.0 5)          // compile error: operands must be all integers or all floats
```

`mul` ('multiplication')

```
func main
    (mul 3 5)            // 15
    (mul 3 5 -14)        // -112
    (mul 3.0 -5.0)       // -15.0
    (mul 3.0 5)          // compile error: operands must be all integers or all floats
```

`div` ('division')

```
func main
    (div 9 3)            // 3
    (div 3 5)            // 0
    (div 3 5 -14)        // compile error: must have only two operands
    (div 9.0 3.0)        // 3.0
    (div 3.0 5.0)        // 0.6
```

`mod` ('modulus')

```
func main
    (mod 15 4)           // 3 (remainder of division)
    (mod 15 3)           // 0 (remainder of division)
    (mod 9.0 3.0)        // compile error: operands must be integers
```

## logic operators

`and`

```
func main
    (and false false false)            // false (all operands were false)
    (and true true true true)          // true (all operands were true)
    (and false true false false)       // false (not all operands were true)
```

`or`

```
func main
    (or false false false)            // false (all operands were false)
    (or true true true true)          // true (at least one operand was true)
    (or false true false false)       // true (at least one operand was true)
```

`not`

```
func main
    (not false)                // true (opposite of false)
    (not true)                 // false (opposite of true)
    (not true true)            // compile error: expecting only one operand
```

## comparison operators

`eq` ('equals')

```
func main
    (eq 53 53 53)                  // true (all operands equal)
    (eq 53 4 53)                   // false (not all operands equal)
    (eq "hi" 53 53)                // compile error: operands must be matching types
    (eq "hi" "hi" "hi" "hi")       // true (all operands equal)
    (eq "hi")                      // compile error: expecting at least two operands
```

`neq` ('not equals')

```
func main
    (neq 53 53 53)                  // false (all operands equal)
    (neq 53 4 53)                   // true (not all operands equal)
    (neq "hi" 53 53)                // true (not all operands equal)
    (neq "hi" "hi" "hi" "hi")       // false (all operands equal)
    (neq "hi")                      // compile error: expecting at least two operands
```

`lt` ('less than')

```
func main
    (lt 1 2 3)                     // true (every operand is less than the operand to its right)
    (lt 1 3 2)                     // false (not every operand is less than the operand to its right)
    (lt 1 3 3)                     // false (not every operand is less than the operand to its right)
    (lt 42.72 53)                  // compile error: operands must be all integers or all floats
    (lt 42.72 53.0)                // true
```

`lte` ('less than or equal')

```
func main
    (lte 1 2 3)                     // true (every operand is less than or equal to the operand to its right)
    (lte 1 3 2)                     // false (not every operand is less than or equal to the operand to its right)
    (lte 1 3 3)                     // true (every operand is less than or equal to the operand to its right)
    (lte 42.72 53)                  // compile error: operands must be all integers or all floats
    (lte 42.72 53.0)                // true
```

`gt` ('greater than')

```
func main
    (gt 3 2 1)                     // true (every operand is greater than the operand to its right)
    (gt 1 3 2)                     // false (not every operand is greater than the operand to its right)
    (gt 4 3 3)                     // false (not every operand is greater than the operand to its right)
    (gt 42.72 53)                  // compile error: operands must be all integers or all floats
    (gt 42.72 53.0)                // false
```

`gte` ('greater than or equal')

```
func main
    (gte 3 2 1)                    // true (every operand is greater than or equal to the operand to its right)
    (gte 1 3 2)                    // false (not every operand is greater than or equal to the operand to its right)
    (gte 4 3 3)                    // true (every operand is greater than or equal to the operand to its right)
    (gt 42.72 53)                  // compile error: operands must be all integers or all floats
    (gt 42.72 53.0)                // false
```

## string operators

`concat` ('concatenation')

```
func main
    (concat "red" 432 "blue")       // "red432blue"
```

`charlist`

```
func main
    (charlist "orange")             // (L<Str> "o" "r" "a" "n" "g" "e")
```

`charslice`

```
func main
    (charlist "orange")             // (S<Str> "o" "r" "a" "n" "g" "e")
```

`getchar` 

```
func main
    (getchar "orange" 0)            // "o"
    (getchar "orange" 1)            // "r"
    (getchar "orange" 6)            // runtime error: index out of bounds
```

`runelist`

```
func main
    (charlist "orange")             // a list of the individual Unicode character codes
                                    // (L<I> 111 82 97 110 103 101)
```

`runeslice`

```
func main
    (charlist "orange")             // a list of the individual Unicode character codes
                                    // (S<I> 111 82 97 110 103 101)
```

`getrune`

```
func main
    (getchar "orange" 0)            // 111 (the Unicode character code for "o")
    (getchar "orange" 1)            // 92 (the Unicode character code for "r")
    (getchar "orange" 6)            // runtime error: index out of bounds
```

## collection operators

`get`

```
func main 
    // return value at an index of a list, array, or slice
    (get (L<Str> "hi" "yo") 0)                   // "hi" (the value at index 0 of the array)
    (get (A<Str 2> "hi" "yo") 1)                 // "yo" (the value at index 1 of the array)
    (get (S<Str> "hi" "yo") 0)                   // "hi" (the value at index 0 of the array)
    // return value of a key of a map
    (get (M<Str I> "foo" 3 "bar" -2) "bar")      // -2 (the value of the key "bar" in the map)
```

`set`

```
func main 
    locals m M<Str I>
    // set value at an index of a list, array, or slice (returns the value set)
    (set (L<Str> "hi" "yo") 0 "bye")             // "bye" (the list now has "bye" at index 0)
    (sget (A<Str 2> "hi" "yo") 1 "avast")        // "avast" (the array now has "avast" at index 1)
    (set (S<Str> "hi" "yo") 0 "sup")             // "sup" (the array now has "sup" at index 0)
    // set value of a key of a map (returns the value set)
    as m (M<Str I> "foo" 3 "bar" -2)
    (set m "bar" 8)                              // 8 (the map now has value 8 for key "bar")
    (set m "ack" 11)                             // 11 (the map now has value 11 for new key "ack")
```

`push`

```
func main 
    locals a L<Str>
    as a (L<Str> "hi" "yo")
    (len a)                                      // 2
    (push a "bonjour")                           // returns nothing (adds the string to the end of the list)
    (len a)                                      // 3
    (push a "bye" "avast")                       // returns nothing (adds the two strings to the end of the list)
    (len a)                                      // 5
```

`slice`

```
func main
    locals a A<I 50> s S<I>
    as s (slice a 20 40)
    (len s)                                // 20 (the length of the slice)
    (cap s)                                // 30 (the capacity of the slice)
    (set s 0 999)                          // 999 (the array now has value 999 at index 20)
    as s (slice s 5 10)                    
    (len s)                                // 5 (the length of the slice)
    (cap s)                                // 25 (the capacity of the slice)
    (set s 0 77)                           // 77 (the array now has value 77 at index 25)
```

`append`

```
func main
    locals x S<I> y S<I>
    as x (S<I> 1 2 3)
    (len x)                                // 3
    (cap x)                                // 3
    as y (append x 4 5)
    (len y)                                // 5
    (cap y)                                // 5 (or possibly greater)    
    (len x)                                // 3
    (cap x)                                // 3
```

## pointer operators

`ref` ('reference')

```
struct Cat
    age I
    name Str

func main
    locals a A<I 10> b I c Cat p P<I>
    as p (ref b)                       // pointer to 'b'
    as p (ref a 3)                     // pointer to index 3 of 'a'
    as p (ref c age)                   // pointer to field 'age' of 'c'
```

`dr` ('dereference')

```
func main
    locals i I p P<I>
    as p (ref i)
    as i 5
    (println i)                        // 5                       
```

`sr` ('set at reference')

```
func main
    locals i I p P<I>
    as p (ref i)
    (sr p 5)
    (println i)                        // 5                       
```

## bitwise operators

`band` ('bitwise and')

```
func main
    locals a Byte b Byte c Byte
    as a (Byte 131)      // 1000_0011
    as b (Byte 25)       // 0001_1001
    as c (band a b)      // 0000_0001 (a 1 in any position where both inputs have a 1)
    (println c)          // prints 1
```

`bor` ('bitwise or')

```
func main
    locals a Byte b Byte c Byte
    as a (Byte 131)                // 1000_0011
    as b (Byte 25)                 // 0001_1001
    as c (bor a b)                 // 1001_1011  (a 1 in any position where either--or both--inputs have a 1)
    (println c)                    // prints 155
```

Above, five of the bits had a 1 bit in one or both of the inputs.

`bxor` ('bitwise exclusive or')

```
func main
    locals a Byte b Byte c Byte
    as a (Byte 131)                // 1000_0011
    as b (Byte 25)                 // 0001_1001
    as c (bxor a b)                // 1001_1010  (a 1 in any position where one input--and only one input--has a 1)
    (println c)                    // prints 154
```


`bneg` ('bitwise negate')

```
func main
    locals a Byte b Byte c Byte
    as a (Byte 131)               // 1000_0011
    as b (bneg a)                 // 0111_1100 (flip all the bits of the input)
    (println b)                   // prints 124
```


## input/output operators

`print`

```
func main
    (print 3 "yo" true)                // prints: 3 yo true
    (print (concat 3 "yo" true))       // prints: 3yotrue
```

`println` ('print line')

```
func main
    (println 3 "yo" true)                // prints: 3 yo true (followed by a newline)
    (println (concat 3 "yo" true))       // prints: 3yotrue (followed by a newline)
```

`prompt`

```
func main
    (prompt "Enter your name:")          // prints "Enter your name:", then waits for the user to hit enter
                                         // returns a string of what the user typed before hitting enter
```

`createFile`

Returns an integer which uniquely identifies the open file. (No guarantee is made about what this integer value will be other than it will be unique amongst all the open files.)

```
func main
    locals file I err Str
    as file err (createFile "myFile.txt")
    if (neq err "")
        (println "Error:" err)
        return
```

`openFile`

Opens a file for both reading and writing.



`readFile`

When reading at end of file, returns `0` and `"EOF"` ('end of file').

`writeFile`

`closeFile`

`seekFile`

`seekFileStart`

`seekFileEnd`



