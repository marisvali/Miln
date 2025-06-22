package gamelib

const NRows = 8
const NCols = 8

type Matrix[T any] struct {
	// This is made public for the sake of serializing and deserializing
	// using the encoding/binary package.
	// Don't access it otherwise.
	Cells [64]T
}

func (m *Matrix[T]) Set(pos Pt, val T) {
	m.Cells[pos.Y.Times(I(NCols)).Plus(pos.X).ToInt64()] = val
}

func (m *Matrix[T]) Get(pos Pt) T {
	return m.Cells[pos.Y.Times(I(NCols)).Plus(pos.X).ToInt64()]
}

func (m *Matrix[T]) InBounds(pt Pt) bool {
	return pt.X.IsNonNegative() &&
		pt.Y.IsNonNegative() &&
		pt.Y.Lt(I(NRows)) &&
		pt.X.Lt(I(NCols))
}

func (m *Matrix[T]) PtToIndex(p Pt) Int {
	return p.Y.Times(I(NCols)).Plus(p.X)
}

func (m *Matrix[T]) IndexToPt(i Int) (p Pt) {
	p.X = i.Mod(I(NCols))
	p.Y = i.DivBy(I(NCols))
	return
}

func (m *Matrix[T]) RandomPos(r *Rand) Pt {
	var pt Pt
	pt.X = r.RInt(ZERO, I(NCols).Minus(ONE))
	pt.Y = r.RInt(ZERO, I(NCols).Minus(ONE))
	return pt
}
