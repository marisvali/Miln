package world

import (
	. "github.com/marisvali/miln/gamelib"
)

type UltraHound struct {
	EnemyBase
}

func NewUltraHound(w WorldData, pos Pt) *UltraHound {
	var h UltraHound
	h.pos = pos
	h.maxHealth = w.UltraHoundMaxHealth
	h.health = h.maxHealth
	h.freezeCooldown = w.UltraHoundFreezeCooldown
	return &h
}

func (h *UltraHound) Clone() Enemy {
	nh := *h
	return &nh
}

func (h *UltraHound) Step(w *World) {
	if h.Vulnerable(w) && h.beamJustHit(w) {
		h.freezeCooldownIdx = h.freezeCooldown
		if w.Player.HitPermissions.CanHitUltraHound {
			h.health.Dec()
		}
	}

	if h.freezeCooldownIdx.IsPositive() {
		h.freezeCooldownIdx.Dec()
		return // Don't move.
	}

	// For ultra hounds, only consider other ultra hounds.
	m := NewMatBool(w.Obstacles.Size())
	for _, enemy := range w.Enemies {
		_, ok := enemy.(*UltraHound)
		if ok && !enemy.Pos().Eq(h.pos) {
			m.Set(enemy.Pos())
		}
	}
	h.move(w, m)
}

func (h *UltraHound) Vulnerable(w *World) bool {
	if h.freezeCooldownIdx.IsPositive() {
		return false
	}
	return w.Player.HitPermissions.CanHitUltraHound
}
