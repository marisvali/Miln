package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"image/color"
	"math"
	. "playful-patterns.com/miln/ints"
)

type Player struct {
	Pos Pt
}

type Enemy struct {
	Pos       Pt
	FuturePos Pt
}

type World struct {
	Player    Player
	Enemy     Enemy
	Obstacles Matrix
	TimeStep  Int
}

type Gui struct {
	defaultFont    font.Face
	imgGround      *ebiten.Image
	imgTree        *ebiten.Image
	imgPlayer      *ebiten.Image
	imgEnemy       *ebiten.Image
	imgEnemyShadow *ebiten.Image
	world          World
	frameIdx       Int
	pathfinding    Pathfinding
	screenSize     Pt
	leftClick      bool
	rightClick     bool
	leftClickPos   Pt
	rightClickPos  Pt
	beamIdx        Int
	beam           Line
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func (g *Gui) Update() error {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) {
		x, y := ebiten.CursorPosition()
		g.leftClick = true
		g.leftClickPos = IPt(x, y)
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton2) {
		x, y := ebiten.CursorPosition()
		g.rightClick = true
		g.rightClickPos = IPt(x, y)
	}

	g.frameIdx.Inc()
	if g.frameIdx.Mod(I(5)).Neq(ZERO) {
		return nil // skip update
	}

	g.world.TimeStep.Inc()
	if g.world.TimeStep.Eq(I(math.MaxInt64)) {
		// Damn.
		Check(fmt.Errorf("got to an unusually large time step: %d", g.world.TimeStep.ToInt64()))
	}

	// Get keyboard input.
	var pressedKeys []ebiten.Key
	pressedKeys = inpututil.AppendPressedKeys(pressedKeys)

	// Move the player.
	//if g.world.TimeStep.Mod(I(3)).Eq(ZERO) {
	//	moveLeft := slices.Contains(pressedKeys, ebiten.KeyA)
	//	moveUp := slices.Contains(pressedKeys, ebiten.KeyW)
	//	moveDown := slices.Contains(pressedKeys, ebiten.KeyS)
	//	moveRight := slices.Contains(pressedKeys, ebiten.KeyD)
	//
	//	if moveLeft {
	//		newPos := g.world.Player.Pos
	//		if g.world.Player.Pos.X.Gt(ZERO) {
	//			newPos.X.Dec()
	//		}
	//		if g.world.Obstacles.Get(newPos.Y, newPos.X).Eq(ZERO) {
	//			g.world.Player.Pos = newPos
	//		}
	//	}
	//
	//	if moveRight {
	//		newPos := g.world.Player.Pos
	//		if g.world.Player.Pos.X.Lt(g.world.Obstacles.NCols().Minus(I(1))) {
	//			newPos.X.Inc()
	//		}
	//		if g.world.Obstacles.Get(newPos.Y, newPos.X).Eq(ZERO) {
	//			g.world.Player.Pos = newPos
	//		}
	//	}
	//
	//	if moveUp {
	//		newPos := g.world.Player.Pos
	//		if g.world.Player.Pos.Y.Gt(ZERO) {
	//			newPos.Y.Dec()
	//		}
	//		if g.world.Obstacles.Get(newPos.Y, newPos.X).Eq(ZERO) {
	//			g.world.Player.Pos = newPos
	//		}
	//	}
	//
	//	if moveDown {
	//		newPos := g.world.Player.Pos
	//		if g.world.Player.Pos.Y.Lt(g.world.Obstacles.NRows().Minus(I(1))) {
	//			newPos.Y.Inc()
	//		}
	//		if g.world.Obstacles.Get(newPos.Y, newPos.X).Eq(ZERO) {
	//			g.world.Player.Pos = newPos
	//		}
	//	}
	//}

	if g.rightClick {
		g.rightClick = false

		enemyPos := g.TileToScreen(g.world.Enemy.Pos)
		dist := enemyPos.Minus(g.rightClickPos).Len()
		if dist.Lt(I(100)) {
			// Hit enemy.
			g.beamIdx = I(15)
			g.beam = Line{g.TileToScreen(g.world.Player.Pos), g.TileToScreen(g.world.Enemy.Pos)}
			if intersects, pt := g.LineObstaclesIntersection(g.beam); intersects {
				g.beam.End = pt
			}
		}
	}

	if g.leftClick {
		g.leftClick = false

		enemyPos := g.TileToScreen(g.world.Enemy.Pos)
		dist := enemyPos.Minus(g.leftClickPos).Len()
		if dist.Geq(I(100)) {
			sz := g.screenSize
			numX := g.world.Obstacles.NCols().ToInt()
			numY := g.world.Obstacles.NRows().ToInt()
			blockWidth := sz.X.ToFloat64() / float64(numX)
			blockHeight := sz.Y.ToFloat64() / float64(numY)

			// Translate from screen coordinates to grid coordinates.
			newPos := IPt(
				int(g.leftClickPos.X.ToFloat64()/blockWidth),
				int(g.leftClickPos.Y.ToFloat64()/blockHeight))
			if g.world.Obstacles.Get(newPos.Y, newPos.X).Eq(ZERO) {
				g.world.Player.Pos = newPos
			}
		}
	}

	// Move the enemy.
	if g.world.TimeStep.Mod(I(40)).Eq(ZERO) {
		path := g.pathfinding.FindPath(g.world.Enemy.Pos, g.world.Player.Pos)
		if len(path) > 1 {
			g.world.Enemy.Pos = path[1]
		}
	} else {
		path := g.pathfinding.FindPath(g.world.Enemy.Pos, g.world.Player.Pos)
		if len(path) > 1 {
			g.world.Enemy.FuturePos = path[1]
		}
	}

	return nil
}

