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
	g.moveCooldownMultiplier = w.GremlinMoveCooldownMultiplier
	g.preparingToAttackCooldown = w.GremlinPreparingToAttackCooldown
	g.attackCooldownMultiplier = w.GremlinAttackCooldownMultiplier
	g.hitCooldown = w.GremlinHitCooldown
	g.hitsPlayer = w.GremlinHitsPlayer
	g.aggroDistance = w.GremlinAggroDistance
	return &g
}

func (g *Gremlin) Clone() Enemy {
	ng := *g
	return &ng
}
