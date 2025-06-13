package world

import (
	"fmt"
	. "github.com/marisvali/miln/gamelib"
)

type Wave struct {
	SecondsAfterLastWave Int `yaml:"SecondsAfterLastWave"`
	NHounds              Int `yaml:"NHounds"`
}

type SpawnPortal struct {
	Rand
	pos           Pt
	Health        Int
	MaxHealth     Int
	SpawnCooldown Cooldown
	Waves         [10]Wave
	WavesLen      int
	frameIdx      Int
	worldParams   WorldParams
}

func NewSpawnPortal(seed Int, p SpawnPortalParams, w WorldParams) (sp SpawnPortal) {
	sp.RSeed(seed)
	sp.pos = p.Pos
	sp.MaxHealth = I(1)
	sp.Health = sp.MaxHealth
	sp.SpawnCooldown = NewCooldown(p.SpawnPortalCooldown)
	sp.Waves = p.Waves
	sp.WavesLen = p.WavesLen
	sp.worldParams = w
	return
}

func (p *SpawnPortal) CurrentWave() *Wave {
	// Compute the frame at which each wave starts.
	waveStarts := [100]Int{}
	startOfLastWave := ZERO
	for i := range p.WavesLen {
		framesAfterLastWave := p.Waves[i].SecondsAfterLastWave.Times(I(60))
		startOfThisWave := startOfLastWave.Plus(framesAfterLastWave)
		waveStarts[i] = startOfThisWave
		startOfLastWave = startOfThisWave
	}
	waveStartsLen := p.WavesLen

	// Find the i so that:
	// waveStarts[i] <= p.frameIdx < waveStarts[i+1]
	// Watch out for the edge cases:
	// p.frameIdx < waveStarts[0]
	// waveStarts[len(waveStarts)-1] <= p.frameIdx

	if p.frameIdx.Lt(waveStarts[0]) {
		// No wave has started yet.
		return nil
	}

	if p.frameIdx.Geq(waveStarts[waveStartsLen-1]) {
		// The last wave is active.
		return &p.Waves[waveStartsLen-1]
	}

	for i := range waveStartsLen {
		if p.frameIdx.Lt(waveStarts[i]) {
			return &p.Waves[i-1]
		}
	}

	Check(fmt.Errorf("should never get here"))
	return nil
}

func (p *SpawnPortal) Step(w *World) {
	p.frameIdx.Inc()
	p.SpawnCooldown.Update()
	if !p.SpawnCooldown.Ready() {
		return // Don't spawn.
	}

	if !w.EnemyMoveCooldown.Ready() {
		return // Only spawn when the enemy cooldown is ready.
	}

	wave := p.CurrentWave()
	if wave == nil {
		// No wave active.
		return
	}

	if wave.NHounds.IsPositive() {
		w.Enemies[w.EnemiesLen] = NewHound(p.RInt63(), p.worldParams, p.pos)
		w.EnemiesLen++
		wave.NHounds.Dec()
	}

	p.SpawnCooldown.Reset()
}

func (p *SpawnPortal) Active() bool {
	wave := p.CurrentWave()
	if wave != &p.Waves[p.WavesLen-1] {
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
