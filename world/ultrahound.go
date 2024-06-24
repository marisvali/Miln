package world

import (
	. "github.com/marisvali/miln/gamelib"
)

type UltraHound struct {
	EnemyBase
}

func NewUltraHound(pos Pt) *UltraHound {
	var h UltraHound
	h.pos = pos
	h.maxHealth = UltraHoundMaxHealth
	h.health = h.maxHealth
	h.freezeCooldown = UltraHoundFreezeCooldown
	h.moveCooldown = UltraHoundMoveCooldown
	h.moveCooldownIdx = h.moveCooldown.DivBy(TWO)
	return &h
}

func (h *UltraHound) Step(w *World) {
	if h.beamJustHit(w) {
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
	m := NewMatrix(w.Obstacles.Size())
	for _, enemy := range w.Enemies {
		_, ok := enemy.(*UltraHound)
		if ok && !enemy.Pos().Eq(h.pos) {
			m.Set(enemy.Pos(), ONE)
		}
	}
	h.move(w, m)
}
