package frame

import (
	"bytes"
	"image"
	"image/draw"
	"image/color"
	"io"
	"fmt"
)

type Tick struct {
	Pen [3]*Select
	P0, P1 int
	Fr     *Frame
	dirty bool
}

var(
	tickBlue  = image.NewUniform(color.RGBA{132, 255, 255, 0})
	tickRed   = image.NewUniform(color.RGBA{132, 254, 128, 0})
	tickGreen = image.NewUniform(color.RGBA{132, 254, 128, 99})
)

func NewTick(f *Frame) *Tick{
	t := &Tick{
		Fr: f,
	}
	back := image.NewRGBA(f.Bounds())
	for i, bg := range []*image.Uniform{tickBlue, tickGreen, tickRed}{
		t.Pen[i] = NewSelect(f.Bounds(),back, bg, Resolver(f))
	}
	return t
}

func (t *Tick) Open(i int){
	t.Close()
	t.P0 = i
	t.P1 = i
	t.Pen[0].Open(i)
}

func (t *Tick) Sweep(i int){
	t.Seek(int64(i), 1)
}

func (t *Tick) Seek(offset int64, whence int) (int64, error){
	return t.Pen[0].Seek(offset, whence)
}

func (t *Tick) Commit(){
	t.P0 = t.Pen[0].a
	t.P1 = t.Pen[0].b
}

func (t *Tick) Cancel(){
	t.Pen[0].a = t.P0
	t.Pen[0].b = t.P1
}

func (t *Tick) Close() error {
	for _, v := range t.Pen{
		if v != nil{
			v.Close()
		}
	}	
	return nil
}

func (t *Tick) In(pt image.Point) bool{
	for i, r := range t.Rects(){
		fmt.Printf("Tick.In: %v in %v (#%d)\n", pt, r, i)
		if pt.In(r){
			return true
		}
	}
	return false
}

func (t *Tick) Resize(size image.Point) {
	for _, v := range t.Pen{
		if v != nil{
			v.Resize(size)
		}
	}	
}

func (t *Tick) Rects() (r []image.Rectangle){
	return t.Pen[0].Rects()
}

// drawrect draws a rectangle over the glyphs p0:p1
func (t *Tick) draw(x, y, xx, yy int, bg image.Image) {
	t.Pen[0].draw(x,y,xx,yy,bg)
}

func (t *Tick) Draw() error {
	if t.P1 == t.P0 {
		pt := t.Fr.PointOf(t.P1)
		r := image.Rect(0, 0, 2, t.Fr.FontHeight()).Add(pt)
		draw.Draw(t.Fr.RGBA(), r, t.Fr.Colors.Text, image.ZP, draw.Over)
	}
		// assuming the underlying selection is already
		// drawn on t.Img, see Sweep.
		for _, v := range t.Pen{
			if v == nil{
				continue
			}
			//r := image.Rect(0, 0, t.Fr.Bounds().Dx(), t.Fr.Bounds().Dy())
			//draw.Draw(t.Fr.RGBA(), r, v.Img, image.ZP, draw.Over)
				for _, r := range v.Rects(){
					draw.Draw(t.Fr.RGBA(), r, v.Img, r.Min, draw.Over)
				}
		}
	return nil
}

func (t *Tick) Insert(p []byte) (err error) {
	if t.P1 != t.P0 {
		t.Delete()
	}
	if len(p) == 0 {
		return nil
	}
	if err = t.Fr.Insert(p, t.P1); err != nil {
		return err
	}
	t.Open(t.P0+len(p))
	return nil
}

func (t *Tick) Delete() (err error) {
	if t.P0 == t.P1 && t.P0 == 0 {
		return nil
	}
	// Either act like the delete button or erase the
	// contents of an active selectiond
	if t.P0 == t.P1 {
		t.P0--
		t.Fr.Delete(t.P0, t.P1)
		t.P1--
	} else {
		if t.P0 > t.P1 {
			t.P0, t.P1 = t.P1, t.P0
		}
		t.Fr.Delete(t.P0, t.P1)
		t.P1 = t.P0
	}
	t.Open(t.P0)
	return nil
}

func (t *Tick) ReadByte() (byte){
	var b [1]byte
	t.Read(b[:])
	return b[0]
}
func (t *Tick) WriteRune(r rune) (err error) {
	return t.Insert([]byte{byte(r)})
}

