# Pigeon

The Pigeon language is a reductively simple programming language for educational purposes. It includes the features common to all popular languages:

## common elements of high-level languages

All commonly used high-level languages have these features:

 - data types
 - operators
 - declaration statements
 - assignment statements
 - branches (`if` statements)
 - loops (usually called `while` statements)
 - functions

In any code, we deal with data, and data comes in different types. Every language has a set of built-in ***data types***, such as types for representing numbers and strings (pieces of text data). 

With our data, we want to perform *operations*, such as adding and subtracting numbers, so every language has a few dozen ***operators*** for use with the language's built-in data types. An operation takes in one or more input values and produces a single output value. For example, an addition operation takes two numbers as input and produces the number which is the sum of the two inputs. Every language contains the arithmetic operators familiar from math, but most languages also include some operators not familiar from math.

To retain values in our code, we need to store them in memory. A ***variable*** is a symbolic name that represents a location in memory that stores a single value. Be clear that what we call variables in mathematics are not exactly the same thing: in a mathematical equation, like `y = 2x`, the variables in a sense represent all possible values at once; in code, we can overwrite the value stored by a variable with a new value, but at any one moment, a variable in code stores just one value.

Code in most languages is written as a series of ***statements***, and these statements are executed one after the other, first-to-last.

A *declaration* statement creates a variable. An *assignment* statement stores a value in a variable.

An `if` statement has a condition and contains other statements. The condition is something like 'is variable x greater than the number 5?' or 'is variable y equal to the number 8?'. The contained statements can be any kind, even other `if` statements. When an `if` is executed, its condition is tested, and if true, the contained statements are executed in order; if the condition tests false, the contained statements are skipped over. Either way, execution continues on to the next statement after the `if`.

A `while` statement is just like an `if`, but with one difference: after the condition tests true and the contained statements are executed, the condition is tested again. If the condition is true again, the contained statements are executed another time. This repeats indefinitely until the condition tests false, in which case execution continues on to the next statement after the `while`.

A ***function*** is a series of statements that we give a name, such that we can run that series of statements in other parts of code by just referring to the name. A function can also receive input values and produce an output value. A function, in a sense, is like an operator created by the programmer.

## comments

It's sometimes useful to leave notes in code, and so we need some way of telling the compiler or interpreter to ignore a chunk of text. These ignored chunks of text are called ***comments***. In Pigeon, a comment starts with // and includes all text through the rest of that line:

```
// These two slashes and everything after them on the line are ignored by the language
```

## data types

Numbers are written as you would expect:

```
3      // integer
-7.3   // floating-point
```

A boolean is a data type with just two values:

```
true
false
```

Much like a bit, what a boolean represents depends entirely upon context.

Strings (pieces of text) are written enclosed in double-quote marks:

```
"hi there"    //string with the characters: h i space t h e r e
```

Because double-quote marks are used to denote the end of the string, you must write `\"` to include a double-quote mark in the string:

```
"foo\"bar"     // string with the characters: f o o " b a r
```

A string cannot span multiple lines. To include a newline in a string, you must write `\n`. The actual ASCII/Unicode character(s) this denotes depends upon the platform: on Windows, it denotes CR and LF (carriage return, line feed); on Linux and Mac, it denotes just LF (line feed).

```
"foo\nbar"     // string with the characters: f o o newline b a r
```

Because backslash is used to denote certain special characters, you must write `\\` to include a backslash in the string:

```
"foo\\bar"     // string with the characters: f o o \ b a r
```

## arithmetic operators

Every operation in Pigeon is written as a pair of parentheses containing the name of the operator followed by the operands (the inputs). The arithmetic operators are `add` (addition), `sub` (subtraction), `mul` (multiplication), and `div` (division):

```
(add 3 5)      //  add 3 to 5, producing 8
(sub 3 5)      //  subtract 5 from 3, producing -2
(mul 3 5)      //  multiply 3 and 5, producing 15
(div 3 5)      //  divide 3 by 5, producing 0.6
```

Pigeon is a high-level language that spares us from having to think about the bit-representation of numbers. The math operators always return the mathmatically correct answer, and you don't have to worry about overflow or underflow. (This convenience does come at significant efficiency costs, but for many tasks, the costs won't really matter.)