func LineVerticalLineIntersection(l, vert Line) (bool, Pt) {
	// Check if the Lines even intersect.

	// Check if l's min X is at the right of vertX.
	minX, maxX := MinMax(l.Start.X, l.End.X)
	vertX := vert.Start.X // we assume vert.Start.X == vert.End.X

	if minX.Gt(vertX) {
		return false, Pt{}
	}

	// Or if l's max X is at the left of vertX.
	if maxX.Lt(vertX) {
		return false, Pt{}
	}

	//// Check if l's minY is under the vertMaxY.
	//minY, maxY := MinMax(l.Start.Y, l.End.Y)
	//vertMinY, vertMaxY := MinMax(vert.Start.Y, vert.End.Y)
	//
	//if minY.Gt(vertMaxY) {
	//	return false, Pt{}
	//}
	//
	//// Or if l's max Y is above vertMinY.
	//if maxY.Lt(vertMinY) {
	//	return false, Pt{}
	//}

	vertMinY, vertMaxY := MinMax(vert.Start.Y, vert.End.Y)

	// We know the intersection point will have the X coordinate equal to vertX.
	// We just need to compute the Y coordinate.
	// We have to move along the Y axis the same proportion that we moved along
	// the X axis in order to get to the intersection point.

	//factor := (vertX - l.Start.X) / (l.End.X - l.Start.X) // will always be positive
	//y := l.Start.Y + factor * (l.End.Y - l.Start.Y) // l.End.Y - l.Start.Y will
	// have the proper sign so that Y gets updated in the right direction
	//y := l.Start.Y + (vertX - l.Start.X) / (l.End.X - l.Start.X) * (l.End.Y - l.Start.Y)
	//y := l.Start.Y + (vertX - l.Start.X) * (l.End.Y - l.Start.Y) / (l.End.X - l.Start.X)
	var y Int
	if l.End.X.Eq(l.Start.X) {
		y = l.Start.Y
	} else {
		y = l.Start.Y.Plus((vertX.Minus(l.Start.X)).Times(l.End.Y.Minus(l.Start.Y)).DivBy(l.End.X.Minus(l.Start.X)))
	}

	if y.Lt(vertMinY) || y.Gt(vertMaxY) {
		return false, Pt{}
	} else {
		return true, Pt{vertX, y}
	}
}

