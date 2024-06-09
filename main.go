package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"image"
	"image/color"
	_ "image/png"
	"math"
	"os"
	. "playful-patterns.com/miln/geometry"
	. "playful-patterns.com/miln/ints"
	pathfinding2 "playful-patterns.com/miln/pathfinding"
	"playful-patterns.com/miln/point"
	"playful-patterns.com/miln/utils"
	. "playful-patterns.com/miln/world"
)

var EnemyCooldown Int = I(40)
var PlayerCooldown Int = I(15)
var BlockSize Int = I(80)

func RandomLevel1() (m utils.Matrix, pos1 []point.Pt, pos2 []point.Pt) {
	m.Init(I(10), I(10))
	for i := 0; i < 10; i++ {
		m.Set(RInt(ZERO, m.NRows().Minus(ONE)), RInt(ZERO, m.NCols().Minus(ONE)), ONE)
	}
	pos1 = append(pos1, point.IPt(0, 0))
	pos2 = append(pos2, point.IPt(2, 2))
	return
}

type Gui struct {
	defaultFont     font.Face
	imgGround       *ebiten.Image
	imgTree         *ebiten.Image
	imgPlayer       *ebiten.Image
	imgEnemy        *ebiten.Image
	imgBeam         *ebiten.Image
	imgShadow       *ebiten.Image
	world           World
	frameIdx        Int
	pathfinding     pathfinding2.Pathfinding
	screenSize      point.Pt
	leftClick       bool
	rightClick      bool
	leftClickPos    point.Pt
	rightClickPos   point.Pt
	beamIdx         Int
	beamMax         Int
	beamHitsEnemy   bool
	beamEnd         point.Pt
	folderWatcher   utils.FolderWatcher
	attackableTiles []point.Pt
}

