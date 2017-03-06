package frame

import "io"
import "fmt"

type Box struct {
	data  []byte
	width int
}

func (b *Box) Len() int {
	return len(b.data)
}

func (b *Box) Cap() int {
	return cap(b.data)
}

func (b *Box) Width() int {
	return b.width
}

type Boxes struct {
	Box     []*Box
	measure func([]byte) int
}

func NewBoxes(measure func([]byte) int) *Boxes {
	b := &Boxes{measure: measure}
	b.Add(25)
	return b
}

func (b *Boxes) WriteAt(p []byte, off int64) (n int, err error) {
	b.Insert(p, int(off))
	return len(p), nil
}

func (b *Boxes) ReadAt(p []byte, off int64) (n int, err error) {
	bn, err := b.Find(0, 0, int(off))
	fmt.Println("offset %d found in box %d", off, bn)
	if err != nil {
		return n, err
	}
	bp := b.Box[bn]
	for need := len(p); need > 0; {
		m := copy(p[n:], bp.data[:min(len(bp.data), need)])
		n += m
		need -= m
		bn++
		if bn >= len(b.Box) {
			return n, io.ErrUnexpectedEOF
		}
		bp = b.Box[bn]
	}
	return n, nil
}

func (b *Boxes) Insert(p []byte, off int) {
	n, err := b.Find(0, 0, off)
	if err != nil {
		panic("insert")
	}
	if n >= len(b.Box) {
		n = len(b.Box) - 1
		b.Add(1)
	}
	bp := b.Box[n]
	bp.data = append(bp.data, p...)
	bp.width = b.measure(bp.data)
}

// Find starts at box n, assuming offset i, and
// advances to offset j. It returns the box number
// containing offset j, guaraneed to align on a box
// boundary.
func (b *Boxes) Find(n, i, j int) (int, error) {
	fmt.Printf("Boxes.Find: n=%d i=%d j=%d\n", n, i, j)
	for _, v := range b.Box {
		w := len(v.data)
		if i+w > j {
			break
		}
		i += w
	}
	if i < j {
		fmt.Printf("Boxes.Find: split %d,%d\n", n+1, j-i)
		b.Split(n+1, j-i)
	}
	fmt.Printf("Boxes.Find: ret %d\n", n+1)
	return n + 1, nil
}

func (b *Boxes) Split(n int, at int) {
	fmt.Printf("Boxes.Split: n=%d at=%d\n", n, at)
	b.Dup(n)
	//b.Truncate(n, len(box.data)-at+1)
	b.Truncate(n, at)
	b.Chop(n+1, at)
}

func (b *Boxes) Add(n int) {
	boxes := make([]*Box, n)
	for i := range boxes {
		boxes[i] = &Box{data: make([]byte, 0, 25)}
	}
	b.Box = append(b.Box, boxes...)
}

func (b *Boxes) Dup(n int) *Box {
	bp0 := b.Box[n]
	bp1 := &Box{
		width: bp0.width,
		data:  append([]byte{}, bp0.data...),
	}
	b.Box = append(b.Box[:n+1], append([]*Box{bp1}, b.Box[n+1:]...)...)
	//fmt.Printf("%#v\n", b.Box)
	return bp1
}

// Truncate discards all but the first n bytes of box bn
func (b *Boxes) Truncate(bn, n int) {
	box := b.Box[bn]
	box.data = box.data[:n]
	box.width = b.measure(box.data)
}

// Chop discards the first n bytes of box bn
func (b *Boxes) Chop(bn, n int) {
	box := b.Box[bn]
	copy(box.data, box.data[n:])
	box.data = box.data[:len(box.data)-n]
	box.width = b.measure(box.data)
}

func (b *Boxes) Merge(n int) {
	sp := b.Box[n:]
	sp[0].data = append(sp[0].data, sp[1].data...)
	sp[0].width += sp[1].width
}

func (b *Boxes) Delete(n0, n1 int) {
	dn := len(b.Box[n1:]) - len(b.Box[n0:])
	copy(b.Box[n0:], b.Box[n1:])
	b.Box = b.Box[:len(b.Box)-dn]
}

func (b *Boxes) Dump() {
	fmt.Println("dumping boxes")
	for bn, b := range b.Box {
		fmt.Printf("%4d: %q\n", bn, b.Bytes())
	}
}

func (b Box) Bytes() []byte {
	return b.data
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