func LineHorizontalLineIntersection(l, horiz Line) (bool, Pt) {
	// Check if the Lines even intersect.

	// Check if l's minY is under the vertY.
	minY, maxY := MinMax(l.Start.Y, l.End.Y)
	vertY := horiz.Start.Y // we assume vert.Start.Y == vert.End.Y

	if minY.Gt(vertY) {
		return false, Pt{}
	}

	// Or if l's max Y is above vertY.
	if maxY.Lt(vertY) {
		return false, Pt{}
	}

	//// Check if l's min X is at the right of vertMaxX.
	//minX, maxX := MinMax(l.Start.X, l.End.X)
	//vertMinX, vertMaxX := MinMax(horiz.Start.X, horiz.End.X)
	//
	//if minX.Gt(vertMaxX) {
	//	return false, Pt{}
	//}
	//
	//// Or if l's max X is at the left of vertMinX.
	//if maxX.Lt(vertMinX) {
	//	return false, Pt{}
	//}

	vertMinX, vertMaxX := MinMax(horiz.Start.X, horiz.End.X)

	// We know the intersection point will have the Y coordinate equal to vertY.
	// We just need to compute the X coordinate.
	// We have to move along the X axis the same proportion that we moved along
	// the Y axis in order to get to the intersection point.

	//factor := (vertY - l.Start.Y) / (l.End.Y - l.Start.Y) // will always be positive
	//x := l.Start.X + factor * (l.End.X - l.Start.X) // l.End.X - l.Start.X will
	// have the proper sign so that Y gets updated in the right direction
	//x := l.Start.X + (vertY - l.Start.Y) / (l.End.Y - l.Start.Y) * (l.End.X - l.Start.X)
	//x := l.Start.X + (vertY - l.Start.Y) * (l.End.X - l.Start.X) / (l.End.Y - l.Start.Y)
	var x Int
	if l.End.Y.Eq(l.Start.Y) {
		x = l.Start.X
	} else {
		x = l.Start.X.Plus((vertY.Minus(l.Start.Y)).Times(l.End.X.Minus(l.Start.X)).DivBy(l.End.Y.Minus(l.Start.Y)))
	}

	if x.Lt(vertMinX) || x.Gt(vertMaxX) {
		return false, Pt{}
	} else {
		return true, Pt{x, vertY}
	}
}

func colorHex(hexVal int) color.Color {
	if hexVal < 0x000000 || hexVal > 0xFFFFFF {
		panic(fmt.Sprintf("Invalid HEX value for color: %d", hexVal))
	}
	r := uint8(hexVal & 0xFF0000 >> 16)
	g := uint8(hexVal & 0x00FF00 >> 8)
	b := uint8(hexVal & 0x0000FF)
	return color.RGBA{
		R: r,
		G: g,
		B: b,
		A: 255,
	}
}

func DrawPixel(screen *ebiten.Image, pt Pt, color color.Color) {
	size := I(2)
	for ax := pt.X.Minus(size); ax.Leq(pt.X.Plus(size)); ax.Inc() {
		for ay := pt.Y.Minus(size); ay.Leq(pt.Y.Plus(size)); ay.Inc() {
			screen.Set(ax.ToInt(), ay.ToInt(), color)
		}
	}
}

type Line struct {
	Start Pt
	End   Pt
}

type Circle struct {
	Center   Pt
	Diameter Int
}

type Square struct {
	Center Pt
	Size   Int
}

func GetClosestPoint(pts []Pt, refPt Pt) (bool, Pt) {
	if len(pts) == 0 {
		return false, Pt{}
	}

	// Find the point closest to the reference point.
	minDist := refPt.SquaredDistTo(pts[0])
	minIdx := 0
	for idx := 1; idx < len(pts); idx++ {
		dist := refPt.SquaredDistTo(pts[idx])
		if dist.Lt(minDist) {
			minDist = dist
			minIdx = idx
		}
	}
	return true, pts[minIdx]
}

