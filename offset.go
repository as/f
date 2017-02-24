package frame

import (
	"image"
)

// Origin returns the insertion point of the first
// glyph in the frame
func (f *Frame) Origin() image.Point {
	return f.Bounds().Min.Add(f.origin)
}

// IndexOf computes the index of the glyph containing pt
func (f *Frame) IndexOf(pt image.Point) (i int) {
	//defer func() { fmt.Printf("IndexOf: pt=%v i=%d (%c)\n", pt, i, f.s[i])}()
	pt = f.alignY(pt)
	dot := NewDot(f.Origin(), f.Option.Wrap, f.Font)
	s := f.s[:f.nbytes]
	for i = 0; i < len(s); i++{
		switch{
		case dot.Y < pt.Y:
			// nothing special
		case dot.Y == pt.Y:
			// same line
			if dot.X+dot.Advance(rune(s[i]))/2 >= pt.X{
				return i
			}
		case dot.Y > pt.Y:
			// advanced too far
			if i-1 < 0{
				// bug fix for crash: happened when selecting
				// and dragging all the way to the top
				return i
			}
			if s[i-1] == '\n'{
				// a hard newline
				return i-1
			} else {
				// line wrapped
				return i
			}
		}
		dot.Insert(rune(s[i]))
	}
	return i
}

// PointOf computes the point of origin for glyph i
func (f *Frame) PointOf(i int) (pt image.Point) {
	//	defer func(){fmt.Printf("PointOf: pt=%v i=%d (%c)\n", pt, i, f.s[i])}()
	if i < 0 {
		i = 0
	}
	s := f.s[:i]
	dot := NewDot(f.Origin(), f.Option.Wrap, f.Font)
	for j := 0; j < len(s); j++ {
		dot.Insert(rune(s[j]))
	}
	return dot.Point
}

// PointWalk walks from index s to index e. It returns the point of
// origin for glyph at index e.

// IndexWalk walks from index s at point sp to the terminus, ep.
// It returns the index of the glyph under ep.
