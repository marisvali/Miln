package world

import (
	. "github.com/marisvali/miln/gamelib"
)

type Pillar struct {
	EnemyBase
}

func NewPillar(pos Pt) *Pillar {
	var p Pillar
	p.pos = pos
	p.maxHealth = PillarMaxHealth
	p.health = p.maxHealth
	p.freezeCooldown = PillarFreezeCooldown
	p.moveCooldown = PillarMoveCooldown
	p.moveCooldownIdx = p.moveCooldown.DivBy(TWO)
	return &p
}

func (p *Pillar) Step(w *World) {
	if p.beamJustHit(w) {
		p.freezeCooldownIdx = p.freezeCooldown
		if w.Player.HitPermissions.CanHitPillar {
			p.health.Dec()
			if p.health == ZERO {
				// pillar turns into obstacle
				w.Obstacles.Set(p.pos)
			}
		}
	}

	if p.freezeCooldownIdx.IsPositive() {
		p.freezeCooldownIdx.Dec()
		return // Don't move.
	}

	p.move(w, getObstaclesAndEnemies(w))
}
