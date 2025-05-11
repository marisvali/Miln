package go_basics

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math"
	"reflect"
	"testing"
	"unsafe"
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
	// var a, b BigStruct
	var v []BigStruct
	v = append(v, BigStruct{})
	v = append(v, BigStruct{})
	for i := range v {
		// for i := range v {
		// for i := 0; i < len(v); i++ {
		print(i)
		print(v[i].v2)
		// print(c.v2)
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

	// m := []reflect.Kind{reflect.String, reflect.Int, reflect.Float32}
	var m map[reflect.Kind]int
	m = make(map[reflect.Kind]int)
	m[reflect.Int] = 2
	m[reflect.String] = 5
	println(m[reflect.Int])

	m2 := make(map[reflect.Type]int)
	m2[reflect.TypeOf(A{})] = 13
	// m2[A] = 13

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

type Type1 struct {
	A float64
}

type Type2 struct {
	B bool
}

func funcx(x interface{}) {
	println(x)
}

func TestInterfaces(t *testing.T) {
	var t1 Type1
	var t2 Type2
	var t3 int
	funcx(t1)
	funcx(t2)
	funcx(t3)
	assert.True(t, true)
}

func TestString(t *testing.T) {
	var s1, s2 string
	s1 = "绝对不是。"
	println(len(s1))
	println(string(s1[9:15]))
	println(s1, s2)

	var s3, s4 string
	s3 = "a"
	s3 = s3 + "B"
	println(s3)
	println(s4)
	assert.True(t, true)
}

func TestArray(t *testing.T) {
	a1 := [3]int{1, 2, 3}
	a2 := [3]int{5, 6, 7}
	fmt.Println(cap(a2), a2)
	a2 = a1
	a2[0] = 4
	if a1 == a2 {
		fmt.Println("Equal")
	} else {
		fmt.Println("Not equal")
	}
	assert.True(t, true)
}

func TestSlice(t *testing.T) {
	primes := [5]int{2, 3, 5, 7, 11}
	fmt.Println(cap(primes), primes)
	var s []int = primes[1:4]
	fmt.Println(cap(s), s)
	s = append(s, 99)
	fmt.Println(cap(s), s)
	s = append(s, 124)
	fmt.Println(cap(s), s)
	s = append(s, 932)
	fmt.Println(cap(s), s)
	fmt.Println(cap(primes), primes)
}

func TestSlice2(t *testing.T) {
	arr := [2]int{2, 3}
	s := arr[0:0]
	fmt.Println(cap(arr), arr)
	for i := 0; i < 10; i++ {
		s = append(s, 932)
	}
	s[0] = 13
	fmt.Println(cap(arr), arr)
	fmt.Println(cap(s), s)
}

func TestArrayOfString(t *testing.T) {
	a := [3]string{"abc", "tjoijoijoijokoijojoijoijoijoijiiuhiuhiuhjoijres", "barki"}
	fmt.Println(cap(a), len(a), a)
	for i := range a {
		addr := uintptr(unsafe.Pointer(&a[i]))
		fmt.Println(addr)
	}
	assert.True(t, true)
}

func TestMap(t *testing.T) {
	m := make(map[string]int)

	m["Answer"] = 42
	fmt.Println("The value:", m["Answer"])

	m["Answer"] = 48
	fmt.Println("The value:", m["Answer"])

	delete(m, "Answer")
	fmt.Println("The value:", m["Answer"])

	v, ok := m["Answer"]
	fmt.Println("The value:", v, "Present?", ok)

	vx := m["Answer"]
	fmt.Println("The value:", vx)

	v2, ok2 := m["Answer"]
	fmt.Println("The value:", v2, "Present?", ok2)
}

type ff func(int, int) int

func f1(a, b int) int {
	return a + b
}

func f2(a, b int) int {
	return a * b
}

func TestFuncAsType(t *testing.T) {
	a := 13
	b := 5
	var x ff
	if a < b {
		x = f1
	} else {
		x = f2
	}
	println(x(a, b))
}

type Abser interface {
	Abs() float64
	f1()
	f2()
	f3()
}

func TestInterfaces1(t *testing.T) {
	var a Abser

	fx := MyFloat(-math.Sqrt2)
	v := Vertex{3.1, 4.1, 1.1, 2.1, 3.1}

	a = fx // a MyFloat implements Abser
	println(a)
	a = v // a *Vertex implements Abser
	println(a)

	println(unsafe.Sizeof(a))
	println(a)

	fmt.Println(a.Abs())
}

type MyFloat float64

func (f MyFloat) Abs() float64 {
	if f < 0 {
		return float64(-f)
	}
	return float64(f)
}

func (f MyFloat) f1() {

}
func (f MyFloat) f2() {

}
func (f MyFloat) f3() {

}

type Vertex struct {
	X, Y, Z, T, Q float64
}

func (v Vertex) Abs() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z + v.T + v.Q)
}
func (v Vertex) f1() {

}
func (v Vertex) f2() {

}
func (v Vertex) f3() {

}

type ValueReceiver struct {
	X int
	M []int
}

func (v ValueReceiver) Do() {
	v.X = v.X + 1
	v.M[1] = 13
}

func TestValueReceiver(t *testing.T) {
	var v ValueReceiver
	v.M = make([]int, 3)
	fmt.Println(v.X)
	fmt.Println(v.M[1])
	v.Do()
	fmt.Println(v.X)
	fmt.Println(v.M[1])
}
