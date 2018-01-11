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
        (super this)                               // super() 
        {this things (new A<Object> 10)}
        {this size 0}

    method append o Object : Integer
        if (eq [this size] [things length])
            (expand this [this size])
        {this things [this size] o}
        {this size (inc [this size])}
        return [this size]

    method getSize : Integer
        return [this size]

    method getThing i Integer : Object
        if (or (lt idx 0) (gte idx [this size]))
            throw (new IndexOutOfBoundsException "Index out of bounds!")
        return [this things idx]

    method setThing o Object idx I
        if (or (lt idx 0) (gte idx [this size]))
            throw (new IndexOutOfBoundsException "Index out of bounds!")
        {this things idx o}
        
    method expand toSize I
        if (gt (add [this size] toSize) [this things length])
            var newArray A<Object> (new A<Object> (add [this size] toSize))
            forinc i I 0 (lt i [this things length])
                {newArray i [this things i]}
            {this things newArray}

    method remove idx I : Object
        var temp Object (getThing this idx)
        forinc i idx (lt i (dec [this size]))
            {this things i [this things (inc i)]}
        {this size (dec [this size])}
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


// JavaPigeon

// will need to decide what set of stdlib classes to include as built-ins


class MyClass extends Object

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


// when running a JavaPigeon program, you specify the class whose main you wish to start off execution


var strs A<Str> (new A<Str> 5)


as i (inc i)


(println [System out] "hi")            // maybe just keep print and prompt operators

interface Dog
    method foo String Cat : Dog


(foo x 3 6)                    // x.foo(3, 6)

global s Str "hi"

func foo s Str i I : Cat
    return (new Cat i s)





try

catch Exception ex


catch bla bla bla

finally



get and set operators for accessing array indexes

var cats A<Cat> (new A<Cat> 9)      // new array of 9 Cats
{cats 3 10}                         // cats[3] = 10
var x I [cats 3]



(len s)          // use len on arrays and strings instead of .length property


class Cat extends Object implements Foo Bar    // implements Foo and Bar
    field i I
    field c Cat
    field s String

    constructor
        return this                    // every constructor must explicitly return this

    constructor i I s String
        {this i i}
        {this s s}
        return this

    method foo s String : Cat
        return (new Cat)

    method bar i I                   // method returns nothing
        return


var c Cat (new Cat)
as c (foo c "hi")



// casting is done by using the type name as operator

var i I 30
var d D (cast D i)     // cast i to Double


no coercions: must always be explicit like Go



multi-dimensional arrays

(new A<A<I>> 5 10)          // new int[5][10]
(new A<A<I>> 5)             // new int[5][]

exception classes

enums

generics

overloading (compile error on any ambiguous call, forcing user to do casts to resolve ambiguities)



// downcast

var m Mammal (new Cat)
var c Cat (cast Cat m)


(instanceof Cat m)    // returns true or false
                      
(cast Cat m)          // returns same instance but now compiler accepts it as a Cat
                      // by 'm' as a Cat, throws exception if 'm' is not a Cat



primitives vs wrappers

I -> Integer
F -> Float
D -> Double
C -> Character
S -> Short
L -> Long
B -> Byte
Bool ->Boolean




no packages, no imports, only one file
class member order: fields, constructors, methods



this is never implicit
no abstract
no interface inheritance
no checked exceptions, no throws clauses (in the generated Java, every method throws Throwable)
constructor return is not implicit: must always write it at end, and any early return must return this




    var ints A<I> (new A<I> 8)   // new array with 8 ints
    // newarr creates array with specified values
    (newarr A<I> 7 2 9 11 -3)              
    (newarr A<Str> 3 "hi" "Yo" "bye")
                                







enums
generics
annotations


var x D 9.0
(new Double x)    // to get wrapped from primitive











class BinaryConverter {
    
    public static void main(String[] args){
        for(int i = -5; i < 33; i++){
            System.out.println(i + ": " + toBinary(i));
            System.out.println(i);
            //always another way
            System.out.println(i + ": " + Integer.toBinaryString(i));
        }
    }
    
    /*
     * pre: none
     * post: returns a String with base10Num in base 2
     */
    public static String toBinary(int base10Num){
        boolean isNeg = base10Num < 0;
        base10Num = Math.abs(base10Num);        
        String result = "";
        
        while(base10Num > 1){
            result = (base10Num % 2) + result;
            base10Num /= 2;
        }
        assert base10Num == 0 || base10Num == 1 : "value is not <= 1: " + base10Num;
        
        result = base10Num + result;
        assert all0sAnd1s(result);
        
        if( isNeg )
            result = "-" + result;
        return result;
    }
    
    /*
     * pre: cal != null
     * post: return true if val consists only of characters 1 and 0, false otherwise
     */
    public static boolean all0sAnd1s(String val){
        assert val != null : "Failed precondition all0sAnd1s. parameter cannot be null";
        boolean all = true;
        int i = 0;
        char c;
        
        while(all && i < val.length()){
            c = val.charAt(i);
            all = c == '0' || c == '1';
            i++;
        }
        return all;
    }
}



class BinaryConverter
    
    staticmethod main args A<Str>
        forinc i I -5 33
            (println (concat i ": " (toBinary BinaryConverter i)))
            (println i)
            //always another way
            (println i (concat ": " (toBinaryString Integer i)))
    
    staticmethod toBinary base10Num I : String
        as isNeg Bool (lt base10Num 0)
        as base10Num (abs Math base10Num)
        as result String ""
        
        while (gt base10Num 1)
            as result (add (mod base10Num 2) result)
            as base10Num (div base10Num 2)
        
        assert (or (eq base10Num 0) (eq base10Num 1)) (concat "value is not <= 1: " base10Num)
        
        as result (concat base10Num result)
        assert (all0sAnd1s BinaryConverter result)
        
        if isNeg
            as result (concat "-" result)
        return result
    
    staticmethod all0sAnd1s val String : Bool
        assert (neq val null) "Failed precondition all0sAnd1s. parameter cannot be null"
        as all Bool true
        as i I 0
        as c C
        
        while (and all (lt i (len val)))
            as c (charAt val i)
            as all (or (eq c '0') (eq c '1'))
            as i (inc i)
        return all



public class MineSweeper
{	private int[][] myTruth;
	private boolean[][] myShow;
	
	public void cellPicked(int row, int col)
	{	if( inBounds(row, col) && !myShow[row][col] )
		{	myShow[row][col] = true;
		
			if( myTruth[row][col] == 0)
			{	for(int r = -1; r <= 1; r++)
					for(int c = -1; c <= 1; c++)
						cellPicked(row + r, col + c);
			}	
		}
	}
	
	public boolean inBounds(int row, int col)
	{	return 0 <= row && row < myTruth.length && 0 <= col && col < myTruth[0].length;
	}
}



class MineSweeper
	field myTruth A<A<I>>
	field myShow A<A<Bool>>
	
	method cellPicked row I col I
		if (and (inBounds this row col) (not [myShow row col]))
			{myShow row col true}
		
			if (eq [myTruth row col] 0)
				forinc r I -1 2
					forinc c I -1 2
						(cellPicked this (add row r) (add col c))
			
	
	method inBounds row I col I
		return (and (lte 0 row) 
            ,(lt row (len myTruth))
            ,(lte 0 col)
            ,(lt col (len [myTruth 0])))


