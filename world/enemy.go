package world

import (
	. "github.com/marisvali/miln/gamelib"
)

var GremlinMoveCooldown = I(30)
var GremlinFreezeCooldown = I(30)
var GremlinMaxHealth = I(1)
var HoundMoveCooldown = I(50)
var HoundFreezeCooldown = I(30)
var HoundMaxHealth = I(4)
var PillarMoveCooldown = I(60)
var PillarFreezeCooldown = I(200)
var PillarMaxHealth = I(3)
var KingMoveCooldown = I(30)
var KingFreezeCooldown = I(30)
var KingMaxHealth = I(5)
var QuestionMaxHealth = I(1)
var SpawnPortalCooldown = I(60)

type Enemy interface {
	Step(w *World)
	Pos() Pt
	Alive() bool
	FreezeCooldownIdx() Int
	FreezeCooldown() Int
	Health() Int
	MaxHealth() Int
}

type EnemyBase struct {
	pos               Pt
	health            Int
	maxHealth         Int
	freezeCooldownIdx Int
	freezeCooldown    Int
	moveCooldownIdx   Int
	moveCooldown      Int
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

func (e *EnemyBase) goToPlayer(w *World, m Matrix) {
	path := FindPath(e.pos, w.Player.Pos, m)
	if len(path) > 1 {
		e.pos = path[1]
		if e.pos.Eq(w.Player.Pos) {
			w.Player.Hit()
		}
	}
}

func getObstaclesAndEnemies(w *World) (m Matrix) {
	m = w.Obstacles.Clone()
	for _, enemy := range w.Enemies {
		m.Set(enemy.Pos(), ONE)
	}
	return
}

func (e *EnemyBase) beamJustHit(w *World) bool {
	if !w.Beam.Idx.Eq(w.BeamMax) { // the fact that this is required shows me
		// I need to structure this stuff differently.
		return false
	}
	return w.WorldPosToTile(w.Beam.End) == e.pos
}

func (e *EnemyBase) move(w *World, m Matrix) {
	if e.moveCooldownIdx.IsPositive() {
		e.moveCooldownIdx.Dec()
	}
	if e.moveCooldown.IsPositive() && e.moveCooldownIdx == ZERO && w.Player.OnMap {
		e.moveCooldownIdx = e.moveCooldown
		e.goToPlayer(w, m)
	}
}
