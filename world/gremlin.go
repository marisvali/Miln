package world

import (
	. "github.com/marisvali/miln/gamelib"
)

type Gremlin struct {
	EnemyBase
}

func NewGremlin(w WorldData, pos Pt) *Gremlin {
	var g Gremlin
	g.pos = pos
	g.maxHealth = w.GremlinMaxHealth
	g.health = g.maxHealth
	g.freezeCooldown = w.GremlinFreezeCooldown
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
