package gamelib

import (
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"image/color"
)

// Col creates a color from the red, green, blue, alpha components.
// Use this instead of color.RGBA{r, g, b, a} or color.NRGBA{r, g, b, a}.
// color.RGBA{r, g, b, a} is most likely not what you want (has unexpected
// behavior if you don't read its specification very carefully).
// color.NRGBA{r, g, b, a} is most likely what you want, but it is slightly
// inefficient.
// Col has the behavior you expect and is efficient.
func Col(r, g, b, a uint8) color.Color {
	// Detailed explanation:
	// - The first instinct is to create a color using color.RGBA{r, g, b, a}.
	// This is almost never what you want to do if your alpha isn't 255. As the
	// specification for color.RGBA says (emphasis mine):
	// "RGBA represents a traditional 32-bit **ALPHA-PREMULTIPLIED** color,
	// having 8 bits for each of red, green, blue and alpha.
	// An alpha-premultiplied color component **C HAS BEEN SCALED** by alpha
	// (A), so has **VALID VALUES 0 <= C <= A**."
	// This is not what you expect! You expect a color of {255, 0, 0, 100} to be
	// a transparent red. This is WRONG if you do color.RGBA{255, 0, 0, 100}.
	// The red must be scaled by the alpha, so what you should do is
	// color.RGBA{100, 0, 0, 100}. But nobody bothers to specify colors this
	// way, when coding. But if you don't specify colors like this, you will get
	// strange effects like your alpha being ignored. Then you might think
	// something is wrong with the draw function and the blending options.
	// But no, the default blending options of ebiten.Image.DrawImage() are
	// exactly what you want, no need to tweak them.
	// - There is a color struct that behaves like you expect: color.NRGBA.
	// However, if you use it in a tight loop, like a DrawLine(), you call the
	// color.NRGBA.RGBA() function, which converts from non-alpha-premultiplied
	// to alpha-premultiplied. Doing this every time is silly, so might as well
	// have a color that is already alpha-premultiplied. So, Col() gets the
	// input that seems natural to a coder (non-alpha-premultiplied) and returns
	// an alpha-premultiplied color.
	// - There is already a conversion function in the color package, so use
	// that.
	return color.RGBA64Model.Convert(color.NRGBA{r, g, b, a})
	// PS: worrying about the efficiency of the color structure is kind of silly
	// if you're going to use the color.Color interface. And you're going to use
	// it, because that's what most functions take. But a RGBA64 color struct
	// has 8 bytes. A color.Color variable has 16 bytes: pointer to the value
	// and 8 bytes indicating the type.
	// Using RGBA64 instead of NRGBA gets you from 3.571 ns/op to 2.156 ns/op.
	// But using RGBA64 directly instead of going through color.Color gets you
	// from 2.156 ns/op to 0.7282 ns/op.
	// But even though this function returns a RGBA64, all graphics functions
	// will just receive color.Color, and it will defeat the whole point.
	// So, you can't worry too much about efficiency when working with colors.
	// The most important thing about Col() is correctness.
	// Why not just use NRGBA everywhere then? It even has the nice effect that
	// you can directly check color components like .R or .G and compare them
	// to some number (e.g. col.R > 10). And the values are not
	// alpha-premultiplied so you can actually reason about them. The issue is
	// that you could only do that for color values you create, but not things
	// colors obtained from images (e.g. imgBeam.At(0, 0)). So now you have to
	// deal with treating some colors one way and some colors another way. So,
	// since the color.Color interface decided to make alpha-premultiplied the
	// default, I guess it's better to be uniform than have extra convenience
	// some of the time, but always remember there's two systems.
}

func DrawSpriteXY(screen *ebiten.Image, img *ebiten.Image,
	x float64, y float64) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(screen.Bounds().Min.X)+x, float64(screen.Bounds().Min.Y)+y)
	screen.DrawImage(img, op)
}

func DrawSpriteAlpha(screen *ebiten.Image, img *ebiten.Image,
	x float64, y float64, targetWidth float64, targetHeight float64, alpha uint8) {
	op := &ebiten.DrawImageOptions{}

	// Resize image to fit the target size we want to draw.
	// This kind of scaling is very useful during development when the final
	// sizes are not decided, and thus it's impossible to have final sprites.
	// For an actual release, scaling should be avoided.
	imgSize := img.Bounds().Size()
	newDx := targetWidth / float64(imgSize.X)
	newDy := targetHeight / float64(imgSize.Y)
	op.GeoM.Scale(newDx, newDy)
	op.GeoM.Translate(float64(screen.Bounds().Min.X)+x, float64(screen.Bounds().Min.Y)+y)
	op.ColorScale.SetA(float32(alpha) / 255)
	screen.DrawImage(img, op)
}

func DrawSprite(screen *ebiten.Image, img *ebiten.Image,
	x float64, y float64, targetWidth float64, targetHeight float64) {
	op := &ebiten.DrawImageOptions{}

	// Resize image to fit the target size we want to draw.
	// This kind of scaling is very useful during development when the final
	// sizes are not decided, and thus it's impossible to have final sprites.
	// For an actual release, scaling should be avoided.
	imgSize := img.Bounds().Size()
	newDx := targetWidth / float64(imgSize.X)
	newDy := targetHeight / float64(imgSize.Y)
	op.GeoM.Scale(newDx, newDy)
	op.GeoM.Translate(float64(screen.Bounds().Min.X)+x, float64(screen.Bounds().Min.Y)+y)
	screen.DrawImage(img, op)
}

