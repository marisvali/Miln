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
	h.freezeCooldown = w.HoundFreezeCooldown
	return &h
}

func (h *Hound) Clone() Enemy {
	nh := *h
	return &nh
}

func (h *Hound) Step(w *World) {
	if h.Vulnerable(w) && h.beamJustHit(w) {
		h.freezeCooldownIdx = h.freezeCooldown
		if w.Player.HitPermissions.CanHitHound {
			h.health.Dec()
		}
	}

	if h.freezeCooldownIdx.IsPositive() {
		h.freezeCooldownIdx.Dec()
		return // Don't move.
	}

	h.move(w, getObstaclesAndEnemies(w))
}

func (h *Hound) Vulnerable(w *World) bool {
	if h.freezeCooldownIdx.IsPositive() {
		return false
	}
	return w.Player.HitPermissions.CanHitHound
}
