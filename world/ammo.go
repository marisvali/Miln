package world

import (
	. "github.com/marisvali/miln/gamelib"
)

type Ammo struct {
	Pos   Pt
	Count Int
}

func (w *World) SpawnAmmos() {
	// Spawn new ammos
	for {
		if len(w.Ammos) == 1 {
			break
		}

		if w.Player.AmmoCount.IsPositive() {
			break
		}

		pt := w.Obstacles.RandomPos()
		if !w.Obstacles.Get(pt).IsZero() {
			continue
		}
		if w.Player.Pos == pt {
			continue
		}
		invalid := false
		for i := range w.Ammos {
			if w.Ammos[i].Pos == pt {
				invalid = true
				break
			}
		}
		if invalid {
			continue
		}
		ammo := Ammo{
			Pos:   pt,
			Count: I(3),
		}
		w.Ammos = append(w.Ammos, ammo)
	}
}
