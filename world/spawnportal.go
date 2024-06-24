package world

import (
	. "github.com/marisvali/miln/gamelib"
)

type SpawnPortal struct {
	Pos                  Pt
	Health               Int
	MaxHealth            Int
	MaxTimeout           Int
	TimeoutIdx           Int
	nGremlinsLeftToSpawn Int
	nHoundsLeftToSpawn   Int
}

func NewSpawnPortal(pos Pt) (p SpawnPortal) {
	p.Pos = pos
	p.MaxHealth = I(1)
	p.Health = p.MaxHealth
	p.MaxTimeout = spawnPortalCooldown
	p.nGremlinsLeftToSpawn = I(4)
	p.nHoundsLeftToSpawn = I(3)
	return
}

func (p *SpawnPortal) Step(w *World) {
	if w.Beam.Idx.Eq(w.BeamMax) { // the fact that this is required shows me
		// I need to structure this stuff differently.
		beamEndTile := w.WorldPosToTile(w.Beam.End)
		if beamEndTile.Eq(p.Pos) {
			// We have been shot.
			if w.Player.HitPermissions.CanHitPortal {
				p.Health.Dec()
			}
		}
	}

	if p.TimeoutIdx.IsPositive() {
		p.TimeoutIdx.Dec()
		return // Don't spawn.
	}

	// Spawn guy.
	// Check if there is already a guy here.
	occupied := false
	for _, enemy := range w.Enemies {
		if enemy.Pos == p.Pos {
			occupied = true
			break
		}
	}
	if occupied {
		return // Don't spawn.
	}

	//w.Enemies = append(w.Enemies, NewEnemy(RInt(I(0), I(2)), p.Pos))
	// Spawn only gremlins.
	//if p.nGremlinsLeftToSpawn.IsPositive() && p.nHoundsLeftToSpawn.IsPositive() {
	//	if RInt(I(0), I(4)) == ZERO {
	//		w.Enemies = append(w.Enemies, NewEnemy(TWO, p.Pos))
	//		p.nHoundsLeftToSpawn.Dec()
	//	} else {
	//		w.Enemies = append(w.Enemies, NewEnemy(ZERO, p.Pos))
	//		p.nGremlinsLeftToSpawn.Dec()
	//	}
	//} else if p.nGremlinsLeftToSpawn.IsPositive() {
	//	w.Enemies = append(w.Enemies, NewEnemy(ZERO, p.Pos))
	//	p.nGremlinsLeftToSpawn.Dec()
	//} else if p.nHoundsLeftToSpawn.IsPositive() {
	//	w.Enemies = append(w.Enemies, NewEnemy(TWO, p.Pos))
	//	p.nHoundsLeftToSpawn.Dec()
	//} else {
	//	if !w.KingSpawned {
	//		w.Enemies = append(w.Enemies, NewEnemy(I(4), p.Pos))
	//		w.KingSpawned = true
	//	}
	//}
	if p.nHoundsLeftToSpawn.IsPositive() {
		w.Enemies = append(w.Enemies, NewEnemy(TWO, p.Pos))
		p.nHoundsLeftToSpawn.Dec()
	} else if p.nGremlinsLeftToSpawn.IsPositive() {
		w.Enemies = append(w.Enemies, NewEnemy(ZERO, p.Pos))
		p.nGremlinsLeftToSpawn.Dec()
	} else if !w.KingSpawned {
		w.Enemies = append(w.Enemies, NewEnemy(I(4), p.Pos))
		w.KingSpawned = true
	} else {
		w.Enemies = append(w.Enemies, NewEnemy(ZERO, p.Pos))
	}

	//w.Enemies = append(w.Enemies, NewEnemy(I(0), p.Pos))
	p.TimeoutIdx = p.MaxTimeout
}
