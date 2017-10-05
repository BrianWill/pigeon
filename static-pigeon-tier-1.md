# StaticPigeon (tier 1)

## static typing vs. dynamic typing

While DynamicPigeon is a dynamically-typed language, StaticPigeon is a statically-typed language 

In a statically-typed language, each variable (including each function parameter) is marked by a designated type such that only values of the designated type can be assigned to the variable. Functions are also marked by a 'return type', such that you must always return values of that type (and only that type) from the function.

The compiler will refuse to compile the code if you:

 - use the wrong type of operands in an operation
 - assign the wrong type of value to a variable
 - pass the wrong type of argument to a function
 - return the wrong type of value from a function

In a dynamically typed language, the code will compile and execute regardless of such problems. However, when an operation in a dynamic language is executed with the wrong type(s) of operands, an error occurs, aborting execution.

Static typing has the advantage of detecting all ***type errors*** at compile time, before the code even runs. With dynamic typing, a type error may lurk undetected in some uncommonly executed branch of code. On the other hand, static typing can require more thinking about types up front, which may feel onerous or inhibiting. Some programmers prefer static typing; others prefer dynamic typing.

Here's an example function in StaticPigeon:

```
// function amber has two parameters, x (a string) and y (a boolean), and returns a boolean
func amber x Str y Bool : Bool 
    (println x y)
    return (not y)

func main
    locals answer Bool                // local variable answer is a Boolean
    (println answer)                  // print "false"
    as answer (amber "hi" false)      // OK
    (println answer)                  // print "true"
    (amber 4 true)                    // compile error: first argument to amber must be a string
```

Note a few things:

- all data type names in StaticPigeon start with uppercase letters, and all other names begin with lowercase letters
- each parameter is followed by its type
- after a function's parameters, a colon precedes the function's return type
- the colon may be omited if a function returns nothing (as in the case of main)
- because amber is declared to return a boolean, it must return a boolean
- like function parameters, each variable declared in `locals` is followed by its type
- a variable starts out with the default value for its type (the default boolean value is false)

## number types

In DynamicPigeon, all numbers are 64-bit floating-point. In StaticPigeon, the 64-bit floating-point number type is called `F`, and we also have a 64-bit integer type called `I`. Though both are numbers, the compiler considers them to be different things. Number literals with a decimal point are considered floats, and number literals without are considered integers. The arithmetic operators work on both floats and integers, but a single operation can only have operands of one type:

```
func main
    locals i I f F         // i is an integer, f is a float
    as i 5                 // OK
    as f -2.7              // OK
    as i 5.0               // compile error: 5.0 is considered a float
    as f -2                // compile error: -2 is considered an integer
    as i (add i 7)         // OK
    as i (add i f)         // compile error: cannot add an integer and float together
    as f (add f 8.2)       // OK
```

## multi-return functions and multiple assignment

A function in StaticPigeon may be declared to return multiple values. The values returned from such a function can only be received in an assignment statement with multiple targets:

```
// zelda returns both an integer and a string
func zelda : I Str
    return 3 "hi"

func main
    locals x I y Str
    as x y (zelda)              // assign 3 to x, "hi" to y
    (zelda)                     // OK
    as x (add (zelda) 3)        // compile error: zelda cannot be called where only one value is expected
``` 

## structs

In StaticPigeon, we can define our own data types called `structs` (as in 'structures'). Structs are defined at the top-level of code (meaning outside any function). A struct is a composite of one or more named elements of data, called 'fields':

```
// define a struct called Ronald with two fields
struct Ronald
    foo Str          // a field of type Str called foo
    bar F            // a field of type F called bar
```

Having defined a struct, we can create values of that type by using its name like an operator and supplying a value for each field (in the order they are defined). We can access the fields of a struct value with the dot operator: 

```
func main
    locals r Ronald
    as r (Ronald "hi" 4.6)     // assign to r a Ronald value where foo is "hi" and bar is 4.6
    (println r.foo)            // print "hi"
    as r.bar 8.1               // assign 8.1 to field bar of the variable r
```

