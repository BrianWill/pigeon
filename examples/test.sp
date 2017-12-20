// assume Fruit is an interface implemented by structs Banana, Orange, and others
func foo f Fruit 
    // the order of the cases does not matter
    typeswitch f
    case b Banana
        (print "banana")
        // ... executed if 'f' references an Banana
    case o Orange
        (print "orange")
        // ... executed if 'f' references an Orange
    // the default clause (if present) must come last
    default 
        (print "not banana or orange")
        // ... executed if 'f' references something other than a Banana or Orange

struct Banana
    x I

struct Orange
    y I

struct Apple    
    z Str

method foo o Orange
    (println o)

method foo b Banana
    (println b)

method foo a Apple
    (println a)

interface Fruit
    foo

// func main
//     locals b Banana o Orange a Apple
//     (foo a)

func main
    (println (eq 53 53 53))                  // true (all operands equal)
    (println (eq 53 4 53))                   // false (not all operands equal)
    // (println (eq "hi" 53 53))                // compile error: operands must be matching types
    (println (eq "hi" "hi" "hi" "yo" "hi"))       // true (all operands equal)
    // (println (eq "hi"))                      // compile error: expecting at least two operands