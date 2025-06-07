package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
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

// How much does a slice of 1000 bytes contribute to the size of the struct it
// belongs to?
// 24 bytes (so just the header).
func TestStructSizeWithSlice(t *testing.T) {
	m2 := my2{}
	m2.data = make([]byte, 1000)
	fmt.Println(unsafe.Sizeof(m2))
	assert.Equal(t, int(unsafe.Sizeof(m2)), 2*8+24)
}

// How much does an array of 1000 bytes contribute to the size of the struct it
// belongs to?
// 1000 bytes (so all of it).
func TestStructSizeWithArray(t *testing.T) {
	m := my{}
	fmt.Println(unsafe.Sizeof(m))
	assert.Equal(t, int(unsafe.Sizeof(m)), 2*8+1000)
}

// Does assigning an array copy it, or copy just a reference to the same array?
// Assigning an array copies it.
func TestArrayAssignment(t *testing.T) {
	x := my{}
	x.data[2] = 13
	y := x
	assert.Equal(t, x.data[2], y.data[2])
}

// Can I create a slice based on an existing array?
// Yes.
func TestArrayToSlice(t *testing.T) {
	var d [10]int
	s := d[:]
	fmt.Println(len(s))
	fmt.Println(s)
	assert.Equal(t, len(d), len(s))
}

// Can I create a slice based on an existing array, initialize it at zero-length
// and then rely on the fact that it will keep using the same underlying array
// if I append to the slice, until I reach the limit of the array?
// Yes.
func TestGrowingSliceBasedOnArray(t *testing.T) {
	var d [10]int
	s := d[:0]
	assert.Equal(t, 0, len(s))

	// Add just one and see.
	s = append(s, 13)
	assert.Equal(t, 1, len(s))
	assert.Equal(t, 13, s[0])
	assert.Equal(t, 13, d[0])

	// Add until d is full.
	for i := range len(d) - 1 {
		s = append(s, 13+i)
	}
	assert.Equal(t, 10, len(s))
	assert.Equal(t, 21, s[9])
	assert.Equal(t, 21, d[9])
}

// What happens if I create a slice s based on array d of size n and then I grow
// s until it has size n+1?
// When the slice reaches the limit of the array, a new array is allocated under
// the hood and the slice now points to it.
func TestGrowingSliceBasedOnArrayOverLimit(t *testing.T) {
	var d [10]int
	s := d[:0]

	for i := range len(d) {
		s = append(s, i)
	}

	// Check that array and slice are the same.
	assert.Equal(t, 10, len(d))
	for i := range len(d) {
		assert.Equal(t, d[i], s[i])
	}

	// Check that changing an element in the slice changes it in the array.
	s[3] = 13
	assert.Equal(t, 13, d[3])

	// Grow the slice beyond the length of d.
	s = append(s, 17)
	assert.Equal(t, 11, len(s))

	// Check that changing an element in the slice DOES NOT change it in the
	// array.
	s[3] = 14
	assert.Equal(t, 13, d[3])
	assert.NotEqual(t, d[3], s[3])
}
