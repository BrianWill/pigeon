# Pigeon, an educational programming language

Pigeon ([reference](docs/dynamic-pigeon-reference.md), [notes](docs/dynamic-pigeon-notes.md)) teaches basic programming concepts common to all mainstream languages:

- variables
- operations
- control flow (if, while)
- local and global variables
- functions
- lists and hashmaps

GoPigeon ([reference](docs/static-pigeon-reference.md), [notes](docs/static-pigeon-notes.md)) builds upon Pigeon to teach most of the core concepts of the Go language (*sans* goroutines and channels):

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

Once a student is comfortable with GoPigeon, learning Go is mostly a matter of adjusting to different syntax. ([Learning Go after Pigeon](docs/go-lang.md)) 

# Installation and usage

Because Pigeon compiles to Go, you'll need to install the [Go tools](https://golang.org/doc/install) if you haven't already. Once the Go tool is installed, you can install the Pigeon compiler with one command:

```
go get github.com/BrianWill/pigeon
```

This will put the Pigeon executable in your Go bin directory. Assuming the Go bin directory is on your shell path, you can then run the compiler from any directory like so:

```
pigeon somefile.pigeon        # compile and run somefile.pigeon as a Pigeon program
pigeon somefile.gopigeon      # compile and run somefile.gopigeon as a GoPigeon program
```
