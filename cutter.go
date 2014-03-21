/*
Package cutter provides a function to crop image.

By default, the original image will be cropped at the
given size from the top left corner.

		croppedImg, err := cutter.Crop(img, cutter.Config{
		  Width:  250,
		  Height: 500,
		})

It is possible to specify the top left position:

		croppedImg, err := cutter.Crop(img, cutter.Config{
		  Width:  250,
		  Height: 500,
		  Anchor: image.Point{100, 100},
		  Mode:   TopLeft, // optional, default value
		})

The Anchor property can represents the center of the cropped image
instead of the top left corner:


		croppedImg, err := cutter.Crop(img, cutter.Config{
		  Width: 250,
		  Height: 500,
		  Mode: Centered,
		})

The default crop use the specified dimension, but it is possible
to use Width and Heigth as a ratio instead. In this case,
the resulting image will be as big as possible to fit the asked ratio
from the anchor position.

		croppedImg, err := cutter.Crop(baseImage, cutter.Config{
		  Width: 4,
		  Height: 3,
		  Mode: Centered,
		  Options: Ratio,
		})
*/
package cutter

import (
	"image"
)

// Config	is used to defined
// the way the crop should be realized.
type Config struct {
	Width, Height int
	Anchor        image.Point // The Anchor Point in the source image
	Mode          AnchorMode  // Which point in the resulting image the Anchor Point is referring to
	Options       Option
}

// AnchorMode is an enumeration of the position an anchor can represent.
type AnchorMode int

const (
	// TopLeft defines the Anchor Point
	// as the top left of the cropped picture.
	TopLeft AnchorMode = iota
	// Centered defines the Anchor Point
	// as the center of the cropped picture.
	Centered = iota
)

// Option flags to modify the way the crop is done.
type Option int

const (
	// Ratio flag is use when Width and Height
	// must be used to compute a ratio rather
	// than absolute size in pixels.
	Ratio Option = 1 << iota
)

// Crop retrieves an image that is a
// cropped copy of the original img.
//
// The crop is made given the informations provided in config.
func Crop(img image.Image, c Config) (image.Image, error) {
	maxBounds := c.maxBounds(img.Bounds())
	size := c.computeSize(maxBounds, image.Point{c.Width, c.Height})
	cr := c.computedCropArea(img.Bounds(), size)
	cr = img.Bounds().Intersect(cr)
	result := image.NewRGBA(cr)
	for x, dx := cr.Min.X, cr.Max.X; x < dx; x++ {
		for y, dy := cr.Min.Y, cr.Max.Y; y < dy; y++ {
			result.Set(x, y, result.ColorModel().Convert(img.At(x, y)))
		}
	}
	return result, nil
}

func (c Config) maxBounds(bounds image.Rectangle) (r image.Rectangle) {
	if c.Mode == Centered {
		anchor := c.centeredMin(bounds)
		w := min(anchor.X-bounds.Min.X, bounds.Max.X-anchor.X)
		h := min(anchor.Y-bounds.Min.Y, bounds.Max.Y-anchor.Y)
		r = image.Rect(anchor.X-w, anchor.Y-h, anchor.X+w, anchor.Y+h)
	} else {
		r = image.Rect(c.Anchor.X, c.Anchor.Y, bounds.Max.X, bounds.Max.Y)
	}
	return
}

// computeSize retrieve the effective size of the cropped image.
// It is defined by Height, Width, and Ratio option.
func (c Config) computeSize(bounds image.Rectangle, ratio image.Point) (p image.Point) {
	if c.Options&Ratio == Ratio {
		// Ratio option is on, so we take the biggest size available that fit the given ratio.
		if float64(ratio.X)/float64(bounds.Dx()) > float64(ratio.Y)/float64(bounds.Dy()) {
			p = image.Point{bounds.Dx(), (bounds.Dx() / ratio.X) * ratio.Y}
		} else {
			p = image.Point{(bounds.Dy() / ratio.Y) * ratio.X, bounds.Dy()}
		}
	} else {
		p = image.Point{ratio.X, ratio.Y}
	}
	return
}

// computedCropArea retrieve the theorical crop area.
// It is defined by Height, Width, Mode and
func (c Config) computedCropArea(bounds image.Rectangle, size image.Point) (r image.Rectangle) {
	min := bounds.Min
	switch c.Mode {
	case Centered:
		rMin := c.centeredMin(bounds)
		r = image.Rect(rMin.X-size.X/2, rMin.Y-size.Y/2, rMin.X+size.X/2, rMin.Y+size.Y/2)
	default: // TopLeft
		rMin := image.Point{min.X + c.Anchor.X, min.Y + c.Anchor.Y}
		r = image.Rect(rMin.X, rMin.Y, rMin.X+size.X, rMin.Y+size.Y)
	}
	return
}

func (c *Config) centeredMin(bounds image.Rectangle) (rMin image.Point) {
	min := bounds.Min
	if c.Anchor.X == 0 && c.Anchor.Y == 0 {
		rMin = image.Point{
			X: min.X + bounds.Dx()/2,
			Y: min.Y + bounds.Dy()/2,
		}
	} else {
		rMin = image.Point{
			X: min.X + c.Anchor.X,
			Y: min.Y + c.Anchor.Y,
		}
	}
	return
}

func min(a, b int) (r int) {
	if a < b {
		r = a
	} else {
		r = b
	}
	return
}
