package gamelib

import (
	"github.com/hajimehoshi/ebiten/v2"
	_ "image/png"
	"io/fs"
	"slices"
	"strconv"
)

var AnimationFps = I(10)

// Animation represents an instance of a running animation.
// It is cheap to copy this struct. You should make copies for every
// instance of an animation that you need.
// The idea is that once the images are loaded, there's no need to change
// this data. So you can just copy around the references to the images.
type Animation struct {
	imgs     []*ebiten.Image
	imgIndex Int
	frameIdx Int
}

func NewAnimation(fsys fs.FS, name string) (a Animation) {
	count := 1
	for {
		fullName := name + strconv.Itoa(count) + ".png"
		if !FileExists(fsys, fullName) {
			break
		}

		img := LoadImage(fsys, fullName)
		a.imgs = append(a.imgs, img)
		count++
	}

	// If no files exist following the format "player1.png", "player2.png" ..
	// try just loading "player.png".
	if count == 1 {
		fullName := name + ".png"
		img := LoadImage(fsys, fullName)
		a.imgs = append(a.imgs, img)
	}
	a.imgIndex = ZERO
	return
}

func (a *Animation) Step() {
	a.frameIdx.Inc()
	if a.frameIdx.Eq(I(60)) {
		a.frameIdx = ZERO
	}

	framesPerImage := I(60).DivBy(AnimationFps)
	if a.frameIdx.Mod(framesPerImage).Eq(ZERO) {
		a.imgIndex.Inc()
	}
	if a.imgIndex.Eq(I(len(a.imgs))) {
		a.imgIndex = ZERO
	}
}

func (a *Animation) Img() *ebiten.Image {
	return a.imgs[a.imgIndex.ToInt()]
}

func (a *Animation) Valid() bool {
	return len(a.imgs) > 0
}

// Same returns true if this animation contains the same images as "other".
func (a *Animation) Same(other Animation) bool {
	return slices.Equal(a.imgs, other.imgs)
}
