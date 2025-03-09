package world

import (
	"fmt"
	. "github.com/marisvali/miln/gamelib"
)

type Wave struct {
	SecondsAfterLastWave Int
	NGremlins            Int
	NHounds              Int
	NUltraHounds         Int
	NKings               Int
}

func NewWave(wd WaveData) (w Wave) {
	w.SecondsAfterLastWave = wd.SecondsAfterLastWave
	w.NGremlins = RInt(wd.NGremlinMin, wd.NGremlinMax)
	w.NHounds = RInt(wd.NHoundMin, wd.NHoundMax)
	w.NUltraHounds = RInt(wd.NUltraHoundMin, wd.NUltraHoundMax)
	w.NKings = RInt(wd.NKingMin, wd.NKingMax)
	return
}

type SpawnPortal struct {
	pos        Pt
	Health     Int
	MaxHealth  Int
	MaxTimeout Int
	TimeoutIdx Int
	Waves      []Wave
	frameIdx   Int
	worldData  WorldData
}

func NewSpawnPortal(w WorldData, pos Pt, cooldown Int, waves []Wave) (p SpawnPortal) {
	p.pos = pos
	p.MaxHealth = I(1)
	p.Health = p.MaxHealth
	p.MaxTimeout = cooldown
	p.Waves = waves
	p.worldData = w
	return
}

func (p *SpawnPortal) CurrentWave() *Wave {
	// Compute the frame at which each wave starts.
	waveStarts := []Int{}
	startOfLastWave := ZERO
	for i := range p.Waves {
		framesAfterLastWave := p.Waves[i].SecondsAfterLastWave.Times(I(60))
		startOfThisWave := startOfLastWave.Plus(framesAfterLastWave)
		waveStarts = append(waveStarts, startOfThisWave)
		startOfLastWave = startOfThisWave
	}

	// Find the i so that:
	// waveStarts[i] <= p.frameIdx < waveStarts[i+1]
	// Watch out for the edge cases:
	// p.frameIdx < waveStarts[0]
	// waveStarts[len(waveStarts)-1] <= p.frameIdx

	if p.frameIdx.Lt(waveStarts[0]) {
		// No wave has started yet.
		return nil
	}

	if p.frameIdx.Geq(waveStarts[len(waveStarts)-1]) {
		// The last wave is active.
		return &p.Waves[len(waveStarts)-1]
	}

	for i := range waveStarts {
		if p.frameIdx.Lt(waveStarts[i]) {
			return &p.Waves[i-1]
		}
	}

	Check(fmt.Errorf("should never get here"))
	return nil
}

func (p *SpawnPortal) Step(w *World) {
	p.frameIdx.Inc()

	if p.TimeoutIdx.IsPositive() {
		p.TimeoutIdx.Dec()
		return // Don't spawn.
	}

	if !(p.TimeoutIdx.IsZero() && w.EnemyMoveCooldownIdx.IsZero()) {
		return // Only spawn when the enemy cooldown is at max.
	}

	wave := p.CurrentWave()
	if wave == nil {
		// No wave active.
		return
	}

	// if wave.NUltraHounds.IsPositive() {
	// 	w.Enemies = append(w.Enemies, NewUltraHound(p.worldData, p.pos))
	// 	wave.NUltraHounds.Dec()
	// } else if wave.NHounds.IsPositive() {
	// 	w.Enemies = append(w.Enemies, NewHound(p.worldData, p.pos))
	// 	wave.NHounds.Dec()
	// } else if wave.NGremlins.IsPositive() {
	// 	w.Enemies = append(w.Enemies, NewGremlin(p.worldData, p.pos))
	// 	wave.NGremlins.Dec()
	// } else if wave.NKings.IsPositive() {
	// 	w.Enemies = append(w.Enemies, NewKing(p.worldData, p.pos))
	// 	wave.NKings.Dec()
	// } else {
	// 	// Nothing spawned, so don't trigger the cooldown.
	// 	return
	// }

	ng := wave.NGremlins
	nh := wave.NHounds
	nu := wave.NUltraHounds
	nk := wave.NKings
	total := ng.Plus(nh).Plus(nu).Plus(nk)
	if total.IsZero() {
		return
	}
	spawn := RInt(ZERO, total.Minus(ONE))
	if spawn.Lt(ng) {
		w.Enemies = append(w.Enemies, NewGremlin(p.worldData, p.pos))
		wave.NGremlins.Dec()
	} else if spawn.Lt(ng.Plus(nh)) {
		w.Enemies = append(w.Enemies, NewHound(p.worldData, p.pos))
		wave.NHounds.Dec()
	} else if spawn.Lt(ng.Plus(nh).Plus(nu)) {
		w.Enemies = append(w.Enemies, NewUltraHound(p.worldData, p.pos))
		wave.NUltraHounds.Dec()
	} else if spawn.Lt(ng.Plus(nh).Plus(nu).Plus(nk)) {
		w.Enemies = append(w.Enemies, NewKing(p.worldData, p.pos))
		wave.NKings.Dec()
	}

	p.TimeoutIdx = p.MaxTimeout
}

func (p *SpawnPortal) Active() bool {
	wave := p.CurrentWave()
	if wave != &p.Waves[len(p.Waves)-1] {
		// We are not at the last wave yet.
		return true
	}

	if wave.NGremlins.Gt(ZERO) ||
		wave.NHounds.Gt(ZERO) ||
		wave.NUltraHounds.Gt(ZERO) ||
		wave.NKings.Gt(ZERO) {
		return true
	}
	return false
}

func (p *SpawnPortal) Pos() Pt {
	return p.pos
}

func (p *SpawnPortal) State() string { return "NotUsed" }
