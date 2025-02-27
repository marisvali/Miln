package world

import (
	. "github.com/marisvali/miln/gamelib"
)

type King struct {
	EnemyBase
	worldData WorldData
}

func NewKing(w WorldData, pos Pt) *King {
	var k King
	k.pos = pos
	k.maxHealth = w.KingMaxHealth
	k.health = w.KingMaxHealth
	k.freezeCooldown = w.KingFreezeCooldown
	k.worldData = w
	return &k
}

func (k *King) Clone() Enemy {
	nk := *k
	return &nk
}

func (k *King) Step(w *World) {
	if k.Vulnerable(w) && k.beamJustHit(w) {
		k.freezeCooldownIdx = k.freezeCooldown
		if w.Player.HitPermissions.CanHitKing {
			k.health.Dec()
			if k.health == ZERO {
				w.Keys = append(w.Keys, NewUltraHoundKey(k.pos))
			}
			w.Enemies = append(w.Enemies, NewHound(k.worldData, k.pos))
		}
	}

	if k.freezeCooldownIdx.IsPositive() {
		k.freezeCooldownIdx.Dec()
		return // Don't move.
	}

	k.move(w, getObstaclesAndEnemies(w))
}

func (k *King) Vulnerable(w *World) bool {
	if k.freezeCooldownIdx.IsPositive() {
		return false
	}
	return w.Player.HitPermissions.CanHitKing
}
