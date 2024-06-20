package gamelib

import (
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"image/color"
)

func DrawSpriteXY(screen *ebiten.Image, img *ebiten.Image,
	x float64, y float64) {
	op := &ebiten.DrawImageOptions{}

	op.Blend.BlendFactorSourceRGB = ebiten.BlendFactorSourceAlpha
	op.Blend.BlendFactorSourceAlpha = ebiten.BlendFactorSourceAlpha
	op.Blend.BlendFactorDestinationRGB = ebiten.BlendFactorOneMinusSourceAlpha
	op.Blend.BlendFactorDestinationAlpha = ebiten.BlendFactorOneMinusSourceAlpha
	op.Blend.BlendOperationAlpha = ebiten.BlendOperationAdd
	op.Blend.BlendOperationRGB = ebiten.BlendOperationAdd

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

	op.Blend.BlendFactorSourceRGB = ebiten.BlendFactorSourceAlpha
	op.Blend.BlendFactorSourceAlpha = ebiten.BlendFactorSourceAlpha
	op.Blend.BlendFactorDestinationRGB = ebiten.BlendFactorOneMinusSourceAlpha
	op.Blend.BlendFactorDestinationAlpha = ebiten.BlendFactorOneMinusSourceAlpha
	op.Blend.BlendOperationAlpha = ebiten.BlendOperationAdd
	op.Blend.BlendOperationRGB = ebiten.BlendOperationAdd

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

	op.Blend.BlendFactorSourceRGB = ebiten.BlendFactorSourceAlpha
	op.Blend.BlendFactorSourceAlpha = ebiten.BlendFactorSourceAlpha
	op.Blend.BlendFactorDestinationRGB = ebiten.BlendFactorOneMinusSourceAlpha
	op.Blend.BlendFactorDestinationAlpha = ebiten.BlendFactorOneMinusSourceAlpha
	op.Blend.BlendOperationAlpha = ebiten.BlendOperationAdd
	op.Blend.BlendOperationRGB = ebiten.BlendOperationAdd

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

	op.Blend.BlendFactorSourceRGB = ebiten.BlendFactorSourceAlpha
	op.Blend.BlendFactorSourceAlpha = ebiten.BlendFactorSourceAlpha
	op.Blend.BlendFactorDestinationRGB = ebiten.BlendFactorOneMinusSourceAlpha
	op.Blend.BlendFactorDestinationAlpha = ebiten.BlendFactorOneMinusSourceAlpha
	op.Blend.BlendOperationAlpha = ebiten.BlendOperationAdd
	op.Blend.BlendOperationRGB = ebiten.BlendOperationAdd

	op.GeoM.Translate(r.Min().X.ToFloat64(), r.Min().Y.ToFloat64())
	screen.DrawImage(img, op)
}

func DrawFilledSquare(screen *ebiten.Image, s Square, col color.Color) {
	img := ebiten.NewImage(s.Size.ToInt(), s.Size.ToInt())
	img.Fill(col)

	op := &ebiten.DrawImageOptions{}

	op.Blend.BlendFactorSourceRGB = ebiten.BlendFactorSourceAlpha
	op.Blend.BlendFactorSourceAlpha = ebiten.BlendFactorSourceAlpha
	op.Blend.BlendFactorDestinationRGB = ebiten.BlendFactorOneMinusSourceAlpha
	op.Blend.BlendFactorDestinationAlpha = ebiten.BlendFactorOneMinusSourceAlpha
	op.Blend.BlendOperationAlpha = ebiten.BlendOperationAdd
	op.Blend.BlendOperationRGB = ebiten.BlendOperationAdd

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
