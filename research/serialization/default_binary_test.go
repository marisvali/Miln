package main

import (
	"bytes"
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStructWithArray1_default(t *testing.T) {
	type SArr struct {
		X [2]int64
	}

	var v1 SArr
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, v1)
	assert.NoError(t, err)

	var v2 SArr
	err = binary.Read(buf, binary.LittleEndian, &v2)
	assert.NoError(t, err)
	assert.Equal(t, v1, v2)
}

func TestStructWithArray2_default(t *testing.T) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, s1Arr)
	assert.NoError(t, err)
	var s2Arr S3Arr
	err = binary.Read(buf, binary.LittleEndian, &s2Arr)
	assert.NoError(t, err)
	assert.Equal(t, s1Arr, s2Arr)
}
