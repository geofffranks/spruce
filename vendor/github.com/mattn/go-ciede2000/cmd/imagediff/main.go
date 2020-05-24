package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"

	"github.com/mattn/go-ciede2000"
)

func pixelization(img image.Image, pixelsize int) image.Image {
	dest := image.NewRGBA(img.Bounds())
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			c := img.At(int(x/10)*10, int(y/10)*10)
			dest.Set(x, y, c)
		}
	}
	return dest
}

func loadImage(filename string) (image.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("failed to load image %q: %v", filename, err)
	}
	return img, nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage of %s: [image1] [image2]\n", os.Args[0])
		os.Exit(1)
	}
	img1, err := loadImage(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	img2, err := loadImage(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}

	bounds := img1.Bounds()
	if !bounds.Eq(img2.Bounds()) {
		log.Fatal("image1 and image2 should be same bounds")
	}

	pixelsize := 1
	diff := 0.0
	for y := bounds.Min.Y; y < bounds.Max.Y; y += pixelsize {
		for x := bounds.Min.X; x < bounds.Max.X; x += pixelsize {
			diff += ciede2000.Diff(img1.At(x, y), img2.At(x, y))
		}
	}
	diff /= float64(bounds.Dx() * bounds.Dy())
	fmt.Println(diff)
}