func DrawPixel(screen *ebiten.Image, pt Pt, color color.Color) {
	size := I(2)
	m := screen.Bounds().Min
	for ax := pt.X.Minus(size); ax.Leq(pt.X.Plus(size)); ax.Inc() {
		for ay := pt.Y.Minus(size); ay.Leq(pt.Y.Plus(size)); ay.Inc() {
			screen.Set(m.X+ax.ToInt(), m.Y+ay.ToInt(), color)
		}
	}
}

func DrawLine(screen *ebiten.Image, l Line, color color.Color) {
	x1 := l.Start.X
	y1 := l.Start.Y
	x2 := l.End.X
	y2 := l.End.Y
	if x1.Gt(x2) {
		x1, x2 = x2, x1
		y1, y2 = y2, y1
	}

	dx := x2.Minus(x1)
	dy := y2.Minus(y1)
	if dx.IsZero() && dy.IsZero() {
		return // No line to draw.
	}

	if dx.Abs().Gt(dy.Abs()) {
		inc := dx.DivBy(dx.Abs())
		for x := x1; x.Neq(x2); x.Add(inc) {
			y := y1.Plus(x.Minus(x1).Times(dy).DivBy(dx))
			DrawPixel(screen, Pt{x, y}, color)
		}
	} else {
		inc := dy.DivBy(dy.Abs())
		for y := y1; y.Neq(y2); y.Add(inc) {
			x := x1.Plus(y.Minus(y1).Times(dx).DivBy(dy))
			DrawPixel(screen, Pt{x, y}, color)
		}
	}
}

func DrawFilledRect(screen *ebiten.Image, r Rectangle, col color.Color) {
	img := ebiten.NewImage(r.Width().ToInt(), r.Height().ToInt())
	img.Fill(col)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(r.Min().X.ToFloat64(), r.Min().Y.ToFloat64())
	screen.DrawImage(img, op)
}

func DrawFilledSquare(screen *ebiten.Image, s Square, col color.Color) {
	img := ebiten.NewImage(s.Size.ToInt(), s.Size.ToInt())
	img.Fill(col)

	op := &ebiten.DrawImageOptions{}
	x := s.Center.X.Minus(s.Size.DivBy(TWO)).ToFloat64()
	y := s.Center.Y.Minus(s.Size.DivBy(TWO)).ToFloat64()
	op.GeoM.Translate(x, y)
	screen.DrawImage(img, op)
}

func DrawSquare(screen *ebiten.Image, s Square, color color.Color) {
	halfSize := s.Size.DivBy(I(2)).Plus(s.Size.Mod(I(2)))

	// square corners
	upperLeftCorner := Pt{s.Center.X.Minus(halfSize), s.Center.Y.Minus(halfSize)}
	lowerLeftCorner := Pt{s.Center.X.Minus(halfSize), s.Center.Y.Plus(halfSize)}
	upperRightCorner := Pt{s.Center.X.Plus(halfSize), s.Center.Y.Minus(halfSize)}
	lowerRightCorner := Pt{s.Center.X.Plus(halfSize), s.Center.Y.Plus(halfSize)}

	DrawLine(screen, Line{upperLeftCorner, upperRightCorner}, color)
	DrawLine(screen, Line{upperLeftCorner, lowerLeftCorner}, color)
	DrawLine(screen, Line{lowerLeftCorner, lowerRightCorner}, color)
	DrawLine(screen, Line{lowerRightCorner, upperRightCorner}, color)
}

func ToImagePoint(pt Pt) image.Point {
	return image.Point{pt.X.ToInt(), pt.Y.ToInt()}
}

func FromImagePoint(pt image.Point) Pt {
	return IPt(pt.X, pt.Y)
}

func ToImageRectangle(r Rectangle) image.Rectangle {
	return image.Rectangle{ToImagePoint(r.Min()), ToImagePoint(r.Max())}
}

func FromImageRectangle(r image.Rectangle) Rectangle {
	return Rectangle{FromImagePoint(r.Min), FromImagePoint(r.Max)}
}

func SubImage(screen *ebiten.Image, r Rectangle) *ebiten.Image {
	// Do this because when dealing with sub-images in general I think in
	// relative coordinates. So for img2 = img1.SubImage(pt1, pt2) I now expect
	// that img2.At(0, 0) indicates the same pixel as img1.At(pt1). Ebitengine
	// doesn't do it like that. I still need to use img2.At(pt1) to indicate
	// pixel img1.At(pt1). I don't know why Ebitengine does it like that.
	// Personally, I'm used to a different style, one of the main reasons for
	// working with subimages, for me, is to be able to think in local
	// coordinates instead of global ones.
	minPt := FromImagePoint(screen.Bounds().Min)
	r.Corner1.Add(minPt)
	r.Corner2.Add(minPt)
	return screen.SubImage(ToImageRectangle(r)).(*ebiten.Image)
}
