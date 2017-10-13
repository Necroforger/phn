package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/anthonynsimon/bild/blur"

	"github.com/Necroforger/phn"

	"github.com/nfnt/resize"
)

// Flags
var (
	HelpFlag         = flag.Bool("help", false, "Display help")
	Width            = flag.Uint("w", 0, "Width to resize both images to")
	Height           = flag.Uint("h", 0, "Height to resize both images to")
	OutputPath       = flag.String("o", "", "Output file name")
	Depth            = flag.Uint("d", 10, "Colour depth")
	ModeDecode       = flag.Bool("decode", false, "[encoded] [original] Decode the hidden image.")
	ModeEstimate     = flag.Bool("estimate", false, "[encoded] Decode with estimate mode")
	BackgroundColour = flag.String("c", "", "Background colour of solid image. Gray by default")
)

func getImage(path string) (image.Image, error) {
	var (
		source io.ReadCloser
		err    error
	)
	if strings.HasPrefix(path, "http://") ||
		strings.HasPrefix(path, "https://") {
		resp, err := http.Get(path)
		if err != nil {
			return nil, err
		}
		source = resp.Body
	} else {
		source, err = os.Open(path)
		if err != nil {
			return nil, err
		}
	}
	defer source.Close()

	img, _, err := image.Decode(source)
	return img, err
}

func createUniformImage(clr color.Color, b image.Rectangle) *image.RGBA {
	out := image.NewRGBA(b)
	draw.Draw(out, b, image.NewUniform(clr), image.ZP, draw.Src)
	return out
}

func handle(err error) {
	if err != nil {
		panic(err)
	}
}

func cloneAsRGBA(src image.Image) *image.RGBA {
	bounds := src.Bounds()
	img := image.NewRGBA(bounds)
	draw.Draw(img, bounds, src, bounds.Min, draw.Src)
	return img
}

// func xorCipherImage(key []byte, img image.Image) *image.RGBA {
// 	src := cloneAsRGBA(img)
// 	b := src.Bounds()
// 	dst := image.NewRGBA(b)
// 	width, height := b.Dx(), b.Dy()

// 	parallel(height, func(start, end int) {
// 		for y := start; y < end; y++ {
// 			for x := 0; x < width; x++ {
// 				idx := y*src.Stride + x*4
// 				dst.Pix[idx+0] = src.Pix[idx+0] ^ key[(idx+0)%len(key)]
// 				dst.Pix[idx+1] = src.Pix[idx+1] ^ key[(idx+1)%len(key)]
// 				dst.Pix[idx+2] = src.Pix[idx+2] ^ key[(idx+2)%len(key)]
// 				dst.Pix[idx+3] = src.Pix[idx+3]
// 			}
// 		}
// 	})

// 	return dst
// }

// func parallel(height int, fn func(start, end int)) {
// 	procs := runtime.GOMAXPROCS(0)
// 	partSize := height / procs
// 	if procs <= 1 || partSize < procs {
// 		fn(0, height)
// 		return
// 	}
// 	var wg sync.WaitGroup
// 	for i := height; i > 0; i -= partSize {
// 		start := i - partSize
// 		end := i
// 		if start < 0 {
// 			start = 0
// 		}
// 		wg.Add(1)
// 		go func() {
// 			fn(start, end)
// 			wg.Done()
// 		}()
// 	}
// 	wg.Wait()
// }

func hexToRGB(h string) (r uint8, g uint8, b uint8) {
	r, g, b = 127, 127, 127
	data, err := hex.DecodeString(h)
	if err != nil {
		return
	}

	if len(data) > 0 {
		r = data[0]
	}
	if len(data) > 1 {
		g = data[1]
	}
	if len(data) > 2 {
		b = data[2]
	}

	return
}

func main() {
	var (
		img1, img2 image.Image
		dst        image.Image
		err        error
	)

	flag.Parse()

	if *HelpFlag {
		fmt.Print("usage: phn [hidden image] [source image]\n" +
			"if the second result is left blank, a grey image will be generated\n" +
			"with the same bounds as the first image\n")
		flag.PrintDefaults()
		return
	}

	if len(flag.Args()) == 0 {
		log.Println("Please enter at least one image image path")
		return
	}

	// Decode images
	// If no image path is provided use a uniformly coloured background.
	img1, err = getImage(flag.Arg(0))
	handle(err)

	if len(flag.Args()) > 1 {
		img2, err = getImage(flag.Arg(1))
		handle(err)
	} else {
		r, g, b := hexToRGB(*BackgroundColour)
		img2 = createUniformImage(color.RGBA{R: r, G: g, B: b, A: 255}, img1.Bounds())
	}

	// Set output destination
	if *OutputPath == "" {
		*OutputPath = "output.png"
	}

	out, err := os.Create(*OutputPath)
	handle(err)
	defer out.Close()

	if *Width > 0 || *Height > 0 {
		img2 = resize.Resize(*Width, *Height, img2, resize.NearestNeighbor)
		img1 = resize.Resize(*Width, *Height, img1, resize.NearestNeighbor)
	}

	switch {
	case *ModeDecode:
		dst = phn.Decode(img2, img1, uint8(*Depth))
	case *ModeEstimate:
		dst = phn.Decode(img1, blur.Gaussian(img1, 10), uint8(*Depth))
	default:
		dst = phn.Encode(img2, img1, uint8(*Depth))
	}

	handle(png.Encode(out, dst))
}