The arithmetic operators expect only numbers as inputs. The wrong type of input will trigger an error that aborts the program:

```
(sub 3 "hi")    // error: cannot subtract a string
```

An operation can itself be used as an operand, in which case the value it produces is used as input to the containing operation:

```
(mul 2 (add 3 5))    // multiply 2 with the result of (add 3 5), producing 16
```

The `mod` (modulus) operator returns the remainder of division:

```
(mod 11 3)     // produces 2  (11 divided by 3 has remainder 2)
(mod 12 3)     // produces 0  (12 divided by 3 has no remainder)
```

### equality and logic operators

The `eq` (equality) operator produces the boolean value `true` if its operands are equal; otherwise, it produces `false`:

```
(eq 2 2 2)        // all operands are equal, produces true
(eq 3 8 3)        // not all operands are equal, produces false
(eq "hi" "hi")    // produces true
(eq "hi" "bye")   // produces false
```

The `not` operator takes a single boolean operand and returns the opposite:

```
(not false)       // produces true
(not true)        // produces false
```

The `neq` (not equal) operator is a convenient way to combine `not` and `eq` in one operation:

```
(not (eq 2 2 2))   // produces false
(neq 2 2 2)        // produces false
``` 

The `and` operator returns true only if every operand is true:

```
(and true true true)     // true
(and true false true)    // false
(and false false false)  // false
```

The `or` operator returns true if any operand is true:

```
(or true true true)     // true
(or true false true)    // true
(or false false false)  // false (no operands are true)
```

# relational operators

The `gt` (greater than) operator returns true only if every operand is greater than the operand to its right:

```
(gt 8 5 2)         // true
(gt 8 5 6)         // false
(gt 8 8 2)         // false (8 is not greater than 8)
```

The `gte` (greater than or equal) operator returns true only if every operand is greater than or equal to the operand to its right:

```
(gt 8 5 2)         // true
(gt 8 5 6)         // false
(gt 8 8 2)         // true
```

The `lt` (less than) operator returns true only if every operand is less than the operand to its right:

```
(lt 2 5 8)         // true
(lt 7 5 8)         // false
(lt 7 7 8)         // false (7 is not less than 7)
```

The `lte` (less than or equal) operator returns true only if every operand is less than or equal to the operand to its right:

```
(lte 2 5 8)         // true
(lte 7 5 8)         // false
(lte 7 7 8)         // true
```

## concat

The `concat` (concatenate) operator produces a string that is the concatenation of the operands:

```
(concat "FOO" "BAR")                         // produces "FOOBAR"
(concat "rubber " "baby" " buggy bumper")    // produces "rubber baby buggy bumper"
```

## input/output

