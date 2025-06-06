package main

import (
	"fmt"
	"unsafe"
)

type my struct {
	x    int
	y    int
	data [1000]byte
}

type my2 struct {
	x    int
	y    int
	data []byte
}

func main() {
	m := my{}
	m2 := my2{}
	m2.data = make([]byte, 1000)
	fmt.Println(unsafe.Sizeof(m))
	fmt.Println(unsafe.Sizeof(m2))

	x := my{}
	x.data[2] = 13
	y := x
	fmt.Println(y.data[2])
}
