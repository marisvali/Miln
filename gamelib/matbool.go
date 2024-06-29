package gamelib

import "fmt"

type MatBool struct {
	Matrix[bool]
}

func NewMatBool(size Pt) (m MatBool) {
	m.size = size
	m.cells = make([]bool, size.Y.Times(size.X).ToInt64())
	return m
}

func (m *MatBool) Clone() (c MatBool) {
	c.size = m.size
	c.cells = append(c.cells, m.cells...)
	return
}

func (m *MatBool) At(pos Pt) bool {
	return m.Get(pos)
}

func (m *MatBool) Set(pos Pt) {
	m.Matrix.Set(pos, true)
}

func (m *MatBool) Clear(pos Pt) {
	m.Matrix.Set(pos, false)
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
	m.Set(p)
	return
}

// ConnectedPositions returns a bool matrix that shows all the positions
// connected to the start point. A position is connected if there is a path
// between it and the start point, where all the elements of the path have the
// same value as the start point.
func (m MatBool) ConnectedPositions(start Pt) (res MatBool) {
	goodVal := m.Get(start)
	queue := []Pt{}
	queue = append(queue, start)
	res = NewMatBool(m.size)
	res.Set(start)
	i := 0
	dirs := Directions8()
	for i < len(queue) {
		pt := queue[i]
		i++
		for _, d := range dirs {
			newPt := pt.Plus(d)
			if m.InBounds(newPt) && !res.At(newPt) && m.Get(newPt) == goodVal {
				res.Set(newPt)
				queue = append(queue, newPt)
			}
		}
	}
	return
}
