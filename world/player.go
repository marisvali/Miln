package world

import (
	. "github.com/marisvali/miln/gamelib"
)

type Player struct {
	pos                        Pt
	OnMap                      bool
	MaxHealth                  Int
	AmmoCount                  Int
	AmmoLimit                  Int
	JustHit                    bool
	Health                     Int
	CooldownAfterGettingHit    Int
	CooldownAfterGettingHitIdx Int
	Energy                     Int
	state                      string
}

func NewPlayer() (p Player) {
	p.MaxHealth = I(3)
	p.Health = p.MaxHealth
	p.CooldownAfterGettingHit = I(40)
	return
}

// ComputeFreePositions returns a matrix that indicates the positions where the
// player can move to from his current state.
func (p *Player) ComputeFreePositions(w *World) (free MatBool) {
	if p.OnMap {
		free = w.VisibleTiles.Clone()
	} else {
		free = NewMatBool(w.Obstacles.Size())
		free.SetAll()
	}

	free.Subtract(w.Obstacles)
	free.Subtract(w.EnemyPositions())
	return
}

func (p *Player) Step(w *World, input PlayerInput) {
	// See about the beam.
	if w.Beam.Idx.Gt(ZERO) {
		w.Beam.Idx.Dec()
	}

	// Update the player's state.
	if w.Beam.Idx.Gt(ZERO) {
		p.state = "Shooting"
	} else {
		p.state = "Resting"
	}

	if p.CooldownAfterGettingHitIdx.Gt(ZERO) {
		p.CooldownAfterGettingHitIdx.Dec()
		return
	}

	if input.Move {
		free := p.ComputeFreePositions(w)
		if free.At(input.MovePt) {
			p.pos = input.MovePt
			p.OnMap = true

			// Collect ammos.
			newAmmos := make([]Ammo, 0)
			for i := range w.Ammos {
				if w.Ammos[i].Pos == w.Player.pos {
					w.Player.AmmoCount.Add(w.Ammos[i].Count)
					if w.Player.AmmoCount.Gt(w.Player.AmmoLimit) {
						w.Player.AmmoCount = w.Player.AmmoLimit
					}
				} else {
					newAmmos = append(newAmmos, w.Ammos[i])
				}
			}
			w.Ammos = newAmmos
		}
	}

	if input.Shoot &&
		w.VisibleTiles.At(input.ShootPt) &&
		(!w.UseAmmo || w.UseAmmo && w.Player.AmmoCount.IsPositive()) {

		shotEnemies := []*Enemy{}
		for i := range w.Enemies {
			if w.Enemies[i].Pos().Eq(input.ShootPt) {
				shotEnemies = append(shotEnemies, &w.Enemies[i])
			}
		}

		if len(shotEnemies) > 0 {
			w.Beam.Idx = w.BeamMax // show beam
			w.Beam.End = w.TileToWorldPos(input.ShootPt)
			if w.UseAmmo {
				w.Player.AmmoCount.Dec()
			}
		}
	}
}

func (p *Player) Hit() {
	p.JustHit = true
	p.OnMap = false
	p.Health.Dec()
	p.CooldownAfterGettingHitIdx = p.CooldownAfterGettingHit
}

func (p *Player) Pos() Pt {
	return p.pos
}

func (p *Player) SetPos(pos Pt) {
	p.pos = pos
}

func (p *Player) State() string {
	return p.state
}
