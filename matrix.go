package main

import (
	. "playful-patterns.com/miln/ints"
)

type Matrix struct {
	cells []Int
	nRows Int
	nCols Int
}

func (m *Matrix) Clone() (c Matrix) {
	c.nRows = m.nRows
	c.nCols = m.nCols
	c.cells = append(c.cells, m.cells...)
	return
}

func (m *Matrix) Init(nRows, nCols Int) {
	m.nRows = nRows
	m.nCols = nCols
	m.cells = make([]Int, nRows.Times(nCols).ToInt64())
}

func (m *Matrix) Set(row, col, val Int) {
	m.cells[row.Times(m.nCols).Plus(col).ToInt64()] = val
}

func (m *Matrix) Get(row, col Int) Int {
	return m.cells[row.Times(m.nCols).Plus(col).ToInt64()]
}

func (m *Matrix) InBounds(pt Pt) bool {
	return pt.X.IsNonNegative() &&
		pt.Y.IsNonNegative() &&
		pt.Y.Lt(m.nRows) &&
		pt.X.Lt(m.nCols)
}

func (m *Matrix) NRows() Int {
	return m.nRows
}

func (m *Matrix) NCols() Int {
	return m.nCols
}

func (m *Matrix) PtToIndex(p Pt) Int {
	return p.Y.Times(m.nCols).Plus(p.X)
}

func (m *Matrix) IndexToPt(i Int) (p Pt) {
	p.X = i.Mod(m.nCols)
	p.Y = i.DivBy(m.nCols)
	return
}
