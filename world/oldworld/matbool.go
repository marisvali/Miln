package oldworld

import (
	"fmt"
	"slices"
	"strings"
)

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
	if m.size != other.size {
		Check(fmt.Errorf("trying to combine matrices of different sizes: "+
			"(%d, %d) and (%d, %d)", m.size.X.ToInt(), m.size.Y.ToInt(),
			other.size.X.ToInt(), other.size.Y.ToInt()))
	}

	for i := range m.cells {
		m.cells[i] = m.cells[i] || other.cells[i]
	}
}

// Subtract performs the subtraction operation between the two sets represented
// by the matrices. As in, what's true in other becomes false in m.
func (m *MatBool) Subtract(other MatBool) {
	if m.size != other.size {
		Check(fmt.Errorf("trying to combine matrices of different sizes: "+
			"(%d, %d) and (%d, %d)", m.size.X.ToInt(), m.size.Y.ToInt(),
			other.size.X.ToInt(), other.size.Y.ToInt()))
	}

	for i := range m.cells {
		m.cells[i] = m.cells[i] && !other.cells[i]
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

// Negate changes the matrix so that each position has the opposite value (true
// becomes false, false becomes true).
func (m *MatBool) Negate() {
	for i := range m.cells {
		m.cells[i] = !m.cells[i]
	}
	return
}

func (m MatBool) ToSlice() (s []Pt) {
	nCols := m.size.X.ToInt()
	for i := range m.cells {
		if m.cells[i] {
			s = append(s, IPt(i%nCols, i/nCols))
		}
	}
	return
}

func (m *MatBool) FromSlice(s []Pt) {
	for i := range s {
		m.Set(s[i])
	}
}

func (m *MatBool) Equal(o MatBool) bool {
	return m.size == o.size && slices.Equal(m.cells, o.cells)
}

func (m MatBool) MarshalYAML() ([]byte, error) {
	var s string
	nCols := m.size.X.ToInt()
	nRows := m.size.Y.ToInt()

	if nRows == 0 {
		return []byte("empty"), nil
	}

	for i := 0; i < nRows; i++ {
		row := m.cells[nCols*i : nCols*(i+1)]
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
		m.size = Pt{}
		m.cells = []bool{}
		return nil
	}

	rows := strings.Split(s, "\n")
	for rowIdx := range rows {
		trimmedRow := strings.TrimSpace(rows[rowIdx])
		innerRow := trimmedRow[3 : len(trimmedRow)-1]
		tokens := strings.Split(innerRow, ",")
		zero := Pt{}
		if m.size == zero {
			m.size = IPt(len(tokens), len(rows))
			m.cells = make([]bool, m.size.Y.Times(m.size.X).ToInt64())
		}
		for cellIdx, token := range tokens {
			if strings.TrimSpace(token) == "X" {
				m.Set(IPt(cellIdx, rowIdx))
			}
		}
	}
	return nil
}
