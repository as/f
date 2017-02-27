package frame

import (
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomedium"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/mouse"
	"image"
	"image/color"
	"image/draw"
	//"fmt"
	"time"
)

var (
	AcmeColors = &Colors{
		Back:  image.NewUniform(color.RGBA{0, 0, 0, 0}),
		Text:  image.NewUniform(color.RGBA{255, 0, 0, 255}),
		HText: image.NewUniform(color.RGBA{0, 0, 0, 255}),
		HBack: image.NewUniform(color.RGBA{255, 0, 0, 128}),
	}
	DefaultColors  = defaultColors
	DarkGrayColors = &Colors{
		Back:  image.NewUniform(color.RGBA{33, 33, 33, 4}),
		Text:  image.NewUniform(color.RGBA{0, 128 + 64, 128 + 64, 255}),
		HText: image.NewUniform(color.RGBA{0, 0, 0, 255}),
		HBack: image.NewUniform(color.RGBA{0, 128, 128, 64}),
	}
	GrayColors = &Colors{
		Back:  image.NewUniform(color.RGBA{48, 48, 48, 0}),
		Text:  image.NewUniform(color.RGBA{99, 99, 99, 255}),
		HText: image.NewUniform(color.RGBA{0, 0, 0, 255}),
		HBack: image.NewUniform(color.RGBA{0, 128, 128, 64}),
	}
	defaultColors = &Colors{
		Back:  image.NewUniform(color.RGBA{33, 33, 33, 0}),
		Text:  image.NewUniform(color.RGBA{0, 255, 255, 255}),
		HText: image.NewUniform(color.RGBA{0, 0, 0, 255}),
		HBack: image.NewUniform(color.RGBA{33, 255, 255, 0}),
	}
	defaultOption = &Option{
		Font:   NewFont(parseDefaultFont(12)),
		Wrap:   80,
		Colors: *defaultColors,
	}
	largeScale = &Option{
		Font:   NewFont(parseDefaultFont(24)),
		Wrap:   80,
		Colors: *defaultColors,
	}
)

type Frame struct {
	disp   *image.RGBA
	origin image.Point
	size   image.Point

	Option
	Tick *Tick

	s          []byte
	width      int
	nbytes     int
	dirty      bool
	dirtyrange []Range
	selecting  bool
	menu       bool

	Cache Cache
	// cache for the transformation
	cached draw.Image
	last   rune

	lastmouse  mouse.Event
	mousecache image.Point
	Menu       *Menu
	Mouse      *Mouse

	dot *Dot
}

type Colors struct {
	Text, Back   image.Image
	HText, HBack image.Image
}

type Option struct {
	// Font is the font face for the frame
	*Font

	// Number of glyphs drawn on one line before wrapping
	Wrap int

	// Multiplicative scale factor for X and Y coordinates
	// (1, 1) means no scale.
	//Scale image.Point

	// Colors define the text and background colors for the rame
	// Text: glyph color
	// Back: background color
	// HText: highlighted glyph color
	// HBack: highlighted background color
	Colors Colors

	fontheight int
}

