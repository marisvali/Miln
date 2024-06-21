package world

import (
	. "github.com/marisvali/miln/gamelib"
	_ "image/png"
)

type Enemy struct {
	Pos       Pt
	Health    Int
	MaxHealth Int
	FrozenIdx Int
	MaxFrozen Int
	Type      Int
}

func NewEnemy(enemyType Int, pos Pt) Enemy {
	e := Enemy{}
	e.Type = enemyType
	e.Pos = pos
	e.MaxHealth = enemyHealths[e.Type.ToInt()]
	e.Health = e.MaxHealth
	e.MaxFrozen = enemyFrozenCooldowns[e.Type.ToInt()]
	return e
}

func (e *Enemy) Step(w *World) {
	if w.Beam.Idx.Eq(w.BeamMax) { // the fact that this is required shows me
		// I need to structure this stuff differently.
		beamEndTile := w.WorldPosToTile(w.Beam.End)
		if beamEndTile.Eq(e.Pos) {
			// We have been shot.
			e.FrozenIdx = e.MaxFrozen
			if w.Player.HitPermissions.CanHitEnemy[e.Type.ToInt()] {
				e.Health.Dec()
			}
		}
	}

	if e.FrozenIdx.IsPositive() {
		e.FrozenIdx.Dec()
		return // Don't move.
	}
	if w.TimeStep.Plus(ONE).Mod(enemyCooldowns[e.Type.ToInt()]).Neq(ZERO) {
		return
	}

	// Move.
	if w.Player.OnMap {
		// Clone obstacle matrix and put (other) enemies on it.
		allObstacles := w.Obstacles.Clone()
		for _, enemy := range w.Enemies {
			if !enemy.Pos.Eq(e.Pos) {
				allObstacles.Set(enemy.Pos, TWO)
			}
		}

		path := FindPath(e.Pos, w.Player.Pos, allObstacles)
		if len(path) > 1 {
			e.Pos = path[1]
			if e.Pos.Eq(w.Player.Pos) {
				w.Player.Hit()
			}
		}
	}
}
