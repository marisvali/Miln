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
