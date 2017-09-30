# Pigeon, an educational programming language

More accurately, *two* educational programming languages:

- DynamicPigeon
- StaticPigeon

DynamicPigeon ([docs](dynamic-pigeon.md)) teaches basic programming concepts common to all mainstream languages:

- variables
- operations
- control flow (if, while)
- local and global variables
- functions
- lists and hashmaps

StaticPigeon builds upon DynamicPigeon to teach most of the core concepts of the Go language. Once a student is comfortable with StaticPigeon, learning Go should then mostly be a matter of adjusting to different syntax.

StaticPigeon is logically split into two tiers for easier learning.

The first tier ([docs](static-pigeon-tier-1.md)) includes:

- static typing
- multi-return functions
- structs
- methods
- interfaces
- pointers
- typeswitchs
- foreach loops
- break and continue statements
- imports
- standard library file and util packages

The second tier ([docs](static-pigeon-tier-2.md)) includes:

- arrays and slices
- function variables, nested functions, closures
- goroutines
- channels, `select` statements
- bitwise operations
- standard library networking packages

# Installation and usage

Because Pigeon compiles to Go, you'll need to install the [Go tool](https://golang.org/doc/install) if you haven't already. Once the Go tool is installed, you can install the Pigeon compiler with one command:

```
go get http://github.com/BrianWill/pigeon
```

This will put the pigeon executable in your Go bin directory. Assuming the Go bin directory is on your shell path, you can then run the compiler from any directory like so:

```
pigeon somefile.dpigeon      # compile and run a DynamicPigeon program
pigeon somefile.spigeon      # compile and run a StaticPigeon program
```
