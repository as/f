package frame

import (
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/mouse"
	"image"
	"image/draw"
)

type Direction int

const (
	DirUp   Direction = -1
	DirNone Direction = 0
	DirDown Direction = 1
)

type Handler interface {
	Draw()
	screen.Buffer
	Handle(interface{})
	Dirty() bool
}

type Sc struct {
	disp *image.RGBA
	src  Handler

	origin image.Point
	size   image.Point
	sp     image.Point
	width  int

	Bar     image.Rectangle
	holding mouse.Button

	dirty bool

	BarColor, FrameColor image.Image
}

func NewSc(src Handler, origin, size image.Point, width int, bar, frame image.Image) *Sc {
	r := image.Rectangle{origin, origin.Add(size)}
	return &Sc{
		dirty:      true,
		disp:       image.NewRGBA(r),
		size:       size,
		src:        src,
		origin:     origin,
		sp:         src.Bounds().Min,
		width:      width,
		BarColor:   bar,
		FrameColor: frame,
		Bar:        image.Rect(0, 0, width, r.Max.Y),
	}
}

func (s *Sc) RGBA() *image.RGBA {
	return s.disp
}
func (s *Sc) Dirty() bool {
	return s.dirty || s.src.Dirty()
}

func (s *Sc) Clicksb(pt image.Point, dir Direction) {
	rat := float64(s.src.Bounds().Dy()) / float64(s.Bounds().Dy())
	dy := int(float64(pt.Y-s.Bounds().Min.Y) * rat)
	switch dir {
	case DirDown:
		s.sp.Y += dy
	case DirUp:
		s.sp.Y -= dy
	default:
		s.sp.Y = dy
	}
	if s.sp.Y > s.src.Bounds().Max.Y {
		s.sp.Y = s.src.Bounds().Max.Y - s.Bounds().Max.Y
	} else if s.sp.Y < s.src.Bounds().Min.Y {
		s.sp.Y = s.src.Bounds().Min.Y
	}
	s.updatebar()
	s.dirty = true
}

func (s *Sc) updatebar() {
	r := s.Bounds()
	sp := s.sp
	src := s.src.Bounds()
	rat := float64(r.Dy()) / float64(src.Dy())
	s.Bar.Min.Y = int(float64(sp.Y-src.Min.Y) * rat)
	s.Bar.Max.Y = int(float64(sp.Y+s.Bounds().Dy()) * rat)
}

func (s *Sc) Project(x, y float32) (float32, float32) {
	x += float32(-s.width)
	y += float32(s.sp.Y)
	return x, y
}

func (s *Sc) Handle(e interface{}) {

	switch e := e.(type) {
	case key.Event:
		switch e.Code {
		case key.CodeUpArrow:
			s.sp.Y -= 10
			s.dirty = true
		case key.CodeDownArrow:
			s.sp.Y += 10
			s.dirty = true
		}
		s.src.Handle(e)
	case mouse.Event:
		pt := image.Pt(int(e.X), int(e.Y))
		if pt.In(s.EmbedBounds()) {
			e.X, e.Y = s.Project(e.X, e.Y)
			s.src.Handle(e)
		} else {
			if e.Direction == mouse.DirPress {
				switch e.Button {
				case mouse.ButtonLeft:
					s.Clicksb(pt, DirUp)
				case mouse.ButtonRight:
					s.Clicksb(pt, DirDown)
				case mouse.ButtonMiddle:
					s.Clicksb(pt, DirNone)
				}
				s.holding = e.Button
			} else if e.Direction == mouse.DirRelease {
				s.holding = mouse.ButtonNone
			} else if s.holding == mouse.ButtonMiddle {
				s.Clicksb(pt, DirNone)
			}
		}
		return
	default:
		//s.src.Handle(e)
	}
}

func (s *Sc) Draw() {
	r := image.Rectangle{image.ZP, s.size}
	draw.Draw(s.disp, image.Rect(0, 0, s.width, r.Max.Y), s.FrameColor, image.ZP, draw.Src)
	draw.Draw(s.disp, s.Bar, s.BarColor, image.ZP, draw.Src)
	drawBorder(s.disp, image.Rect(0, 0, s.width, r.Max.Y), GrayColors.Back, image.ZP, 2)
	if s.src.Dirty() {
		s.src.Draw()
	}
	draw.Draw(s.disp, s.EmbedBounds(), s.src.RGBA(), s.sp, draw.Src)
	s.dirty = false
}
func mul(r image.Rectangle, rat float64) image.Rectangle {
	return image.Rect(
		r.Min.X,
		int(float64(r.Min.Y)*rat),
		r.Max.X,
		int(float64(r.Max.Y)*rat),
	)
}

func (s *Sc) EmbedBounds() image.Rectangle {
	r := s.Bounds()
	return image.Rect(s.width+1, 0, r.Max.X+s.width, r.Max.Y)
}

func (s *Sc) Bounds() image.Rectangle {
	return image.Rectangle{image.ZP, s.size}
}
func (s *Sc) Release() {

}
func (s *Sc) Size() image.Point {
	return s.Bounds().Max
}

type Scroller interface {
	Image() draw.Image
	Bounds() image.Rectangle
	Visible() image.Rectangle
	Shift(pt image.Point)
}

func drawBorder(dst draw.Image, r image.Rectangle, src image.Image, sp image.Point, thick int) {
	draw.Draw(dst, image.Rect(r.Min.X, r.Min.Y, r.Max.X, r.Min.Y+thick), src, sp, draw.Src)
	draw.Draw(dst, image.Rect(r.Min.X, r.Max.Y-thick, r.Max.X, r.Max.Y), src, sp, draw.Src)
	draw.Draw(dst, image.Rect(r.Min.X, r.Min.Y, r.Min.X+thick, r.Max.Y), src, sp, draw.Src)
	draw.Draw(dst, image.Rect(r.Max.X-thick, r.Min.Y, r.Max.X, r.Max.Y), src, sp, draw.Src)
}
