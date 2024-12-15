package gamelib

import (
	"errors"
	"github.com/hajimehoshi/ebiten/v2"
	_ "image/png"
	"os"
	"strconv"
)

var AnimationFps = I(10)

type Animation struct {
	imgs     []*ebiten.Image
	imgIndex Int
	frameIdx Int
}

func NewAnimation(name string) (a Animation) {
	count := 1
	for {
		fullName := name + strconv.Itoa(count) + ".png"
		if _, err := os.Stat(fullName); errors.Is(err, os.ErrNotExist) {
			break
		}

		img := LoadImage(fullName)
		a.imgs = append(a.imgs, img)
		count++
	}

	// If no files exist following the format "player1.png", "player2.png" ..
	// try just loading "player.png".
	if count == 1 {
		fullName := name + ".png"
		img := LoadImage(fullName)
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
