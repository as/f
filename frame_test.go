package frame

import (
	"testing"
)

func newBoxesFixed() *Boxes {
	return NewBoxes(func(b []byte) int {
		return len(b) * 4
	})
}

func TestDup(t *testing.T) {
	b := newBoxesFixed()
	want := "mink"
	b.Insert([]byte(want), 0)
	N := 2
	for i := 1; i < N+1; i++ {
		b.Dup(1)
	}
	for i := 1; i < N+1; i++ {
		if have := string(b.Box[i].data); have != want {
			t.Logf("box #%d: want %q have %q\n", i, want, have)
			t.FailNow()
		}
	}
}
func TestTruncate(t *testing.T) {
	b := newBoxesFixed()
	ck := func(bn int, want string) {
		if have := string(b.Box[bn].data); have != want {
			t.Logf("box #%d: want %q have %q\n", bn, want, have)
			t.FailNow()
		}
	}
	b.Insert([]byte("supermink"), 0)
	b.Truncate(1, 1)
	ck(1, "s")
}

func TestChop(t *testing.T) {
	b := newBoxesFixed()
	ck := func(bn int, want string) {
		if have := string(b.Box[bn].data); have != want {
			t.Logf("box #%d: want %q have %q\n", bn, want, have)
			t.FailNow()
		}
	}
	b.Insert([]byte("supermink"), 0)
	b.Chop(1, 1)
	ck(1, "upermink")
}

func TestSplit(t *testing.T) {
	b := newBoxesFixed()
	ck := func(bn int, want string) {
		if have := string(b.Box[bn].data); have != want {
			t.Logf("box #%d: want %q have %q\n", bn, want, have)
			t.FailNow()
		}
	}
	b.Insert([]byte("supermink"), 0)
	b.Split(1, len("super"))
	ck(1, "super")
	ck(2, "mink")
	b.Split(1, len("sup"))
	ck(1, "sup")
	ck(2, "er")
	ck(3, "mink")
	b.Split(3, len("mi"))
	ck(1, "sup")
	ck(2, "er")
	ck(3, "mi")
	ck(4, "nk")
}

func TestFind(t *testing.T) {
	b := newBoxesFixed()
	b.Insert([]byte("mink"), 0)
	want := 1
	have, _ := b.Find(0, 0, 0)
	if have != want {
		t.Logf("want %v have %v\n", want, have)
		t.Fail()
	}
}

func TestMerge(t *testing.T) {
	b := newBoxesFixed()
	ck := func(bn int, want string) {
		if have := string(b.Box[bn].data); have != want {
			t.Logf("box #%d: want %q have %q\n", bn, want, have)
			t.FailNow()
		}
	}
	b.Insert([]byte("super"), 0)
	ck(1, "super")
	b.Add(1)
	b.Box[2].data = []byte("mink")
	b.Box[2].width = 4 * 4
	ck(2, "mink")

	b.Merge(1)
	ck(1, "supermink")
}

func TestBoxInsert00(t *testing.T) {
	b := newBoxesFixed()
	ck := func(bn int, want string) {
		if have := string(b.Box[bn].data); have != want {
			t.Logf("box #%d: want %q have %q\n", bn, want, have)
			t.FailNow()
		}
	}
	b.Insert([]byte("mink"), 0)
	ck(1, "mink")
}
func TestBoxInsert01(t *testing.T) {
	b := newBoxesFixed()
	ck := func(bn int, want string) {
		if have := string(b.Box[bn].data); have != want {
			t.Logf("box #%d: want %q have %q\n", bn, want, have)
			t.FailNow()
		}
	}
	b.Insert([]byte("mink"), 0)
	ck(1, "mink")
	b.Insert([]byte("mink"), 1)
	ck(1, "mmink")
	ck(2, "ink")
}

func TestBoxWidth(t *testing.T) {
	b := newBoxesFixed()
	ck := func(bn int, want int) {
		if have := b.Box[bn].Width(); have != want {
			t.Logf("box #%d: width: want %d have %d\n", bn, want, have)
			t.FailNow()
		}
	}
	b.Insert([]byte("mink"), 0)
	ck(1, 4*4)
	b.Insert([]byte("nk or"), 2)
	ck(1, len("mink or")*4)
}
