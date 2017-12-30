# GoPigeon (notes for new programmers)

## static typing vs. dynamic typing

While DynamicPigeon is a dynamically-typed language, GoPigeon is a statically-typed language 

In a statically-typed language, each variable (including each function parameter) is marked by a designated type such that only values of the designated type can be assigned to the variable. Functions are also marked by a 'return type', such that you must always return values of that type (and only that type) from the function.

The compiler will refuse to compile the code if you:

 - use the wrong type of operands in an operation
 - assign the wrong type of value to a variable
 - pass the wrong type of argument to a function
 - return the wrong type of value from a function

In a dynamically typed language, the code will compile and execute regardless of such problems. However, when an operation in a dynamic language is executed with the wrong type(s) of operands, an error occurs, aborting execution.

Static typing has the advantage of detecting all ***type errors*** at compile time, before the code even runs. With dynamic typing, a type error may lurk undetected in some uncommonly executed branch of code. On the other hand, static typing can require more thinking about types up front, which may feel onerous or inhibiting. Some programmers prefer static typing; others prefer dynamic typing.

Here's an example function in GoPigeon:

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

- all data type names in GoPigeon start with uppercase letters, and all other names begin with lowercase letters
- each parameter is followed by its type
- after a function's parameters, a colon precedes the function's return type
- the colon may be omited if a function returns nothing (as in the case of main)
- because amber is declared to return a boolean, it must return a boolean
- like function parameters, each variable declared in `locals` is followed by its type
- a variable starts out with the default value for its type (the default boolean value is `false`)



## number types

In DynamicPigeon, all numbers are 64-bit floating-point. In GoPigeon, the 64-bit floating-point number type is called `F`, and we also have a 64-bit integer type called `I`. Though both are numbers, the compiler considers them to be different things. Number literals with a decimal point are considered floats, and number literals without are considered integers. The arithmetic operators work on both floats and integers, but a single operation can only have operands of one type:

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

A function in GoPigeon may be declared to return multiple values. The values returned from such a function can only be received in an assignment statement with multiple targets:

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

In GoPigeon, we can define our own data types called `structs` (as in 'structures'). Structs are defined at the top-level of code (meaning outside any function). A struct is a composite of one or more named elements of data, called 'fields':

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



## arrays

Like a list, an array is a value made up of multiple values of the same type. The difference is that arrays are fixed in size, and in fact the size is integral to its type:

```
func main
    locals x A<I 3>                // x is an array of 3 integers
        ,y A<I 5>                  // y is an array of 5 integers
        ,z A<I 5>                  // z is an array of 5 integers
    as x y                         // compile error: x and y are not the same type of array
    as y z                         // OK
```

Each element of the array is known by its numeric index. The first element is at index 0, the second at index 1, etc. The last element’s index is effectively always one less than the length of the array:

```
func main
    locals x A<I 3>    
    (set x 2 57)                   // set index 2 of x to 57
```

When we create an array, the size must be a constant expression (meaning the expression can’t include variables or function calls).

Accessing an index out of bounds with a constant expression triggers a compile error. Accessing an index out of bounds with a runtime expression triggers a panic (a runtime error, discussed later):

```
func main
    locals x A<I 5>
    (set x 26 9)           // compile error: index out of bounds
```

Assigning one array to another copies all the elements by their respective indexes. An array variable can only be assigned arrays of the same type and size:

```
func main
    locals x A<I 3>
        ,y A<I 3>
        ,z A<I 8>
    as x y                  // assign index 0 of y to index 0 of x, index 1 of y to index 1 of x, etc.
    as y z                  // compile error: cannot assign an A<I 8> to an A<I 3>
```

We can also compare arrays of the same type and size with `eq`. The equality test returns true if all of the respective elements are equal:

```
func main
    locals x A<I 3>
        ,y A<I 3>
    (println (eq x y))       // print true (all elements of both arrays are currently 0)
```

We can create an array value by using the array type like an operator:

```
func main
    locals x A<I 3>
    (println (get x 1))       // print 0
    as x (A<I 3> 5 -24 10)
    (println (get x 1))       // print -24
```

Functions can take arrays as inputs and return arrays as output:

```
// returns the sum of all values in the array
func sum nums A<I 10> : I
    locals val I
    foreach i I v I nums
        as val (add val v)
    return val

func main
    locals arr A<I 10>
    as arr (A<I 10> 1 2 3 4 5 6 7 8 10)
    (println (sum arr))                    // print 55
```

Be clear that when sum is called, the whole array argument is copied to the array parameter. The argument variable and parameter variable are separate arrays, each made up of 10 int values.

## slices

A slice value represents a subsection of an array. Each slice value has three components: a reference to an element within an array, a length (a number of elements), and a capacity (the count of elements from the referenced element through the end of the array.)

Given an array, we get a slice value representing a subsection of the array using the `slice` operator, and we use `get` to access the values of the array subsection that it represents:

```
func main
    locals arr A<I 10> s S<I>
    as arr (A<I 10> 10 20 30 40 50 60 70 80 80 100)
    as s (slice arr 3 7)    // slice referencing index 3 of the array, with 
                            // length 4 (because 7 - 3 is 4) and capacity 7 (because 10 - 3 is 7)
    (set s 1 -999)    
    (print (get s 0))       // 40
    (print (get s 1))       // -999
    (print (get s 2))       // 60
    (print (get s 3))       // 70
    (print (get s 4))       // panic! out of bounds (index must be less than length)
```

In effect, a slice represents length-number of elements starting from the referenced element. (The capacity is needed for the `append` operator, discussed shortly.)

```
func main
    locals arr A<I 10> s S<I>
    as arr (A<I 10> 10 20 30 40 50 60 70 80 80 100)
    as s (slice arr 3 7)
    (set s 1 -999)    
    (print (get s 0))       // -999
    (print (get s 1))       // 50
    (print (get s 2))       // 60
    (print (get s 3))       // 70
    (print (get s 4))       // panic! out of bounds (index must be less than length)
```

It’s perfectly possible for a slice to start at the beginning of an array. In fact, a slice can represent the whole of an array:

```
func main
    locals arr A<I 10> s S<I> s2 S<I>
    as arr (A<I 10> 10 20 30 40 50 60 70 80 80 100)
    as s (slice arr 0 7)           // slice referencing index 0 of the array, with length 7 and capacity 10
    as s2 (slice arr 0 10)         // slice referencing index 0 of the array, with length 10 and capacity 10
```

Note that multiple slice values can represent overlapping subsections of the same array. Consequently, changes *via* one slice can affect other slices:

```
func main
    locals arr A<I 10> s S<I> s2 S<I>
    as arr (A<I 10> 10 20 30 40 50 60 70 80 80 100)
    as s (slice arr 4 9)
    as s2 (slice arr 8 10)
    (set s 4 -999)
    (println (get s2 0))          // -999
```

Note that slices are typed, *e.g.* an integer slice is different from a boolean slice which is different from a string slice, *etc.* The length and capacity of a slice is not part of its type, so we can assign a slice of any length or capacity to a slice variable.

We can create a slice with a new underlying array by using the slice type as an operator:

```
func main
    locals s S<I>
    as s (S<I> 10 20 30 40 50)      // create a slice referencing start of a new 
                                    // underlying array, with length 5 and capacity 5
```

We can use the slice operator to get a new slice from a slice. The new slice represents a subsection of the same array as the original:

```
func main
    locals arr A<I 10> s S<I> s2 S<I>
    as arr (A<I 10> 10 20 30 40 50 60 70 80 90 100)
    as s (slice arr 2 8)
    as s2 (slice s 3 5)          // same subsection as (slice arr 5 7)
    (println z s2 0)             // 60
```

The `len` (‘length’) operator returns the length of a slice, and the `cap` (‘capacity’) operator returns the capacity of a slice:

```
func main
    locals s S<I>
    as s (S<I> 1 2 3 4)
    (println (len s))            // 4
    (println (cap s))            // 4 (or possibly something greater!)
```

(For reasons discussed in a moment, a newly created slice may have a capacity larger than the minimum required to accomodate the length.)

The `append` operator takes a slice and one or more values to append to the slice. If the slice has enough capacity after the end of its length to store the values, the values are assigned into the existing array, and a slice with a bigger length is returned:

```
func main
    locals arr A<I 10> s S<I>
    as s (slice arr 0 5)             // len 5, cap 10
    as s (append s 46 900 -70) 
    (println (len s))                // 8
    (println (cap s))                // 10
    (println (get s 5))              // 46
    (println (get s 6))              // 900
    (println (get s 7))              // -70
    (println (get s 8))              // panic: index out of bounds
```