## methods

A method is a special kind of function in which the first parameter must be a struct of some kind, and the method is said to belong to that struct type. Methods are called like functions, but with a dot before the method name:

```
struct Cat
    age I
    weight F
    name Str

method eat c Cat food F : F
    as c.weight (add c.weight food)
    return c.weight

func main
    locals c Cat
    as c (Cat 10 8.9 "Mittens")
    (.eat c 0.7)
    (println c.weight)                  // 9.6
```

Whereas we can only have one function of a particular name, multiple structs can all have methods with the same name, *e.g.* a *Dog* struct could also have a method *eat*.

By themselves, methods are simply a minor stylistic alternative to functions, but *interfaces* (discussed next) make them more consequential.

## interfaces

An interface specifies a set of method names, along with parameter lists and return types for the named methods:

```
interface Jack
    foo I Str : Str         // method named foo; takes an int and a string; returns a string
    bar                     // method named bar; takes no arguments; returns nothing
```

Any struct which has all the methods specified in an interface is considered to *implement* that interface:

```
// struct Cat implements Jack
method foo c Cat i I s Str : Str
    // ... do stuff

method bar c Cat
    // ... do stuff
```

If a struct has additional methods not included in an interface, that doesn’t affect whether the type implements the interface. A single struct can implement any number of interfaces. Implementing one interface does not affect whether it implements another.

We can cast a value to an interface type if the value’s type implements the interface. Assuming Cat implements Jack, we can cast (convert) a Cat value to a Jack value. When we do such a cast, the returned interface value is made up of two references:

- a reference to the value of the implementing struct
- a reference to the implementing struct's method table.

When we assign a value of an implementing struct to an interface variable, it is implicitly cast to a value of that interface type:

```
func main
    locals j Jack c Cat
    as j (Jack c)            // create a Jack value referencing the Cat value and the Cat method table
    as j c                   // same as previous statement, but cast left implicit
    (.bar j)                 // calls method bar of Cat
```

The default value of an interface variable is made up of two references to nothing. We can assign an interface variable its default value by assigning it nil:

```
func main
    locals j Jack c Cat
    as j nil                 // create a Jack value referencing the Cat value and the Cat method table
    (.bar j)                 // panic (runtime error)
```

Calling methods on a nil interface value triggers a panic: without a referenced value, there is no referenced method, and so no actual method to call!

## typeswitch

Given an interface value, we can use a `typeswitch` to branch on its referenced value’s concrete type. A typeswitch has one or more clauses, and only the matching clause (if any) executes. Here, this function takes an interface value, but what the function does depends upon the concrete type referenced by the interface value:

```
// assume Jack is an interface type with implementors Cat, Dog, Bird, and others
func alice j Jack
    typeswitch v j
    case Cat
        // ... clause executed if j holds a Cat; v in this clause is a Cat value
    case Dog
        // ... clause executed if j holds a Dog; v in this clause is a Dog value
    case Bird
        // ... clause executed if j holds a Bird; v in this clause is a Bird value
    default
        // ... clause executed if j holds neither a Cat, Dog, nor Bird; v in this clause is a Jack value
```

Including a default clause is optional.

## foreach loops

