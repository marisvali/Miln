package world

import (
	. "github.com/marisvali/miln/gamelib"
)

type Hound struct {
	EnemyBase
}

func NewHound(pos Pt) *Hound {
	var h Hound
	h.pos = pos
	h.maxHealth = HoundMaxHealth
	h.health = h.maxHealth
	h.freezeCooldown = HoundFreezeCooldown
	h.moveCooldown = HoundMoveCooldown
	h.moveCooldownIdx = h.moveCooldown.DivBy(TWO)
	return &h
}

func (h *Hound) Step(w *World) {
	if h.beamJustHit(w) {
		h.freezeCooldownIdx = h.freezeCooldown
		if w.Player.HitPermissions.CanHitHound {
			h.health.Dec()
		}
	}

	if h.freezeCooldownIdx.IsPositive() {
		h.freezeCooldownIdx.Dec()
		return // Don't move.
	}

	// For hounds, only consider other hounds.
	m := NewMatBool(w.Obstacles.Size())
	for _, enemy := range w.Enemies {
		_, ok := enemy.(*Hound)
		if ok && !enemy.Pos().Eq(h.pos) {
			m.Set(enemy.Pos())
		}
	}
	h.move(w, m)
}
