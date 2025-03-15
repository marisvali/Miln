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
	p.hitCooldown = w.PillarFreezeCooldown
	return &p
}

func (p *Pillar) Clone() Enemy {
	np := *p
	return &np
}

func (p *Pillar) Step(w *World) {
	if p.Vulnerable(w) && p.beamJustHit(w) {
		p.hitCooldownIdx = p.hitCooldown
		if w.Player.HitPermissions.CanHitPillar {
			p.health.Dec()
			if p.health == ZERO {
				// pillar turns into obstacle
				w.Obstacles.Set(p.pos)
			}
		}
	}

	if p.hitCooldownIdx.IsPositive() {
		p.hitCooldownIdx.Dec()
		return // Don't move.
	}

	// p.move(w, getObstaclesAndEnemies(w))
}

func (p *Pillar) Vulnerable(w *World) bool {
	if p.hitCooldownIdx.IsPositive() {
		return false
	}
	return w.Player.HitPermissions.CanHitGremlin
}
