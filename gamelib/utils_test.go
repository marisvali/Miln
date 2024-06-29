package gamelib

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestZip(t *testing.T) {
	data1 := []byte("some kind of data1 aaaaaaaaaaaaaaaaaaaaaaa")
	data2 := []byte("some kind of data1 baaaaaaaaaaaaaaaaaaaaaa")
	data3 := []byte("some kind of data1 baaaaaaaaaaaaaaaaaaaaaa")
	zippedData1 := Zip(data1)
	unzippedData1 := Unzip(zippedData1)
	zippedData2 := Zip(data2)
	unzippedData2 := Unzip(zippedData2)
	zippedData3 := Zip(data3)
	assert.Equal(t, data1, unzippedData1)
	assert.Equal(t, data2, unzippedData2)
	assert.NotEqual(t, zippedData1, zippedData2)
	assert.Equal(t, zippedData2, zippedData3)
}

func TestMatrixFromString(t *testing.T) {
	mapping := map[byte]Int{'x': ONE}

	expected1 := NewMatrix[Int](IPt(3, 2))
	expected1.Set(IPt(0, 1), ONE)
	result1 := MatrixFromString("\nabc\nxyz", mapping)
	assert.Equal(t, expected1, result1)

	expected2 := NewMatrix[Int](IPt(7, 3))
	expected2.Set(IPt(4, 0), ONE)
	expected2.Set(IPt(2, 1), ONE)
	result2 := MatrixFromString(`
----x--
--x----
-------
`, mapping)
	assert.Equal(t, expected2, result2)
}

func TestFloodFill(t *testing.T) {
	m1 := MatrixFromString(`
x---x--
--x-xx-
--x----
`, map[byte]Int{'x': ONE})

	expected1 := MatrixFromString(`
x---x--
--x-xx-
--x----
`, map[byte]Int{'x': ONE, '-': TWO})

	result1 := m1.Clone()
	FloodFill(result1, IPt(1, 0), TWO)
	assert.Equal(t, expected1, result1)

	expected2 := MatrixFromString(`
x---a--
--x-aa-
--x----
`, map[byte]Int{'x': ONE, 'a': TWO})
	result2 := m1.Clone()
	FloodFill(result2, IPt(5, 1), TWO)
	assert.Equal(t, expected2, result2)
}
