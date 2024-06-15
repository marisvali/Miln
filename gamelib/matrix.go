package gamelib

type Matrix struct {
	cells []Int
	size  Pt
}

func (m *Matrix) Clone() (c Matrix) {
	c.size = m.size
	c.cells = append(c.cells, m.cells...)
	return
}

func (m *Matrix) Init(size Pt) {
	m.size = size
	m.cells = make([]Int, size.Y.Times(size.X).ToInt64())
}

func (m *Matrix) Set(pos Pt, val Int) {
	m.cells[pos.Y.Times(m.size.X).Plus(pos.X).ToInt64()] = val
}

func (m *Matrix) Get(pos Pt) Int {
	return m.cells[pos.Y.Times(m.size.X).Plus(pos.X).ToInt64()]
}

func (m *Matrix) InBounds(pt Pt) bool {
	return pt.X.IsNonNegative() &&
		pt.Y.IsNonNegative() &&
		pt.Y.Lt(m.size.Y) &&
		pt.X.Lt(m.size.X)
}

func (m *Matrix) Size() Pt {
	return m.size
}

func (m *Matrix) PtToIndex(p Pt) Int {
	return p.Y.Times(m.size.X).Plus(p.X)
}

func (m *Matrix) IndexToPt(i Int) (p Pt) {
	p.X = i.Mod(m.size.X)
	p.Y = i.DivBy(m.size.X)
	return
}

func (m *Matrix) RPos() Pt {
	var pt Pt
	pt.X = RInt(ZERO, m.Size().X.Minus(ONE))
	pt.Y = RInt(ZERO, m.Size().Y.Minus(ONE))
	return pt
}
