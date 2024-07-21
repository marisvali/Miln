package world

import (
	. "github.com/marisvali/miln/gamelib"
)

type Vision interface {
	Compute(start Pt, obstacles MatBool) (attackableTiles MatBool)
}

type Vision1 struct {
	blockSize Int
}

func NewVision1() *Vision1 {
	var v Vision1
	v.blockSize = I(1000)
	return &v
}

func (v *Vision1) tileToWorldPos(pt Pt) Pt {
	half := v.blockSize.DivBy(TWO)
	offset := Pt{half, half}
	return pt.Times(v.blockSize).Plus(offset)
}

func (v *Vision1) worldPosToTile(pt Pt) Pt {
	return pt.DivBy(v.blockSize)
}

func (v *Vision1) Compute(start Pt, obstacles MatBool) (attackableTiles MatBool) {
	// Compute which tiles are attackable.
	attackableTiles = NewMatBool(obstacles.Size())

	rows := obstacles.Size().Y
	cols := obstacles.Size().X

	// Get a list of squares.
	squares := []Square{}
	for y := ZERO; y.Lt(rows); y.Inc() {
		for x := ZERO; x.Lt(cols); x.Inc() {
			pt := Pt{x, y}
			if obstacles.At(pt) {
				center := v.tileToWorldPos(pt)
				size := v.blockSize.Times(I(98)).DivBy(I(100))
				squares = append(squares, Square{center, size})
			}
		}
	}

	// Draw a line from the player's pos to each of the tiles and test if that
	// line intersects the squares.
	lineStart := v.tileToWorldPos(start)
	for y := ZERO; y.Lt(rows); y.Inc() {
		for x := ZERO; x.Lt(cols); x.Inc() {
			// Check if tile can be attacked.
			lineEnd := v.tileToWorldPos(Pt{x, y})
			l := Line{lineStart, lineEnd}
			if intersects, pt := LineSquaresIntersection(l, squares); intersects {
				obstacleTile := v.worldPosToTile(pt)
				if obstacleTile.Eq(Pt{x, y}) {
					attackableTiles.Set(Pt{x, y})
				} else {
					attackableTiles.Clear(Pt{x, y})
				}
			} else {
				attackableTiles.Set(Pt{x, y})
			}
		}
	}
	return
}
