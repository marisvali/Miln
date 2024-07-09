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

func NewSpawnPortal(pos Pt, cooldown Int, nGremlins Int, nHounds Int, nUltraHounds Int, nKings Int) (p SpawnPortal) {
	p.Pos = pos
	p.MaxHealth = I(1)
	p.Health = p.MaxHealth
	p.MaxTimeout = cooldown
	p.nGremlinsLeftToSpawn = nGremlins
	p.nHoundsLeftToSpawn = nHounds
	p.nUltraHoundsLeftToSpawn = nUltraHounds
	p.nKingsLeftToSpawn = nKings
	return
}

func (p *SpawnPortal) Step(w *World) {
	if p.TimeoutIdx.IsPositive() {
		p.TimeoutIdx.Dec()
		return // Don't spawn.
	}

	if p.nUltraHoundsLeftToSpawn.IsPositive() {
		w.Enemies = append(w.Enemies, NewUltraHound(p.Pos))
		p.nUltraHoundsLeftToSpawn.Dec()
	} else if p.nHoundsLeftToSpawn.IsPositive() {
		w.Enemies = append(w.Enemies, NewHound(p.Pos))
		p.nHoundsLeftToSpawn.Dec()
	} else if p.nGremlinsLeftToSpawn.IsPositive() {
		w.Enemies = append(w.Enemies, NewGremlin(p.Pos))
		p.nGremlinsLeftToSpawn.Dec()
	} else if p.nKingsLeftToSpawn.IsPositive() {
		w.Enemies = append(w.Enemies, NewKing(p.Pos))
		p.nKingsLeftToSpawn.Dec()
	}

	// ng := p.nGremlinsLeftToSpawn
	// nh := p.nHoundsLeftToSpawn
	// nu := p.nUltraHoundsLeftToSpawn
	// nk := p.nKingsLeftToSpawn
	// total := ng.Plus(nh).Plus(nu).Plus(nk)
	// if total.IsZero() {
	// 	return
	// }
	// spawn := RInt(ZERO, total.Minus(ONE))
	// if spawn.Lt(ng) {
	// 	w.Enemies = append(w.Enemies, NewGremlin(p.Pos))
	// 	p.nGremlinsLeftToSpawn.Dec()
	// } else if spawn.Lt(ng.Plus(nh)) {
	// 	w.Enemies = append(w.Enemies, NewHound(p.Pos))
	// 	p.nHoundsLeftToSpawn.Dec()
	// } else if spawn.Lt(ng.Plus(nh).Plus(nu)) {
	// 	w.Enemies = append(w.Enemies, NewUltraHound(p.Pos))
	// 	p.nUltraHoundsLeftToSpawn.Dec()
	// } else if spawn.Lt(ng.Plus(nh).Plus(nu).Plus(nk)) {
	// 	w.Enemies = append(w.Enemies, NewKing(p.Pos))
	// 	p.nKingsLeftToSpawn.Dec()
	// }

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
