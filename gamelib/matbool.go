package gamelib

import (
	"strings"
)

type MatBool struct {
	Matrix[bool]
}

func (m *MatBool) At(pos Pt) bool {
	return m.Get(pos)
}

func (m *MatBool) Set(pos Pt) {
	m.Matrix.Set(pos, true)
}

func (m *MatBool) SetAll() {
	for i := range m.cells {
		m.cells[i] = true
	}
}

func (m *MatBool) Clear(pos Pt) {
	m.Matrix.Set(pos, false)
}

func (m *MatBool) ClearAll() {
	for i := range m.cells {
		m.cells[i] = false
	}
}

// Add performs the union operation between the two sets represented by the
// matrices.
func (m *MatBool) Add(other MatBool) {
	for i := range m.cells {
		m.cells[i] = m.cells[i] || other.cells[i]
	}
}

// Subtract performs the subtraction operation between the two sets represented
// by the matrices. As in, what's true in other becomes false in m.
func (m *MatBool) Subtract(other MatBool) {
	for i := range m.cells {
		m.cells[i] = m.cells[i] && !other.cells[i]
	}
}

// IntersectWith performs the intersection operation between the two sets
// represented by the matrices.
func (m *MatBool) IntersectWith(other MatBool) {
	for i := range m.cells {
		m.cells[i] = m.cells[i] && other.cells[i]
	}
}

// Negate changes the matrix so that each position has the opposite value (true
// becomes false, false becomes true).
func (m *MatBool) Negate() {
	for i := range m.cells {
		m.cells[i] = !m.cells[i]
	}
	return
}

func (m MatBool) RandomUnoccupiedPos(r *Rand) (p Pt) {
	for {
		p = m.RandomPos(r)
		if !m.Get(p) {
			return
		}
	}
}

func (m *MatBool) OccupyRandomPos(r *Rand) (p Pt) {
	p = m.RandomUnoccupiedPos(r)
	m.Set(p)
	return
}

// ConnectedPositions returns a bool matrix that shows all the positions
// connected to the start point. A position is connected if there is a path
// between it and the start point, where all the elements of the path have the
// same value as the start point.
// res.At(pt) == true if pt is connected to start
// res.At(pt) == false if pt is NOT connected to start
type posqueueArray struct {
	N int64
	V [NCols * NRows]Pt
}

var posqueue posqueueArray

func (m MatBool) ConnectedPositions(start Pt) (res MatBool) {
	goodVal := m.Get(start)
	posqueue.V[0] = start
	posqueue.N = 1
	res.Set(start)
	i := int64(0)
	dirs := Directions8()
	for i < posqueue.N {
		pt := posqueue.V[i]
		i++
		for _, d := range dirs {
			newPt := pt.Plus(d)
			if m.InBounds(newPt) && !res.At(newPt) && m.Get(newPt) == goodVal {
				res.Set(newPt)
				posqueue.V[posqueue.N] = newPt
				posqueue.N++
			}
		}
	}
	return
}

var MatBoolArray posqueueArray

func (m MatBool) ToSlice() []Pt {
	MatBoolArray.N = 0
	for i := range m.cells {
		if m.cells[i] {
			MatBoolArray.V[MatBoolArray.N] = IPt(i%NCols, i/NCols)
			MatBoolArray.N++
		}
	}
	return MatBoolArray.V[:MatBoolArray.N]
}

func (m *MatBool) FromSlice(s []Pt) {
	for i := range s {
		m.Set(s[i])
	}
}

func (m MatBool) MarshalYAML() ([]byte, error) {
	var s string

	for i := 0; i < NRows; i++ {
		row := m.cells[NCols*i : NCols*(i+1)]
		var rowS string
		for j := 0; j < len(row)-1; j++ {
			if row[j] {
				rowS += "X,"
			} else {
				rowS += ".,"
			}
		}
		if row[len(row)-1] {
			rowS += "X"
		} else {
			rowS += "."
		}
		s += "- [" + rowS + "]\n"
	}
	return []byte(s), nil
}

func (m *MatBool) UnmarshalYAML(b []byte) error {
	s := string(b)

	if strings.TrimSpace(s) == "empty" {
		m.cells = [64]bool{}
		return nil
	}

	rows := strings.Split(s, "\n")
	for rowIdx := range rows {
		trimmedRow := strings.TrimSpace(rows[rowIdx])
		innerRow := trimmedRow[3 : len(trimmedRow)-1]
		tokens := strings.Split(innerRow, ",")
		for cellIdx, token := range tokens {
			if strings.TrimSpace(token) == "X" {
				m.Set(IPt(cellIdx, rowIdx))
			}
		}
	}
	return nil
}
