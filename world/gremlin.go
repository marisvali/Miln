package world

import (
	. "github.com/marisvali/miln/gamelib"
)

type Gremlin struct {
	EnemyBase
}

func NewGremlin(pos Pt) *Gremlin {
	var g Gremlin
	g.pos = pos
	g.maxHealth = GremlinMaxHealth
	g.health = g.maxHealth
	g.freezeCooldown = GremlinFreezeCooldown
	g.moveCooldown = GremlinMoveCooldown
	g.moveCooldownIdx = g.moveCooldown.DivBy(TWO)
	return &g
}

func (g *Gremlin) Clone() Enemy {
	ng := *g
	return &ng
}

func (g *Gremlin) Step(w *World) {
	if g.Vulnerable(w) && g.beamJustHit(w) {
		g.freezeCooldownIdx = g.freezeCooldown
		if w.Player.HitPermissions.CanHitGremlin {
			g.health.Dec()
		}
	}

	if g.freezeCooldownIdx.IsPositive() {
		g.freezeCooldownIdx.Dec()
		return // Don't move.
	}

	g.move(w, getObstaclesAndEnemies(w))
}

func (g *Gremlin) Vulnerable(w *World) bool {
	if g.freezeCooldownIdx.IsPositive() {
		return false
	}
	return w.Player.HitPermissions.CanHitGremlin
}
