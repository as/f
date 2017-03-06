package frame

import (
	"bytes"
	"golang.org/x/image/font"
	"image"
)

type Font struct {
	font.Face
	height int
}

func NewFont(face font.Face) *Font {
	if face == nil {
		panic("NewFont: nil font face")
	}
	return &Font{
		Face: face,
	}
}

func (f *Font) Height() int {
	if f.Face == nil {
		return 0
	}
	if f.height == 0 {
		f.height = int(f.Metrics().Height>>6) + 1
	}
	return f.height
}

type Dot struct {
	image.Point
	origin image.Point
	maxw   int
	font   *Font
}

func NewDot(origin image.Point, maxw int, font *Font) *Dot {
	if font == nil {
		panic("NewDot: font is nil")
	}
	return &Dot{
		Point:  origin,
		origin: origin,
		maxw:   maxw,
		font:   font,
	}
}

func (d *Dot) advance(r rune) int {
	dx, _ := d.font.GlyphAdvance(r)
	return int(dx >> 6)
}

func (d *Dot) Advance(r rune) int {
	if r == '\t' {
		return d.advance(' ') * 4
	}
	return d.advance(r)
}

func (d *Dot) Visible(r rune) bool {
	switch r {
	case '\t', '\n':
		return false
	}
	return true
}

func (d *Dot) Newline() {
	d.X = d.origin.X
	d.Y += d.Height()
}

func (d *Dot) Origin() image.Point {
	return d.origin
}

// fits returns the number of pixels that would be advance
// if r were printed, or -1 if r doesn't fit on the line
func (d *Dot) fits(r rune) int {
	adv := d.Advance(r)
	if d.Width()+adv > d.maxw {
		return -1
	}
	return adv
}

func (d *Dot) fitsbox(b *Box) int {
	adv := b.Width()
	if d.Width()+adv > d.maxw {
		return -1
	}
	return adv
}

// Insert advances dot by the width of r, or starts a new
// line if r doesn't fit
func (d *Dot) Insert(r rune) image.Point {
	if adv := d.fits(r); adv == -1 || r == rune('\n') {
		d.Newline()
	} else {
		d.X += adv
	}
	return d.Point
}

func (d *Dot) InsertBox(b *Box) image.Point {
	if adv := d.fitsbox(b); adv == -1 {
		d.Newline()
	} else {
		d.X += adv
	}
	return d.Point
}

// IndexOf computes the index of the glyph containing pt
func (dot *Dot) indexOf(box *Box, pt image.Point) (i int) {
	//defer func() { fmt.Printf("IndexOf: pt=%v i=%d (%c)\n", pt, i, f.s[i])}()
	pt = dot.alignY(pt)
	s := box.Bytes()
	for i = 0; i < len(s); i++ {
		switch {
		case dot.Y < pt.Y:
			// nothing special
		case dot.Y == pt.Y:
			// same line
			if dot.X+dot.Advance(rune(s[i]))/2 >= pt.X {
				return i
			}
		case dot.Y > pt.Y:
			// advanced too far
			if i-1 < 0 {
				// bug fix for crash: happened when selecting
				// and dragging all the way to the top
				return i
			}
			if s[i-1] == '\n' {
				// a hard newline
				return i - 1
			} else {
				// line wrapped
				return i
			}
		}
		dot.Insert(rune(s[i]))
	}
	return i
}

// Width returns the amount of horizontal pixels covered by dot
// starting from the origin
func (d *Dot) Width() int {
	return d.X - d.origin.X
}

func (d *Dot) Height() int {
	return d.font.Height()
}

func nlpos(p []byte) (i int) {
	return bytes.Index(p, NL)
}

func (d *Dot) alignY(pt image.Point) image.Point {
	return alignY(d.Height(), d.Origin(), pt)
}
