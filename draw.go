package frame

import (
	"bytes"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
	"math"
	"fmt"
)

func (f *Frame) Draw() {
	f.Redraw(f.selecting)
}

type Drawer interface {
	measure([]byte) int
	drawtext(image.Point, int, []byte) int
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

func XEllipse(dst draw.Image, c image.Point, src image.Image, a, b, thick float64, sp image.Point, alpha, phi int) {
	th := int(thick)
	ath := th
	dt := 1 / a * float64(ath/2)
	for theta := float64(0); theta <= 2*math.Pi; theta += dt {
		x, y := int(math.Cos(theta)*a), int(math.Sin(theta)*b)
		r := image.Rect(0, 0, th, th).Add(c)
		draw.Draw(dst, r.Add(image.Pt(x-ath, y-ath)), src, sp, draw.Src)
	}
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

// Redraw redraws the entire frame. The caller should check
// that the frame is Dirty before calling this in a tight
// loop
func (f *Frame) Redraw(selecting bool) {
	dy := f.Origin().Y
	draw.Draw(f.disp, f.Bounds(), f.Colors.Back, image.ZP, draw.Src)
	dx := f.Origin().X
	for s := f.s[:f.nbytes]; ; {
		i := len(s)
		if i == 0 {
			break
		}
		if dx >= f.Wrap {
			dx = f.Origin().X
		}
		j := bytes.Index(s[:i], NL)
		if j >= 0 {
			i = j
			if dx >= f.Wrap {
				dx = f.Origin().X
				dy += f.FontHeight()
			}
		}
		if dy >= f.Bounds().Max.Y {
			f.size.Y *= 2
			f.disp = image.NewRGBA(image.Rect(0, 0, f.size.X, f.size.Y))
			f.Redraw(selecting)
			return
		}

		dx += f.drawtext(image.Pt(f.Origin().X, dy), 75-dx-f.Origin().X, s[:i], )		
		if dx >= f.Wrap {
			dx = f.Origin().X
			dy += f.FontHeight()
		}
		
		if j >= 0 {
			i++
		}
		if i == len(s) {
			break
		}
		s = s[i:]
	}
	if selecting {
		f.Tick.Sweep(f.Tick.P1)
	}
	f.Tick.Draw()
	if f.Menu.visible {
		f.Menu.Draw(f.disp)
		//f.drawmenu(f.mousecache)
	}
	f.dirty = false
}

// drawtext draws the slice s at position p and returns
// the horizontal displacement dx without line wrapping
func (f *Frame) drawtext(pt image.Point, width int, s []byte) (dx int) {
	defer func() { fmt.Printf("drawtext %q @ %v drew %d pix\n", s, pt, dx) }()
	return f.stringbg(f.disp, pt, f.Colors.Text, image.ZP, f.Font, s, width, f.Colors.Text, image.ZP)
}

func (f *Frame) measure(s []byte) int {
	return int(font.MeasureBytes(f.Font, s) >> 6)
}

func (f *Frame) stringbg(dst draw.Image, p image.Point, src image.Image,sp image.Point, font font.Face, s []byte, width int, bg image.Image, bgp image.Point) int {
	h := f.FontHeight()
	h = int(float64(h) - float64(h)/float64(5))
	for _, v := range s {
		fp := fixed.P(p.X, p.Y)
		dr, mask, maskp, advance, ok := font.Glyph(fp, rune(v))
		if !ok {
			break
		}
		dr.Min.Y += h
		dr.Max.Y += h
		draw.DrawMask(dst, dr, src, sp, mask, maskp, draw.Over)
		dx := int((advance + f.Font.Kern(f.last, rune(v))) >> 6)
		p.X += dx
		f.last = rune(v)
		width -= dx
		if width < 1{
			break
		}
	}
	return int(p.X)
}

// drawsel draws a highlight over points p through q. A highlight
// is a rectanguloid over three intersecting rectangles representing
// the highlight bounds.
func (t *Tick) drawsel(p, q image.Point, bg image.Image) {
	h := t.Fr.FontHeight()
	m := t.Fr.Bounds().Max
	o := t.Fr.Origin()

	// selection spans the same line
	if p.Y == q.Y {
		t.draw(p.X, p.Y, q.X, p.Y+h, bg)
		return
	}

	// draw up to three rectangles for the
	// selection

	t.draw(p.X, p.Y, m.X, p.Y+h, bg)
	p.Y += h
	if p.Y != q.Y {
		t.draw(o.X, p.Y, m.X, q.Y, bg)
	}
	t.draw(o.X, p.Y, q.X, q.Y+h, bg)
}

func (t *Tick) fill(x, y, xx, yy int) {
	t.drawrect(image.Pt(x, y), image.Pt(xx, yy))
}
func (t *Tick) unfill(x, y, xx, yy int) {
	t.deleterect(image.Pt(x, y), image.Pt(xx, yy))
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// drawrect draws a rectangle over the glyphs p0:p1
func (t *Tick) draw(x, y, xx, yy int, bg image.Image) {
	r := image.Rect(x, y, xx, yy)
	draw.Draw(t.Img, r, bg, image.ZP, draw.Src)
}

// drawrect draws a rectangle over the glyphs p0:p1
func (t *Tick) drawrect(pt0, pt1 image.Point) {
	r := image.Rect(pt0.X, pt0.Y, pt1.X, pt1.Y)
	draw.Draw(t.Img, r, t.Fr.Colors.HBack, image.ZP, draw.Src)
}

// delete draws a rectangle over the glyphs p0:p1
func (t *Tick) deleterect(pt0, pt1 image.Point) {
	r := image.Rect(pt0.X, pt0.Y, pt1.X, pt1.Y)
	draw.Draw(t.Img, r, image.Transparent, image.ZP, draw.Src)
}
