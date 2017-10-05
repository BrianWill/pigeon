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


## function variables


## local functions



## closures


## goroutines



## channels



## `select` statements

## bitwise operators


## the standard library http package