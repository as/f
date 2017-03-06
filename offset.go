package frame

import (
	"fmt"
	"image"
)

// Origin returns the insertion point of the first
// glyph in the frame
func (f *Frame) Origin() image.Point {
	return f.Bounds().Min.Add(f.origin)
}

func (f *Frame) Box(bn int) *Box {
	return f.boxes.Box[bn]
}

func (f *Frame) IndexOf(pt image.Point) (bn, offset int) {
	pt = f.alignY(pt)
	dot := NewDot(f.Origin(), f.Option.Wrap, f.Font)
	var box *Box
	for bn, box = range f.Boxes() {
		switch {
		case dot.Y < pt.Y:
			// nothing special
		case dot.Y == pt.Y:
			// point intersects box
			dot.indexOf(f.Box(bn), pt)
		case dot.Y > pt.Y:
			// advanced too far
			if bn-1 < 0 {
				// bug fix for crash: happened when selecting
				// and dragging all the way to the top
				return bn, offset
			}
		}
		dot.InsertBox(box)
		offset += box.Len()
	}
	return
}

// PointOf computes the point of origin for glyph x
func (f *Frame) PointOf(x int) (pt image.Point) {
	defer func() { fmt.Printf("PointOf: pt=%v x=%d\n", pt, x) }()
	pt = f.alignY(pt)
	dot := NewDot(f.Origin(), f.Option.Wrap, f.Font)
	i := 0
	bn := 0
	var box *Box
	for bn, box = range f.Boxes() {
		if i+box.Len() == x {
			return dot.Point
		}
		if i+box.Len() > x {
			break
		}
		dot.InsertBox(box)
		i += box.Len()
	}
	bn, _ = f.boxes.Find(bn-1, i, x)
	dot.InsertBox(f.Box(bn))
	return dot.Point
}

// PointOf computes the point of origin for glyph i
func (f *Frame) zPointOf(i int) (pt image.Point) {
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
