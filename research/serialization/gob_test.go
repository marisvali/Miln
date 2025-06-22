package main

import (
	"bytes"
	"encoding/gob"
	"github.com/stretchr/testify/assert"
	"log"
	"slices"
	"testing"
)

func TestSerializeS3_gob(t *testing.T) {
	var buf bytes.Buffer        // Stand-in for a buf connection
	enc := gob.NewEncoder(&buf) // Will write to buf.
	dec := gob.NewDecoder(&buf) // Will read from buf.

	// Encode (send) some values.
	err := enc.Encode(s1)
	if err != nil {
		log.Fatal("encode error:", err)
	}

	// Decode (receive) and print the values.
	var s2 S3
	err = dec.Decode(&s2)
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

func TestSerializeWorldData_gob(t *testing.T) {
	var buf bytes.Buffer        // Stand-in for a buf connection
	enc := gob.NewEncoder(&buf) // Will write to buf.
	dec := gob.NewDecoder(&buf) // Will read from buf.

	// Encode (send) some values.
	err := enc.Encode(w1)
	if err != nil {
		log.Fatal("encode error:", err)
	}

	// Decode (receive) and print the values.
	var w2 WorldData
	err = dec.Decode(&w2)
	if err != nil {
		log.Fatal("decode error:", err)
	}

	assert.Equal(t, w1, w2)
}

func TestSerializeInterface_gob(t *testing.T) {
	var buf bytes.Buffer        // Stand-in for a buf connection
	enc := gob.NewEncoder(&buf) // Will write to buf.
	dec := gob.NewDecoder(&buf) // Will read from buf.

	// Encode (send) some values.
	gob.Register(&S9{})
	var v S9
	v.X = 149
	var u1 U1
	u1.Z = 3
	u1.I = &v
	err := enc.Encode(u1)
	if err != nil {
		log.Fatal("encode error:", err)
	}

	// Decode (receive) and print the values.
	var u2 U1
	err = dec.Decode(&u2)
	if err != nil {
		log.Fatal("decode error:", err)
	}

	assert.Equal(t, u1, u2)
}

func TestSerializeSameDataDifferentNames_gob(t *testing.T) {
	var buf bytes.Buffer        // Stand-in for a buf connection
	enc := gob.NewEncoder(&buf) // Will write to buf.
	dec := gob.NewDecoder(&buf) // Will read from buf.

	err := enc.Encode(s1)
	assert.NoError(t, err)
	var s2Twin S3Twin
	err = dec.Decode(&s2Twin)
	assert.NoError(t, err)
	assert.Equal(t, s1Twin, s2Twin)
}

// BenchmarkSerializeWorldData_gob-12    	   21307	     56471 ns/op
func BenchmarkSerializeWorldData_gob(b *testing.B) {
	for b.Loop() {
		var buf bytes.Buffer        // Stand-in for a buf connection
		enc := gob.NewEncoder(&buf) // Will write to buf.
		dec := gob.NewDecoder(&buf) // Will read from buf.

		// Encode (send) some values.
		_ = enc.Encode(w1)

		// Decode (receive) and print the values.
		var w2 WorldData
		_ = dec.Decode(&w2)
	}
}
