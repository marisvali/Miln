package world

import (
	. "github.com/marisvali/miln/gamelib"
)

type SpawnPortal struct {
	Pos        Pt
	Health     Int
	MaxHealth  Int
	MaxTimeout Int
	TimeoutIdx Int
}

func NewSpawnPortal(pos Pt) (p SpawnPortal) {
	p.Pos = pos
	p.MaxHealth = I(1)
	p.Health = p.MaxHealth
	p.MaxTimeout = spawnPortalCooldown
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
	w.Enemies = append(w.Enemies, NewEnemy(I(0), p.Pos))
	p.TimeoutIdx = p.MaxTimeout
}
