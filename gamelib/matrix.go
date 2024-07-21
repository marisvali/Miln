package gamelib

type Matrix[T any] struct {
	cells []T
	size  Pt
}

func (m *Matrix[T]) Clone() (c Matrix[T]) {
	c.size = m.size
	c.cells = append(c.cells, m.cells...)
	return
}

func NewMatrix[T any](size Pt) (m Matrix[T]) {
	m.size = size
	m.cells = make([]T, size.Y.Times(size.X).ToInt64())
	return m
}

func (m *Matrix[T]) Set(pos Pt, val T) {
	m.cells[pos.Y.Times(m.size.X).Plus(pos.X).ToInt64()] = val
}

func (m *Matrix[T]) Get(pos Pt) T {
	return m.cells[pos.Y.Times(m.size.X).Plus(pos.X).ToInt64()]
}

func (m *Matrix[T]) InBounds(pt Pt) bool {
	return pt.X.IsNonNegative() &&
		pt.Y.IsNonNegative() &&
		pt.Y.Lt(m.size.Y) &&
		pt.X.Lt(m.size.X)
}

func (m *Matrix[T]) Size() Pt {
	return m.size
}

func (m *Matrix[T]) PtToIndex(p Pt) Int {
	return p.Y.Times(m.size.X).Plus(p.X)
}

func (m *Matrix[T]) IndexToPt(i Int) (p Pt) {
	p.X = i.Mod(m.size.X)
	p.Y = i.DivBy(m.size.X)
	return
}

func (m *Matrix[T]) RandomPos() Pt {
	var pt Pt
	pt.X = RInt(ZERO, m.Size().X.Minus(ONE))
	pt.Y = RInt(ZERO, m.Size().Y.Minus(ONE))
	return pt
}
