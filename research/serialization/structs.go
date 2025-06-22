package main

import . "github.com/marisvali/miln/gamelib"

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

var s1 = S3{
	T: S1{X: "treq", Y: []Int{I(93), I(1), I(4)}},
	V: S2{H: true, Z: []S1{
		{X: "meq", Y: []Int{I(93), I(1), I(4)}},
		{X: "dewf", Y: []Int{I(2), I(189), I(401)}},
		{X: "perfq", Y: []Int{I(333), I(221), I(400)}},
	}},
}

type S1Twin struct {
	XTwin string
	YTwin []Int
}

type S2Twin struct {
	ZTwin []S1Twin
	HTwin bool
}

type S3Twin struct {
	TTwin S1Twin
	VTwin S2Twin
}

var s1Twin = S3Twin{
	TTwin: S1Twin{XTwin: "treq", YTwin: []Int{I(93), I(1), I(4)}},
	VTwin: S2Twin{HTwin: true, ZTwin: []S1Twin{
		{XTwin: "meq", YTwin: []Int{I(93), I(1), I(4)}},
		{XTwin: "dewf", YTwin: []Int{I(2), I(189), I(401)}},
		{XTwin: "perfq", YTwin: []Int{I(333), I(221), I(400)}},
	}},
}

type S1Arr struct {
	X Int
	Y [3]Int
}

type S2Arr struct {
	Z [3]S1Arr
	H bool
}

type S3Arr struct {
	T S1Arr
	V S2Arr
}

var s1Arr = S3Arr{
	T: S1Arr{X: I(12), Y: [3]Int{I(93), I(1), I(4)}},
	V: S2Arr{H: true, Z: [3]S1Arr{
		{X: I(12), Y: [3]Int{I(93), I(1), I(4)}},
		{X: I(12), Y: [3]Int{I(2), I(189), I(401)}},
		{X: I(12), Y: [3]Int{I(333), I(221), I(400)}},
	}},
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

var w1 = WorldData{
	EnemyMoveCooldown: I(341),
	EnemyParamsPath:   "dj23d",
	NEntities: NEntities{
		SpawnPortalDatas: []SpawnPortalData{
			{Waves: []WaveData{{SecondsAfterLastWave: I(194)}, {NHoundMin: I(12)}}},
			{Waves: []WaveData{{NHoundMax: I(111222)}}},
		},
	},
	EnemyParams: EnemyParams{HoundHitsPlayer: true},
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