Performing input and output ultimately requires [system calls](https://en.wikipedia.org/wiki/System_call). Because Pigeon is a simple, educational language, it has no facility for performing system calls, but it does provide two operators for doing very basic input/output.

The `print` operator displays text on screen:

```
(print "hello")    // display "hello" on screen
(print 35)         // display "35" on screen
```

The `prompt` operator takes no operands. It prompts the user on screen to type something and hit enter. The text entered by the user is returned as a string:

```
(print (prompt))       // wait for the user to type something and hit enter; display what they typed
```

(Note that, unlike every other operation, `prompt` waits for user action before execution continues.)

# functions

A Pigeon program is primarily composed of functions. As discussed, a function is a chunk of statements that is given a name such that we can execute the chunk anywhere else in code by just writing the name.

A function definition in Pigeon starts with the reserved word `func`, then the name you've chosen for the function. The body (the statements to execute when the function is called) are written indented on the next lines. For example:

```
// a function named 'david' with a body of two statements
func david 
    (print (add 3 5))    // display result of adding 3 to 5
    (print "hi")         // display "hi"
```

To execute a function---*a.k.a.* 'call' *a.k.a.* 'invoke' a function---we use parens around the function name, as if it were an operator:

```
func heidi
    (david)              // call david
    (print "yo")
```

## the main function

Execution of a Pigeon program begins by calling the function named `main`:

```
func jill
    (print "hi")

func main 
    // execution begins here
    (jill)
    (ted)

func ted
    (print "bye")
```

## `return` statements

A function call returns a value. When encountered, a `return` statement ends the current function call and specifies the value to return:

```
func karen
    (print "hi") 
    return 9

func main
    (print (karen))            // print "hi", then 9
```

A function with no `return` statement at the end implicitly ends with a `return` statement returning the special value `nil` (which represents 'nothing').

```
func jack
    (print "hi")

func erin
    (print "yo")
    return null         

func main
    (print (jack))            // print "hi", then nil
    (print (erin))            // print "yo", then nil
```

## parameters and arguments

A function may have ***parameter variables***. Parameters are denoted as names listed after the function name. When called, an ***argument*** (a value) must be provided for each parameter:

```
// function chris has two parameters: orange and banana
func chris orange banana
    (print banana orange)

func main
    // in this call to chris, orange will have the value 1, and banana will have the value 2
    (chris 1 2)         // print 2 1
```

Note that the arguments to a function call are assigned to the corresponding parameter variables in the same position: the first argument is assigned to the first parameter, the second argument is assigned to the second parameter, *etc.*

```
// function jane has two parameters: a and b
func jane a b
    return (mul a (add 2 b))

func main
    (print (jane 2 7))    // print 18
    (print (jane 4 6))    // print 32
```

## variables

A variable is a symbolic name representing a location in memory that stores a value.

In each function call, a parameter gets its initial value from the corresponding argument to the call, but we can change a variable's value with an assignment statement. An assignment statement starts with the word `as` followed by a variable name and a value to assign to the variable:

```
func ian x y
    as x 8                // assign the value 8 to x
    return (add x y)

func main
    (print (ian 4 7))     // print 15
```

## `locals` statements

In addition to its parameters, a function can have other variables that start with the initial value `nil`. These variables are created with a `locals` statement (which must be the first statement of the function):

```
func john a b
    locals c d             // create variables c and d, both with the initial value nil
    (print c)              // print nil
    as c (add a b)         // assign result of adding a and b to c
    as d (mul c 5)         // assign result of multiplying c and 5 to d
    return d               // return the value of d

func main
    (print (john 2 4))     // print 30
```

The variables of a function do not exist in other functions:

```
func lisa x
    return (add x y)        // illegal! no variable y exists in lisa

func main
    locals y
    as y 8
    (lisa)
```

If two functions have a variable of the same name, they are wholly separate variables that just happen to have the same name:

```
func lisa x              // this x belongs to lisa
    return (add x 7)

func main
    locals x                 // this totally separate x belongs to main
    as x 3
    (print (lisa x))
```

Each call to a function creates its own set of the local variables, and each set disappears when its call ends. So, say, local variable 'x' in one call is separate from local variable 'x' in another call to that same function.

(Variables in Pigeon do not actually store values *directly*. Instead, assignment stores the *address* of the value. This distinction will be significant when we deal with lists and maps later.)

## global variables

A variable created outside any function is ***global*** (rather than ***local***), meaning it exists for the whole duration of the program and is accessible anywhere in code.

We can create a global variable with a `global` statement:

```
global x 6                   // create global x with initial value 6

func lisa x              // this x belongs to lisa
    return (add x 7)

func main
    (print x)                // print 6
    as x 10                  // assign 10 to global x
    (print (lisa x))         // print 17
```

Note that, because function 'lisa' has its own 'x', it cannot use the global 'x'. If we need to use global 'x' in 'lisa', we must rename either variable to avoid the name conflict.

## expressions

An ***expression*** is anything which *evaluates* into a value:

 - a literal evaluates into the value it represents
 - a variable evaluates into the value it references at that moment in time
 - an operation evaluates into the value which it returns
 - a function call evaluates into the value which it returns

For example, these are all expressions:

```
3                     # evaluates into the number 3
"yo"                  # evaluates into the string "yo"
foo                   # evaluates into whatever variable 'foo' references at this moment
(add 4 2)             # evaluates into the number 6
(mul 2 (sub 9 1))     # evaluates into the number 16
```

(Note that operations are expressions built out of other expressions.)

## reserved words

A ***reserved word*** is any word given special significance in the language. Most of the reserved words in Pigeon are the operators (`add`, `sub`, `mul`, *etc.*), and three others are the values `true`, `false`, and `null`. The several remaining reserved words each have their own particular meaning and syntax, *e.g.* `if`, `while`, `return`.

You cannot create variables with reserved word names:

```
as sub 3      # error
```

## `if` and `while` statements

An `if` statement in Pigeon begins with the word `if` followed by a condition and a body.

The condition is an expression (a value, variable, or operation) which must return a boolean value. The body is one or more indented statements.

```
func aaron x
    // if 'x' is not greater than 5, the condition is false, and so the 
    // two indented print statements are skipped over
    if (gt x 5)
        (print "hi")
        (print "bye")
    (print "yo")     // this statement is not part of the 'if' body
```

(The lines in the body must be indented by the same number of spaces and tabs. It's generally best to indent by a single tab or by 4 spaces.)

Because an `if` is itself a kind of statement, an `if` can be nested inside the body of another `if`:

```
func aaron x
    // this outer 'if' has a body of three statements: a print, an if, and another print
    if (gt x 5)
        (print "hi")
        // this inner 'if' has a body of one statement: a print
        if (lt x 10)
            # execution only reaches here if x is greater than 5 and less than 10
            (print "ahoy")
        (print "bye")
```

A `while` statement is written just like `if` but starts with the word `while`:

```
// this program prints: 0 1 2 3 4 done
func main
    locals x
    as x 0
    while (lt x 5)
        (print x)
        as x (add x 1)   // increase the value of 'x' by 1
    (print "done")
```

Again, `if` and `while` are kinds of statements and their bodies are composed of statements: an `if` body can contain other `if`'s and `while`'s; a `while` body can contain other `while`'s and `if`'s.

## `else` and `elseif` clauses

Very often we want to branch between two mutually exclusive cases, meaning we want to do one thing or the other but never both. We can arrange this with two successive `if` statements with logically inverse conditions:

```
func aaron x
    if (gt x 3)
        (print "hi")
    if (not (gt x 3))
        (print "bye")
```

Above, if *x* is greater than 3, we'll print "hi", otherwise we'll print "bye". Always one of the bodies gets executed, but never both.

Alternatively, we can immediately follow an `if` statement with an accompanying `else` clause, which has its own body and executes only when the condition tested false. This code is functionally equivalent to the above:

```
func aaron x
    if (gt x 3)
        (print "hi")
    else            
        (print "bye")
```

Sometimes we wish to branch between *more than two* mutually exclusive cases. We can arrange this by nesting `if`s inside a waterfall of `else` clauses:

```
func aaron x
    if (eq x 3) 
        (print "cat")
    else
        if (eq x 5)
            (print "dog")
        else
            if (eq x 9)
                (print "bird")
```

Above, only of the *print* operations executes, depending on the value of *x*. (The value of *x* might not equal 3, 5, or 9, in which case none of *print* operations execute.)

To express the above less verbosely, we can use `elseif` clauses:

```
func aaron x
    if (eq x 3) 
        (print "cat")
    elseif (eq x 5)
        (print "dog")
    elseif (eq x 9)
        (print "bird")
```

The conditions are tested in order: when a condition tests true, its body executes, and all the others are skipped over. *Only one body ever runs.* 

We can put an `else` clause at the end, whose body will execute when none of the conditions test true:

```
func aaron x
    if (eq x 3) 
        (print "cat")
    elseif (eq x 5)
        (print "dog")
    elseif (eq x 9)
        (print "bird")
    else 
        (print "moose")
```


## lists

A ***list*** in Pigeon is a value which itself is made up of any number of other values. The values in a list are known by their numeric indexes. The first element has index 0, the second has index 1, the third has index 2, the fourth has index 3, *etc.*

The `list` operator returns a new list made up of the operands:

```
function main
    locals x
    (print (list))                  // print a list with no values
    (print (list 6 "hi" 78))        // print a list with three values: 6, "hi", 78
```

The `get` operator retrieves the value of a list at a given index:

```
function main
    locals x
    as x (list 6 "hi" 78)
    (print (get foo 2))         // print 78 (the value at index 2 of the list stored in variable x)
```

The `set` operator modifies the value of a list at a given index:

```
function main
    locals x
    as foo (list 6 "hi" 78)
    (print (get foo 0))          // print 6
    (set foo 0 "bye")    
    (print (get foo 0))          // print "bye"
```

The `len` (length) operator returns the number of values within a list:

```
function main
    locals x y
    as x (list)              
    as y (list 6 "hi" 78)    
    (print (len x))              // print 0
    (print (len y))              // print 3
```

The `append` operator adds a value to the end of a list, increasing the list's length by one:

```
function main
    locals x
    as x (list 6 "hi" 78)
    (append foo 900)
    (print (get foo 3))                // print 900
    (print (len foo))                  // print 4
```

## references vs. values

You may have wondered how a single variable can store any kind of value. Strings, numbers, booleans, and lists all take up different amounts of memory, so how can one location in memory store any kind of thing? Well, a variable in Pigeon does not actually store a value directly. Instead, a variable stores a memory address, a ***reference*** to where an actual value is located elsewhere in memory. 

Likewise, a list stores references rather than directly store actual values.

When multiple references refer to the same unmodifiable value, there's nothing to worry about. When multiple references refer to a modifiable value, however, it's important to be aware that changes *via* one reference will show up when we read the value *via* another reference.

Nearly all operations in Pigeon leave their input values unmodified. The `add` operator, for example, produces a new, separate output value, leaving its input values unmolested. The only two exceptions are `set` and `append`, which modify their list input. A list is the only kind of value in Pigeon that can be modified:

```
function main
    locals x y
    as x (list 6 "hi" 78)
    as y x                     // y will now reference the same list as x
    (set y 1 "yo")             // modify index 1 of the list referenced by y
    (print (get x 1))          // print "yo"
```

Again, a list actually stores references, not values directly, so it's possible for multiple list indexes to reference the same value:

```
function main
    locals x y
    as x (list 6 "hi" 78)
    as y (list)
    (append y x)                
    (append y "yo")
    (append y x)                
    (print (get y 0))           // print the same list referenced by x
    (print (get y 1))           // print "yo"
    (print (get y 2))           // print the same list referenced by x
```

It's even possible for a list index to reference the same list that contains it, *i.e.* a list can contain itself!

```
function main
    locals foo
    as foo (list 6 "hi" 78)
    (append foo foo)
    (print (get foo 3))           // print the same list referenced by foo
```

## maps

A ***map*** in Pigeon is a value which itself is made up of ***key-value pairs***. Each key-value pair consists of a key (a string or number) and its associated value (which may be of any type). Each key in a map must be unique amongst the other keys because it is the key which uniquely identifies each key-value pair.

The `map` operator creates a new map. The `set` operator creates or modifies a key-value pair in the map. The `get` operator retrieves the value associated with a given key. The `len` operator returns the number of key-value pairs in a map:

```
function main
    locals x 
    as x (map)                // assign a new, empty map to x
    (set x "hi" 6)
    (set x "bye" "yo")
    (print (get x "hi"))      // prints 6
    (print (len x))           // prints 2
```

We can specify key-value pairs when we create a map:

```
function main
    locals x
    // a new map with two key-value pairs: key "hi" with value 6 and key "bye" with value "yo"
    as x (map "hi" 6 "bye" "yo")
```

Like a list, a map is mutable, so if two variables reference the same map, changes *via* either variable affect the same map.

## special `[]` syntax for `get` and `set`

Because getting and setting from lists and maps is particularly common, we have shorthand syntax for `get` and `set`:

```
function main
    locals x
    as x (list 86 "hi")
    (print [x 0])               // (print (get x 0)) 
    as [x 1] "bye"              // (set x 1 "bye")
    (print [x (sub 9 8)])       // (print (get x (sub 9 8))) 
```

We can use any expression in the brackets to specify the index/key.

Notice that an assignment statement is used in place of a `set` operation. (This is not really any more compact than just using `set`, but we introduce it here because this mirrors how `set` is expressed in most other languages.)






