package world

import (
	. "github.com/marisvali/miln/gamelib"
	_ "image/png"
)

type Player struct {
	Pos            Pt
	OnMap          bool
	TimeoutIdx     Int
	MaxHealth      Int
	AmmoCount      Int
	JustHit        bool
	Health         Int
	HitPermissions HitPermissions
}

func NewPlayer() (p Player) {
	p.MaxHealth = I(300)
	p.Health = p.MaxHealth
	p.HitPermissions = NewHitPermissions()
	return
}

func (p *Player) Step(w *World, input *PlayerInput) {
	if w.Player.TimeoutIdx.Gt(ZERO) {
		w.Player.TimeoutIdx.Dec()
	}

	onEnemy := false
	for i := range w.Enemies {
		if w.Enemies[i].Pos == input.MovePt {
			onEnemy = true
			break
		}
	}

	if input.Move && w.Player.TimeoutIdx.Eq(ZERO) &&
		w.Obstacles.Get(input.MovePt).Eq(ZERO) &&
		(w.AttackableTiles.Get(input.MovePt).Neq(ZERO) || !w.Player.OnMap) &&
		!onEnemy {
		w.Player.Pos = input.MovePt
		w.Player.OnMap = true
		w.Player.TimeoutIdx = playerCooldown

		// Collect ammos.
		newAmmos := make([]Ammo, 0)
		for i := range w.Ammos {
			if w.Ammos[i].Pos == w.Player.Pos {
				w.Player.AmmoCount.Add(w.Ammos[i].Count)
			} else {
				newAmmos = append(newAmmos, w.Ammos[i])
			}
		}
		w.Ammos = newAmmos

		// Collect keys.
		newKeys := make([]Key, 0)
		for i := range w.Keys {
			if w.Keys[i].Pos == w.Player.Pos {
				w.Player.HitPermissions.Add(w.Keys[i].Permissions)
			} else {
				newKeys = append(newKeys, w.Keys[i])
			}
		}
		w.Keys = newKeys
	}

	// See about the beam.
	if w.Beam.Idx.Gt(ZERO) {
		w.Beam.Idx.Dec()
	}
	if input.Shoot &&
		w.Player.TimeoutIdx.Eq(ZERO) &&
		!w.AttackableTiles.Get(input.ShootPt).IsZero() {

		shotEnemies := []*Enemy{}
		for i, _ := range w.Enemies {
			if w.Enemies[i].Pos.Eq(input.ShootPt) {
				shotEnemies = append(shotEnemies, &w.Enemies[i])
			}
		}

		shotPortals := []*SpawnPortal{}
		for i, _ := range w.SpawnPortals {
			if w.SpawnPortals[i].Pos.Eq(input.ShootPt) {
				shotPortals = append(shotPortals, &w.SpawnPortals[i])
			}
		}

		if len(shotEnemies) > 0 || len(shotPortals) > 0 {
			w.Beam.Idx = w.BeamMax // show beam
			w.Player.TimeoutIdx = playerCooldown
			w.Beam.End = w.TileToWorldPos(input.ShootPt)
		}
	}
}

func (p *Player) Hit() {
	p.JustHit = true
	p.OnMap = false
	p.Health.Dec()
}
