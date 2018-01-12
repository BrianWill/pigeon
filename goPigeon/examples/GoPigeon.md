
# GoPigeon

## number types

Whereas StaticPigeon has one number type, GoPigeon has ten number types:

- I8 (8-bit signed integer)
- I16 (16-bit signed integer)
- I32 (32-bit signed integer)
- I64 (64-bit signed integer)
- U8 (8-bit unsigned integer)
- U16 (16-bit unsigned integer)
- U32 (32-bit unsigned integer)
- U64 (64-bit unsigned integer)
- F32 (32-bit floating point)
- F64 (64-bit floating point)

A value of one number type is not a valid value of another number type:

```
var x F32
var y I16
as x y          // cannot assign an F32 to an I16
as y x          // cannot assign an I16 to an F32 variable
```

A number literal does not have any specific type and so can be assigned to any number type variable (as long as the value is in the type's range):

```
var a I8
var b F64

as a 92
as b 92

as a 702.45     // compile error: 702.45 is not a valid I8 value
as b 702.45

as a 300        // compile error: 300 is not a valid I8 value
as b 300      
```

When declaring a variable in an `as` statement, a number literal is assumed to be an I64 or an F64:

```
as i' 52        // 'i' is an I64 variable
as f' 52.41     // 'f' is an F64 variable
```

Given a value of one type, you can get its equivalent (or near-equivalent) as another type by using the type like an operator:

```
var x F32
var y U16
as x (F32 y)     // OK: get the F32 equivalent of 'y' and assign to 'x'
as y (U16 x)     // OK: get the U16 equivalent of 'x' and assign to 'y'
```

When converting from larger-range types to smaller-range types, the bytes get truncated. For example, to get a U8 (unsigned 8-bit integer) from an I64 (signed 64-bit integer), StaticPigeon drops the first three bytes, leaving just the least-significant byte. Be clear, then, that there is not necessarily a useful, logical relationship between the original value and the new value. It's the programmer's responsibility to use number conversions in sensible ways.

Arithmetic operations require operands of the same number type and return a value of that type:

```
var a U32
var b U32
var c I64

as a (add a b)            // OK: the operation returns a U32
as a (add a c)            // compile error: the operand types do not match
as a (add a (U32 c))      // OK: both operands are U32

as c (mul a b)            // compile error: the operation returns a U32, but an I64 is expected
as c (I64 (mul a b))      // OK
```

## arrays

Rather than lists, StaticPigeon has arrays. Unlike a list, an array is fixed in size: once created, an array can neither grow nor shrink. In fact, the size is considered integral to the type:

```
var x A<S 7>              // 'x' is an array of 7 strings
var y A<S 4>              
var z A<S 7>              
as x y                    // compile error: A<S 7> and A<S 4> are different types
as x z                    // OK: copy z[0] to x[0], z[1] to x[1], z[2] to z[3], etc.
```

The values of an array default to the default value of their type, *e.g.* the values of a string array default to an empty string.

Unlike list variables in BasicPigeon, an array variable in StaticPigeon is not a reference: an array variable directly represents an array stored in memory.

The size of the array can only be specified by a compile-time expression:

```
var x A<Bool (add 7 2)>    // OK: the size expression can be computed at compile time

as n' 5
var y A<Bool (add n 2)>    // compile error: the size expression cannot be computed at compile time
```

We can create an array by using its type like an operator. Optionally, we can specify its initial values:

```
as a' (A<I32 4>)                      // the 4 values default to zeros
as b' (A<I32 4> -65 43 818 -900)      // the 4 values are -65, 43, 818, -900
as c' (A<I32 4> -65 43)               // compile error: must provide 4 values or none at all
```

To access the elements of an array, we use the `get`, `set`, and `[]` operators, just like with lists in BasicPigeon.

Using `ref` and `[]`, we can create a pointer that represents the address of an individual element within an array:

```
as arr' (A<I32 4>)
var p P<I32>
as p (ref arr[0])            // pointer to the first element of the array
as p (ref arr[1])            // pointer to the second element of the array
```

## slices

A ***slice*** value represents a subsection of an array. Each slice value has three components: 

- a reference to an element within an array
- a *length* (a number of elements)
- a *capacity* (the count of elements from the referenced element through the end of the array)

Given an array, we use the `slice` operator to get a slice value representing a subsection of the array. We use `get`, `set`, and `[]` on the slice to access the values of the array subsection that it represents:

```
as arr' (A<I32 9> 10 20 30 40 50 60 70 80 90)
var sl Sl<I32>
as sl (slice arr 3 7)    // slice referencing index 3 of the array, with 
                          // length 4 (because 7 - 3 is 4) and capacity 7 (because 10 - 3 is 7)
as a' sl[0]      // 40
as b' sl[1]      // 50
as c' sl[2]      // 60
as d' sl[3]      // 70
as e' sl[4]      // panic! out of bounds (index must be less than length)
```

In effect, a slice represents *length*-number of elements starting from the referenced element. (The capacity is needed for the *append* built-in function, discussed shortly.)

```
as arr' (A<I32 9> 10 20 30 40 50 60 70 80 90)
as sl' (slice arr 3 7)
as s[0] -999
as z' arr[3]              // -999
```

It's perfectly possible for a slice to start at the beginning of an array. In fact, a slice can represent the whole of an array:

```
as arr' (A<I32 12> 10 20 30 40 50 60 70 80 90 100 110 120)
as s' (slice arr 0 7)     // slice referencing index 0 of the array, with length 7 and capacity 10
as s2' (slice arr 0 10)   // slice referencing index 0 of the array, with length 10 and capacity 10
```

Note that multiple slice values can represent overlapping subsections of the same array. Consequently, changes *via* one slice can affect other slices:

```
as arr' (A<I32 10> 10 20 30 40 50 60 70 80 90 100)
as s' (slice arr 4 9)
as s2' (slice arr 8 10)
as s[4] -999      
as z' s2[0]               // -999
```

Slices are typed, *e.g.* an `Sl<I32>` is different from a `Sl<Bool>` slice which is different from a `Sl<S>` slice, *etc.* The length and capacity of a slice is not part of its type, so we can assign a slice of any length or capacity to a slice variable:

```
as arr' (A<I32 10> 10 20 30 40 50 60 70 80 90 100)
var foo Sl<I32>
as foo (slice arr 2 5)
as foo (slice arr 5 9)

var bar Sl<S>
as bar (slice arr 2 7)     // compile error: cannot assign Sl<I32> to an Sl<S> variable
```

If we specify just one number to the `slice` operator, it defaults to the length of the array:

```
as arr' (A<I32 10> 10 20 30 40 50 60 70 80 90 100)
as sl' (slice arr 3)       // as s' (slice arr 0 10)
```

Using a slice type like an operator creates an slice value with a new underlying array:

```
var sl' Sl<I32>
as sl (Sl<I32 5> 10 20 30 40 50)    // create a slice referencing first element 
                                    // of a new underlying array, with length 5 and capacity 5
```

The size of the array can be specified by a runtime expression:

```
as n' 30
as sl' (Sl<I32 n>)
```

We can use the slice operator to get a new slice from a slice. The new slice represents a subsection of the same array as the original:

```
as arr' (A<I32 10> 10 20 30 40 50 60 70 80 90 100)
as sl' (slice arr 2 8)       
as sl2' (slice sl 3 5)        // same subsection as (slice arr 5 7)
as z' sl2[0]                  // 60
```

The `len` operator returns the length of a slice:

```
as sl' (Sl<I32 4> 1 2 3 4)
as z' (len sl)                // 4
```

The `cap` operator returns the capacity of a slice:

```
as arr' (Sl<I32 4> 1 2 3 4)
as z' (cap sl)                // 4 (or possibly something greater!)
```

(For reasons discussed in a moment, GoPigeon may give a newly created slice a capacity larger than the minimum required to accomodate the length.)

The `append` operator takes a slice and one or more values to append to the slice. If the slice has enough capacity after the end of its length to store the values, the values are assigned into the existing array, and a slice with a bigger length is returned:

```
var arr A<I32 10>
as sl' (slice arr 0 5)            // len 5, cap 10
as sl (append s 46 900 -70)
as a' (len sl)                    // 8
as b' (cap sl)                    // 10
as c' sl[5]                       // 46
as d' sl[6]                       // 900
as e' sl[7]                       // -70
as f' sl[8]                       // panic: index out of bounds
```

However, if there is not enough capacity at the end to store all of the new values, *append* will 

 1. create a new array that is big enough to store the existing slice values plus all the new values
 2. copy the values in the existing slice to the new array
 3. the new values are copied into the new array after the existing values
 4. return a slice referencing the first index of this new array, with the new length and capacity:

```
var arr A<I32 6>
as sl' (slice arr 0 5)            // len 5, cap 6
as sl (append sl 46 900 -70)
as a' (len sl)                    // 8
as b' cap(sl)                     // 8 (or possibly something greater!)
as c' sl[5]                       // 46
as d' sl[6]                       // 900
as e' sl[7]                       // -70
as f' sl[8]                       // panic: index out of bounds
```

When we append something to a slice, it's very common that we'll append more stuff to the slice soon thereafter. Because creating new arrays and copying elements is expensive, *append* will often create new arrays bigger than immediately necessary so as to avoid having to create new arrays in subsequent appends on the slice.

The `copy` operator copies elements of one slice to another slice of the same type. The returned value is the number of elements copied, which is equal to the shorter of the two lengths:

```
as foo' (Sl<I32 5> 10 20 30 40 50)
as bar' (Sl<I32 3 7>)                // a slice of I32's with a length of 3 and a capacity of 7
as i' (copy bar foo)                 // 3 (the number of elements copied)
as a' bar[0]                         // 10
as b' bar[1]                         // 20
as c' bar[2]                         // 30
```

## slice default values

The zero value of a slice is represented by the reserved word `nil`. A `nil` slice value references nothing and has a length and capacity of zero. Accesing elements of a `nil` slice triggers a panic, but we can append to a `nil` slice:

```
var foo Sl<I32>              // defaults to nil
as foo nil                   // assign nil to 'foo'
as a' foo[0]                 // panic: index out of bounds
as b' (len foo)              // 0
as foo (append foo 8 10)     // returns a slice with length 2 and a capacity of at least 2
```

(As we'll see shortly, the reserved word `nil` is also used to represent the zero value for some other types as well.)


# strings

A string value is composed of a reference and a length. The reference points to where the actual text data is stored, and the length indicates the number of bytes of text data.

So when we assign a string to a variable, the text data is a chunk of bytes somewhere in memory, but the variable itself stores just a reference to that chunk and an integer representing the length of the chunk.

```
// the text data is somewhere in memory; the string variable stores 
// a reference to that chunk and its length
as s' "hello"
```

Because string text data is stored as UTF8, some characters may take up more than one byte, and so the number of bytes may not be the same as the number of characters. The `runeCount` operator returns the number of characters in a string:

```
s := "hello"

as a' (len s)          // 5 (English characters are single-byte characters)
as b' (runeCount s)    // 5

as s "世界"
as c' (l               // 6
as d' (runeCount s)    // 2
```

## indexing strings

Use `get` or `[]` on strings to read individual bytes (`U8`) of the text data:

```
as s' "hello"
var b U8
as b s[0]               // 104  (lowercase 'h' in Unicode)
as b s[1]               // 101  (lowercase 'e' in Unicode)

s = "世界"
as b s[0]               // 228   (the first byte of three-byte character '世')
as b s[1]               // 184   (the second byte of three-byte character '世')
as b s[3]               // 231   (the first byte of three-byte character '界')
```

The bytes of a string cannot be modified:

```go
as s "hello"
as s[0] 65              // compile error: cannot modify bytes of a string
```

Using the `substr` operator on a string returns a string with the specified range of text data bytes:

```
as s' "hello"
as s (substr s 1 4)            // "ell"

as s "世界"
as s2' (substr s 0 3)          // "世"
as s2 (substr s 2 4)           // "���"  (the 3 bytes do not form valid UTF8 characters)
```

(How exactly invalid UTF8 text data gets displayed depends upon the program displaying the data.)

## converting between strings and byte slices

We can convert between strings and byte slices. Because strings are meant to be immutable, converting between strings and byte slices always copies the bytes. Effectively, a string and a byte slice never share bytes:

```
as s' "hello"
var b Sl<U8>
as b (Sl<U8> s)              // the new Sl<U8> has its own copy of the string's bytes
as b[0] 72                   // does not affect the string
as s2' s                     // "hello"
as s (S b)                   // "Hello"     (72 is capital 'H' in Unicode)
as b[0] 90                   // does not affect either string
as s2 s                      // "Hello"
```

## casting strings to I32 slices

Converting a string to a slice of I32's produces a slice wherein each I32 represents the codepoint of each character in the string:

```
as s' "世界"
as a' (len s)             // 6
as chars' (Sl<I32> s)
as b' (len r)             // 2
as c' r[0]                // 19990 (the codepoint of 世)
as d' r[1]                // 30028 (the codepoint of 界)
```

## character literals

We can express the integer values of Unicode characters by enclosing a character in single-quotes:

```
as x' 'H'                // as x' 72 
as y' '世'               // as y' 19990
as z' (add 'H' '世')     // as z' (add 72 19990)
```

Like other integer literals, these character literals have no specific type:

```
var i I32
as i 'H'        // as i 72
var j U8
as j '世'       // compile error: 19990 is not a valid uint8 value
```

## local functions

We can create functions inside other functions (or methods) with `localfunc` statements:

```
func foo
    localfunc bar a I32 b I32 : I32
        return (add a b)
    (bar)
```

A `localfunc` statement does two things:

- creating a local variable of a function type
- creating a function and assigning it to the variable

Above, `bar` is a local variable of type `F<I32 I32: I32>`, and `bar` is assigned the function defined in the `localfunc` statement.


## closures

A local function can read and write the variables of the enclosing function call in which the local function is created: 

```
func main
    // main has four local variables: 'a', 'b', 'bar', and 'z'
    as a' 3
    as b' 11
    localfunc bar : I32 
        // this function has its own local 'x', but we can also 
        // use 'a', 'b', and 'bar' of the enclosing function call
        as x' 2
        return (add x a)
    as z' (mul (bar) b)      // 55
```

In fact, even when the enclosing function call returns, the local function can continue to use the enclosing call's variables *even though a call's local variables normally disappear after the call returns*. In other words, the local function can *retain* local variables of the enclosing function (or method) calls. A ***closure*** is a value that references a function and a set of retained variables:

```
// 'foo' returns a function taking no parameters and returning an int
func foo : F< : I32>
    as a' 2
    localfunc bar : I32
        // 'a' is from the enclosing call
        as a (add a 3)
        return a
    return bar
    
func main
    var x F< : I32>
    var y F< : I32>

    as x (foo)   // assign closure to 'x' (function returned by 'foo' retains variable 'a')
    (x)          // 5
    (x)          // 8
    (x)          // 11

    as y (foo)   // assign a different closure to 'y' (same function but a different retained variable 'a')
    (y)          // 5
    (y)          // 8
    (y)          // 11

    (x)          // 14
    (x)          // 17
    (y)          // 14
    (y)          // 17
```

## multi-threading

As we've discussed earlier, a process starts off with one thread of execution, but *via* system calls, a process can spawn additional threads of execution. These threads all share the same process memory, but the OS schedules these threads independently. 

Each CPU core can run one thread at a time, *e.g.* given a 4 core CPU, the CPU can run 4 threads simultaneously. At any moment, the running threads may or may not be from the same process. (It all depends on which threads the OS deems most worthy to currently use the CPU.)

 - given *N* CPU cores, *N* threads can run simultaneously
 - while one thread blocks, the OS scheduler will generally let another thread run in its place

If computers were infinitely fast---if any amount of code could be fully executed instantaneously---we'd have no real reason to parcel out the work of our programs to multiple threads. Sadly, in the real world, all code takes some amount of time to execute, so to speed things up, we sometimes want to multi-thread our programs:

## goroutines

Using the *syscall* package, we can spawn separate threads. However, it is generally better to use Go's special mechanism for multi-threading called ***goroutines***. A goroutine is a thread of execution managed by the Go runtime rather than by the OS. In my Go program, I can simultaneously have many, many goroutines (thousands or even hundreds of thousands): the Go runtime creates some number of actual OS threads (usually one for each CPU core in the system) and then schedules the goroutines to run in the threads.

So say we have 4 OS threads and 100 goroutines. The OS decides which (if any) of the 4 threads should run at any moment, and the Go runtime decides which goroutines should run in these 4 threads.

Why goroutines? Why not just create 100 OS threads? In short, creating and managing goroutines requires less overhead compared to creating and managing OS threads. Whereas creating thousands of OS threads is inadvisable, creating thousands of goroutines incurs relatively reasonable overhead costs. ([This blog post explains more details](http://tleyden.github.io/blog/2014/10/30/goroutines-vs-threads/).)

To create a goroutine, we use a `go` statement, specifying a function call to kick off the new goroutine:

```
func foo
    (print "foo")

func main
    (print "before")
    go (foo)                  // spawn a new goroutine which starts execution by calling (foo)
    (print "after")
```

This program, like any Go program, starts with a call to *(main)* in its *main* goroutine. After printing `"before"`, *(main)* spawns another goroutine, which calls *(foo)*. The *main* and *foo* goroutines continue execution independently: the *main* goroutine completes when its call to *(main)* returns; the *foo* goroutine completes when its call to *(foo)* returns. However, the *main* goroutine is special in that, when it completes, the program will terminate even if other goroutines have not yet completed. (As we'll see later, there are ways to ensure a goroutine will wait for the completion of other goroutines). 

Nothing is guaranteed about when and for how long the goroutines get time to run on the CPU. In some cases, a goroutine will start execution immediatly after spawning; in other cases, it won't. In some cases, the goroutine which spawns another will continue running for some time; in other cases, it will wait some time before being resumed. All of this depends on the choices of the Go runtime and the OS scheduler. Goroutines will be paused and resumed at times the programmer can neither determine nor predict.

So in our example above, it cannot be said whether `"foo"` or `"after"` will be printed first. Sometimes `"foo"` will be printed first; other times `"after"` will be printed first. Even if running the program a million times prints `"foo"` first, we cannot say `"foo"` will always be printed first: it may just happen that the Go runtime and OS schedulers almost always make the same choices because other OS threads aren't taking up CPU time. But if other running OS threads were to steal CPU time at the right moments, the schedulers would make different choices, causing `"after"` to be printed first.

Lastly, be clear that the arguments in the call of a `go` statement are evaluated in the current goroutine, not in the new goroutine:

```
go (foo (bar))     // (bar) is called in this goroutine before the new goroutine is created
```

## shared state

Multi-threading gets hard when threads share *state* (*i.e.* data that can be modified in the course of execution). If a piece of data accessible in multiple threads is modified by one thread, the other threads may be affected by that modification when they read the data. When a shared piece of data is modified by a thread in a way that the other threads are not expecting, the logic of those threads may break. In other words, shared state can easily cause bugs.

Generally, the whole point of sharing state is to allow changes in one thread to affect other threads. However, there are typically chunks of code during which a thread expects no other thread to modify one or more pieces of state. For example, a thread may expect global variable *foo* to remain unmolested by other threads for the duration of any call to *(bar)*. We would say then that *bar* is a ***critical section***: a chunk of code during which one or more pieces of shared state should not be modified by other threads. A critical section expects to have some chunk of shared state all to itself for its duration.

Enter ***synchronization primitives***, which arrange exclusive access to some chunk of shared state for the duration of a critical section. There are many kinds of synchronization primitives, but the most common is called a ***lock*** or ***mutex*** ('mutual exclusion').

For a piece of shared state, we create a lock to govern its access:

 - before using the state, we *assert* the lock
 - when done with the state, we *release* the lock
 - if another thread has asserted the lock without yet releasing it, asserting the lock fails, in which case our thread should not read or write the data and instead do something else or wait before trying to assert the lock again

Note that 'lock' is a misleading name: in the real world, a lock physically restrains access; in code, a lock merely indicates whether you *should* access the associated state. As long as all threads remember to properly assert and release locks, everything is fine, but doing this in practice is not always easy.

The Go standard library *"sync"* package provides locks and a few other synchronization primitives.

## channels

Go's ***channels*** offer another way to communicate between and coordinate goroutines.

What programmers call a *queue* is a list in which items are read from the list strictly in the order in which they were added to the list. Think of a checkout line at a store: the first person in line is the first person to make it through; the last person in line is the last person to make it through.

A channel is a queue in which:

 - the queue has a limited capacity
 - if the queue is full, adding an item will block until space is available once some other goroutine removes a value
 - if the queue is empty, retrieving a value will block until a value is available once some other goroutine adds a value

Adding a value to a channel is called *sending*; retrieving a value is called *receiving*.

Like arrays and slices, channels are typed: a channel of `I32`'s, for example, can only store `I32`'s. A channel variable is merely a reference to a channel; channel values are created by using the channel type like an operator and specifying a size; an uninitialized channel variable defaults to referencing a new channel with capacity 0:

```
var ch Ch<I32>              // create variable 'ch', referencing a channel of I32's with capacity 0
as ch (Ch<I32> 10)          // assign to 'ch' a new channel of I32's with capacity 10
```

The `send` operator sends a value to a channel; the `rcv` operator receives a value from a channel:

```
as ch' (Ch<I32> 10)
(send ch 3)                    // send the value 3 to the channel referenced by 'ch'
(send ch 5)                    // send the value 5
(send ch 2)                    // send the value 2
as a' (rcv ch)                 // receive the value 3 from the channel referenced by 'ch'
as b' (rcv ch)                 // receive the value 5
as c' (rcv ch)                 // receive the value 2
```

Again, when we send to a full channel, the goroutine in which we do so will block until another goroutine makes space by receiving from the channel. When we receive from an empty channel, the goroutine in which we do so will block until it has a value to retrieve:

```
func foo ch Ch<I32>
    while true
        // this receive will block until it can retrieve a value from the channel
        (print (rcv ch))

func main
    as ch' (Ch<I32> 2)
    go (foo ch)
    (send ch 3)
    (send ch 5)
    // at this point the channel may be full, so this third send may block
    (send ch 2)
```

Remember that nothing is guaranteed about how far a goroutine has reached in its execution relative to other goroutines. Above, maybe the channel never gets full because the *(foo)* goroutine already receives the first one or two values. But maybe it does get full! This depends on how exactly the goroutines get scheduled. Thanks to the blocking behavior of send and receive, we don't need to worry about the scheduling: this program will always print `3 5 2`.

It's possible---and in fact most common---to create a channel with a capacity of `0`. Such a channel is always empty *and* full, so every send will block until another goroutine receives from the channel, and every receive will block until another goroutine sends a value to the channel:

```
func foo ch Ch<I32>
    while true
        // this receive will block until it can retrieve a value from the channel
        (print (rcv ch))

func main
    as ch' (Ch<I32> 2)
    go (foo ch)
    // each send will block until the other goroutine receives from the channel
    (send ch 3)
    (send ch 5)
    (send ch 2)
```

Be clear that receiving from a channel returns a *copy* of the sent value. Just like assigning a value to a variable actually copies the value to the variable, sending to a channel copies the value into the channel. Now, if the value sent through a channel is a reference of some kind (*e.g.* a slice or a pointer), then the sender and receiver can end up sharing state. Sometimes that's what we want, but more commonly we use channels to communicate and coordinate between threads by sharing copies, not by sharing state.

> ***Sharing copies is safe: I can do whatever I want with my copy without affecting your copy. Sharing state is dangerous: I might change the state in ways you aren't expecting.***

Enumerating all the possible ways channels can be useful would be very difficult. This [blog post](https://blog.golang.org/pipelines) describes some examples.

## closing channels

The *close* operator closes a channel. We can still receive from a closed channel, but sending to a closed channel triggers a panic. Once a closed channel has no more values to receive, any subsequent receive operations will return the default value of the type without ever blocking:

```
as ch' (Ch<I32> 3)
(send ch 1)
(send ch 2)
(send ch 3)
(close ch)
as a' (rcv ch)   // 1
as b' (rcv ch)   // 2
as c' (rcv ch)   // 3
as d' (rcv ch)   // 0
as e' (rcv ch)   // 0
```

To distinguish between a default value sent through a channel and a zero value indicating the channel has closed, the `rcv` operator can return two values. The first returned value is the value read from the channel, and the second is a boolean (`true` indicating the value was sent):

```go
as ch' (Ch<I32> 3)
// ...
as val' ok' (rcv ch)      // 'ok' will be true if the value was sent
```

Closing a channel which has already been closed triggers a panic.

## `foreach` with channels

A `foreach` loop is a convenient way to read from a channel until it closes. Each iteration receives a value from the channel (and will block accordingly, like any normal receive operation). Once the channel is closed and empty, the loop ends.

```
as ch' (Ch<I32> 10)
(send ch 6)
(send ch 4)
(close ch)

// after two iterations, loop ends because the channel is closed
foreach v ch
    (print v)
```

The loop above is simply a more compact way to write the below:

```
as v' ok' (rcv ch)
while ok 
    (print v)
    as v ok (rcv ch)
```

## `select` statements

A `select` statement allows a goroutine to wait for multiple send or receive operations until one of them stops blocking. Only one of the awaited operations completes.

Each case in a `select` has a send or receive operation. Unlike in a `switch`, there is no significance to the order of the cases. Execution blocks until one of the send or receive operations is ready to run. If multiple operations are ready to run, `select` picks one to run at 'random'. (Well, more accurately, it's random *from our perspective*: Go makes no guarantees about which of multiple ready cases will run.)

```
var ch Ch<I32>
var ch2 Ch<I32>
var ch3 Ch<I32>

// ... 

select
case as v' (rcv ch)   // assign to new variable 'v' a value received from 'ch'
    // 'v' belongs to the scope of this case (each case is its own scope) 
    // ... do stuff with value received from 'ch'
case (send ch2 7)
    // ... do stuff after having sent 7 to 'ch2'
case as v' (rcv ch3)
    // this case has its own 'v' separate from 'v' of the first case
    // ... do stuff with value received from 'ch3'
```

A `switch` with a `default` case will never block. If no case operation is ready when the `select` is reached, the `default` case will immediately execute:

```
select
case as v' (rcv ch)
    // ...
case (send ch2 7)
    // ...
case as v' (rcv ch3)
    // ...
default
    // ...will immediately execute if the three operations all block
```

So if we wish to send or receive on a single channel without blocking, we can use `select` with a `default` case:

```
var ch Ch<I32>
// ...
select
case as i' (rcv ch)
    // ...read the channel
default
    // ...didn't read the channel because it was blocked
```

## channels of channels

Understand that...

 - channels are values
 - we can send values of any type through channels

Therefore, we can create channels of channels:

```
var ch Ch<Ch<I32>>
as ch (Ch<Ch<I32>> 4)

var ich Ch<I32>
as ich (Ch<I32> 17)

(send ch ich)
```

In fact, just like we can have arrays, slices, and pointers of any degree, we can have channels of any degree:

```
// a channel of channels of channels of channels of channels of ints
var ch Ch<Ch<Ch<Ch<Ch<I32>>>>>       
```

In practice, 2-degree channels are occasionally useful, and someone somewhere has probably used a 3-degree channel once or twice. It's unlikely anyone has ever used a channel of more than 3 degrees.

## variadic functions

A *variadic function* is a function in which the last parameter is a slice preceded by `...`. A variadic function is called not by passing a slice to this last parameter but rather zero or more elements that get automatically bundled into a new slice:

```
// function 'foo' is variadic
// 'b' is a Sl<I32> but gets its argument in a special way
func foo a S ... b Sl<I32>
    // ...

func main
    (foo "hi" 3 2 7)         // passes (Sl<I32> 3 2 7) to parameter 'b'
    (foo "hi" 3)             // passes (Sl<I32> 3) to parameter 'b'
    (foo "hi")               // passes (Sl<I32>) to parameter 'b'
```

This minor syntax allowance simply spares us from creating these new slices explicitly in each call:

```
// non-variadic version of 'foo' 
func foo a S b Sl<I32>
    // ...

func main
    (foo "hi" (Sl<I32> 3 2 7))
    (foo "hi" (Sl<I32> 3))
    (foo "hi" (Sl<I32>))
```

If we want to pass an already existing slice to a variadic function, we can do so using `...` as a suffix on the last argument:

```
func main
    as x' (Sl<I32> 3 2 7)
    (foo "hi" x...)          // passes the slice to parameter 'b'
}
```

## return parameters

The return types of a function can be given associated variables. A `return` statement with no explict values returns the value(s) of the return variable(s). The return variables have default values at the start of the call:

```
// 'bar' has a return variable 'a' of type I32
func bar x I32 : a I32 b S
    // 'a' starts out 0, 'b' starts out ""
    as a 3
    as b "hi"
    if (gt x 7)
        return        // implicitly returns 'a' and 'b'
    return x b     

func main
    as i' s' (bar 10)    // 3, "hi"
    as i s (bar 5)       // 5, "hi"
```

Return variables can occaisonally make a function look a bit cleaner in some cases where the function has many `return` statements. There are also some scenarios involving `defer` statements where return variables are needed.

## bitwise operators


The `band` ('bitwise and') operator performs an 'and' between the bits of two integers of the same type. The result of a bitwise `band` has a 1 bit in any position where both inputs have a 1:

```
var a U8
var b U8
var c U8
as a 131             // 1000_0011
as b 25              // 0001_1001
as c (band a b)      // 0000_0001  (decimal 1)
```

Above, only the least-significant bits of the inputs were both 1's, so all other bits in the result are 0's.

The `bor` ('binary or') operator performs an 'or' between the bits of two integers of the same type. The result of a bitwise `bor` has a 1 bit in any position where either (or both) inputs have a 1:

```
var a U8
var b U8
var c U8
as a 131             // 1000_0011
as b 25              // 0001_1001
as c (bor a b)       // 1001_1011  (decimal 155)
```

Above, five of the bits had a 1 bit in one or both of the inputs.

The `bxor` ('binary exclusive or') operator performs an 'xor' between the bits of two integers of the same type. The result of a `bxor` operation has a 1 bit in any position where one input (and *only* one input) has a 1:

```
var a U8
var b U8
var c U8
as a 131             // 1000_0011
as b 25              // 0001_1001
as c (bxor a b)      // 1001_1010  (decimal 154)
```

Above, the least-signifcant bits of both inputs were 1's, so the result does not have a 1 in that position.

The `bneg` ('binary negation') operator negates the bits of an integer. The result of a `bneg` operation has all the bits of the input flipped:

```
var a U8
var b U8
as a 131             // 1000_0011
as b (bneg a)        // 0111_1100  (decimal 124)
```

## `defer` statements

A `defer` statement defers execution of a function or method call. Every `defer` adds another call to a list belonging to the containing function or method call; when the call ends, its list of defered calls are executed in reverse order (*i.e.* the last defered call runs first). 

```
// prints: "1", then "2", then "3", then "4"
func foo
    (print "1")
    defer (println "4")
    defer (println "3")
    if false
        // this defer statement is never executed, so this call is never defered
        defer (println "never")   
    (println "2")
```

Defering calls can be useful for doing clean-up business, such as making sure a file is closed when execution leaves a call.

## panics

A ***panic*** is triggered by various bad operations. Some example bad operations:

 - accessing an array or slice index that is out of bounds
 - invoking a method *via* a nil interface value
 - sending to a closed channel
 - asserting the wrong type using the single-return form of type assertion

When a panic occurs in a goroutine, execution backs out of the call chain, executing all deferred calls as it goes. For example, say a goroutine executes A, which calls B, which calls C, which panics. If A, B, and C have deferred calls before the panic, the deferred calls will run in reverse order: C, then B, then A. 

Once a panic backs execution out of a goroutine, the whole program aborts regardless whether other goroutines are still executing. 

The `panic` operation triggers a panic in the current goroutine. Deliberately triggering panics is sometimes appropriate, such as when the caller passed bad arguments. (Passing bad arguments is a bug, not an error: we should fix the code to stop passing bad arguments.)

```
func foo a I32 b I32 : I32
    // ...
    if badInput
        (panic)
```

## recovering from panics

We can stop a panic and resume a goroutine's normal execution using the `recover` operation. When called directly from a defered call, recover stops the panic from propagating up to the next call:

```go
func foo
    localfunc fn
        (println "still recovering")
    defer (fn)

    localfunc fn2
        (recover)
        (println "recovering")
    defer (f2)

    (panic)

func main
    (foo)           // prints: "recovering", then "still recovering"
    // ... execution continues normally
```

Above, we recover in a defered call of `foo`, so execution resumes normally where `foo` was called. But what if `foo` returned a value?

```
func foo : I32
    localfunc fn
        (recover)
    defer (fn)

    (panic)
    return 3


func main
    as z' (foo)         // 0
```

Here, the recovered call returns a zero value. Using return variables, defered calls can set the return value to something else:

```
func foo : a I32
    localfunc fn
        (recover)
        as a 5
    defer (fn)

    (panic)
    return 3

func main
    as z' (foo)         // 5
```

We can pass a single value of any type to *panic*. This value is then returned by *recover* (as an empty interface value):

```
func foo a I32
    localfunc fn
        as a (I32 (recover))
    defer (fn)

    (panic 7)
    return 3

func main
    as z' (foo)         // 7
```

If no value is passed to *panic*, *recover* returns `nil`.

Using *recover* outside a defered call during a panic does nothing and returns `nil`:

```
func main
    as z' (recover)     // does nothing and returns empty interface value nil
```

If a panic is triggered while a panic is already in progress, the defered call where the second panic occurs aborts, but otherwise the panic continues as normal.










Go features not in GoPigeon
    embedding
    unidirectional channels
    const
    iota
    goto
    break / continue labels
    method binding (like `meth2func` but with concrete value instead of type)




