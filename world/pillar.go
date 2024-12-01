package world

import (
	. "github.com/marisvali/miln/gamelib"
)

type Pillar struct {
	EnemyBase
}

func NewPillar(w WorldData, pos Pt) *Pillar {
	var p Pillar
	p.pos = pos
	p.maxHealth = w.PillarMaxHealth
	p.health = p.maxHealth
	p.freezeCooldown = w.PillarFreezeCooldown
	p.moveCooldown = w.PillarMoveCooldown
	p.moveCooldownIdx = p.moveCooldown.DivBy(TWO)
	return &p
}

func (p *Pillar) Clone() Enemy {
	np := *p
	return &np
}

func (p *Pillar) Step(w *World) {
	if p.Vulnerable(w) && p.beamJustHit(w) {
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

func (p *Pillar) Vulnerable(w *World) bool {
	if p.freezeCooldownIdx.IsPositive() {
		return false
	}
	return w.Player.HitPermissions.CanHitGremlin
}
