package main

import "fmt"

func cat() {
	fmt.Println("meow")
	z := recover() // stops panic
	fmt.Println(z)
}

func dog() {
	fmt.Println("bark")
}

func foo() int {
	defer fmt.Println("yo")
	defer cat()
	defer dog()
	panic(nil)
	fmt.Println("end foo")
	return 5
}

// recover is executed in deferred call of foo, so execution resumes after the call to foo
func main() {
	z := foo() // 0 (recovered call returns zero values)
	fmt.Println("after", z)
}
