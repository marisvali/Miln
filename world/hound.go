package world

import (
	. "github.com/marisvali/miln/gamelib"
)

type Hound struct {
	EnemyBase
}

func NewHound(w WorldData, pos Pt) *Hound {
	var h Hound
	h.pos = pos
	h.maxHealth = w.HoundMaxHealth
	h.health = h.maxHealth
	h.hitCooldown = w.HoundFreezeCooldown
	h.moveCooldownMultiplier = w.HoundMoveCooldownMultiplier
	h.moveCooldownIdx = h.moveCooldownMultiplier
	h.hitsPlayer = w.HoundHitsPlayer
	h.aggroDistance = w.HoundAggroDistance
	return &h
}

func (h *Hound) Clone() Enemy {
	nh := *h
	return &nh
}

func (h *Hound) Step(w *World) {
	if h.Vulnerable(w) && h.beamJustHit(w) {
		h.hitCooldownIdx = h.hitCooldown
		if w.Player.HitPermissions.CanHitHound {
			h.health.Dec()
		}
	}

	if h.hitCooldownIdx.IsPositive() {
		// Only move after hitCooldownIdx is down to ZERO
		// AND also w.EnemyMoveCooldownIdx is at maximum.
		if h.hitCooldownIdx.Gt(I(10)) {
			h.hitCooldownIdx.Dec()
		} else if w.EnemyMoveCooldownIdx.Eq(w.EnemyMoveCooldown.Minus(ONE)) {
			h.hitCooldownIdx = ZERO
		}
		return // Don't move.
	}

	// h.move(w, getObstaclesAndEnemies(w))
}

func (h *Hound) Vulnerable(w *World) bool {
	if h.hitCooldownIdx.IsPositive() {
		return false
	}
	return w.Player.HitPermissions.CanHitHound
}