func LineSquareIntersection(l Line, s Square) (bool, Pt) {
	half := s.Size.DivBy(TWO)
	p1 := Pt{s.Center.X.Minus(half), s.Center.Y.Minus(half)}
	p2 := Pt{s.Center.X.Plus(half), s.Center.Y.Minus(half)}
	p3 := Pt{s.Center.X.Plus(half), s.Center.Y.Plus(half)}
	p4 := Pt{s.Center.X.Minus(half), s.Center.Y.Plus(half)}

	l1 := Line{p1, p2}
	l2 := Line{p2, p3}
	l3 := Line{p3, p4}
	l4 := Line{p4, p1}

	ipts := []Pt{}
	if intersects, ipt := LineHorizontalLineIntersection(l, l1); intersects {
		ipts = append(ipts, ipt)
	}
	if intersects, ipt := LineVerticalLineIntersection(l, l2); intersects {
		ipts = append(ipts, ipt)
	}
	if intersects, ipt := LineHorizontalLineIntersection(l, l3); intersects {
		ipts = append(ipts, ipt)
	}
	if intersects, ipt := LineVerticalLineIntersection(l, l4); intersects {
		ipts = append(ipts, ipt)
	}

	return GetClosestPoint(ipts, l.Start)
}

func EqualFloats(f1, f2 float64) bool {
	return math.Abs(f1-f2) < 0.000001
}

func (g *Gui) LineObstaclesIntersection(l Line) (bool, Pt) {
	sz := g.screenSize
	numX := g.world.Obstacles.NCols().ToInt()
	numY := g.world.Obstacles.NRows().ToInt()
	blockWidth := sz.X.ToFloat64() / float64(numX)
	blockHeight := sz.Y.ToFloat64() / float64(numY)

	rows := g.world.Obstacles.NRows()
	cols := g.world.Obstacles.NCols()
	ipts := []Pt{}
	for y := ZERO; y.Lt(rows); y.Inc() {
		for x := ZERO; x.Lt(cols); x.Inc() {
			if !g.world.Obstacles.Get(y, x).IsZero() {
				if !EqualFloats(blockWidth, blockHeight) {
					panic(fmt.Errorf("blocks are not squares"))
				}
				s := Square{g.TileToScreen(Pt{x, y}), I(int(blockWidth * 0.9))}
				if intersects, ipt := LineSquareIntersection(l, s); intersects {
					ipts = append(ipts, ipt)
				}
			}
		}
	}

	return GetClosestPoint(ipts, l.Start)
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

func (g *Gui) DrawTile(screen *ebiten.Image, img *ebiten.Image, pos Pt) {
	sz := screen.Bounds().Size()
	numX := g.world.Obstacles.NCols().ToInt()
	numY := g.world.Obstacles.NRows().ToInt()
	blockWidth := float64(sz.X) / float64(numX)
	blockHeight := float64(sz.Y) / float64(numY)
	margin := float64(1)
	posX := pos.X.ToFloat64() * blockWidth
	posY := pos.Y.ToFloat64() * blockHeight
	DrawSprite(screen, img, posX+margin, posY+margin, blockWidth-2*margin, blockHeight-2*margin)
}

func (g *Gui) TileToScreen(pos Pt) Pt {
	sz := g.screenSize
	numX := g.world.Obstacles.NCols()
	numY := g.world.Obstacles.NRows()
	blockWidth := sz.X.DivBy(numX)
	blockHeight := sz.Y.DivBy(numY)
	x := pos.X.Times(blockWidth).Plus(blockWidth.DivBy(TWO))
	y := pos.Y.Times(blockHeight).Plus(blockHeight.DivBy(TWO))
	return Pt{x, y}
}

func (g *Gui) Draw(screen *ebiten.Image) {
	// Draw background.
	screen.Fill(color.RGBA{0, 0, 0, 0})

	// Draw ground and trees.
	rows := g.world.Obstacles.NRows()
	cols := g.world.Obstacles.NCols()
	for y := ZERO; y.Lt(rows); y.Inc() {
		for x := ZERO; x.Lt(cols); x.Inc() {
			if g.world.Obstacles.Get(y, x).Eq(ZERO) {
				g.DrawTile(screen, g.imgGround, Pt{x, y})
			} else {
				g.DrawTile(screen, g.imgTree, Pt{x, y})
			}
		}
	}

	// Draw player.
	g.DrawTile(screen, g.imgPlayer, g.world.Player.Pos)

	// Draw enemy.
	g.DrawTile(screen, g.imgEnemyShadow, g.world.Enemy.FuturePos)
	g.DrawTile(screen, g.imgEnemy, g.world.Enemy.Pos)

	// Draw beam.
	if g.beamIdx.Gt(ZERO) {
		DrawLine(screen, g.beam, intToCol(4))
		g.beamIdx.Dec()
	}

	// Output TPS (ticks per second, which is like frames per second).
	//ebitenutil.DebugPrint(screen, fmt.Sprintf("ActualTPS: %f", ebiten.ActualTPS()))
}

func (g *Gui) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	g.screenSize = IPt(outsideWidth, outsideHeight)
	return outsideWidth, outsideHeight
}

