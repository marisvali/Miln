package closures

import (
	"fmt"
	"testing"
)

// Q: Does a function closure/lambda function/anonymous function copy variables
// from the context, or does it reference them (pointer/reference semantics)?
// A: It uses reference semantics.
func TestInterfaceCasting(t *testing.T) {
	i := 1
	fmt.Println(i)
	func() {
		fmt.Println("hi")
		fmt.Println(i)
		i = 2
		fmt.Println(i)
	}()
	fmt.Println("hello")
	fmt.Println(i)
}
