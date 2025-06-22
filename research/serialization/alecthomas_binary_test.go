package main

import (
	"github.com/alecthomas/binary"
	. "github.com/marisvali/miln/gamelib"
	"github.com/stretchr/testify/assert"
	"log"
	"slices"
	"testing"
)

func TestSerializeS3_alecthomas(t *testing.T) {
	// Encode (send) some values.
	buf, err := binary.Marshal(s1)
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

func TestSerializeWorldData_alecthomas(t *testing.T) {
	// Encode (send) some values.
	var w1 WorldData
	w1.EnemyMoveCooldown = I(341)
	w1.EnemyParamsPath = "dj23d"
	w1.SpawnPortalDatas = make([]SpawnPortalData, 2)
	w1.SpawnPortalDatas[0].Waves = make([]WaveData, 2)
	w1.SpawnPortalDatas[0].Waves[0].SecondsAfterLastWave = I(194)
	w1.SpawnPortalDatas[0].Waves[1].NHoundMin = I(12)
	w1.SpawnPortalDatas[1].Waves = make([]WaveData, 1)
	w1.SpawnPortalDatas[0].Waves[0].NHoundMax = I(111222)
	w1.HoundHitsPlayer = true

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

func TestSerializeInterface_alecthomas(t *testing.T) {
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

func TestSerializeSameDataDifferentNames_alecthomas(t *testing.T) {
	buf, err := binary.Marshal(s1)
	assert.NoError(t, err)
	var s2Twin S3Twin
	err = binary.Unmarshal(buf, &s2Twin)
	assert.NoError(t, err)
	assert.Equal(t, s1Twin, s2Twin)
}

// BenchmarkSerializeWorldData_alecthomas-12    	   86332	     13243 ns/op
func BenchmarkSerializeWorldData_alecthomas(b *testing.B) {
	for b.Loop() {
		// Encode (send) some values.
		buf, _ := binary.Marshal(w1)

		// Decode (receive) and print the values.
		var w2 WorldData
		_ = binary.Unmarshal(buf, &w2)
	}
}
