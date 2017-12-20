# Pigeon, an educational programming language

More accurately, Pigeon is *two* educational programming languages:

- DynamicPigeon
- StaticPigeon

DynamicPigeon ([reference](docs/dynamic-pigeon-reference.md), [notes](docs/dynamic-pigeon-notes.md)) teaches basic programming concepts common to all mainstream languages:

- variables
- operations
- control flow (if, while)
- local and global variables
- functions
- lists and hashmaps

StaticPigeon ([reference](docs/static-pigeon-reference.md), [notes](docs/static-pigeon-notes.md)) builds upon DynamicPigeon to teach most of the core concepts of the Go language (*sans* goroutines and channels):

- static typing
- arrays and slices
- multi-return functions
- structs
- methods
- interfaces
- typeswitches
- pointers
- bitwise operations
- reading/writing files

Once a student is comfortable with StaticPigeon, learning Go is mostly a matter of adjusting to different syntax. ([Learning Go after Pigeon](docs/go-lang.md)) 

# Installation and usage

Because Pigeon compiles to Go, you'll need to install the [Go tools](https://golang.org/doc/install) if you haven't already. Once the Go tool is installed, you can install the Pigeon compiler with one command:

```
go get github.com/BrianWill/pigeon
```

This will put the Pigeon executable in your Go bin directory. Assuming the Go bin directory is on your shell path, you can then run the compiler from any directory like so:

```
pigeon somefile.dpigeon      # compile and run somefile.dpigeon as a DynamicPigeon program
pigeon somefile.spigeon      # compile and run somefile.spigeon as a StaticPigeon program
```
