package world

import (
	"fmt"
	. "github.com/marisvali/miln/gamelib"
	"slices"
)

type Wave struct {
	SecondsAfterLastWave Int
	NHounds              Int
}

func NewWave(wd WaveData) (w Wave) {
	w.SecondsAfterLastWave = wd.SecondsAfterLastWave
	w.NHounds = RInt(wd.NHoundMin, wd.NHoundMax)
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

func (p *SpawnPortal) Clone() SpawnPortal {
	clone := *p
	clone.Waves = slices.Clone(p.Waves)
	return clone
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

	if wave.NHounds.IsPositive() {
		w.Enemies = append(w.Enemies, NewHound(p.worldData, p.pos))
		wave.NHounds.Dec()
	}

	p.TimeoutIdx = p.MaxTimeout
}

func (p *SpawnPortal) Active() bool {
	wave := p.CurrentWave()
	if wave != &p.Waves[len(p.Waves)-1] {
		// We are not at the last wave yet.
		return true
	}

	if wave.NHounds.Gt(ZERO) {
		return true
	}
	return false
}

func (p *SpawnPortal) Pos() Pt {
	return p.pos
}

func (p *SpawnPortal) State() string { return "NotUsed" }
