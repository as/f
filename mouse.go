package frame
// The chord implementation kinda sucks
// make it so that 'if one is held down and two is pressed'
// instead of tying to pack everything into a bit vector

import (
	"golang.org/x/mobile/event/mouse"
	"time"
	"image"
	"fmt"
)

type Sender interface{
	Send(i interface{})
	SendFirst(i interface{})
}

func NewMouse(delay time.Duration, events Sender, f *Frame) *Mouse{
	m := &Mouse{
		Last: []Click{Click{}, Click{}},
		doubled: delay,
		Machine: NewMachine(events, f),
	}
	m.Sink = m.Machine.Run()
	return m
}

type Chord struct{
	Start int
	Seq int
	Step int
}
	
type Click struct{
	Button mouse.Button
	At image.Point
	Time time.Time
}

type Mouse struct{
	Chord Chord
	Last []Click
	Down mouse.Button
	At image.Point
	
	doubled time.Duration
	last time.Time
	
	*Machine
}

// State is the state of the machine
type State int

const(
	StateNone State = iota
	StateSelect
	StateSweep
	StateSnarf
	StateInsert
	StateCommit
)

// StateFn is a state function that expresses a state
// transition. All StateFns return the next state
// as a transitionary StateFn
type StateFn func(*Machine, mouse.Event) StateFn

// Action executes a procedure on the event of
// a specific state transition
type Action    func(mouse.Event)

type SelectEvent struct{
	mouse.Event
}
type SweepEvent struct{
	mouse.Event
}
type SnarfEvent struct{
	mouse.Event
}
type InsertEvent struct{
	mouse.Event
}
type CommitEvent struct{
	mouse.Event
}

// Machine is the conduit that state transitions happen
// though. It contains a Skink chan for input mouse events
// that drive the StateFns
type Machine struct{
	Sink chan mouse.Event
	
	f *Frame
	
	// Should only send events, no recieving.
	Sender
}

// NewMachine initialize a new state machine with no-op
// functions for all chording events.
func NewMachine(deque Sender, f *Frame) *Machine{
	return &Machine{
		Sink: make(chan mouse.Event),
		f: f,
		Sender: deque,
	}
}


func (m *Machine) Run() chan mouse.Event {
	go func(){
		fn := none
		for e := range m.Sink{
			fn = fn(m, e)
		}
	}()
	return m.Sink
}
func none(m *Machine, e mouse.Event) StateFn {
	if e.Direction == mouse.DirPress || e.Button == mouse.ButtonLeft {
		return selecting(m, e)
	}
	return none
}

func terminus(e mouse.Event) bool{
	return e.Direction == mouse.DirRelease && e.Button == mouse.ButtonLeft
}

func selecting(m *Machine, e mouse.Event) StateFn {
	if e.Direction == mouse.DirPress && e.Button == mouse.ButtonLeft {
		m.f.selecting = true
		m.Send(SelectEvent{Event: e})
		return selecting
	} else if terminus(e) {
		return commit(m, e)
	}
	return sweeping
}
func sweeping(m *Machine, e mouse.Event) StateFn{
	if e.Direction == mouse.DirNone{
		
		m.Send(SweepEvent{Event: e})
		return sweeping
	}
	if terminus(e) {
		return none
	}
	if e.Direction == mouse.DirPress {
		switch e.Button{
		case mouse.ButtonMiddle:
			return snarfing(m, e)
		case mouse.ButtonRight:
			return inserting(m, e)
		}
	}	
	return sweeping
}
func snarfing(m *Machine, e mouse.Event) StateFn {
	if e.Direction == mouse.DirNone{
		return snarfing
	}
	if e.Direction == mouse.DirPress {
		if e.Button == mouse.ButtonMiddle {
			m.f.selecting = false
			m.Send(SnarfEvent{Event: e})
			return snarfing
		}
		if e.Button == mouse.ButtonRight {
			return inserting(m, e)
		}
	} else if terminus(e){
		return commit(m, e)
	}
	return snarfing
}
func inserting(m *Machine, e mouse.Event) StateFn {
	if e.Direction == mouse.DirNone{
		return inserting
	}
	if e.Direction == mouse.DirPress {
		if e.Button == mouse.ButtonMiddle {
			return snarfing(m, e)
		}
		if e.Button == mouse.ButtonRight {
			m.f.selecting = false
			m.Send(InsertEvent{Event: e})
			return inserting
		}
	} else if terminus(e){
		return commit(m, e)
	}
	return inserting
}
func commit(m *Machine, e mouse.Event) StateFn {
	m.Send(CommitEvent{Event: e})
	return none
}

func (m *Mouse) Process(e mouse.Event){
	m.Sink <- e
	return
	if e.Direction == mouse.DirNone && e.Button == mouse.ButtonNone {
		return
	}
	c := Click{
		Button: m.Down,
		At: image.Pt(int(e.X), int(e.Y)),
		Time: time.Now(),
	}
	if e.Direction == mouse.DirPress{
		c.Button = e.Button
		if m.Chord.Seq == 0{
			m.Chord.Start = int(e.Button)
		}
		m.Chord.Seq = int(m.Chord.Seq << 8 | int(e.Button))
		m.Chord.Step++
		if m.Chord.Step > 4{
			m.Chord.Start = 0
			m.Chord.Seq = 0
			m.Chord.Step = 0
		}
		fmt.Printf("Chord Sequence %x\n", m.Chord.Seq)
	} else if e.Direction == mouse.DirRelease{
		c.Button = mouse.ButtonNone
		if int(e.Button) != m.Chord.Start || m.Chord.Step < 2 {
			m.Chord.Seq = 0
		}
	} else if e.Direction == mouse.DirNone{
		c.Button = m.Last[0].Button
	}
	m.Down = c.Button
	if e.Direction == mouse.DirPress && e.Button != mouse.ButtonNone{
		m.Last = append([]Click{c}, m.Last...)
	}
}

func (m *Mouse) Pt() image.Point{
	return m.At
}

// Double returns true if and only if the previous
// event is part of a double click
func (m *Mouse) Double() bool{
	a, b := m.Last[0], m.Last[1]
	if a.Button == mouse.ButtonNone{
		return false
	}
	if a.Button != b.Button{
		return false
	}
	if m.Last[0].Time == m.last{
		return false
	}
	if m.Last[1].Time == m.last{
		return false
	}
	if a.Time.Sub(b.Time) <= m.doubled {
		m.last = a.Time
		return true
	}
	return false
}