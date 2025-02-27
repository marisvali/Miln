package world

import (
	. "github.com/marisvali/miln/gamelib"
)

type Enemy interface {
	Step(w *World)
	Pos() Pt
	Alive() bool
	FreezeCooldownIdx() Int
	FreezeCooldown() Int
	Health() Int
	MaxHealth() Int
	Clone() Enemy
	Vulnerable(w *World) bool
	State() string
}

type EnemyBase struct {
	pos               Pt
	health            Int
	maxHealth         Int
	freezeCooldownIdx Int
	freezeCooldown    Int
}

func (e *EnemyBase) Pos() Pt {
	return e.pos
}

func (e *EnemyBase) FreezeCooldownIdx() Int {
	return e.freezeCooldownIdx
}

func (e *EnemyBase) FreezeCooldown() Int {
	return e.freezeCooldown
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

func (e *EnemyBase) State() string { return "NotUsed" }

func (e *EnemyBase) goToPlayer(w *World, m MatBool) {
	path := FindPath(e.pos, w.Player.Pos(), m.Matrix, false)
	if len(path) > 1 {
		e.pos = path[1]
		if e.pos.Eq(w.Player.Pos()) {
			w.Player.Hit()
		}
	}
}

func getObstaclesAndEnemies(w *World) (m MatBool) {
	m = w.Obstacles.Clone()
	m.Add(w.EnemyPositions())
	return
}

func (e *EnemyBase) beamJustHit(w *World) bool {
	if !w.Beam.Idx.Eq(w.BeamMax) { // the fact that this is required shows me
		// I need to structure this stuff differently.
		return false
	}
	return w.WorldPosToTile(w.Beam.End) == e.pos
}

func (e *EnemyBase) move(w *World, m MatBool) {
	if w.EnemyMoveCooldown.IsPositive() && w.EnemyMoveCooldownIdx == ZERO && w.Player.OnMap {
		e.goToPlayer(w, m)
	}
}
