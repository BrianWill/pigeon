# Pigeon, an educational programming language

More accurately, Pigeon is *two* educational programming languages:

- DynamicPigeon
- StaticPigeon

DynamicPigeon ([docs](dynamic-pigeon.md)) teaches basic programming concepts common to all mainstream languages:

- variables
- operations
- control flow (if, while)
- local and global variables
- functions
- lists and hashmaps

StaticPigeon builds upon DynamicPigeon to teach most of the core concepts of the Go language. Once a student is comfortable with StaticPigeon, learning Go should then be mostly a matter of adjusting to different syntax.

The StaticPigeon documentation is split into three parts:

Part one ([docs](static-pigeon-tier-1.md)):

- static typing
- multi-return functions
- structs
- methods
- interfaces
- pointers
- typeswitches
- break and continue statements

Part two ([docs](static-pigeon-tier-2.md)):

- arrays and slices
- imports
- standard library file package
- foreach loops
- bitwise operations

Part three ([docs](static-pigeon-tier-3.md)):

- function variables, nested functions, closures
- goroutines
- channels, `select` statements
- standard library networking packages

Once comfortable with StaticPigeon, students should learn Go ([docs](go-lang.md)).

# Installation and usage

Because Pigeon compiles to Go, you'll need to install the [Go tool](https://golang.org/doc/install) if you haven't already. Once the Go tool is installed, you can install the Pigeon compiler with one command:

```
go get http://github.com/BrianWill/pigeon
```

This will put the Pigeon executable in your Go bin directory. Assuming the Go bin directory is on your shell path, you can then run the compiler from any directory like so:

```
pigeon somefile.dpigeon      # compile and run a DynamicPigeon program
pigeon somefile.spigeon      # compile and run a StaticPigeon program
```
