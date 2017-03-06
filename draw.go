package frame

import (
	"fmt"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
)

func (f *Frame) Draw(force bool) {
	f.Tick.Draw()
	if !force {
		for _, r := range f.DirtyRange() {
			f.RedrawRange(r.I, r.J)
		}
	} else {
		f.Redraw(f.selecting)
	}
	f.CleanRange()
}

type Drawer interface {
	measure([]byte) int
	drawtext(image.Point, int, []byte) (int, int)
}

var NL = []byte{'\n'}

var MenuColor = image.NewUniform(color.RGBA{128, 128, 128, 255})

func (f *Frame) drawmenu(pt image.Point) {
	for _, v := range []image.Point{
		image.Pt(0, 34),
		image.Pt(-70, 0),
		image.Pt(70, 0),
		image.Pt(0, -34),
	} {
		Ellipse(f.disp, pt.Add(v), MenuColor, 50, 20, 5, image.ZP, 5, 5)
	}
	//Ellipse(f.disp, pt, B, 99, 99, 1, image.ZP, 1, 5)
	//Ellipse(f.disp, pt, B, 94, 94, 1, image.ZP, 1, 5)
	//Ellipse(f.disp, pt, B, 94, 94, 1, image.ZP, 1, 5)
}

func (f *Frame) Resize(size image.Point) {
	f.Tick.Resize(size)
	f.resize(size)
}

func (f *Frame) reallocY(dot image.Point) bool {
	if dot.Y >= f.Bounds().Max.Y {
		f.resize(image.Pt(f.size.X, f.size.Y*2))
		return true
	}
	return false
}

func (f *Frame) resize(size image.Point) {
	f.size = size
	f.disp = image.NewRGBA(image.Rect(0, 0, f.size.X, f.size.Y))
	f.Redraw(f.selecting)
	f.Option.Wrap = f.size.X - 2*f.Origin().X
	return
}

// Redraw redraws the entire frame. The caller should check
// that the frame is Dirty before calling this in a tight
// loop
func (f *Frame) Redraw(selecting bool) {
	draw.Draw(f.disp, f.Bounds(), f.Colors.Back, image.ZP, draw.Src)
	f.RedrawRange(0, f.nbytes)
}

func (f *Frame) RedrawRange(i, j int) {
	f.boxes.Dump()
	bi, _ := f.boxes.Find(0, 0, i)
	bj, _ := f.boxes.Find(bi, 0, j)
	fmt.Printf("note: i=%d j=%d bi=%d bj=%d\n", i, j, bi, bj)
	bi = 0
	bj = len(f.boxes.Box)
	f.RedrawBox(bi, bj)
}

func note(i, j int) {
	fmt.Printf("note: i=%d j=%d\n", i, j)
}

func (f *Frame) RedrawBox(i, j int) {
	dot := NewDot(f.Origin(), f.Option.Wrap, f.Font)
	note(i, j)
	for _, box := range f.Boxes()[i:j] {
		i, sp := 0, dot.Point
		w := sp.X
		dot.InsertBox(box)
		r := image.Rect(sp.X, sp.Y, w, sp.Y+dot.font.Height())
		draw.Draw(f.disp, r, f.Colors.Back, image.ZP, draw.Src)
		s := box.Bytes()
		if i-1 >= 0 && i-1 < len(s) && s[i-1] == '\n' {
			f.RedrawBytes(dot.Point, dot.maxw, s)
		} else {
			f.RedrawBytes(dot.Point, dot.maxw, s)
		}
	}
}

func (f *Frame) RedrawBytes(origin image.Point, width int, s []byte) {
	fmt.Printf("RedrawBytes: origin=%v width=%v s=%q\n", origin, width, s)
	dot := NewDot(origin, width, f.Font)
	for s := s; len(s) != 0; {
		i, sp := 0, dot.Point
		w := sp.X
		for ; i < len(s) && sp.Y == dot.Y; i++ {
			w = dot.X
			dot.Insert(rune(s[i]))
		}
		if i-1 >= 0 && i-1 < len(s) && s[i-1] == '\n' {
			f.drawtext(sp, dot.maxw-w, s[:i-1])
		} else {
			f.drawtext(sp, dot.maxw-w, s[:i])
		}
		s = s[i:]
	}
}