However, if there is not enough capacity at the end to store all of the new values, append will:

1. create a new array that is big enough to store the existing slice values plus all the new values
2. copy the values in the existing slice to the new array
3. copy the new values into the new array after the existing values
4. return a slice referencing the first index of this new array, with the new length and capacity

```
func main
    locals arr A<I 6> s S<I>
    as s (slice arr 0 5)             // len 5, cap 6
    as s (append s 46 900 -70) 
    (println (len s))                // 8
    (println (cap s))                // 8 (or possibly something greater!)
    (println (get s 5))              // 46
    (println (get s 6))              // 900
    (println (get s 7))              // -70
    (println (get s 8))              // panic: index out of bounds
```

When we append something to a slice, it’s very common that we’ll append more stuff to the slice soon thereafter. Because creating new arrays and copying elements is expensive, append will often create new arrays bigger than immediately necessary so as to avoid having to create new arrays in subsequent appends on the slice.

The `make` operator creates a slice with an underlying array of a specified size. The values of the array start out as the default of the type:

```
func main
    locals s S<I>
    as s (make S<I> 6)
    (println (get s 0))         // 0
    (println (len s))           // 0
    (println (cap s))           // 0
```

The `copy` operator copies elements of one slice to another slice of the same type. The returned value is the number of elements copied, which is equal to the shorter of the two lengths:

```
func main
    locals foo S<I> bar S<I>
    foo foo := []int{10, 20, 30, 40, 50}
bar := make([]int, 3, 7)
i := copy(bar, foo)             // 3 (the number of elements copied)
a := bar[0]                     // 10
b := bar[1]                     // 20
c := bar[2]                     // 30
```

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

## bitwise operators

The `band` operator performs a 'bitwise and’ between two integers or two bytes. The result of a 'bitwise and’ has a 1 in any position where both inputs have a 1:

```
func main
    locals a Byte b Byte c Byte
    as a (Byte 131)      // 1000_0011
    as b (Byte 25)       // 0001_1001
    as c (band a b)      // 0000_0001
    (println c)          // prints 1
```

Above, only the least-significant bits of the inputs were both 1’s, so all other bits in the result are 0’s.

The `bor` operator performs a 'bitwise or' between two integers or two bytes. The result of a 'bitwise or' has a 1 in any position where either (or both) inputs have a 1:

```
func main
    locals a Byte b Byte c Byte
    as a (Byte 131)                // 1000_0011
    as b (Byte 25)                 // 0001_1001
    as c (bor a b)                 // 1001_1011  
    (println c)                    // prints 155
```

Above, five of the bits had a 1 bit in one or both of the inputs.

The `bxor` operator performs a 'bitwise exclusive or' between two integers or two bytes. The result of a 'bitwise exclusive or' has a 1 in any position where one input (and only one input) has a 1:

```
func main
    locals a Byte b Byte c Byte
    as a (Byte 131)                // 1000_0011
    as b (Byte 25)                 // 0001_1001
    as c (bxor a b)                // 1001_1010  
    (println c)                    // prints 154
```

Above, the least-signifcant bits of both inputs were 1’s, so the result does not have a 1 in that position.

The `bneg` operator performs a 'bitwise negation' on an integer or byte. The result of a 'bitwise negation' has all the bits of the input flipped:

```
func main
    locals a Byte b Byte c Byte
    as a (Byte 131)               // 1000_0011
    as b (bneg a)                 // 0111_1100
    (println b)                   // prints 124
```


#### struct File

#### func createFile

#### func openFile

#### func readFile

#### func writeFile

#### func closeFile

#### func seekFile

#### func seekFileStart

#### func seekFileEnd




## function variables

A function’s type signature is the list of its parameter types and return types:

```
// this function's signature: take a string and a float, return an integer
func foo a Str b F : I
    // ...
```
A function variable references functions with a specified signature. Invoking a function variable invokes whatever function it currently references:

```
func foo a I b Byte
    // ...

func bar a I b Byte : Str
    // ... 
    
func main
    // f is a variable that can reference functions which take an integer and a byte and return a string
    locals f Fn<I Byte : Str>     
    as f bar
    (f 8 2)         // calls bar
    as f foo        // compile error: foo does not have the right signature
```

