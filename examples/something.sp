
func sum a N b N : N
    return (add a b)

func giveNums : N N
    locals a N
    as a 5
    return a 7

func doNothing
    (print "hi")

struct Foo
    bar Str
    cat N

interface Roger
    foo N Str : N Foo

method foo f Foo a N b Str : N Foo
    (print "hi")
    return 3 (Foo "hi" 5)

method doStuff f Foo apple N : Str
    return "hi"

// typeswitch
// break, continue
// check lists, _newList

func testing
    locals n N p P<N>
    as p (ref n)
    as n (dr p)

func main
    locals a Foo b L<Str> c M<N Str> d Roger e Bool
    (append b "hi")
    as a (Foo "hi" 3)
    foreach i N v Str b
        (print i v)
    as d a
    (.foo d 5 "hi")
    (.doStuff a 4)
    typeswitch d
    case a Foo
        (println "d is Foo")
    default
        (println "d is not Foo")
    as b (L<Str> "yo" "byte")
    as c (M<N Str> 5 "hi" 9 "yo")
    as b[0] "asdf"
    as c[3] "hi"
    (print "bla")