package gamelib

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Add(t *testing.T) {
	result := NewMatBool(IPt(4, 4))
	result.Set(IPt(1, 2))
	result.Set(IPt(3, 3))

	m2 := NewMatBool(IPt(4, 4))
	m2.Set(IPt(1, 2))
	m2.Set(IPt(2, 2))
	m2.Set(IPt(0, 0))

	expected := NewMatBool(IPt(4, 4))
	expected.Set(IPt(1, 2))
	expected.Set(IPt(2, 2))
	expected.Set(IPt(0, 0))
	expected.Set(IPt(3, 3))

	result.Add(m2)

	assert.Equal(t, result, expected)
}

func Test_Intersect(t *testing.T) {
	result := NewMatBool(IPt(4, 4))
	result.Set(IPt(1, 2))
	result.Set(IPt(3, 3))

	m2 := NewMatBool(IPt(4, 4))
	m2.Set(IPt(1, 2))
	m2.Set(IPt(2, 2))
	m2.Set(IPt(0, 0))

	expected := NewMatBool(IPt(4, 4))
	expected.Set(IPt(1, 2))

	result.IntersectWith(m2)

	assert.Equal(t, result, expected)
}