The default value of a function variable is a reference to nothing, and it is denoted nil. Invoking a nil function value triggers a panic:

```
func main
    locals f Fn<I Byte : Str>         
    (f 8 2)                           // panic: cannot invoke nil
```

A function can receive functions as inputs and return them as outputs:

```
// jared takes a function (taking an integer and returning a string) for its first parameter, takes an integer 
// for its second parameter, and returns a function that returns a byte
func jared a Fn<I : Str> b I : Fn<: Byte> 
    // ...    
```

## local functions and closures

Functions can be created inside other functions with a `localfunc` statement. A local function only exists in its enclosing function.

All `localfunc` statements must appear after the `locals` statement (if any) but before all other statements of the function. 

```
func main
    // create a local function named evan
    localfunc evan a I : I
        return (add a 3)
    (println (evan 8))          // 11
```

Aside from not being accessible outside its enclosing function, the special thing about a local function is that it can access the variables of the call to the enclosing function:

```
func main
    locals a I b I
    localfunc bar : I
        // this function has its own local x, but we can also use 
        // a and b of the enclosing function call
        locals x I
        as x 2
        return (add x a)
    as a 3
    as b 11
    (println (mul (bar) b))     // prints 55
```

In fact, even when the enclosing function call returns, the nested function can continue to use the enclosing call’s variables even though a call’s local variables normally disappear after the call returns. In other words, the nested function can retain local variables of the enclosing function (or method) calls. A closure is a value that references a function and a set of retained variables:

```
// foo returns a function taking no parameters and returning an int
func foo : Fn<: I>
    locals x I
    localfunc f : I
        // x is from enclosing call
        as x (add x 3)
        return x
    as x 2
    return f

func main
    locals a Fn<: I> b Fn<: I>

    as a (foo)    // assign closure to a (function returned by foo retains variable x)
    (a)           // 5
    (a)           // 8
    (a)           // 11

    as b (foo)    // assign a different closure to b (same function but a different retained variable x)
    (b)           // 5
    (b)           // 8
    (b)           // 11

    (b)           // 14
    (b)           // 17
    (b)           // 14
    (b)           // 17
```

## multi-threading with goroutines

