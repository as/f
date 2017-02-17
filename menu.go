package frame

import (
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"
)

type MenuEvent struct {
	image.Point
	Strings []string
}

type Opt struct {
	Inset         int
	Width, Height int
	FGColor       image.Image
	BGColor       image.Image
	BGColor2      image.Image
	BRColor       image.Image
}

var DefaultOpt = Opt{
	Inset: -2,
	Width: 100, Height: 33,
	FGColor:  image.NewUniform(color.RGBA{44, 44, 44, 0}),
	BGColor:  image.NewUniform(color.RGBA{22, 22, 22, 0}),
	BGColor2: image.NewUniform(color.RGBA{33, 33, 33, 0}),
	BRColor:  image.NewUniform(color.RGBA{22, 22, 22, 0}),
}

type Menu struct {
	Item []*Item
	Sel  *Item

	sp             image.Point
	visible, dirty bool

	*Opt
	sender Sender
	drawer Drawer
}

type Item struct {
	Name string
	Menu *Menu

	sp image.Point
	*Opt
}

func NewMenuJSON(json []byte, dr Drawer, se Sender) *Menu {
	return nil
}

func NewMenuFS(basedir string, dr Drawer, se Sender) *Menu {
	if basedir == "" {
		return nil
	}
	fd, err := os.Open(basedir)
	if err != nil {
		return nil
	}
	defer fd.Close()
	fi, err := fd.Stat()
	if err != nil || !fi.IsDir() {
		return nil
	}
	m := &Menu{
		drawer: dr,
		sender: se,
	}
	flist, _ := fd.Readdirnames(-1)
	for _, nm := range flist {
		it := &Item{
			Name: nm,
		}
		mm := NewMenuFS(filepath.Join(basedir, nm), dr, se)
		it.Menu = mm
		m.Item = append(m.Item, it)
	}
	return m
}

func (m *Menu) Strings() []string {
	if m.Sel == nil {
		return nil
	}
	if m.Sel.Menu == nil {
		return []string{m.Sel.Name}
	}
	return append([]string{m.Sel.Name}, m.Sel.Menu.Strings()...)
}

func (m *Menu) String() string {
	if m.Sel == nil {
		return ""
	}
	if m.Sel.Menu == nil {
		return m.Sel.Name
	}
	return m.Sel.Name + "/" + m.Sel.Menu.String()
}

func (m *Menu) Visible() bool {
	return m.visible
}

func (m *Menu) Unselect() {
	if m.Sel == nil {
		return
	}
	x := m.Sel
	m.Sel = nil
	if x.Menu != nil {
		x.Menu.Unselect()
	}
}

func (m *Menu) Hit(pt image.Point) {
	m.hit(pt)
	m.sender.Send(MenuEvent{Point: pt, Strings: m.Strings()})
}

func (m *Menu) hit(pt image.Point) {
	sp := m.sp
	size := m.Size()
	dy := size.Y/(len(m.Item)) - 2
	x2 := sp.X + size.X
	for _, v := range m.Item {
		y2 := sp.Y + dy
		r := image.Rect(sp.X, sp.Y, x2, y2+1)
		if pt.In(r) {
			m.Sel = v
			if m.Sel.Menu != nil {
				m.Sel.Menu.sp = image.Pt(x2+1, sp.Y)
			}
			return
		}
		//v.sp = sp
		//v.size = image.Pt(x2, y2+1)
		sp.Y += dy + 2
	}

	if m.Sel != nil && m.Sel.Menu != nil {
		m.Sel.Menu.hit(pt)
	}
}

func (m *Menu) Draw(dst draw.Image) {
	if m == nil {
		return
	}
	if m.Opt == nil {
		m.Opt = &DefaultOpt
	}
	size := m.Size()
	drawBorder(dst, m.Bounds().Inset(-2), m.Opt.BRColor, image.ZP, 1)
	draw.Draw(dst, m.Bounds(), m.BGColor, image.ZP, draw.Src)
	sp := m.sp
	dy := size.Y/(len(m.Item)) - 2
	x2 := sp.X + size.X
	for i, v := range m.Item {
		y2 := sp.Y + dy
		r := image.Rect(sp.X, sp.Y, x2, y2+1)
		bg, bord := m.FGColor, m.BRColor
		if m.Item[i] == m.Sel {
			bg = m.BGColor2
		}
		drawBorder(dst, r.Inset(-2), bord, image.ZP, 2)
		draw.Draw(dst, r, bg, image.ZP, draw.Src)

		dx := m.drawer.measure([]byte(v.Name))
		bw := m.Opt.Width
		csp := (dx+bw)/2 - dx
		m.drawer.drawtext(sp.Add(image.Pt(4+csp, 4)), bw, []byte(v.Name))

		if m.Item[i] == m.Sel {

			v.Menu.Draw(dst)
		}
		sp.Y += dy + 2
	}
}

func (m *Menu) Size() image.Point {
	N := len(m.Item)
	if m.Opt == nil {
		m.Opt = &DefaultOpt
	}
	h := m.Opt.Height
	return image.Pt(m.Opt.Width, h*N)
}

func (m *Menu) Bounds() image.Rectangle {
	size := m.Size()
	return image.Rect(m.sp.X, m.sp.Y, size.X+m.sp.X, size.Y+m.sp.Y)
}
