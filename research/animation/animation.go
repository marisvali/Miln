package main

import (
	"errors"
	"github.com/hajimehoshi/ebiten/v2"
	x "github.com/marisvali/miln/gamelib"
	_ "image/png"
	"os"
	"strconv"
)

var AnimationFps = x.I(10)

type Animation struct {
	imgs     []*ebiten.Image
	imgIndex x.Int
	frameIdx x.Int
}

func NewAnimation(name string) (a Animation) {
	count := 1
	for {
		fullName := name + strconv.Itoa(count) + ".png"
		if _, err := os.Stat(fullName); errors.Is(err, os.ErrNotExist) {
			break
		}

		img := x.LoadImage(fullName)
		a.imgs = append(a.imgs, img)
		count++
	}
	a.imgIndex = x.I(len(a.imgs))
	return
}

func (a *Animation) Img() *ebiten.Image {
	a.frameIdx.Inc()
	if a.frameIdx.Eq(x.I(60)) {
		a.frameIdx = x.ZERO
	}

	framesPerImage := x.I(60).DivBy(AnimationFps)
	if a.frameIdx.Mod(framesPerImage).Eq(x.ZERO) {
		a.imgIndex.Inc()
	}
	if a.imgIndex.Eq(x.I(len(a.imgs))) {
		a.imgIndex = x.ZERO
	}
	return a.imgs[a.imgIndex.ToInt()]
}
