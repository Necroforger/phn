// Package phn is used to hide an image inside another image.
// The hidden image can be decoded by retrieving the absolute value of
// the difference of each pixel. The top-left pixel of the image is
// used to store the colour depth of the encoded image.
package phn

import (
	"image"
	"image/draw"
	"math/rand"
	"runtime"
	"sync"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

// Encode generates a hidden image from a supplied source image
// and a given hidden image.
//     source: source image
//     hidden: image to hide inside source.
//     depth : colour depth of the hidden image.
func Encode(source image.Image, hidden image.Image, depth uint8) *image.RGBA {
	src := cloneAsRGBA(source)
	sbounds := src.Bounds()
	srcW, srcH := sbounds.Dx(), sbounds.Dy()

	if sbounds.Empty() {
		return &image.RGBA{}
	}

	hid := cloneWithBounds(hidden, sbounds)
	dst := image.NewRGBA(sbounds)

	level := func(n, low, high uint8) uint8 {
		return uint8((float64(n)/255.0)*float64(high-low)) + low
	}

	// calc returns the destination pixel value.
	//     hp : hidden pixel
	//     sp : visible pixel
	calc := func(sp, hp uint8) uint8 {
		hp = level(hp, 0, depth)
		if 255-hp < sp {
			return sp - hp
		}
		if hp > sp {
			return sp + hp
		}
		if rand.Float64() >= 0.5 {
			return sp - hp
		}
		return sp + hp
	}

	parallel(srcH, func(start, end int) {
		for y := start; y < end; y++ {
			for x := 0; x < srcW; x++ {
				pos := y*src.Stride + x*4
				dst.Pix[pos+0] = calc(src.Pix[pos+0], hid.Pix[pos+0])
				dst.Pix[pos+1] = calc(src.Pix[pos+1], hid.Pix[pos+1])
				dst.Pix[pos+2] = calc(src.Pix[pos+2], hid.Pix[pos+2])
				dst.Pix[pos+3] = src.Pix[pos+3]
			}
		}
	})

	// Set depth pixel
	pos := (srcH-1)*src.Stride + (srcW-1)*4
	dst.Pix[pos+0] = calc(src.Pix[pos+0], 255)

	return dst
}

// Decode decodes a hidden image from the supplied source image.
// The image is extracted by converting the absolute value of the
// difference of the rgb values of the source image and the hidden image
// from a scale between x/depth to x/255. If the depth is left as 0, it will
// be inferred from the bottom-right pixel of the image, which will set its rgb values
// to the depth.
//     encoded : Encoded image
//     source  : original source image
//     depth   : Colour depth of encoded image. Set as 0 for default.
func Decode(encoded image.Image, source image.Image, depth uint8) *image.RGBA {
	enc := cloneAsRGBA(encoded)
	ebounds := enc.Bounds()
	encW, encH := ebounds.Dx(), ebounds.Dy()

	if ebounds.Empty() {
		return &image.RGBA{}
	}
	src := cloneWithBounds(source, ebounds)
	dst := image.NewRGBA(ebounds)

	abs := func(a int) uint8 {
		if a < 0 {
			return uint8(-a)
		}
		return uint8(a)
	}

	level := func(n, low, high uint8) uint8 {
		return uint8((float64(n)/float64(depth))*float64(high-low)) + low
	}

	if depth == 0 {
		pos := (encH-1)*src.Stride + (encW-1)*4
		depth = abs(int(src.Pix[pos+0]) - int(enc.Pix[pos+0]))
		// Default to 10
		if depth == 0 {
			depth = 10
		}
	}

	parallel(encH, func(start, end int) {
		for y := start; y < end; y++ {
			for x := 0; x < encW; x++ {
				pos := y*enc.Stride + x*4
				dst.Pix[pos+0] = level(abs(int(src.Pix[pos+0])-int(enc.Pix[pos+0])), 0, 255)
				dst.Pix[pos+1] = level(abs(int(src.Pix[pos+1])-int(enc.Pix[pos+1])), 0, 255)
				dst.Pix[pos+2] = level(abs(int(src.Pix[pos+2])-int(enc.Pix[pos+2])), 0, 255)
				dst.Pix[pos+3] = src.Pix[pos+3]
			}
		}
	})

	return dst
}

func cloneAsRGBA(src image.Image) *image.RGBA {
	bounds := src.Bounds()
	img := image.NewRGBA(bounds)
	draw.Draw(img, bounds, src, bounds.Min, draw.Src)
	return img
}

func cloneWithBounds(src image.Image, bounds image.Rectangle) *image.RGBA {
	img := image.NewRGBA(bounds)
	draw.Draw(img, bounds, src, bounds.Min, draw.Src)
	return img
}

func parallel(height int, fn func(start, end int)) {
	procs := runtime.GOMAXPROCS(0)
	partSize := height / procs
	if procs <= 1 || partSize < procs {
		fn(0, height)
		return
	}
	var wg sync.WaitGroup
	for i := height; i > 0; i -= partSize {
		start := i - partSize
		end := i
		if start < 0 {
			start = 0
		}
		wg.Add(1)
		go func() {
			fn(start, end)
			wg.Done()
		}()
	}
	wg.Wait()
}
