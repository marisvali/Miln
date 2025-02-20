package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	. "github.com/marisvali/miln/gamelib"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	_ "image/png"
	"log"
	"os"
)

type Game struct {
	imgEnemy  *ebiten.Image
	imgGround *ebiten.Image
	imgShadow *ebiten.Image
	img34     *ebiten.Image
	once      bool
}

func (g *Game) Draw(screen *ebiten.Image) {

	if g.once {
		fmt.Println("erk", erk.At(0, 0))
		col1 := g.imgShadow.At(3, 3)
		col3 := g.imgShadow.RGBA64At(3, 3)
		colorModel := g.imgShadow.ColorModel()
		fmt.Println("colorModel", colorModel)
		fmt.Println("col1 from shadow", col1)
		fmt.Println("col3 from shadow", col3)

		colorModel2 := g.imgGround.ColorModel()
		fmt.Println("colorModel2", colorModel2)

		alt := LoadImage("shadow.png")
		fmt.Println("alt", alt.At(0, 0))

		col2 := g.imgGround.At(3, 3)
		fmt.Println("ground", col2)

		fmt.Println("alt2", g.img34.At(0, 0))
		fmt.Println("alt3", g.img34.At(0, 0))
	}

	screen.Fill(color.RGBA{255, 255, 255, 255})
	// DrawSprite(screen, g.imgGround, 0, 0, 400, 400)
	DrawSprite(screen, g.imgEnemy, 0, 0, 400, 400)
	DrawSprite(screen, g.imgShadow, 0, 0, 400, 400)

	// Draw beam.
	beamScreen := ebiten.NewImage(screen.Bounds().Dx(), screen.Bounds().Dy())
	// beamCol := color.RGBA{uint8(255), uint8(255), uint8(0), 150}
	// beamCol.R = (uint8)(int64(beamCol.R) * int64(beamCol.A) / int64(255)) // pre-multiply alpha
	// beamCol.G = (uint8)(int64(beamCol.G) * int64(beamCol.A) / int64(255)) // pre-multiply alpha
	// beamCol.B = (uint8)(int64(beamCol.B) * int64(beamCol.A) / int64(255)) // pre-multiply alpha

	beamCol := color.NRGBA{uint8(255), uint8(255), uint8(0), 150}
	beam := Line{IPt(30, 30), IPt(310, 310)}
	DrawLine(beamScreen, beam, beamCol)
	// m1, m2 := screen.Bounds().Min, screen.Bounds().Max
	// for ax := m1.X; ax < m2.X; ax++ {
	// 	for ay := m1.Y; ay < m2.Y; ay++ {
	// 		beamScreen.Set(m1.X+ax, m1.Y+ay, beamCol)
	// 	}
	// }

	col1 := g.imgShadow.At(0, 0)
	col2 := beamScreen.At(0, 0)
	DrawSprite(screen, beamScreen, 0, 0, 400, 400)
	if g.once {
		fmt.Println(col1, col2)
		g.once = false
		SaveImage("test1.png", g.imgShadow)
		SaveImage("test2.png", beamScreen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 400, 400
}

func (g *Game) Update() error {
	return nil
}

var erk *ebiten.Image

func loadImage(filename string) *ebiten.Image {
	// Open PNG file
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Decode PNG
	img, err := png.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	// Convert to *image.NRGBA (ensuring non-paletted format)
	nrgba := toNRGBA(img)

	// Convert to premultiplied alpha
	premultiplied := premultiplyAlpha(nrgba)

	// Convert to *ebiten.Image
	return ebiten.NewImageFromImage(premultiplied)
}

// Convert any image.Image to *image.NRGBA
func toNRGBA(src image.Image) *image.NRGBA {
	bounds := src.Bounds()
	nrgba := image.NewNRGBA(bounds)
	draw.Draw(nrgba, bounds, src, bounds.Min, draw.Src)
	return nrgba
}

// Convert an image to premultiplied alpha
func premultiplyAlpha(img *image.NRGBA) *image.NRGBA {
	bounds := img.Bounds()
	out := image.NewNRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			if a > 0 {
				r = (r * a) / 0xffff
				g = (g * a) / 0xffff
				b = (b * a) / 0xffff
			}
			out.Set(x, y, color.NRGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			})
		}
	}
	return out
}

func main() {
	g := Game{}
	g.once = true
	g.imgEnemy = LoadImage("enemy2.png")
	g.imgGround = LoadImage("ground.png")
	g.imgShadow = LoadImage("shadow.png")
	g.img34 = loadImage("shadow.png")

	file, err := os.Open("shadow.png")
	defer func(file *os.File) { Check(file.Close()) }(file)
	Check(err)

	img, _, err := image.Decode(file)
	fmt.Println("img", img.At(0, 0))
	Check(err)
	erk = ebiten.NewImageFromImage(img)

	ebiten.SetWindowSize(400, 400)
	ebiten.SetWindowTitle("Blending")
	if err := ebiten.RunGame(&g); err != nil {
		log.Fatal(err)
	}
}
