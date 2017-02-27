package frame

import (
	"fmt"
	"image"
	"image/draw"
	"os"
)

type Resolver interface {
	PointOf(int) image.Point
	IndexOf(image.Point) int
	Height() int
	Origin() image.Point
}

// Select consists of 3 rectangles
type Select struct {
	Img   *image.RGBA // todo: use New
	Color image.Image

	// anchor index and index of last sweep
	a int
	b int

	// cache of points for a,b
	at image.Point
	bt image.Point

	res Resolver

	// avoid redrawing the entire selection on the
	// canvas by caching the three rectangles representing
	// the selection
	R [3]image.Rectangle
}

func NewSelect(r image.Rectangle, img *image.RGBA, color image.Image, res Resolver) *Select {
	return &Select{
		Img:   img,
		Color: color,
		res:   res,
	}
}

func (s *Select) Open(i int) {
	s.a = i
	s.at = s.res.PointOf(i)
	s.b = i
}

func (s *Select) Seek(offset int64, whence int) (int64, error) {
	s.Sweep(int(offset))
	return offset, nil
}

func (s *Select) Close() {
	s.a, s.b = 0, 0
	s.at, s.bt = image.ZP, image.ZP
	s.Clear()

	//
	// TODO: Huge memory leak here
}

func (s *Select) Clear() {
	r := s.Img.Bounds()
	draw.Draw(s.Img, r, image.Transparent, image.ZP, draw.Src)
}

//
// Rects returns the rectangles representing the active selection. cap(r) == 3
func (s *Select) Rects() (r []image.Rectangle) {
	fmt.Printf("%v,%v\n", s.at, s.bt)
	return s.rects(s.at, s.bt)
}

func (s *Select) DeltaRects() (r []image.Rectangle) {
	panic("unimplemented")
}

func (s *Select) rects(p, q image.Point) (r []image.Rectangle) {
	h := s.res.Height()
	m := s.Img.Bounds().Max
	o := s.res.Origin()
	add := func(x0, y0, x1, y1 int) {
		r = append(r, image.Rect(x0, y0, x1, y1))
	}
	if p.Y == q.Y {
		add(p.X, p.Y, q.X, p.Y+h)
		return
	}
	add(p.X, p.Y, m.X, p.Y+h)
	p.Y += h
	if p.Y != q.Y {
		add(o.X, p.Y, m.X, q.Y)
	}
	add(o.X, p.Y, q.X, q.Y+h)
	return
}

func (s *Select) Resize(size image.Point) {
	r := image.Rect(0, 0, size.X, size.Y)
	if !s.Img.Bounds().Max.In(r) {
		s.Img = image.NewRGBA(r)
		draw.Draw(s.Img, s.Img.Bounds(), s.Img, image.ZP, draw.Src)
	}
}

/*
func init(){
	r0 := image.Rect(0,0,100,100)
	r1 := image.Rect(50,0,100,100)
	r2 := image.Rect(0 ,50,100,100)
	r3 := image.Rect(0 ,0,50,100)
	r4 := image.Rect(0 ,0,0,50)
	x01, _ := Xor(r0,r1)
	x02, _ := Xor(r0,r2)
	fmt.Println(r1, r3)
	panic(".")
}
*/
func Xor(r0, r1 image.Rectangle) (image.Rectangle, error) {
	switch {
	case r0 == image.ZR && r1 == image.ZR:
		return image.ZR, fmt.Errorf("xor: rectangle has infinite dimension")
	case r0 == image.ZR:
		return r1, nil
	case r1 == image.ZR:
		return r0, nil
	case r0 == r1:
		return image.ZR, nil
	case r0.Dy() == r1.Dy():
		y0 := r0.Min.Y
		y1 := r0.Max.Y
		if r0.Min == r1.Min { // erase from left
			x0 := min(r0.Dx(), r1.Dx()) + r0.Min.X
			x1 := max(r0.Max.X, r1.Max.X)
			return image.Rect(x0, y0, x1, y1), nil
		}
		if r0.Max == r1.Max { // erase from right
			x0 := min(r0.Min.X, r1.Min.X)
			x1 := r0.Max.X - max(r0.Min.X, r1.Min.X)
			return image.Rect(x0, y0, x1, y1), nil
		}
	case r0.Dx() == r1.Dx():
		x0 := r0.Min.X
		x1 := r0.Max.X
		if r0.Min == r1.Min { // erase from top
			y0 := min(r0.Dy(), r1.Dy()) + r0.Min.Y
			y1 := max(r0.Max.Y, r1.Max.Y)
			return image.Rect(x0, y0, x1, y1), nil
		}
		if r0.Max == r1.Max { // erase from botton
			y0 := max(r0.Min.Y, r1.Min.Y)
			y1 := r0.Max.Y - max(r0.Min.Y, r1.Min.Y)
			return image.Rect(x0, y0, x1, y1), nil
		}
	}
	return image.ZR, fmt.Errorf("xor: can't XOR %s ^ %s", r0, r1)
}

