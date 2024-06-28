package go_basics

import (
	"github.com/stretchr/testify/assert"
	"reflect"
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

type A struct {
	x int
}

type B struct {
	x int
}

func TestMapOfType(t *testing.T) {
	// regular map
	var v map[string]int
	v = make(map[string]int)
	v["ceva"] = 3
	v["b"] = 5
	println(v["b"])

	//m := []reflect.Kind{reflect.String, reflect.Int, reflect.Float32}
	var m map[reflect.Kind]int
	m = make(map[reflect.Kind]int)
	m[reflect.Int] = 2
	m[reflect.String] = 5
	println(m[reflect.Int])

	m2 := make(map[reflect.Type]int)
	m2[reflect.TypeOf(A{})] = 13
	//m2[A] = 13

	m3 := make(map[interface{}]int)
	m3[A{}] = 4
	m3[B{}] = 5
	println(m3[A{}])
	println(m3[B{}])
	println(m3[A{}])
	println(m3[B{}])

	assert.True(t, true)
}

func f(x int) (int, int) {
	return x + 2, x * 2
}

func TestCreateVars(t *testing.T) {
	a, b := f(5)
	println(a, b)
	a, c := f(7)
	println(a, c)
	assert.True(t, true)
}
