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
				if e.Health == ZERO {
					// Drop things on death.
					if e.Type == ZERO {
						// gremlin
						// Random chance to drop pillar key, if it wasn't dropped before.
						if !w.PillarKeyDropped && RInt(I(0), I(10)) == ZERO {
							w.Keys = append(w.Keys, NewPillarKey(e.Pos))
							w.PillarKeyDropped = true
						}
					} else if e.Type == ONE {
						// pillar
						// One of the last pillars drops the hound key, if it wasn't dropped before.
						nPillars := ZERO
						for i := range w.Enemies {
							if w.Enemies[i].Type == ONE {
								nPillars.Inc()
							}
						}
						if !w.HoundKeyDropped && nPillars.Lt(TWO) {
							w.Keys = append(w.Keys, NewHoundKey(e.Pos))
							w.HoundKeyDropped = true
						}
					} else if e.Type == TWO {
						// hound
						// One of the last hounds drops the portal key, if it wasn't dropped before.
						nHounds := ZERO
						for i := range w.Enemies {
							if w.Enemies[i].Type == TWO {
								nHounds.Inc()
							}
						}
						if !w.PortalKeyDropped && nHounds.Lt(TWO) {
							w.Keys = append(w.Keys, NewPortalKey(e.Pos))
							w.PortalKeyDropped = true
						}
					}
				}
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
		if e.Type == TWO {
			// For hounds, only put other hounds on it.
			for _, enemy := range w.Enemies {
				if enemy.Type == TWO && !enemy.Pos.Eq(e.Pos) {
					allObstacles.Set(enemy.Pos, ONE)
				}
			}
		} else {
			// For other enemies, put everyone on it.
			for _, enemy := range w.Enemies {
				if !enemy.Pos.Eq(e.Pos) {
					allObstacles.Set(enemy.Pos, ONE)
				}
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