// drawtext draws the slice s at position p and returns
// the horizontal displacement dx without line wrapping
func (f *Frame) drawtext(pt image.Point, width int, s []byte) (dx int, i int) {
	//defer func() { fmt.Printf("drawtext %q @ %v drew %d pix\n", s, pt, dx) }()
	return f.stringbg(f.disp, pt, f.Colors.Text, image.ZP, f.Font, s, width, f.Colors.Text, image.ZP)
}

func (f *Frame) measure(s []byte) int {
	return int(font.MeasureBytes(f.Font, s) >> 6)
}

func (f *Frame) stringbg(dst draw.Image, p image.Point, src image.Image, sp image.Point, font font.Face, s []byte, width int, bg image.Image, bgp image.Point) (int, int) {
	h := f.Font.Height()
	h = int(float64(h) - float64(h)/float64(5))
	i := 0
	if f.dot == nil {
		f.dot = NewDot(f.Origin(), f.Option.Wrap, f.Font)
	}
	for _, v := range s {
		fp := fixed.P(p.X, p.Y)

		if f.dot.Visible(rune(v)) {
			dr, mask, maskp, _, ok := font.Glyph(fp, rune(v))
			if !ok {
				break
			}
			dr.Min.Y += h
			dr.Max.Y += h
			draw.DrawMask(dst, dr, src, sp, mask, maskp, draw.Over)
		}

		dx := f.dot.Advance(rune(v))
		//dx := int((advance + f.Font.Kern(f.last, rune(v))) >> 6)
		p.X += dx
		i++
		f.last = rune(v)
		width -= dx
		if width < 1 {
			break
		}
	}
	return int(p.X), i
}

// drawsel draws a highlight over points p through q. A highlight
// is a rectanguloid over three intersecting rectangles representing
// the highlight bounds.
func (t *Tick) drawsel(p, q image.Point, bg image.Image) {
	t.Pen[0].Draw(p, q, bg)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Ellipse draws a filled ellipse at center point c
// and eccentricity a and b. The thick argument is ignored
// (until a line drawing function is available)
//
// The method uses an efficient integer-based rasterization
// technique originally described in:
//
// McIlroy, M.D.: There is no royal road to programs: a trilogy
// on raster ellipses and programming methodology,
// Computer Science TR155, AT&T Bell Laboratories, 1990
//
func Ellipse(dst draw.Image, c image.Point, src image.Image, a, b, thick int, sp image.Point, alpha, phi int) {
	xc, yc := c.X, c.Y
	var (
		x, y       = 0, b
		a2, b2     = a * a, b * b
		crit1      = -(a2/4 + a%2 + b2)
		crit2      = -(b2/4 + b%2 + a2)
		crit3      = -(b2/4 + b%2)
		t          = -a2 * y
		dxt, dyt   = 2 * b2 * x, -2 * a2 * y
		d2xt, d2yt = 2 * b2, 2 * a2
		incx       = func() { x++; dxt += d2xt; t += dxt }
		incy       = func() { y--; dyt += d2yt; t += dyt }
	)
	point := func(x, y int) {
		draw.Draw(dst, image.Rect(x, y, x+1, yc), src, sp, draw.Over)
	}

	for y >= 0 && x <= a {
		point(xc+x, yc+y)
		if x != 0 || y != 0 {
			point(xc-x, yc-y)
		}
		if x != 0 && y != 0 {
			point(xc+x, yc-y)
			point(xc-x, yc+y)
		}
		if t+b2*x <= crit1 || t+a2*y <= crit3 {
			incx()
		} else if t-a2*y > crit2 {
			incy()
		} else {
			incx()
			incy()
		}
	}
}