func intToCol(ival int64) color.Color {
	switch ival {
	case 0:
		return color.RGBA{25, 25, 25, 255}
	case 1:
		return color.RGBA{150, 0, 0, 255}
	case 2:
		return color.RGBA{0, 150, 0, 255}
	case 3:
		return color.RGBA{0, 0, 150, 255}
	case 4:
		return color.RGBA{150, 150, 0, 255}
	case 5:
		return color.RGBA{0, 150, 150, 255}
	case 6:
		return color.RGBA{150, 0, 150, 255}
	case 7:
		return color.RGBA{100, 150, 100, 255}
	}
	return color.Black
}

//func (g *Game) DrawFilledSquare(screen *ebiten.Image, s Square, col color.Color) {
//	size := WorldToScreen(s.Size)
//	x := WorldToScreen(s.Center.X) - size/2
//	y := WorldToScreen(s.Center.Y) - size/2
//
//	if g.img == nil {
//		g.img = ebiten.NewImage(int(size), int(size))
//	}
//	g.img.Fill(col)
//	op := &ebiten.DrawImageOptions{}
//	op.GeoM.Translate(x, y)
//	screen.DrawImage(g.img, op)
//}
//
//func DrawSquare(screen *ebiten.Image, s Square, color color.Color) {
//	halfSize := s.Size.DivBy(I(2)).Plus(s.Size.Mod(I(2)))
//
//	// square corners
//	upperLeftCorner := Pt{s.Center.X.Minus(halfSize), s.Center.Y.Minus(halfSize)}
//	lowerLeftCorner := Pt{s.Center.X.Minus(halfSize), s.Center.Y.Plus(halfSize)}
//	upperRightCorner := Pt{s.Center.X.Plus(halfSize), s.Center.Y.Minus(halfSize)}
//	lowerRightCorner := Pt{s.Center.X.Plus(halfSize), s.Center.Y.Plus(halfSize)}
//
//	DrawLine(screen, Line{upperLeftCorner, upperRightCorner}, color)
//	DrawLine(screen, Line{upperLeftCorner, lowerLeftCorner}, color)
//	DrawLine(screen, Line{lowerLeftCorner, lowerRightCorner}, color)
//	DrawLine(screen, Line{lowerRightCorner, upperRightCorner}, color)
//}
//
//func (g *Game) DrawMatrix(screen *ebiten.Image, m Matrix, squareSize Int) {
//	for y := I(0); y.Lt(m.NRows()); y.Inc() {
//		for x := I(0); x.Lt(m.NCols()); x.Inc() {
//			var s Square
//			s.Center.X = x.Times(squareSize).Plus(squareSize.DivBy(I(2)))
//			s.Center.Y = y.Times(squareSize).Plus(squareSize.DivBy(I(2)))
//			s.Size = squareSize
//
//			var col color.Color
//			col = color.RGBA{160, 160, 160, 0}
//			DrawSquare(screen, s, col)
//
//			mVal := m.Get(y, x).ToInt64()
//			g.DrawFilledSquare(screen, s, intToCol(mVal))
//		}
//	}
//}

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

	op.GeoM.Translate(x, y)
	screen.DrawImage(img, op)
}

