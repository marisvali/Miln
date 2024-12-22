package main

import (
	"fmt"
	"unsafe"
)

func main() {
	var v [1000]int
	v[0] = 3
	// println(v[0])
	// println(v[132])
	// var x *int
	// x = &v[30]

	println(&v[0])
	println(&v[999])
	a1 := &v[0]
	a2 := &v[999]
	ptr1 := unsafe.Pointer(a1)
	ptr2 := unsafe.Pointer(a2)

	// Calculate the difference in bytes
	diff := uintptr(ptr2) - uintptr(ptr1)

	// Print the difference
	fmt.Printf("Difference between pointers: %d bytes\n", diff)

	// Calculate the difference in terms of elements
	elementSize := unsafe.Sizeof(*a1)
	fmt.Printf("Difference in elements: %d\n", diff/elementSize)
}
