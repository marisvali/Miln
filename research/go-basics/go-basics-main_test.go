package go_basics

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type SomeStruct struct {
	X int
	Y int
	Z int
}

func TestEqual(t *testing.T) {
	var a, b SomeStruct
	a.X = 2
	a.Y = 3
	a.Z = 4
	b.X = 2
	b.Y = 3
	b.Z = 4
	assert.True(t, a == b)
}

type BigStruct struct {
	v1 [100000000]int
	v2 int
}

func TestCopy(t *testing.T) {
	//var a, b BigStruct
	var v []BigStruct
	v = append(v, BigStruct{})
	v = append(v, BigStruct{})
	for i, _ := range v {
		//for i := range v {
		//for i := 0; i < len(v); i++ {
		print(i)
		print(v[i].v2)
		//print(c.v2)
	}
	assert.True(t, true)
}
