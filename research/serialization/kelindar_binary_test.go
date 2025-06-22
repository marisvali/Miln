package main

import (
	"bytes"
	"github.com/kelindar/binary"
	"github.com/stretchr/testify/assert"
	"log"
	"slices"
	"testing"
)

func TestSerializeS3_kelindar(t *testing.T) {
	// Encode (send) some values.
	buf, err := binary.Marshal(&s1)
	if err != nil {
		log.Fatal("encode error:", err)
	}

	// Decode (receive) and print the values.
	var s2 S3
	err = binary.Unmarshal(buf, &s2)
	if err != nil {
		log.Fatal("decode error:", err)
	}

	assert.Equal(t, s1.T.X, s2.T.X)
	assert.True(t, slices.Equal(s1.T.Y, s2.T.Y))
	assert.Equal(t, s1.V.H, s2.V.H)
	assert.Equal(t, s1.V.Z[0].X, s2.V.Z[0].X)
	assert.True(t, slices.Equal(s1.V.Z[0].Y, s2.V.Z[0].Y))
	assert.Equal(t, s1.V.Z[1].X, s2.V.Z[1].X)
	assert.True(t, slices.Equal(s1.V.Z[1].Y, s2.V.Z[1].Y))
	assert.Equal(t, s1.V.Z[2].X, s2.V.Z[2].X)
	assert.True(t, slices.Equal(s1.V.Z[2].Y, s2.V.Z[2].Y))
}

func TestSerializeWorldData_kelindar(t *testing.T) {
	// Encode (send) some values.
	buf, err := binary.Marshal(w1)
	if err != nil {
		log.Fatal("encode error:", err)
	}

	// Decode (receive) and print the values.
	var w2 WorldData
	err = binary.Unmarshal(buf, &w2)
	if err != nil {
		log.Fatal("decode error:", err)
	}

	assert.Equal(t, w1, w2)
}

func TestSerializeInterface_kelindar(t *testing.T) {
	// Encode (send) some values.
	var v S9
	v.X = 149
	var u1 U1
	u1.Z = 3
	u1.I = &v
	buf, err := binary.Marshal(&u1)
	if err != nil {
		log.Fatal("encode error:", err)
	}

	// Decode (receive) and print the values.
	var u2 U1
	err = binary.Unmarshal(buf, &u2)
	if err != nil {
		log.Fatal("decode error:", err)
	}

	assert.Equal(t, u1, u2)
}

func TestSerializeSameDataDifferentNames_kelindar(t *testing.T) {
	buf, err := binary.Marshal(s1)
	assert.NoError(t, err)
	var s2Twin S3Twin
	err = binary.Unmarshal(buf, &s2Twin)
	assert.NoError(t, err)
	assert.Equal(t, s1Twin, s2Twin)
}

func TestDeseriaizingInSteps_kelindar(t *testing.T) {
	buf, err := binary.Marshal(s1)
	assert.NoError(t, err)
	var s2 S3
	buf2 := bytes.NewBuffer(buf)
	d1 := binary.NewDecoder(buf2)
	err = d1.Decode(&s2.T)
	assert.NoError(t, err)
	d2 := binary.NewDecoder(buf2)
	err = d2.Decode(&s2.V)
	assert.NoError(t, err)
	assert.Equal(t, s1, s2)
}

func TestStructWithArray1_kelindar(t *testing.T) {
	type SArr struct {
		X [2]int64
	}

	var v1 SArr
	buf, err := binary.Marshal(v1)
	assert.NoError(t, err)

	var v2 SArr
	err = binary.Unmarshal(buf, &v2)
	assert.NoError(t, err)
	assert.Equal(t, v1, v2)
}

func TestStructWithArray2_kelindar(t *testing.T) {
	buf, err := binary.Marshal(s1Arr)
	assert.NoError(t, err)
	var s2Arr S3Arr
	err = binary.Unmarshal(buf, &s2Arr)
	assert.NoError(t, err)
	assert.Equal(t, s1Arr, s2Arr)
}

// BenchmarkSerializeWorldData_kelindar-12    	  478526	      2504 ns/op
func BenchmarkSerializeWorldData_kelindar(b *testing.B) {
	for b.Loop() {
		// Encode (send) some values.
		buf, _ := binary.Marshal(w1)

		// Decode (receive) and print the values.
		var w2 WorldData
		_ = binary.Unmarshal(buf, &w2)
	}
}
