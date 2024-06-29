package gamelib

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Add(t *testing.T) {
	result := NewBoolMatrix(IPt(4, 4))
	result.Occupy(IPt(1, 2))
	result.Occupy(IPt(3, 3))

	m2 := NewBoolMatrix(IPt(4, 4))
	m2.Occupy(IPt(1, 2))
	m2.Occupy(IPt(2, 2))
	m2.Occupy(IPt(0, 0))

	expected := NewBoolMatrix(IPt(4, 4))
	expected.Occupy(IPt(1, 2))
	expected.Occupy(IPt(2, 2))
	expected.Occupy(IPt(0, 0))
	expected.Occupy(IPt(3, 3))

	result.Add(m2)

	assert.Equal(t, result, expected)
}

func Test_Intersect(t *testing.T) {
	result := NewBoolMatrix(IPt(4, 4))
	result.Occupy(IPt(1, 2))
	result.Occupy(IPt(3, 3))

	m2 := NewBoolMatrix(IPt(4, 4))
	m2.Occupy(IPt(1, 2))
	m2.Occupy(IPt(2, 2))
	m2.Occupy(IPt(0, 0))

	expected := NewBoolMatrix(IPt(4, 4))
	expected.Occupy(IPt(1, 2))

	result.IntersectWith(m2)

	assert.Equal(t, result, expected)
}