A foreach loop makes it convenient to loop through the elements of a map, list, array, or slice (we'll introduce arrays and slices later). We specify two variables which will exist only in the body of the foreach: the first stores the index/key, the second stores the value:

```
func main
    locals x L<I>
    as x (L<I> 6 2 14)
    // prints 0 6, then 1 2, then 2 14
    foreach i I v V x
        (println i v)    
```

Because maps have no sense of order, no guarantee is made about the order in which foreach will iterate through the key-value pairs:

```
func main
    locals x M<Str I> s Str
    as x (M<Str I> "hi" 3 "yo" 87)
    // prints (but not necessarily in this order): "hi" 3, then "yo" 87
    foreach s Str v V x
        (println s v)    
```

## pointers

A pointer represents a reference, *i.e.* a memory address. There is no single pointer type: rather, there is a pointer tyep for every other type in the language. An int pointer represents a memory address where an int is stored; a string pointer represents a memory address where a string is stored; *etc.* A pointer type is denoted as `P<X>`, where X is the type of pointer, *e.g.* `P<I>` is an int pointer.

The pointed to location can only be a variable, a field within a struct variable, or an index within an array or slice (discussed later).

The `ref` ('reference') operator creates a pointer to a given location. The `dr` ('dereference') operator returns the value at a location represented by a pointer:

```
func main
    locals i I p P<I>
    as p (ref i)             // assign to p a pointer to the location of variable i
    as i 4
    (print (dr p))           // print 4 (the value stored in i)
```

(Be clear that, whereas operators normally take values as operands, the `ref` operator takes a storage location as operand. In the expression `(ref i)`, the operand is the variable i itself, not the value stored in i. The `ref` operator doesn’t care what value is stored at that location: it just wants the address.)

The default value of a pointer variable represents the address of nothing and is represented by the reserved word nil. Dereferencing a nil pointer triggers a panic:

```
func main
    locals p P<F>          // p starts out nil
    (print (dr p))         // panic (runtime error): cannot dereference a nil pointer
```

So why might we want to use pointers? Three reasons:

1) By storing a pointer instead of a value directly, that value can be referenced from multiple places, e.g. pointers A, B, and C can all point to the same value D in memory:

```
func main
    locals i I p1 P<I> p2 P<I> p3 P<I>
    as p1 (ref i)
    as p2 (ref i)
    as p3 (ref i)
    as i 9                 // one assignment affects all three pointers
    (println (dr p1))      // print 9
    (println (dr p2))      // print 9
    (println (dr p3))      // print 9
```

Sharing data this way is sometimes useful because changes to the referenced value are seen everywhere it is referenced. (Shared data can also cause problems if you’re not careful! It’s sometimes easy to forget all the places in code that a piece of data is shared.)

2) By passing a pointer to a function, the function can store a new value at the referenced location:

```
func foo a P<I>
    as (dr a) 11

func main
    locals z I
    as z 5
    (foo (ref z))
    (println z)          // prints 11
```

(Be clear that a function’s parameters are always local to the function call, and so assigning to a pointer parameter is a change seen only within the call. However, if we assign to the dereference of a pointer parameter, the change may be seen outside the call.)

3) Structs and arrays come in all sizes: a few bytes up to thousands or even occasionally millions or billions of bytes. No matter what a pointer points to, a pointer value is always just an address, and all addressses within a single system are the same size. On a 32-bit machine, addresses are 32 bits; on a 64-bit machine, addresses are 64 bits. Therefore, it is often more efficient for functions to have pointer parameters instead of struct or array parameters. A function call argument value is always copied in full to its corresponding parameter: for a large struct or array, that can be a lot of bytes to copy; for a pointer, it’s always just 32 or 64 bits.



## break and continue statements





## imports

We can split the code of our program into multiple source files, called *packages*. To use the code from one package in another, we import the other package, specifying which elements (functions, structs, interfaces, globals) we want to make accessible in this package:

```
// import the package at this path, bringing in mike, Bob, and freida
import "foo/bar/other.spigeon"
    mike
    Bob
    frieda
```

StaticPigeon files must end with the extension `.spigeon`.

Imports must be written at the top-level of code before any other elements.

If the name of an imported element is the same as that of something already defined or imported into the package, we can resolve the name conflict by giving the imported element an alias:

```
import "foo/bar/other.spigeon"
    mike
    Bob
    frieda gina        // import frieda with the alias gina
```

In this package, what is known as *frieda* in *other.spigeon* is known here as *gina*.

## the standard library file package

#### struct File

#### func createFile

#### func openFile

#### func readFile

#### func writeFile

#### func closeFile

#### func seekFile

#### func seekFileStart

#### func seekFileEnd
