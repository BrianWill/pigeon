
// (Str )               // create string from list, array, or slice of strings, or from list, array, or slice of Bytes or Integers

// func prefix s Str pre Str : Bool


//getchar
//getrune
//charlist
//runelist

func substr s Str start I end I : Str
    locals s2 Str
    forinc i I start end
        if (gte i (len s))
            break
        as s2 (concat s2 (getchar s i))
    return s2

func join strings L<Str> separator Str : Str
    locals s Str
    foreach i I v Str strings
        as s (concat s v)
    return s

func trim s Str cutset L<Str> : Str
    locals start I end I startFound Bool
    foreach i I ch Str (charlist s)
        if (containsAny ch cutset)
            if startFound
                break
            as start (inc start)
            as end (inc end)
        else
            as end (inc end)
            as startFound true
    return (substr s start end)

func contains s Str substr Str : Bool
    locals match Bool
    forinc i I 0 (inc (sub (len s) (len substr)))
        as match true
        forinc j I 0 (len substr)
            if (neq (getchar s j) (getchar substr j))
                as match false
                break
        if match
            return true
    return false

func index s Str substr Str : I
    locals match Bool
    forinc i I 0 (inc (sub (len s) (len substr)))
        as match true
        forinc j I 0 (len substr)
            if (neq (getchar s j) (getchar substr j))
                as match false
                break
        if match
            return i
    return -1

func containsAny s Str chars L<Str> : Bool
    forinc i I 0 (inc (sub (len s) (len chars)))
        forinc j I 0 (len chars)
            if (eq (getchar s j) (get chars j))
                return true
    return false

// check that len on strings returns rune count rather than byte count

func foo a L<I>
    (push a 400)

func main
    locals a L<I> b L<I>
    (push a 5)
    as b a
    (push b -99)
    (set b 0 -7)
    (foo a)
    (println a b)
    (println "fun with strings")


