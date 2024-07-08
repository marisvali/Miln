package world

import (
	. "github.com/marisvali/miln/gamelib"
)

type SpawnPortal struct {
	Pos                     Pt
	Health                  Int
	MaxHealth               Int
	MaxTimeout              Int
	TimeoutIdx              Int
	nGremlinsLeftToSpawn    Int
	nHoundsLeftToSpawn      Int
	nUltraHoundsLeftToSpawn Int
	nKingsLeftToSpawn       Int
}

func NewSpawnPortal(pos Pt) (p SpawnPortal) {
	p.Pos = pos
	p.MaxHealth = I(1)
	p.Health = p.MaxHealth
	p.MaxTimeout = SpawnPortalCooldown
	p.nGremlinsLeftToSpawn = RInt(nGremlinMin, nGremlinMax)
	p.nHoundsLeftToSpawn = RInt(nHoundMin, nHoundMax)
	p.nUltraHoundsLeftToSpawn = RInt(nUltraHoundMin, nUltraHoundMax)
	p.nKingsLeftToSpawn = I(0)
	return
}

func (p *SpawnPortal) Step(w *World) {
	if p.TimeoutIdx.IsPositive() {
		p.TimeoutIdx.Dec()
		return // Don't spawn.
	}

	ng := p.nGremlinsLeftToSpawn
	nh := p.nHoundsLeftToSpawn
	nu := p.nUltraHoundsLeftToSpawn
	nk := p.nKingsLeftToSpawn
	total := ng.Plus(nh).Plus(nu).Plus(nk)
	if total.IsZero() {
		return
	}
	spawn := RInt(ZERO, total.Minus(ONE))
	if spawn.Lt(ng) {
		w.Enemies = append(w.Enemies, NewGremlin(p.Pos))
		p.nGremlinsLeftToSpawn.Dec()
	} else if spawn.Lt(ng.Plus(nh)) {
		w.Enemies = append(w.Enemies, NewHound(p.Pos))
		p.nHoundsLeftToSpawn.Dec()
	} else if spawn.Lt(ng.Plus(nh).Plus(nu)) {
		w.Enemies = append(w.Enemies, NewUltraHound(p.Pos))
		p.nUltraHoundsLeftToSpawn.Dec()
	} else if spawn.Lt(ng.Plus(nh).Plus(nu).Plus(nk)) {
		w.Enemies = append(w.Enemies, NewKing(p.Pos))
		p.nKingsLeftToSpawn.Dec()
	}

	p.TimeoutIdx = p.MaxTimeout
}

func (p *SpawnPortal) Active() bool {
	if p.nGremlinsLeftToSpawn.Gt(ZERO) ||
		p.nHoundsLeftToSpawn.Gt(ZERO) ||
		p.nUltraHoundsLeftToSpawn.Gt(ZERO) ||
		p.nKingsLeftToSpawn.Gt(ZERO) {
		return true
	}
	return false
}
