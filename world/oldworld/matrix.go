package oldworld

import (
	"bytes"
	"encoding/gob"
)

type Matrix[T any] struct {
	cells []T
	size  Pt
}

func (m Matrix[T]) GobEncode() ([]byte, error) {
	// Creating a new encoder/decoder every time can be slow for small
	// structs. For a Matrix, it may be ok. But if you ever want to make it
	// more efficient, the direction from the gob package implementers is to
	// do something lightweight in your implementation of the interface. That
	// can mean creating a static buffer and encoder, and resetting the buffer
	// every time so as to not re-allocate buffer memory every time.
	// Or, just make the private fields exportable and comment that no one else
	// should use these fields.
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(m.size); err != nil {
		return nil, err
	}
	if err := encoder.Encode(m.cells); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m *Matrix[T]) GobDecode(b []byte) error {
	// See GobEncode for tips on efficiency.
	buf := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buf)
	if err := decoder.Decode(&m.size); err != nil {
		return err
	}
	if err := decoder.Decode(&m.cells); err != nil {
		return err
	}
	return nil
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
