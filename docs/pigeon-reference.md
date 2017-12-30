# DynamicPigeon reference

Experienced programmers can probably glean everything they need to know from this reference, but beginners should watch the [tutorial videos](http://youtube.com/).

## definitions

A DynamicPigeon program consists of definitions at the top-level of code starting with these reserved words:

### `func`

```
// a function named 'foo' that expects two arguments
func foo x y
    // the body has two statements
    (println x y)
    return 3
```

### `global`

```
global alice 0          // a global variable named 'alice' with initial value 0
global bob "hi"         // a global variable named 'bob' with initial value "hi"
```

## data types

DynamicPigeon has these data types:

- booleans (the values `true` and `false`)
- strings (sequences of text characters, denoted by surrounding double quote marks)
- numbers
- lists
- maps
- functions
- the value `nil` (which represents 'nothing')

### lists

```
func main
    locals a b c d
    as a (list 7 -3 "hi")     // assign to 'a' a new list with values 7, -3, "hi"
    (get a 0)                 // 7
    (get a 2)                 // "hi"
    (set a 0 98.6)         
    (get a 0)                 // 98.6
    (len a)                   // 3
    (push a "yo")
    (len a)                   // 4
    (get a 3)                 // "yo"
```

### maps

```
func main
    locals a 
    as a (map 3 "hi" 9 "yo")       // assign to 'a' a new map with two key-value pairs
    (get a 3)                      // "hi"
    (get a 4)                      // nil (the map has no key 4)
    (get a 9)                      // "yo"
    (len a)                        // 2
    (set 8 "aloha")                // add new key-value pair: key 8 with value "aloha"
    (len a)                        // 3
```

### functions

```
func foo a b
    // ... do something

func bar
    // ... do something

func main
    locals x y
    as x foo
    as y bar
    (x 3 "hi")       // calls 'foo' (because 'x' currently references 'foo') 
    (y)              // calls 'bar' (because 'y' currently references 'bar')
    (x)              // runtime error: the function expects two arguemnts
```

## statements

DynamicPigeon has several kinds of statements:

### `locals`

```
func main
    locals x y           // 'main' has two local variables: 'x' and 'y'
    locals z             // compile error: only the first statement of a function can be a 'locals' statement
    (println x)          // prints nil (the default value for a local variable)
```

### `as`

```
func main
    locals x
    as x 3                    // assign 3 to 'x'
    as x "hi"                 // assign "hi" to 'x'
    (println x)               // prints "hi"
```

### `return`

```
func foo a
    if (gt a 10)
        return "hi"           
    return 3                                  
```

### `if`

```
func foo x
    if (eq x 4.3)
        (println "x equals 4.3")
    elif (eq x 1.6)
        (println "x equals 1.689")
    elif (eq x 7.9)
        (println "x equals 7.9")
    else
        (println "x does not equal 4.3, 1.689, or 7.9")
```

### `while`

```
func main
    locals x                  
    as x 0
    // this loop prints: 0 1 2 3  
    while (lt x 4)
        (println x)
        as x (inc x)         // increase value of 'x' by one
```

### `forinc`, `fordec`

```
func main
    // this loop prints: 0 1 2 3
    forinc x 0 4
        (println x)
```

```
func main
    // this loop prints: 3 2 1 0
    fordec x 4 0
        (println x)
```

### `foreach`

```
func main
    locals fruits
    as fruits (list "banana" "apple" "grape" "orange")
    // this loop prints: 0 banana, 1 apple, 2 grape, 3 orange
    foreach i s fruits
        (println i s)
```

### `break`, `continue`

```
func main
    locals fruits
    as fruits (list "banana" "apple" "grape" "orange")
    // this loop prints: banana apple
    foreach i s fruits
        if (eq s "grape")
            break              // jumps execution out of the loop
        (println s)
```

```
func main
    // this loop prints 1 3 5 7 9
    forinc i 0 10
        if (eq 0 (mod i 2))        // if 'i' is an even number
            continue
        (println i)
```


## arithmetic operators

`add` ('addition')

```
func main
    (add 3 5)            // 8
    (add 3 5 -14)        // -6
    (add 3.0 -5.0)       // -2
```

`sub` ('subtraction')

```
func main
    (sub 3 5)            // -2
    (sub 3 5 -14)        // 12
    (sub 3.0 -5.0)       // 8
```

`mul` ('multiplication')

```
func main
    (mul 3 5)            // 15
    (mul 3 5 -14)        // -112
    (mul 3.0 -5.0)       // -15
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
    (eq "hi" 53 53)                // false (not all operands equal)
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
    (charlist "orange")             // (list "o" "r" "a" "n" "g" "e")
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
                                    // (list 111 82 97 110 103 101)
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
    // return value at an index of a list
    (get (list "hi" "yo") 0)                   // "hi" (the value at index 0 of the array)
    // return value of a key of a map
    (get (map "foo" 3 "bar" -2) "bar")         // -2 (the value of the key "bar" in the map)
```

`set`

```
func main 
    locals m
    // set value at an index of a list (returns the value set)
    (set (list "hi" "yo") 0 "bye")             // "bye" (the list now has "bye" at index 0)
    // set value of a key of a map (returns the value set)
    as m (map "foo" 3 "bar" -2)
    (set m "bar" 8)                             // 8 (the map now has value 8 for key "bar")
    (set m "ack" 11)                            // 11 (the map now has value 11 for new key "ack")
```

`push`

```
func main 
    locals a
    as a (list "hi" "yo")
    (len a)                        // 2
    (push a "bonjour")             // returns nothing (adds the string to the end of the list)
    (len a)                        // 3
    (push a "bye" "avast")         // returns nothing (adds the two strings to the end of the list)
    (len a)                        // 5
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
