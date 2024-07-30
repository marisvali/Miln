package world

import (
	. "github.com/marisvali/miln/gamelib"
)

type King struct {
	EnemyBase
}

func NewKing(pos Pt) *King {
	var k King
	k.pos = pos
	k.maxHealth = KingMaxHealth
	k.health = KingMaxHealth
	k.freezeCooldown = KingFreezeCooldown
	k.moveCooldown = KingMoveCooldown
	k.moveCooldownIdx = k.moveCooldown.DivBy(TWO)
	return &k
}

func (k *King) Clone() Enemy {
	nk := *k
	return &nk
}

func (k *King) Step(w *World) {
	if k.beamJustHit(w) {
		k.freezeCooldownIdx = k.freezeCooldown
		if w.Player.HitPermissions.CanHitKing {
			k.health.Dec()
			if k.health == ZERO {
				w.Keys = append(w.Keys, NewUltraHoundKey(k.pos))
			}
			w.Enemies = append(w.Enemies, NewHound(k.pos))
		}
	}

	if k.freezeCooldownIdx.IsPositive() {
		k.freezeCooldownIdx.Dec()
		return // Don't move.
	}

	k.move(w, getObstaclesAndEnemies(w))
}