// New initializes a new frame on disp. The first glyph of
// text is inserted at point p. If opt is nil, the default
// color, wrapping, and font settings are used.
func New(origin, size image.Point, events Sender, opt *Option) *Frame {
	if opt == nil {
		opt = defaultOption
	}
	f := &Frame{
		dirty:  true,
		size:   size,
		origin: origin,
		Option: *opt,
		s:      make([]byte, 64),
	}
	menu := &Menu{
		drawer: f,
		sender: events,
		Item: []*Item{
			{Name: "explorer"},
			{Name: "Mink"},
			{Name: "Or"},
			{Name: "So"},
			{Name: "You"},
			{Name: "Think"},
			{Name: "More",
				Menu: &Menu{drawer: f,
					Item: []*Item{
						{Name: "Think"},
						{Name: "Again"},
						{Name: "Or",
							Menu: &Menu{
								drawer: f,
								Item: []*Item{
									{Name: "Dont"},
									{Name: "Think"},
								},
							},
						},
					},
				},
			},
		},
	}
	menu = menu
	f.Menu = NewMenuFS(`C:\menu\`, f, events)
	f.disp = image.NewRGBA(f.Bounds())
	f.cached = image.NewRGBA(image.Rectangle{image.ZP, image.Pt(f.FontHeight(), f.FontHeight())})
	f.flushcache()
	f.Mouse = NewMouse(time.Second/3, events, f)
	return f
}

func (o Option) FontHeight() int {
	return o.Font.Height()
}

func ParseDefaultFont(size float64) font.Face {
	return parseDefaultFont(size)
}

func parseDefaultFont(size float64) font.Face {
	f, err := truetype.Parse(gomedium.TTF)
	if err != nil {
		panic(err)
	}
	return truetype.NewFace(f, &truetype.Options{
		Size: size,
	})
}

type Range struct {
	I, J int
}

func (f *Frame) MarkRange(i, j int) {
	f.dirtyrange = append(f.dirtyrange, Range{i, j})
}
func (f *Frame) DirtyRange() []Range {
	return f.dirtyrange
}
func (f *Frame) CleanRange() {
	f.dirtyrange = nil
}

// Insert inserts s starting from index i in the
// the frame buffer.
func (f *Frame) Insert(s []byte, i int) (err error) {
	if i >= len(f.s) {
		i = len(f.s) - 1
	}
	if i < 0 {
		i = 0
	}
	if s == nil {
		return nil
	}
	f.grow(len(s) + 1)
	f.nbytes += len(s)
	copy(f.s[i+len(s):], f.s[i:])
	copy(f.s[i:], s)
	f.MarkRange(i, i+len(s))
	f.dirty = true
	return nil
}

// Delete erases the range [i:j] in the framebuffer
// TODO: fix i == j
func (f *Frame) Delete(i, j int) (err error) {
	if i > j {
		i, j = j, i
	}
	if i < 0 {
		i = 0
	}
	if j >= len(f.s) {
		j = len(f.s) - 1
	}
	copy(f.s[i:], f.s[j:])
	f.nbytes -= j - i
	if f.nbytes < 0 {
		f.nbytes = 0
	}
	f.MarkRange(i, f.nbytes)
	f.dirty = true
	return nil
}

func (f *Frame) Mark() {
	f.dirty = true
}

var Clip []byte

// Handle handles events sent to the frame. The key.Event
// and mouse.Events are handled.
func (f *Frame) Handle(e interface{}) {

	t := f.Tick
	switch e := e.(type) {
	case key.Event:
		f.dirty = true

		if e.Direction != key.DirPress && e.Direction != key.DirNone {
			break
		}
		switch e.Code {
		case key.CodeRightArrow:
			if e.Modifiers != key.ModShift {
				t.P0++
			}
			t.P1++
		case key.CodeLeftArrow:
			if e.Modifiers != key.ModShift {
				t.P0--
			}
			t.P1--
		case key.CodeDeleteBackspace:
			t.Delete()
		case key.CodeReturnEnter:
			t.Write([]byte{'\n'})
		case key.CodeTab:
			t.Write([]byte("\t"))
		default:
			if e.Rune != -1 {
				t.WriteRune(e.Rune)
			}
		}
	case mouse.Event:
		f.Mouse.Process(e)
		return
	}
}

func (f *Frame) Image() draw.Image {
	return f.disp
}

func (f *Frame) Bytes() []byte {
	return f.s
}

func (f *Frame) Bounds() (r image.Rectangle) {
	return image.Rectangle{image.ZP, f.size}
}

func (f *Frame) RGBA() *image.RGBA {
	return f.disp
}

func (f *Frame) Size() image.Point {
	return f.size
}

func (f *Frame) Release() {

}

func (f *Frame) Dirty() bool {
	return f.dirty
}

type Cache struct {
	i     int
	b     byte
	r     image.Rectangle
	valid bool
}

func (c *Cache) PointOf(i int) (pt image.Point, ok bool) {
	return c.r.Min, c.valid && c.i == i
}

func (c *Cache) IndexOf(pt image.Point) (i int, ok bool) {
	return c.i, c.valid && pt.In(c.r.Sub(image.Pt(5, 0)))
}

var HintColor = image.NewUniform(color.RGBA{255, 0, 0, 255})

func (c *Cache) DrawHint(dst draw.Image) {
	drawBorder(dst, c.r, HintColor, image.ZP, 1)
}

func (c *Cache) Set(bounds image.Rectangle, i int, b byte) {
	c.r = bounds
	c.b = b
	c.i = i
	c.valid = false
}
