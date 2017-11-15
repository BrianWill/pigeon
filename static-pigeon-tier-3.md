
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

However, StaticPigeon takes from the Go language a feature called ***goroutines***. A goroutine is a thread of execution managed by the language runtime rather than by the OS. In my StaticPigeon program, I can simultaneously have many, many goroutines (thousands or even hundreds of thousands): the runtime creates some number of actual OS threads (usually one for each CPU core in the system) and then schedules the goroutines to run in the threads.

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

This program, like any other program, starts with a call to main in its original goroutine. After printing "before", main spawns another goroutine, which calls tom. The two goroutines continue execution independently: the original goroutine completes when its call to main returns; the second goroutine completes when its call to tom returns. However, the original goroutine is special in that, when it completes, the program will terminate even if other goroutines have not yet completed. (In Go, there are ways to ensure a goroutine will wait for the completion of other goroutines, but for simplicity we have no such means in StaticPigeon).

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

The Go standard library “sync” package provides locks and a few other synchronization primitives. In StaticPigeon, however, we only can use *channels*, another feature taken from Go.

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