As discussed [here](https://www.youtube.com/watch?v=9-KUm9YpPm0) and [here](https://www.youtube.com/watch?v=9GDX-IyZ_C8), an operating system process (*i.e.* a program) starts off with one thread of execution, but *via* [system calls](https://en.wikipedia.org/wiki/System_call), a process can spawn additional threads of execution. These threads all share the same process memory, but the OS schedules these threads independently.

Each CPU core can run one thread at a time, *e.g.* given a 4-core CPU, the CPU can run 4 threads simultaneously. At any moment, the running threads may or may not be from the same process. (It all depends on which threads the OS deems most worthy to currently use the CPU.)

- given *N* CPU cores, *N* threads can run simultaneously
- while one thread blocks, the OS scheduler will generally let another thread run in its place

If computers were infinitely fast--that is, if any amount of code could be fully executed instantaneously--we’d have no real reason to parcel out the work of our programs to multiple threads. Sadly, in the real world, all code takes some amount of time to execute, so to speed things up, we sometimes want to multi-thread our programs. 

However, GoPigeon takes from the Go language a feature called ***goroutines***. A goroutine is a thread of execution managed by the language runtime rather than by the OS. In my GoPigeon program, I can simultaneously have many, many goroutines (thousands or even hundreds of thousands): the runtime creates some number of actual OS threads (usually one for each CPU core in the system) and then schedules the goroutines to run in the threads.

So say we have 4 OS threads and 100 goroutines. The OS decides which (if any) of the 4 threads should run at any moment, and the runtime decides which goroutines should run in these 4 threads.

Why goroutines? Why not just create 100 OS threads? In short, creating and managing goroutines requires less overhead compared to creating and managing OS threads. Whereas creating thousands of OS threads is inadvisable, creating thousands of goroutines incurs relatively reasonable overhead costs. (This [blog post](http://tleyden.github.io/blog/2014/10/30/goroutines-vs-threads/) explains more details.)

To create a goroutine, we use a `go` statement, specifying a function call to kick off the new goroutine:

```
func tom
    (println "yo")

func main
    (println "before")
    go (tom)                  // spawn a new goroutine which starts execution by calling tom
    (println "after")
```

This program, like any other program, starts with a call to main in its original goroutine. After printing "before", main spawns another goroutine, which calls tom. The two goroutines continue execution independently: the original goroutine completes when its call to main returns; the second goroutine completes when its call to tom returns. However, the original goroutine is special in that, when it completes, the program will terminate even if other goroutines have not yet completed. (In Go, there are ways to ensure a goroutine will wait for the completion of other goroutines, but for simplicity we have no such means in GoPigeon).

Nothing is guaranteed about when and for how long the goroutines get time to run on the CPU. In some cases, a goroutine will start execution immediatly after spawning; in other cases, it won’t. In some cases, the goroutine which spawns another will continue running for some time; in other cases, it will wait some time before being resumed. All of this depends on the choices of the runtime and the OS scheduler. The goroutines will be paused and resumed at times we can neither determine nor predict.

So in our example above, it cannot be said whether "yo" or "after" will be printed first. Sometimes "tom" will be printed first; other times "after" will be printed first. Even if running the program a million times prints "yo" first, we cannot say "yo" will always be printed first: it may just happen that the runtime and OS schedulers almost always make the same choices because other OS threads aren’t taking up CPU time. But if other running OS threads were to steal CPU time at the right moments, the schedulers would make different choices, causing "after" to be printed first.

Lastly, be clear that the arguments in the call of a `go` statement are evaluated in the current goroutine, not in the new goroutine:

```
func main
    go (foo (bar))    // bar is called in the original goroutine before the new goroutine is created

## synchronization with channels

Multi-threading gets hard when threads share state (*i.e.* data that can be modified in the course of execution). If a piece of data accessible in multiple threads is modified by one thread, the other threads may be affected by that modification when they read the data. When a shared piece of data is modified by a thread in a way that the other threads are not expecting, the logic of those threads may break. In other words, shared state can easily cause bugs.

Generally, the whole point of sharing state is to allow changes in one thread to affect other threads. However, there are typically chunks of code during which a thread expects no other thread to modify one or more pieces of state. For example, a thread may expect global variable foo to remain unmolested by other threads for the duration of any call to function bar. We would say then that bar is a *critical section*: a chunk of code during which one or more pieces of shared state should not be modified by other threads. A critical section expects to have some chunk(s) of shared state all to itself for its duration.

Enter synchronization primitives, which arrange exclusive access to some chunk of shared state for the duration of a critical section. There are many kinds of synchronization primitives, but the most common is called a lock or mutex (as in ‘mutual exclusion’).

For a piece of shared state, we create a lock to govern its access:

- before using the state, we assert the lock
- when done with the state, we release the lock
- if another thread has asserted the lock without yet releasing it, asserting the lock fails, in which case our thread should not read or write the data and instead do something else or wait before trying to assert the lock again

Note that ‘lock’ is a misleading name: in the real world, a lock physically restrains access; in code, a lock merely indicates whether a thread *should* access the associated state. As long as all threads remember to properly assert and release locks, everything is fine, but doing this in practice is not always easy.

The Go standard library “sync” package provides locks and a few other synchronization primitives. In GoPigeon, however, we only can use *channels*, another feature taken from Go.

Channels offer another way to communicate between and coordinate goroutines.

What programmers call a queue is a list in which items are read from the list only in the order in which they were added to the list. Think of a checkout line at a store: the first person in line is the first person to make it through; the last person in line is the last person to make it through.

A channel is a queue in which:

- the queue has a limited capacity
- if the queue is full, adding an item will block until space is available (because some other goroutine removed a value)
- if the queue is empty, retrieving a value will block until a value is available (because some other goroutine added a value)

Adding a value to a channel is called *sending*; retrieving a value is called *receiving*.

Like arrays and slices, channels are typed: a channel of integers, for example, can only store integers. A channel variable is merely a reference to a channel; an actual channel is created by using the channel type like an operator, specifying its capacity. The `send` operator sends; the `rcv` operator receives:

```
func main
    locals ch Ch<I>          
    as ch (Ch<I> 10)         // assign to ch a new channel of integers with a capacity of 10
    (send ch 3)
    (send ch 5)
    (send ch 2)
    (println (rcv ch))       // 3
    (println (rcv ch))       // 5
    (println (rcv ch))       // 2
```

Again, when we send to a full channel, the goroutine in which we do so will block until another goroutine makes space by receiving from the channel. When we receive from an empty channel, the goroutine in which we do so will block until it has a value to retrieve:

```
func britney ch Ch<I>
    // an intentional infinite loop!
    while true
        (println (rcv ch))   // this receive will block until it can retrieve a value from the channel

func main
    locals ch Ch<I>          
    as ch (Ch<I> 2)         
    go (britney)
    (send ch 3)
    (send ch 5)
    (send ch 2)      // at this point the channel may be full, so this third send may block
```

Remember that nothing is guaranteed about how far a goroutine has reached in its execution relative to other goroutines. Above, maybe the channel never gets full because the britney goroutine already receives the first one or two values. But maybe it does get full! This depends on how exactly the goroutines get scheduled. Thanks to the blocking behavior of send and receive, we don’t need to worry about the scheduling: this program will always print 3 5 2.

It’s possible (and in fact most common) to create a channel with a capacity of 0. Such a channel is always empty and full, so every send will block until another goroutine receives from the channel, and every receive will block until another goroutine sends a value to the channel:

```
func britney ch Ch<I>
    while true
        // this receive will block until the other thread sends another value
        (println (rcv ch))

func main
    locals ch Ch<I>
    as ch (Ch<I> 0)   // zero capacity
    go (britney ch)
    // each send will block until the other goroutine receives from the channel
    (send ch 3)
    (send ch 5)
    (send ch 2)
```

Again, channel variables are merely references. Like a few other types, the default value for channels is denoted nil. Sending or receiving *via* a nil reference triggers a panic:

```
func main
    locals ch Ch<I>
    (send ch 9)             // panic!
```

Be clear that receiving from a channel returns a copy of the sent value. Just like assigning a value to a variable actually copies the value to the variable, sending to a channel copies the value into the channel. Now, if the value sent through a channel is a reference of some kind (*e.g.* a slice or a pointer), then the sender and receiver can end up sharing state. Sometimes that’s what we want, but more commonly we use channels to communicate and coordinate between threads by sharing copies, not by sharing state.

***Sharing copies is safe: I can do whatever I want with my copy without affecting your copy. Sharing state is dangerous: I might change the state in ways you aren’t expecting.***

Enumerating all the possible ways channels can be useful would be very time-consuming. This [blog post](https://blog.golang.org/pipelines) describes some examples.

## `select` statements

A `select` statement allows a goroutine to wait for multiple send or receive operations until one of them stops blocking. Only one of the awaited operations completes.

Each case in a select has a send or receive operation. There is no significance to the order of the cases. Execution blocks until one of the send or receive operations is ready to run. If multiple operations are ready to run, select picks one to run at ‘random’. (Well, more accurately, it’s random from our perspective: the language makes no guarantees about which of multiple ready cases will run.)

```
func main
    locals ch Ch<I> ch2 Ch<I> ch3 Ch<I>

    // ... init the channels

    select
    rcving v ch    // assign to new variable 'v' a value received from 'ch'
        // 'v' belongs to the scope of this case (each case is its own scope) 
        // ... do stuff with value received from 'ch'
    snding ch2 7
        // ... do stuff after having sent 7 to 'ch2'
    rcving v ch3
        // this case has its own 'v' separate from 'v' of the first case
        // ... do stuff with value received from 'ch3'
```

A switch with a default case will never block. If no case operation is ready when the select is reached, the default case will immediately execute:


```
func main
    locals ch Ch<I> ch2 Ch<I> ch3 Ch<I>

    // ... init the channels

    select
    rcving v ch    // assign to new variable 'v' a value received from 'ch'
        // 'v' belongs to the scope of this case (each case is its own scope) 
        // ... do stuff with value received from 'ch'
    snding ch2 7
        // ... do stuff after having sent 7 to 'ch2'
    rcving v ch3
        // this case has its own 'v' separate from 'v' of the first case
        // ... do stuff with value received from 'ch3'
    default
        // ... this body will immediately execute if the other cases all would block
```

So if we wish to send or receive on a single channel without blocking, we can use select with a default case:

```
func main
    locals ch Ch<I>
    // ...
    select
    rcving i ch
        // ... read the channel
    default
        // ... didn't read the channel because it was blocked
```