func (t *Tick) ck() {
	if t.P1 < t.P0 {
		t.P0, t.P1 = t.P1, t.P0
	}
	nb := len(t.Fr.s)
	if t.P1 >= nb {
		t.P1 = nb - 1
	}
	if t.P0 >= nb {
		t.P0 = nb - 1
	}
	if t.P1 < 0 {
		t.P1 = 0
	}
	if t.P0 < 0 {
		t.P0 = 0
	}
}

func (t *Tick) Size() int {
	return abs(t.P0 - t.P1)
}

func (t *Tick) String() string {
	t.ck()
	println("p0, p1", t.P0, t.P1)
	return string(t.Fr.s[t.P0:t.P1])
}

func (t *Tick) Read(p []byte) (n int, err error) {
	if t.P0 == t.P1 {
		return 0, io.EOF
	}
	t.ck()
	q := t.Fr.s[t.P0:t.P1]
	return copy(p, q), nil
}

func (t *Tick) Write(p []byte) (n int, err error) {
	err = t.Insert(p)
	if err != nil {
		return 0, err
	}
	t.Fr.dirty = true
	return len(p), nil
}

func (t *Tick) Selected() bool{
	fmt.Printf("%v != %v\n", t.P0, t.P1)
	return t.P0 != t.P1
}

func (t *Tick) Next() {
	fmt.Printf("Next(): %#v\n", t.String())
	s := []byte(t.String())
	i := t.find(s, t.P1, len(t.Fr.s), false)
	if i == -1{
		i = t.find(s, 0, t.P1, false)
	}
	if i == -1{
		return
	}
	t.Open(i)
	t.Sweep(i+len(s))
	t.Commit()
}

func (t *Tick) Find(p []byte, back bool) int{
	return t.find(p, t.P1, len(t.Fr.s), back)	
}

func (t *Tick) find(p []byte, i, j int, back bool) int {
	if back {
		panic("unimplemented")
	}
	//fmt.Printf("debug: find: %q check frame[%d:]\n", p, t.P1)
	x := bytes.Index(t.Fr.s[i:j], p)
	if x == -1 {
		return -1
	}
	println("found at index", i, ":", x + i)
	return x + i

}

var Lefts = [...]byte{'(', '{', '[', '<', '"', '\'',  '`'}
var Rights = [...]byte{')', '}', ']', '>', '"', '\'', '`'}
var Free = [...]byte{'"', '\'', '`'}
var AlphaNum = []byte("*&!%-_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func isany(b byte, s []byte) bool{
	for _, v := range s{
		if b == v{
			return true
		}
	}
	return false
}
func (t *Tick) FindAlpha(i int) (int, int){
	j := i
	for ; i != 0 && isany(t.ReadByte(), AlphaNum); i-- {
		t.P0--
		t.P1--
	}
	t.Open(j)
	t.P1 = j+1
	for ; j != t.Fr.nbytes && isany(t.ReadByte(), AlphaNum); j++{
		t.P1++
		t.P0++
	}
	return i, j
}

func (t *Tick) FindSpecial(i int) (int, int){
	fmt.Println("NUMBER", i)
	if i == 0 {
		return i, t.FindOrEOF([]byte{'\n'})
	}
	t.Open(i-1)
	t.Sweep(i)
	t.Commit()
	if t.ReadByte() == '\n'{
		return i, t.FindOrEOF([]byte{'\n'})
	}
	if x := t.FindQuote(); x != -1{
		return i, x
	}
	if x := t.FindParity(); x != -1{
		return i, x
	}
	if isany(t.ReadByte(), AlphaNum){
		return t.FindAlpha(i)
	}
	return i, -1
}

func (t *Tick) FindOrEOF(p []byte) int{
	i := t.Find(p, false)
	if i == -1{
		return t.Fr.nbytes
	}
	return i
}

func (t *Tick) FindQuote() int{
	b := t.ReadByte()
	for _, v := range Free{
		if b != v{
			continue
		}
		return t.Find([]byte{v}, false)
	}
	return -1
}

func (t *Tick) FindParity() int{
	for i := range Lefts{
		j := t.findParity(Lefts[i], Rights[i], false)
		if j != -1{
			return j
		}
	}
	return -1
}

func (t *Tick) findParity(l byte, r byte, back bool) int{
	if back {
		panic("unimplemented")
	}
	b := t.ReadByte()
	if b != l{
		return -1
	}
	push := 1
	//j := -1
	for i, v := range t.Fr.s[t.P1:] {
		if v == l{
			println("\n\n++\n\n")
			push++
		}
		if v == r{
			println("\n\n--\n\n")
			push--
			if push == 0{
				return i+t.P1
			}
		}
	}
	return -1
}