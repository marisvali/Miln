package gamelib

import (
	"github.com/stretchr/testify/assert"
	"os"
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

func RunYamlTest(t *testing.T, m MatBool) {
	fsys := os.DirFS(".")
	filename := "m.txt"
	SaveYAML(filename, m)

	var m2 MatBool
	LoadYAML(fsys, filename, &m2)
	DeleteFile(filename)
	assert.Equal(t, m, m2)
}

func Test_Yaml(t *testing.T) {
	var m MatBool
	m = NewMatBool(IPt(0, 0))
	RunYamlTest(t, m)

	m = NewMatBool(IPt(1, 1))
	RunYamlTest(t, m)

	m = NewMatBool(IPt(1, 1))
	m.Set(IPt(0, 0))
	RunYamlTest(t, m)

	m = NewMatBool(IPt(3, 2))
	m.Set(IPt(1, 1))
	m.Set(IPt(0, 0))
	m.Set(IPt(0, 1))
	RunYamlTest(t, m)

	m = NewMatBool(IPt(3, 1))
	m.Set(IPt(0, 0))
	m.Set(IPt(2, 0))
	RunYamlTest(t, m)

	m = NewMatBool(IPt(1, 3))
	m.Set(IPt(0, 0))
	m.Set(IPt(0, 1))
	m.Set(IPt(0, 2))
	RunYamlTest(t, m)

	m = NewMatBool(IPt(15, 22))
	for y := ZERO; y.Lt(m.size.Y); y.Inc() {
		for x := ZERO; x.Lt(m.size.X); x.Inc() {
			if y.Plus(x).ToInt64()%2 == 0 {
				m.Set(Pt{x, y})
			}
		}
	}
	RunYamlTest(t, m)
}
