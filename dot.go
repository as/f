package frame

import (
	"bytes"
	"image"
	"golang.org/x/image/font"
)

type Font struct{
	font.Face
	height int
}

func NewFont(face font.Face) *Font{
	if face == nil{
		panic("NewFont: nil font face")
	}
	return &Font{
		Face: face,
	}
}

func (f *Font) Height() int{
	if f.Face == nil{
		return 0
	}
	if f.height == 0{
		f.height = int(f.Metrics().Height>>6)+1
	}
	return f.height
}

type Dot struct {
	image.Point
	origin     image.Point
	maxw       int
	font *Font
}

func NewDot(origin image.Point, maxw int, font *Font) *Dot {
	if font == nil{
		panic("NewDot: font is nil")
	}
	return &Dot{
		Point:      origin,
		origin:     origin,
		maxw:       maxw,
		font: font,
	}
}

func (d *Dot) advance(r rune) int{
	dx, _ := d.font.GlyphAdvance(r)
	return int(dx >> 6)
}

func (d *Dot) Advance(r rune) int{
	if r == '\t' {
	 	return d.advance(' ')*4
	}
	return d.advance(r)
}

func (d *Dot) Visible(r rune) bool{
	switch r{
	case '\t', '\n':
		return false
	}
	return true
}

func (d *Dot) Newline() {
	d.X = d.origin.X
	d.Y += d.font.Height()
}

// fits returns the number of pixels that would be advance
// if r were printed, or -1 if r doesn't fit on the line
func (d *Dot) fits(r rune) int{
	adv := d.Advance(r)
	if d.Width() + adv > d.maxw{
		return -1
	}
	return adv
}

// Insert advances dot by the width of r, or starts a new
// line if r doesn't fit
func (d *Dot) Insert(r rune) image.Point{
	if adv := d.fits(r); adv == -1 || r == rune('\n') {
		d.Newline()
	} else {
		d.X += adv
	}
	return d.Point
}

// Width returns the amount of horizontal pixels covered by dot
// starting from the origin
func (d *Dot) Width() int {
	return d.X - d.origin.X
}


func nlpos(p []byte) (i int) {
	return bytes.Index(p, NL)
}
