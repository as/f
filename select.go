package frame

import (
	"image"
	"image/draw"
	"fmt"
)

type Resolver interface {
	PointOf(int) image.Point
	IndexOf(image.Point) int
	Height() int
	Origin() image.Point
}

// Select consists of 3 rectangles
type Select struct {
	Img *image.RGBA // todo: use New
	Color image.Image

	// anchor index and index of last sweep
	a int
	b int

	// cache of points for a,b
	at image.Point
	bt image.Point

	res Resolver
}

func NewSelect(r image.Rectangle, img *image.RGBA, color image.Image, res Resolver) *Select {
	return &Select{
		Img: img,
		Color: color,
		res: res,
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

func (s *Select) Close(){
	s.a, s.b = 0, 0
	s.at, s.bt = image.ZP, image.ZP
	s.Clear()
	
	//
	// TODO: Huge memory leak here
}

func (s *Select) Clear(){
	r := s.Img.Bounds()
	draw.Draw(s.Img, r, image.Transparent, image.ZP, draw.Src)
}

// Rects returns the rectangles representing the active selection. cap(r) == 3 
func (s *Select) Rects() (r []image.Rectangle){
	fmt.Printf("%v,%v\n", s.at, s.bt)
	return s.rects(s.at,s.bt)
}
func (s *Select) rects(p, q image.Point) (r []image.Rectangle){
	h := s.res.Height()
	m := s.Img.Bounds().Max
	o := s.res.Origin()
	add := func(x0,y0,x1,y1 int){
		r = append(r, image.Rect(x0,y0,x1,y1))
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

func (s *Select) Draw(p, q image.Point, bg image.Image) {
	h := s.res.Height()
	m := s.Img.Bounds().Max
	o := s.res.Origin()

	// selection spans the same line
	if p.Y == q.Y {
		s.draw(p.X, p.Y, q.X, p.Y+h, bg)
		return
	}

	// draw up to three rects for selection
	s.draw(p.X, p.Y, m.X, p.Y+h, bg)
	p.Y += h
	if p.Y != q.Y {
		s.draw(o.X, p.Y, m.X, q.Y, bg)
	}
	 s.draw(o.X, p.Y, q.X, q.Y+h, bg)
}

// drawrect draws a rectangle
func (s *Select) draw(x, y, xx, yy int, bg image.Image) {
	r := image.Rect(x, y, xx, yy)
	draw.Draw(s.Img, r, bg, image.ZP, draw.Src)
}

func (s *Select) Sweep(j int) {
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
		//s.Draw(ct, bt, erase)
	case c < a && a <= b: // down and up over
		//s.Draw(at, bt, erase)
		s.Draw(ct, at, bg)
	case c < b && b <= a: // up
		s.Draw(ct, bt, bg)
	case b < c && c < a: // up and down
		//s.Draw(bt, ct, erase)
	case b < a && a < c: // up and down over
		//s.Draw(bt, at, erase)
		s.Draw(at, ct, bg)
	}
	erase=erase
	s.b = c
	s.bt = ct
}

func (s Select) Sp() image.Point { return s.at }
func (s Select) Ep() image.Point { return s.bt }
func (s Select) Addr() (i, j int) {
	return s.a, s.b
}