func Level1() string {
	//	return `
	//xxxxxxxxxxxxxxx
	//x           x x
	//x xx  1  x xx x
	//x x    xxx x  x
	//xxxx          x
	//x  x xxxx     x
	//x         x   x
	//x xxx   xxx   x
	//x     x   x   x
	//x   xxx   xxx x
	//x     x       x
	//x     xxxx    x
	//x xx   x2  xx x
	//x  x          x
	//xxxxxxxxxxxxxxx
	//`
	//	return `
	//xxxxxxxxxxxxxxx
	//x             x
	//x     1       x
	//x        x    x
	//x        x    x
	//x   xxxxxx    x
	//x             x
	//x      xx     x
	//x       x     x
	//x       x     x
	//x       xx    x
	//x xxx         x
	//x   x   2     x
	//x             x
	//xxxxxxxxxxxxxxx
	//`
	//	return `
	//
	//     1
	//        x
	//        x
	//   xxxxxx
	//
	//      xx
	//       x
	//      2x
	//       xx
	//`
	return `
   1  
 xxxxx
      
    xx
     x
    2x
`
}

func LevelFromString(level string) (m Matrix, pos1 []Pt, pos2 []Pt) {
	row := -1
	col := 0
	maxCol := 0
	for i := 0; i < len(level); i++ {
		c := level[i]
		if c == '\n' {
			maxCol = col
			col = 0
			row++
			continue
		}
		col++
	}
	// If the string does not end with an empty line, count the last row.
	if col > 0 {
		row++
	}
	m.Init(I(row), I(maxCol))

	row = -1
	col = 0
	for i := 0; i < len(level); i++ {
		c := level[i]
		if c == '\n' {
			col = 0
			row++
			continue
		} else if c == 'x' {
			m.Set(I(row), I(col), I(1))
		} else if c == '1' {
			pos1 = append(pos1, IPt(col, row))
		} else if c == '2' {
			pos2 = append(pos2, IPt(col, row))
		}
		col++
	}
	return
}

func main() {
	ebiten.SetWindowSize(800, 800)
	ebiten.SetWindowTitle("Miln")
	ebiten.SetWindowPosition(700, 100)

	var g Gui
	g.imgGround = ebiten.NewImage(20, 20)
	g.imgGround.Fill(intToCol(0))
	g.imgTree = ebiten.NewImage(20, 20)
	g.imgTree.Fill(intToCol(1))
	g.imgPlayer = ebiten.NewImage(20, 20)
	g.imgPlayer.Fill(intToCol(2))
	g.imgEnemy = ebiten.NewImage(20, 20)
	g.imgEnemy.Fill(intToCol(3))
	g.imgEnemyShadow = ebiten.NewImage(20, 20)
	c := color.RGBA{0, 0, 150, 30}
	g.imgEnemyShadow.Fill(c)

	var err error
	// Load the Arial font
	fontData, err := opentype.Parse(goregular.TTF)
	Check(err)

	g.defaultFont, err = opentype.NewFace(fontData, &opentype.FaceOptions{
		Size:    24,
		DPI:     72,
		Hinting: font.HintingVertical,
	})
	Check(err)

	//g.world.Obstacles.Init(I(15), I(15))
	pos1 := []Pt{}
	pos2 := []Pt{}
	g.world.Obstacles, pos1, pos2 = LevelFromString(Level1())
	if len(pos1) > 0 {
		g.world.Player.Pos = pos1[0]
	}
	if len(pos2) > 0 {
		g.world.Enemy.Pos = pos2[0]
	}
	g.pathfinding.Initialize(g.world.Obstacles)

	// Start the game.
	err = ebiten.RunGame(&g)
	Check(err)
}
