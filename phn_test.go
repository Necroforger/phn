package phn_test

import (
	"image"
	_ "image/jpeg"
	"image/png"
	"os"
	"testing"

	"github.com/Necroforger/phn"
	"github.com/anthonynsimon/bild/blur"
	"github.com/nfnt/resize"
)

func loadImage(path string) image.Image {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}
	return img
}

func writeImage(img image.Image, path string) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	err = png.Encode(f, img)
	if err != nil {
		panic(err)
	}
}

const (
	sourceimg  = "img1.png"
	hiddenimg  = "img2.jpg"
	destimg    = "dst.png"
	decodedimg = "decoded.png"
)

func TestEncoder(t *testing.T) {
	src, hid := loadImage(sourceimg), loadImage(hiddenimg)
	hid = resize.Resize(uint(src.Bounds().Dx()), 0, hid, resize.Lanczos3)

	dst := phn.Encode(src, hid, 10)
	writeImage(dst, destimg)
}

func TestDecoder(t *testing.T) {
	dst, src := loadImage(destimg), loadImage(sourceimg)
	src = resize.Resize(uint(dst.Bounds().Dx()), 0, src, resize.Lanczos3)

	dec := phn.Decode(dst, src, 0)
	writeImage(dec, decodedimg)
}

func TestEstimation(t *testing.T) {
	dst := loadImage(destimg)
	src := blur.Gaussian(dst, 20.0)

	dec := phn.Decode(dst, src, 10)
	writeImage(dec, decodedimg)
}

// func TestHueShift(t *testing.T) {
// 	src := loadImage("img1.png")

// 	img := &gif.GIF{}
// 	skip := 10
// 	start := -360
// 	end := 360

// 	img.Image = make([]*image.Paletted, int(float64(end-start)/float64(skip)+0.5))
// 	img.Delay = make([]int, int(float64(end-start)/float64(skip)+0.5))

// 	ncpu := runtime.GOMAXPROCS(0)
// 	tokens := make(chan struct{}, ncpu)

// 	for i := 0; i < ncpu; i++ {
// 		tokens <- struct{}{}
// 	}
// 	for i := start; i < end; i += skip {
// 		<-tokens
// 		go func(n int) {
// 			img.Image[(n+360)/skip] = toPaletted(adjust.Hue(src, n))
// 			img.Delay[(n+360)/skip] = 0
// 			fmt.Println(n, " done")
// 			tokens <- struct{}{}
// 		}(i)
// 	}
// 	for i := 0; i < ncpu; i++ {
// 		<-tokens
// 	}
// 	f, _ := os.Create("animated.gif")
// 	gif.EncodeAll(f, img)
// }

// func toPaletted(src image.Image) *image.Paletted {
// 	p := image.NewPaletted(src.Bounds(), palette.Plan9)
// 	b := p.Bounds()
// 	draw.Draw(p, b, src, b.Min, draw.Src)
// 	return p
// }
