package world

import (
	. "github.com/marisvali/miln/gamelib"
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
					} else if e.Type == ONE {
						// pillar turns into obstacle
						w.Obstacles.Set(e.Pos, ONE)
					} else if e.Type == TWO {
						// hound
					} else if e.Type == I(3) {
						nQuestions := ZERO
						for i := range w.Enemies {
							if w.Enemies[i].Type == I(3) {
								nQuestions.Inc()
							}
						}
						if nQuestions == ONE {
							w.SpawnPortals = append(w.SpawnPortals, NewSpawnPortal(e.Pos))
						} else if nQuestions == TWO {
							w.Keys = append(w.Keys, NewPillarKey(e.Pos))
						} else {
							// question mark
							nGremlins := ZERO
							for i := range w.Enemies {
								if w.Enemies[i].Type == ZERO {
									nGremlins.Inc()
								}
							}
							if nQuestions.Mod(I(4)) == ZERO && nGremlins.Leq(I(4)) {
								nHounds := ZERO
								for i := range w.Enemies {
									if w.Enemies[i].Type == TWO {
										nHounds.Inc()
									}
								}
								if nHounds.Lt(ONE) && (RInt(I(0), I(100)).Leq(I(40)) || nQuestions.Leq(I(4))) {
									w.Enemies = append(w.Enemies, NewEnemy(TWO, e.Pos))
								} else {
									w.Enemies = append(w.Enemies, NewEnemy(ONE, e.Pos))
								}
							} else {
								w.Obstacles.Set(e.Pos, ONE)
							}
						}
					} else if e.Type == I(4) {
						w.Keys = append(w.Keys, NewHoundKey(e.Pos))
					}
				}
			}
		}
	}

	if e.FrozenIdx.IsPositive() {
		e.FrozenIdx.Dec()
		return // Don't move.
	}
	if enemyCooldowns[e.Type.ToInt()].IsNegative() {
		return
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
