# StaticPigeon (tier 2)

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
