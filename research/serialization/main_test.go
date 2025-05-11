package main

import (
	"bytes"
	"encoding/gob"
	. "github.com/marisvali/miln/gamelib"
	"github.com/stretchr/testify/assert"
	"log"
	"slices"
	"testing"
)

type S1 struct {
	X string
	Y []Int
}

type S2 struct {
	Z []S1
	H bool
}

type S3 struct {
	T S1
	V S2
}

func TestSerializeS3(t *testing.T) {
	var buf bytes.Buffer        // Stand-in for a buf connection
	enc := gob.NewEncoder(&buf) // Will write to buf.
	dec := gob.NewDecoder(&buf) // Will read from buf.

	var s1 S3
	s1.T.X = "treq"
	s1.T.Y = []Int{I(93), I(1), I(4)}
	s1.V.H = true
	s1.V.Z = make([]S1, 3)
	s1.V.Z[0].X = "meq"
	s1.V.Z[0].Y = []Int{I(93), I(1), I(4)}
	s1.V.Z[1].X = "dewf"
	s1.V.Z[1].Y = []Int{I(2), I(189), I(401)}
	s1.V.Z[2].X = "perfq"
	s1.V.Z[2].Y = []Int{I(333), I(221), I(400)}

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

type WaveData struct {
	SecondsAfterLastWave Int
	NHoundMin            Int
	NHoundMax            Int
}

type SpawnPortalData struct {
	Waves []WaveData
}

type NEntities struct {
	NObstaclesMin    Int
	NObstaclesMax    Int
	SpawnPortalDatas []SpawnPortalData
}

type EnemyParams struct {
	SpawnPortalCooldownMin Int
	SpawnPortalCooldownMax Int

	HoundMaxHealth                 Int
	HoundMoveCooldownMultiplier    Int
	HoundPreparingToAttackCooldown Int
	HoundAttackCooldownMultiplier  Int
	HoundHitCooldown               Int
	HoundHitsPlayer                bool
	HoundAggroDistance             Int
}

type WorldData struct {
	NumRows                 Int
	NumCols                 Int
	NEntitiesPath           string
	EnemyParamsPath         string
	Boardgame               bool
	UseAmmo                 bool
	AmmoLimit               Int
	EnemyMoveCooldown       Int
	EnemiesAggroWhenVisible bool
	NEntities
	EnemyParams
}

func TestSerializeWorldData(t *testing.T) {
	var buf bytes.Buffer        // Stand-in for a buf connection
	enc := gob.NewEncoder(&buf) // Will write to buf.
	dec := gob.NewDecoder(&buf) // Will read from buf.

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

type S9 struct {
	X int
}

func (s *S9) GetX() int {
	return s.X
}

type I1 interface {
	GetX() int
}

type U1 struct {
	Z int
	I I1
}

func TestSerializeInterface(t *testing.T) {
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
