package test;

public class MyArrayList<T> {
    private Object[] things;
    private int size;

    public MyArrayList() {
        super();
        this.things = new Object[10];
        this.size = 0;
    }

    public int append(Object o) {
        if (this.size == things.length) {
            this.expand(this.size);
        }
        this.things[this.size] = o;
        this.size++;
        return this.size;
    }

    public int getSize() {
        return this.size;
    }

    public Object getThing(int idx) {
        if (idx < 0 || idx >= this.size) {
            throw new IndexOutOfBoundsException("Index out of bounds!");
        }
        return this.things[idx];
    }

    public void setThing(Object o, int idx) {
        if (idx < 0 || idx >= this.size) {
            throw new IndexOutOfBoundsException("Index out of bounds!");
        }
        things[idx] = o;
    }

    //
    public void expand(int toSize) {
        if (this.size + toSize > this.things.length) {
            Object[] newArray = new Object[this.size + toSize];
            for (int i = 0; i < this.things.length; i++) {
                newArray[i] = this.things[i];
            }
            this.things = newArray;
        }
    }

    public Object remove(int idx) {
        Object temp = this.getThing(idx);
        for (int i = idx; i < this.size - 1; i++) {
            this.things[i] = this.things[i + 1];
        }
        this.size--;
        return temp;
    }

    public static void main(String[] args) {
        MyArrayList list = new MyArrayList();
        list.append(9);
        list.append("hi");
        list.append("bye");
        list.remove(1);
        System.out.println(list.getSize());
        System.out.println(list.getThing(0));
        System.out.println(list.getThing(1));
        MyArrayList list2 = new MyArrayList();
        System.out.println(list2.getSize());
    }
}

package test

class MyArrayList extends Object
    field things A<Object> 
    field size Integer

    constructor
        (super #)                               // super() 
        {# things (new A<Object> 10)}
        {# size 0}

    method append o Object : Integer
        if (eq [# size] [things length])
            (expand # [# size])
        {# things [# size] o}
        {# size (inc [# size])}
        return [# size]

    method getSize : Integer
        return [th size]

    method getThing i Integer : Object
        if (or (lt idx 0) (gte idx [# size]))
            throw (new IndexOutOfBoundsException "Index out of bounds!")
        return [# things idx]

    method setThing o Object idx I
        if (or (lt idx 0) (gte idx [# size]))
            throw (new IndexOutOfBoundsException "Index out of bounds!")
        {# things idx o}
        
    method expand toSize I
        if (gt (add [# size] toSize) [# things length])
            var newArray A<Object> (new A<Object> (add [# size] toSize))
            forinc i I 0 (lt i [# things length])
                {newArray i [# things i]}
            {# things newArray}

    method remove idx I : Object
        var temp Object (getThing # idx)
        forinc i idx (lt i (dec [# size]))
            {# things i [# things (inc i)]}
        {# size (dec [# size])}
        return temp

    staticmethod main args A<Str>
        var list MyArrayList (new MyArrayList)
        (append list 9)
        (append list "hi")
        (append list "bye")
        (remove list 1)
        (println out (getSize list))          
        (println out (getThing list 0))
        (println out (getThing list 1))
        var list2 MyArrayList (new MyArrayList)
        (println out (getSize list2))









package tcpchat;

import java.io.*;
import java.net.*;

public class ChatServer {

    private static final int PORT = 3000;

    public static void main(String argv[]) throws Exception {
        ServerSocket server = new ServerSocket(ChatServer.PORT);
        while (true) {
            Socket conn = server.accept();
            BufferedReader input = new BufferedReader(new InputStreamReader(conn.getInputStream()));
            DataOutputStream output = new DataOutputStream(conn.getOutputStream());

            String message = input.readLine();
            System.out.println("Received: " + message);
            output.writeBytes(message.toUpperCase() + .\n.);

            conn.close();
        }
    }
}



// Jabiru: name of a bird starting with .ja.
// Jaybird
// **Jacobin (a fancy breed of pigeon)

package foo.bar

import java.io.BufferedReader
import java.io.InputStreamReader
import java.net.ServerSocket
import java.net.Socket
import java.lang.out
import foo.bar.Apple

// cannot import name if it creates conflict, must instead fully qualify name 
// by prepending package name and //, e.g. java.lang.out
// (want to make it visually clear what is a package name and what is a class/interface/func/global)

class MyClass

    staticfield port I 3000

    staticmethod main args A<Str>
        var server ServerSocket (new ServerSocket port)
        while true
            var conn Socket (accept server)
            var isr InputStreamReader (new InputStreamReader (getInputStream conn))
            var input BufferedReader (new BufferedReader isr)
            var output DataOutputStream (new DataOutputStream (getOutputStream conn))

            var message Str (readLine input)
            (println out (add "Received: " message))
            (writeBytes output (add (toUpperCase message)))
            (close conn)


var strs A<Str> (new A<Str> 5)



as i (inc i)


interface Dog
    method foo String Cat : Dog


(foo x 3 6)                    // x.foo(3, 6)

global s Str "hi"

func foo s Str i I : Cat
    return (Cat i s)





try

catch Exception ex


catch bla bla bla

finally



get and set operators for accessing array indexes

var Cat[] cats (new Cat[] 9)      // new array of 9 Cats
(set cats 3 10)                  // cats[3] = 10
var Integer x (get cats 3)



class Cat extends Object implements Foo Bar    // implements Foo and Bar
    field I i
    field Cat c
    field String s

    constructor
        return this          // every constructor must explicitly return this

    constructor I i String s
        as ~i i
        as ~s s
        return this

    method foo String s : Cat
        return (new Cat)

    // override method
    override bar I i         // method returns nothing
        return


var Cat c (new Cat)
as c (.foo c "hi")



// casting is done by using the type name as operator

var I i 30
var D d (D i)     // cast i to Double


no coercions: must always be explicit like Go






// downcast

var Mammal m (new Cat)
var Cat c (Cat m)


(instanceof Cat m)    // returns true or false
(Cat m)               // returns same instance but now compiler accepts it as a Cat
                      // by 'm' as a Cat, throws exception if 'm' is not a Cat

I -> Integer
F -> Float
D -> Double
C -> Character
S -> Short
L -> Long
B -> Byte




no primitives, only primitive-wrappers
one dir = one package
    no rule about correspondence between file names and class names
    files must end in .jacobin
source file directory and name does not matter: only the package statement matters
required written order: package statement, imports, globals, interfaces, classes, funcs (main last)
class member order: fields, constructors, methods
compiler just compiles every .jacobin file under the specified directory (recursively)
overriding methods must be denoted by `override` instead of `method`
no glob import
use [] only to denote array type, not index arrays (use `get` and `set` operators instead)
`this` is never implicit, but we use ~foo as shorthand for `this.foo`

no way to populate array on creation
no abstract
no interface inheritance
no checked exceptions, no throws clauses (in the generated Java, every method throws Throwable)
only types can begin uppercase
dot only means access field or call method
assignment is a statement, not an expression

functions and globals all added as statics to one class in the package
no nested classes


    var I[] ints (new I[] 8)
    (fill ints 7 2 9 11 -3)   // assign these vals to first indexes of the array and returns nothing
                                  // throws an exception if index out of bounds


a constructor is just a method with same name as its class, implicit return 
    type (same as class), and it must return this
(the return is not implicit: must always write it at end, and any early return must return this)
calling a constructor without . implicitly passes a new instance



overloading? methods only, not functions
no enums
no generics
no annotations

can import functions from other packages

(Double x)    // to convert between primitive types, use constructor (x here can be any primitive wrapper)

package names use / instead of . to separate components of the name 
