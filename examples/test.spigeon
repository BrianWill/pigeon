func sum a I64 b I64 : I64
    return (add a b)

func main
    locals ch Ch<I64> something A<I64 10>
    localfunc adam a I64 : I64
        return (sum a 9)
    foreach x I64 y I64 something
        (print "hi")
    select
    sending ch 3
        (print "YO")
    rcving i I64 ch
        (print "YO")
    default
        (print "YO")
    (println (adam 2))