func (g *Gui) Update() error {
	if g.folderWatcher.FolderContentsChanged() {
		g.loadGuiData()
	}

	if g.world.Player.TimeoutIdx.Gt(ZERO) {
		g.world.Player.TimeoutIdx.Dec()
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) && g.world.Player.TimeoutIdx.Eq(ZERO) {
		x, y := ebiten.CursorPosition()

		//enemyPos := g.TileToScreen(g.world.Enemy.Pos)
		//dist := enemyPos.Minus(IPt(x, y)).Len()
		//if dist.Geq(I(100)) {
		{
			sz := g.screenSize
			// The adjustments below are necessary because we can get x or y
			// larger than the screen size when the user clicks around the
			// bottom right corner of the window.
			if x >= sz.X.ToInt() {
				x = sz.X.ToInt() - 1
			}
			if y >= sz.Y.ToInt() {
				y = sz.Y.ToInt() - 1
			}

			numX := g.world.Obstacles.NCols().ToInt()
			numY := g.world.Obstacles.NRows().ToInt()
			blockWidth := sz.X.ToFloat64() / float64(numX)
			blockHeight := sz.Y.ToFloat64() / float64(numY)

			// Translate from screen coordinates to grid coordinates.
			newPos := point.IPt(
				int(float64(x)/blockWidth),
				int(float64(y)/blockHeight))
			if g.world.Obstacles.Get(newPos.Y, newPos.X).Eq(ZERO) {
				g.world.Player.Pos = newPos
				g.world.Player.TimeoutIdx = PlayerCooldown
			}
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton2) && g.world.Player.TimeoutIdx.Eq(ZERO) {
		x, y := ebiten.CursorPosition()

		enemyPos := g.TileToScreen(g.world.Enemy.Pos)
		dist := enemyPos.Minus(point.IPt(x, y)).Len()
		if dist.Lt(I(200)) {
			// Hit enemy.
			g.beamIdx = g.beamMax
			l := Line{g.TileToScreen(g.world.Player.Pos), g.TileToScreen(g.world.Enemy.Pos)}
			if intersects, pt := g.LineObstaclesIntersection(l); intersects {
				g.beamHitsEnemy = false
				g.beamEnd = pt
			} else {
				g.beamHitsEnemy = true
				g.world.Player.TimeoutIdx = PlayerCooldown
			}
		}
	}

	g.frameIdx.Inc()
	//if g.frameIdx.Mod(I(5)).Neq(ZERO) {
	//	return nil // skip update
	//}

	g.world.TimeStep.Inc()
	if g.world.TimeStep.Eq(I(math.MaxInt64)) {
		// Damn.
		utils.Check(fmt.Errorf("got to an unusually large time step: %d", g.world.TimeStep.ToInt64()))
	}

	// Get keyboard input.
	var pressedKeys []ebiten.Key
	pressedKeys = inpututil.AppendPressedKeys(pressedKeys)

	// Move the enemy.
	if g.world.TimeStep.Mod(EnemyCooldown).Eq(ZERO) {
		path := g.pathfinding.FindPath(g.world.Enemy.Pos, g.world.Player.Pos)
		if len(path) > 1 {
			g.world.Enemy.Pos = path[1]
		}
	}

	// Compute which tiles are attackableTiles.
	g.attackableTiles = []point.Pt{}
	rows := g.world.Obstacles.NRows()
	cols := g.world.Obstacles.NCols()
	for y := ZERO; y.Lt(rows); y.Inc() {
		for x := ZERO; x.Lt(cols); x.Inc() {
			// Check if tile can be attacked.
			pt := point.Pt{x, y}
			screenPt := g.TileToScreen(pt)
			l := Line{g.TileToScreen(g.world.Player.Pos), screenPt}
			if intersects, _ := g.LineObstaclesIntersection(l); !intersects {
				g.attackableTiles = append(g.attackableTiles, pt)
			}
		}
	}

	return nil
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

func DrawPixel(screen *ebiten.Image, pt point.Pt, color color.Color) {
	size := I(2)
	for ax := pt.X.Minus(size); ax.Leq(pt.X.Plus(size)); ax.Inc() {
		for ay := pt.Y.Minus(size); ay.Leq(pt.Y.Plus(size)); ay.Inc() {
			screen.Set(ax.ToInt(), ay.ToInt(), color)
		}
	}
}

func EqualFloats(f1, f2 float64) bool {
	return math.Abs(f1-f2) < 0.000001
}

func (g *Gui) LineObstaclesIntersection(l Line) (bool, point.Pt) {
	sz := g.screenSize
	numX := g.world.Obstacles.NCols().ToInt()
	numY := g.world.Obstacles.NRows().ToInt()
	blockWidth := sz.X.ToFloat64() / float64(numX)
	blockHeight := sz.Y.ToFloat64() / float64(numY)

	rows := g.world.Obstacles.NRows()
	cols := g.world.Obstacles.NCols()
	ipts := []point.Pt{}
	for y := ZERO; y.Lt(rows); y.Inc() {
		for x := ZERO; x.Lt(cols); x.Inc() {
			if !g.world.Obstacles.Get(y, x).IsZero() {
				if !EqualFloats(blockWidth, blockHeight) {
					panic(fmt.Errorf("blocks are not squares"))
				}
				s := Square{g.TileToScreen(point.Pt{x, y}), I(int(blockWidth * 0.9))}
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
			DrawPixel(screen, point.Pt{x, y}, color)
		}
	} else {
		inc := dy.DivBy(dy.Abs())
		for y := y1; y.Neq(y2); y.Add(inc) {
			x := x1.Plus(y.Minus(y1).Times(dx).DivBy(dy))
			DrawPixel(screen, point.Pt{x, y}, color)
		}
	}
}

func (g *Gui) DrawTile(screen *ebiten.Image, img *ebiten.Image, pos point.Pt) {
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

func (g *Gui) TileToScreen(pos point.Pt) point.Pt {
	sz := g.screenSize
	numX := g.world.Obstacles.NCols()
	numY := g.world.Obstacles.NRows()
	blockWidth := sz.X.DivBy(numX)
	blockHeight := sz.Y.DivBy(numY)
	x := pos.X.Times(blockWidth).Plus(blockWidth.DivBy(TWO))
	y := pos.Y.Times(blockHeight).Plus(blockHeight.DivBy(TWO))
	return point.Pt{x, y}
}

func (g *Gui) Draw(screen *ebiten.Image) {
	// Draw background.
	screen.Fill(color.RGBA{0, 0, 0, 0})

	// Draw ground and trees.
	rows := g.world.Obstacles.NRows()
	cols := g.world.Obstacles.NCols()
	for y := ZERO; y.Lt(rows); y.Inc() {
		for x := ZERO; x.Lt(cols); x.Inc() {
			g.DrawTile(screen, g.imgGround, point.Pt{x, y})
			if g.world.Obstacles.Get(y, x).Eq(ONE) {
				g.DrawTile(screen, g.imgTree, point.Pt{x, y})
			}
		}
	}

	// Draw player.
	mask := ebiten.NewImageFromImage(g.imgPlayer)
	{
		percent := g.world.Player.TimeoutIdx.Times(I(100)).DivBy(PlayerCooldown)
		var alpha Int
		if percent.Gt(ZERO) {
			alpha = (percent.Plus(I(100))).Times(I(255)).DivBy(I(200))
		} else {
			alpha = ZERO
		}

		sz := mask.Bounds().Size()
		for y := 0; y < sz.Y; y++ {
			for x := 0; x < sz.X; x++ {
				_, _, _, a := mask.At(x, y).RGBA()
				if a > 0 {
					mask.Set(x, y, color.RGBA{0, 0, 0, uint8(alpha.ToInt())})
				}
			}
		}

		totalWidth := I(mask.Bounds().Size().X)
		lineWidth := g.world.Player.TimeoutIdx.Times(totalWidth).DivBy(PlayerCooldown)
		l := Line{point.IPt(0, 0), point.Pt{lineWidth, I(0)}}
		DrawLine(mask, l, color.RGBA{0, 0, 0, 255})
	}
	g.DrawTile(screen, g.imgPlayer, g.world.Player.Pos)
	g.DrawTile(screen, mask, g.world.Player.Pos)
	//g.DrawTile(screen, g.imgPlayer, g.world.Player.Pos)

	// Draw enemy.
	g.DrawTile(screen, g.imgEnemy, g.world.Enemy.Pos)

	// Draw beam.
	beamScreen := ebiten.NewImage(screen.Bounds().Dx(), screen.Bounds().Dy())
	if g.beamIdx.Gt(ZERO) {
		var beam Line
		if g.beamHitsEnemy {
			beam = Line{g.TileToScreen(g.world.Player.Pos), g.TileToScreen(g.world.Enemy.Pos)}
		} else {
			beam = Line{g.TileToScreen(g.world.Player.Pos), g.beamEnd}
		}

		alpha := uint8(g.beamIdx.Times(I(255)).DivBy(g.beamMax).ToInt())
		//alpha = uint8(0)
		colr, colg, colb, _ := g.imgBeam.At(0, 0).RGBA()
		beamCol := color.RGBA{uint8(colr), uint8(colg), uint8(colb), alpha}
		DrawLine(beamScreen, beam, beamCol)
		g.beamIdx.Dec()
	}
	DrawSprite(screen, beamScreen, 0, 0, float64(beamScreen.Bounds().Dx()), float64(beamScreen.Bounds().Dy()))

	// Mark attackable tiles.
	for _, pt := range g.attackableTiles {
		g.DrawTile(screen, g.imgShadow, pt)
	}
	// Output TPS (ticks per second, which is like frames per second).
	//ebitenutil.DebugPrint(screen, fmt.Sprintf("ActualTPS: %f", ebiten.ActualTPS()))
}

func (g *Gui) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	g.screenSize = point.IPt(outsideWidth, outsideHeight)
	return outsideWidth, outsideHeight
}

func intToCol(ival int64) color.Color {
	switch ival {
	case 0:
		return color.RGBA{25, 25, 25, 255}
	case 1:
		return color.RGBA{250, 0, 0, 255}
	case 2:
		return color.RGBA{0, 250, 0, 255}
	case 3:
		return color.RGBA{0, 0, 150, 255}
	case 4:
		return color.RGBA{250, 250, 0, 255}
	case 5:
		return color.RGBA{0, 250, 250, 255}
	case 6:
		return color.RGBA{250, 0, 250, 255}
	case 7:
		return color.RGBA{200, 250, 200, 255}
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

func LevelFromString(level string) (m utils.Matrix, pos1 []point.Pt, pos2 []point.Pt) {
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
			pos1 = append(pos1, point.IPt(col, row))
		} else if c == '2' {
			pos2 = append(pos2, point.IPt(col, row))
		}
		col++
	}
	return
}

func loadImage(str string) *ebiten.Image {
	file, err := os.Open(str)
	defer file.Close()
	utils.Check(err)

	img, _, err := image.Decode(file)
	utils.Check(err)
	if err != nil {
		return nil
	}

	return ebiten.NewImageFromImage(img)
}

func (g *Gui) loadGuiData() {
	// Read from the disk over and over until a full read is possible.
	// This repetition is meant to avoid crashes due to reading files
	// while they are still being written.
	// It's a hack but possibly a quick and very useful one.
	utils.CheckCrashes = false
	for {
		utils.CheckFailed = nil
		g.imgGround = loadImage("data/ground.png")
		g.imgTree = loadImage("data/tree.png")
		g.imgPlayer = loadImage("data/player.png")
		g.imgEnemy = loadImage("data/enemy.png")
		g.imgBeam = loadImage("data/beam.png")
		g.imgShadow = loadImage("data/shadow.png")
		//g.imgShadow.Fill(color.RGBA{0, 0, 0, 100})
		if utils.CheckFailed == nil {
			break
		}
	}
	utils.CheckCrashes = true
}

func main() {
	var g Gui
	g.beamMax = I(15)

	// Obstacles
	//g.world.Obstacles.Init(I(15), I(15))
	pos1 := []point.Pt{}
	pos2 := []point.Pt{}
	//g.world.Obstacles, pos1, pos2 = LevelFromString(Level1())
	g.world.Obstacles, pos1, pos2 = RandomLevel1()
	if len(pos1) > 0 {
		g.world.Player.Pos = pos1[0]
	}
	if len(pos2) > 0 {
		g.world.Enemy.Pos = pos2[0]
	}
	g.pathfinding.Initialize(g.world.Obstacles)

	// screen size
	g.screenSize.X = BlockSize.Times(g.world.Obstacles.NCols())
	g.screenSize.Y = BlockSize.Times(g.world.Obstacles.NRows())

	ebiten.SetWindowSize(g.screenSize.X.ToInt(), g.screenSize.Y.ToInt())
	ebiten.SetWindowTitle("Miln")
	ebiten.SetWindowPosition(100, 100)

	g.folderWatcher.Folder = "data"
	g.loadGuiData()
	//g.imgEnemy.Fill(intToCol(3))

	// font
	var err error
	// Load the Arial font
	fontData, err := opentype.Parse(goregular.TTF)
	utils.Check(err)

	g.defaultFont, err = opentype.NewFace(fontData, &opentype.FaceOptions{
		Size:    24,
		DPI:     72,
		Hinting: font.HintingVertical,
	})
	utils.Check(err)

	// Start the game.
	err = ebiten.RunGame(&g)
	utils.Check(err)
}
