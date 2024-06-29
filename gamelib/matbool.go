package gamelib

import "fmt"

type MatBool struct {
	Matrix[bool]
}

func NewBoolMatrix(size Pt) (m MatBool) {
	m.size = size
	m.cells = make([]bool, size.Y.Times(size.X).ToInt64())
	return m
}

func (m *MatBool) Occupy(pos Pt) {
	m.Set(pos, true)
}

func (m *MatBool) Clear(pos Pt) {
	m.Set(pos, false)
}

// Add performs the union operation between the two sets represented by the
// matrices.
func (m *MatBool) Add(other MatBool) {
	if m.size != other.size {
		Check(fmt.Errorf("trying to combine matrices of different sizes: "+
			"(%d, %d) and (%d, %d)", m.size.X.ToInt(), m.size.Y.ToInt(),
			other.size.X.ToInt(), other.size.Y.ToInt()))
	}

	for i := range m.cells {
		m.cells[i] = m.cells[i] || other.cells[i]
	}
}

// IntersectWith performs the intersection operation between the two sets
// represented by the matrices.
func (m *MatBool) IntersectWith(other MatBool) {
	if m.size != other.size {
		Check(fmt.Errorf("trying to combine matrices of different sizes: "+
			"(%d, %d) and (%d, %d)", m.size.X.ToInt(), m.size.Y.ToInt(),
			other.size.X.ToInt(), other.size.Y.ToInt()))
	}

	for i := range m.cells {
		m.cells[i] = m.cells[i] && other.cells[i]
	}
}

func (m *MatBool) RandomUnoccupiedPos() (p Pt) {
	for {
		p = m.RandomPos()
		if !m.Get(p) {
			return
		}
	}
}

func (m *MatBool) OccupyRandomPos() (p Pt) {
	p = m.RandomUnoccupiedPos()
	m.Occupy(p)
	return
}
