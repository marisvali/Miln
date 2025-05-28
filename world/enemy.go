package world

import (
	. "github.com/marisvali/miln/gamelib"
)

type Enemy interface {
	Step(w *World)
	Pos() Pt
	Alive() bool
	Health() Int
	MaxHealth() Int
	Clone() Enemy
	Vulnerable(w *World) bool
	State() string
	TargetPos() Pt
}

type EnemyState int

const (
	Searching EnemyState = iota
	PreparingToAttack
	Attacking
	Hit
	Dead
)

var enemyStateName = map[EnemyState]string{
	Searching:         "Searching",
	PreparingToAttack: "PreparingToAttack",
	Attacking:         "Attacking",
	Hit:               "Hit",
	Dead:              "Dead",
}

type EnemyBase struct {
	pos       Pt
	health    Int
	maxHealth Int
	targetPos Pt
}

func (e *EnemyBase) Pos() Pt {
	return e.pos
}

func (e *EnemyBase) Health() Int {
	return e.health
}

func (e *EnemyBase) MaxHealth() Int {
	return e.maxHealth
}

func (e *EnemyBase) Alive() bool {
	return e.health.IsPositive()
}

func (e *EnemyBase) TargetPos() Pt { return e.targetPos }

func getObstaclesAndEnemies(w *World) (m MatBool) {
	m = w.Obstacles.Clone()
	m.Add(w.EnemyPositions())
	m.Add(w.EnemyTargetPositions())
	return
}

func (e *EnemyBase) beamJustHit(w *World) bool {
	if !w.Beam.Idx.Eq(w.BeamMax) { // the fact that this is required shows me
		// I need to structure this stuff differently.
		return false
	}
	return w.WorldPosToTile(w.Beam.End) == e.pos
}
