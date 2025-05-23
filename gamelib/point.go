package gamelib

import "fmt"

type Pt struct {
	X Int
	Y Int
}

func IPt(x, y int) Pt {
	return Pt{I(x), I(y)}
}

func (p Pt) SquaredDistTo(other Pt) Int {
	return p.To(other).SquaredLen()
}

func (p *Pt) Add(other Pt) {
	p.X = p.X.Plus(other.X)
	p.Y = p.Y.Plus(other.Y)
}

func (p Pt) Plus(other Pt) Pt {
	return Pt{p.X.Plus(other.X), p.Y.Plus(other.Y)}
}

func (p Pt) Minus(other Pt) Pt {
	return Pt{p.X.Minus(other.X), p.Y.Minus(other.Y)}
}

func (p *Pt) Subtract(other Pt) {
	p.X = p.X.Minus(other.X)
	p.Y = p.Y.Minus(other.Y)
}

func (p Pt) Times(multiply Int) Pt {
	return Pt{p.X.Times(multiply), p.Y.Times(multiply)}
}

func (p Pt) DivBy(divide Int) Pt {
	return Pt{p.X.DivBy(divide), p.Y.DivBy(divide)}
}

// Reflected returns p reflected around vec.
// We don't assume that vec is normalized.
func (p Pt) Reflected(vec Pt) Pt {
	// r = p − 2(p⋅vec)vec
	// where p⋅vec is the dot product and vec is normalized
	l := vec.Len()
	// Compute dot product but normalize vec at the same time.
	// This way we don't normalize vec prematurely. We multiply it by p first,
	// before dividing it by l, this way we don't lose precision.
	dotX := p.X.Times(vec.X).DivBy(l)
	dotY := p.Y.Times(vec.Y).DivBy(l)
	dot := dotX.Plus(dotY)
	// Again, multiply vec by some dot product, and only then divide by the vec
	// length. This way we normalize vec but we don't lose precision.
	intermediate := vec.Times(dot.Times(I(2))).DivBy(l)
	return p.Minus(intermediate)
}

// Reflect p around vec.
// We don't assume that vec is normalized.
func (p *Pt) Reflect(vec Pt) {
	// r = p − 2(p⋅vec)vec
	// where p⋅vec is the dot product and vec is normalized
	l := vec.Len()
	// Compute dot product but normalize vec at the same time.
	// This way we don't normalize vec prematurely. We multiply it by p first,
	// before dividing it by l, this way we don't lose precision.
	dotX := p.X.Times(vec.X).DivBy(l)
	dotY := p.Y.Times(vec.Y).DivBy(l)
	dot := dotX.Plus(dotY)
	// Again, multiply vec by some dot product, and only then divide by the vec
	// length. This way we normalize vec but we don't lose precision.
	intermediate := vec.Times(dot.Times(I(2))).DivBy(l)
	p.Subtract(intermediate)
}

func (p *Pt) Scale(multiply Int, divide Int) {
	p.X = p.X.Times(multiply).DivBy(divide)
	p.Y = p.Y.Times(multiply).DivBy(divide)
}

func (p Pt) SquaredLen() Int {
	return p.X.Sqr().Plus(p.Y.Sqr())
}

func (p Pt) Len() Int {
	return p.SquaredLen().Sqrt()
}

func (p Pt) To(other Pt) Pt {
	return Pt{other.X.Minus(p.X), other.Y.Minus(p.Y)}
}

func (p Pt) Dot(other Pt) Int {
	return p.X.Times(other.X).Plus(p.Y.Times(other.Y))
}

func (p *Pt) SetLen(newLen Int) {
	oldLen := p.Len()
	if oldLen.Eq(I(0)) {
		return
	}
	p.Scale(newLen, oldLen)
}

func (p *Pt) AddLen(extraLen Int) {
	oldLen := p.Len()
	newLen := oldLen.Plus(extraLen)
	if newLen.Leq(I(0)) {
		newLen = I(0)
		return
	}
	p.Scale(newLen, oldLen)
}

func (p Pt) Eq(other Pt) bool {
	return p.X.Eq(other.X) && p.Y.Eq(other.Y)
}

// MarshalYAML turns Pt into a string.
// Useful because if I just let the YAML library do the default marshalling, it
// will turn the X and Y fields into X and "Y" because Y is shorthand for
// "yes/true" in YAML. Plus, it's shorter and easier to read two numbers in
// a single line.
// With this custom marshalling:
// Pos: [5, 19]
// Without this custom marshalling:
// Pos:
// - X: 5
// - "Y": 19
func (p Pt) MarshalYAML() ([]byte, error) {
	s := fmt.Sprintf("[%d, %d]", p.X.ToInt64(), p.Y.ToInt64())
	return []byte(s), nil
}

func (p *Pt) UnmarshalYAML(b []byte) error {
	s := string(b)
	var x, y int64
	n, err := fmt.Sscanf(s, "[%d, %d]", &x, &y)
	p.X, p.Y = I64(x), I64(y)
	if n != 2 {
		Check(fmt.Errorf("failed to get exactly 2 int64 from string %s", s))
	}
	return err
}
