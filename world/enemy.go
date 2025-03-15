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
	MoveCooldownMultiplier() Int
	MoveCooldownIdx() Int
	Health() Int
	MaxHealth() Int
	Clone() Enemy
	Vulnerable(w *World) bool
	State() string
}

type EnemyBase struct {
	pos                    Pt
	health                 Int
	maxHealth              Int
	freezeCooldownIdx      Int
	freezeCooldown         Int
	moveCooldownMultiplier Int
	moveCooldownIdx        Int
	hitsPlayer             bool
	aggroDistance          Int
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

func (e *EnemyBase) MoveCooldownMultiplier() Int {
	return e.moveCooldownMultiplier
}

func (e *EnemyBase) MoveCooldownIdx() Int {
	return e.moveCooldownIdx
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
		if e.hitsPlayer {
			// Move to the position either way and hit player if necessary.
			e.pos = path[1]
			if path[1].Eq(w.Player.Pos()) {
				w.Player.Hit()
			}
		} else {
			// Move to the position only if not occupied by the player.
			if !path[1].Eq(w.Player.Pos()) {
				e.pos = path[1]
			}
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
	// Don't move or consume movement cooldown if the enemy is aggroed by
	// being visible. Otherwise, when you teleport close to an enemy that was
	// previously not aggroed, the enemy instantly jumps at you. This is not
	// what I want.
	if w.EnemiesAggroWhenVisible {
		if !w.Player.OnMap || !w.AttackableTiles.At(e.pos) {
			e.moveCooldownIdx = e.moveCooldownMultiplier
			return
		}
	}

	if w.EnemyMoveCooldown.IsPositive() && w.EnemyMoveCooldownIdx.IsZero() && e.moveCooldownIdx.IsPositive() {
		e.moveCooldownIdx.Dec()
	}

	if e.moveCooldownIdx.IsZero() && w.Player.OnMap {
		if w.EnemiesAggroWhenVisible {
			if w.AttackableTiles.At(e.pos) {
				e.goToPlayer(w, m)
				e.moveCooldownIdx = e.moveCooldownMultiplier
			}
		} else {
			aggroDistance := e.aggroDistance.Times(e.aggroDistance)
			distanceToPlayer := e.Pos().SquaredDistTo(w.Player.Pos())
			if aggroDistance.Geq(distanceToPlayer) {
				e.goToPlayer(w, m)
				e.moveCooldownIdx = e.moveCooldownMultiplier
			}
		}
	}
}