func (s *Select) update(i, x0, y0, x1, y1 int, erase bool) {
	var err error
	if erase {
		s.R[i], err = Xor(s.R[i], image.Rect(x0, y0, x1, y1))
		if err != nil {
			panic(err)
		}
	} else {
		s.R[i] = s.R[i].Union(image.Rect(x0, y0, x1, y1))
	}
}

func (s *Select) update3(p, q image.Point, erase bool) {
	h := s.res.Height()
	m := s.Img.Bounds().Max
	o := s.res.Origin()

	// selection spans the same line
	if p.Y == q.Y {
		s.update(0, p.X, p.Y, q.X, p.Y+h, erase)
		return
	}

	// draw up to three rects for selection
	s.update(0, p.X, p.Y, m.X, p.Y+h, erase)
	p.Y += h
	if p.Y != q.Y {
		s.update(1, o.X, p.Y, m.X, q.Y, erase)
	}
	s.update(2, o.X, p.Y, q.X, q.Y+h, erase)
}

func (s *Select) Update(j int) {
	a := s.a
	b := s.b
	c := j
	if abs(b-c) < 1 {
		return
	}
	pt := s.res.PointOf
	at, bt, ct := pt(a), pt(b), pt(c)
	switch {
	case a <= b && b < c: // down
		s.update3(bt, ct, false)
	case a <= c && c < b: // down and up
		s.update3(ct, bt, true)
	case c < a && a <= b: // down and up over
		s.update3(at, bt, true)
		s.update3(ct, at, false)
	case c < b && b <= a: // up
		s.update3(ct, bt, false)
	case b < c && c < a: // up and down
		s.update3(bt, ct, true)
	case b < a && a < c: // up and down over
		s.update3(bt, at, true)
		s.update3(at, ct, false)
	}
	s.b = c
	s.bt = ct
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func (s *Select) Draw(p, q image.Point, bg image.Image) {
	h := s.res.Height()
	m := s.Img.Bounds().Max
	o := s.res.Origin()

	// selection spans the same line
	if p.Y == q.Y {
		fmt.Fprintf(os.Stderr, "Select.Draw.1a: (%d-%d),(%d-%d)\n", p.X, p.Y, q.X, p.Y+h)
		s.draw(p.X, p.Y, q.X, p.Y+h, bg)
		return
	}

	// draw up to three rects for selection
	fmt.Fprintf(os.Stderr, "Select.Draw.1b: (%d-%d),(%d-%d)\n", p.X, p.Y, q.X, p.Y+h)
	s.draw(p.X, p.Y, m.X, p.Y+h, bg)
	p.Y += h
	if p.Y != q.Y {
		fmt.Fprintf(os.Stderr, "Select.Draw.2b: (%d-%d),(%d-%d)\n", p.X, p.Y, q.X, p.Y+h)
		s.draw(o.X, p.Y, m.X, q.Y, bg)
	}
	fmt.Fprintf(os.Stderr, "Select.Draw.3b: (%d-%d),(%d-%d)\n", p.X, p.Y, q.X, p.Y+h)
	s.draw(o.X, p.Y, q.X, q.Y+h, bg)
}

// drawrect draws a rectangle
func (s *Select) draw(x, y, xx, yy int, bg image.Image) {
	r := image.Rect(x, y, xx, yy)
	draw.Draw(s.Img, r, bg, image.ZP, draw.Src)
}

func (s *Select) Sweep(j int) {
	s.Update(j)
	for i, r := range s.R {
		fmt.Fprintf(os.Stderr, "s.r%d: %s\n", i, r)

	}
	fmt.Fprintf(os.Stderr, "\n")
	return
	a := s.a
	b := s.b
	c := j
	if abs(b-c) < 1 {
		return
	}
	bg, erase := s.Color, image.Transparent
	pt := s.res.PointOf
	at, bt, ct := pt(a), pt(b), pt(c)
	switch {
	case a <= b && b < c: // down
		s.Draw(bt, ct, bg)
	case a <= c && c < b: // down and up
		s.Draw(ct, bt, erase)
	case c < a && a <= b: // down and up over
		s.Draw(at, bt, erase)
		s.Draw(ct, at, bg)
	case c < b && b <= a: // up
		s.Draw(ct, bt, bg)
	case b < c && c < a: // up and down
		s.Draw(bt, ct, erase)
	case b < a && a < c: // up and down over
		s.Draw(bt, at, erase)
		s.Draw(at, ct, bg)
	}
	s.b = c
	s.bt = ct
}

func (s Select) Sp() image.Point { return s.at }
func (s Select) Ep() image.Point { return s.bt }
func (s Select) Addr() (i, j int) {
	return s.a, s.b
}
