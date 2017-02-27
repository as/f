package main

import (
	//	"github.com/as/clip"
	//"golang.org/x/image/font"
	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"image"
	"fmt"
	"image/draw"
	"os/exec"
	"io"
	"log"
	"strings"
	"time"
	"os"
	"sync"

	"github.com/as/frame"
	"github.com/as/cursor"
	window "github.com/as/ms/win"
)

var winSize = image.Pt(550, 1080)

var Clip []byte

var Debug = false
func debugf(fm string, i ...interface{}){
	if Debug{
		log.Printf(fm, i...)
	}
}
func debugln(i ...interface{}){
	if Debug{
		log.Println(i...)
	}
}

func moveMouse(pt image.Point){
	cursor.MoveTo(window.ClientAbs().Min.Add(pt))
}

func main() {
	fmt.Print()
	driver.Main(func(src screen.Screen) {
		var focused, resized bool
		win, _ := src.NewWindow(&screen.NewWindowOptions{winSize.X, winSize.Y})
		tx, _ := src.NewTexture(winSize)
		buf, _ := src.NewBuffer(winSize)
		fr := frame.New(
				image.Pt(25, 25), 
				image.Pt(winSize.X, winSize.Y),
				win,
				&frame.Option{
					Colors: *frame.DefaultColors,
					Font: frame.NewFont(frame.ParseDefaultFont(52)),
					Wrap: 100,
		})
		t := frame.NewTick(fr)
		fr.Tick = t
		Clip = make([]byte, t.Size()+1)
		if len(os.Args) > 1{
			fd, err := os.Open(os.Args[1])
			defer fd.Close()
			if err != nil{
				log.Fatalln(err)
			}
			io.Copy(t, fd)
		}
		
		for {
			switch e := win.NextEvent().(type) {
			case frame.MarkEvent:
				fmt.Printf("frame.MarkEvent: %#v\n", e)
				pt := image.Pt(int(e.X), int(e.Y))
				i := fr.IndexOf(pt)
				switch e.Button{
				case 1:
					t.Open(i)
					fr.Mark()
					win.Send(paint.Event{})
				case 3:
					t.Pen[2].Open(i)
				}
			//	t.SelectAt(fr.IndexOf(pt))
			case frame.SweepEvent:
				debugln("f.Mouse.OnSweep")
				if fr.Menu.Visible(){
					continue
				}
				pt := image.Pt(int(e.X), int(e.Y))
				i := fr.IndexOf(pt)
				switch e.Button{
				case 3:
					t.Pen[2].Sweep(i)
					fr.Mark()
					win.Send(paint.Event{})
				case 1:
					t.Sweep(i)
					fr.Mark()
					debugln("leave f.Mouse.OnSweep")
				}
			case frame.ClickEvent:
				fmt.Printf("frame.ClickEvent: %#v\n", e)
				pt := image.Pt(int(e.X), int(e.Y))
				i := fr.IndexOf(pt)
				switch e.Button{
				case 1:
					if e.Double{
						if i, j := t.FindSpecial(i); j != -1{
							t.Open(i)
							t.Sweep(j)
							t.Commit()
						} else {
							t.Open(i)
						}
					}
				case 2:
					if t.In(pt){
						
						fr.Mark()
					}
				case 3:
					if !t.In(pt){
						if i, j := t.FindSpecial(i); j != -1{
							t.Open(i)
							t.Sweep(j)
							t.Commit()
						} else {
							t.Open(i)
						}					
					}
						t.Next()
						moveMouse(t.Pen[0].Sp())
						fr.Mark()
				}
			case frame.SelectEvent:
				debugln("f.Mouse.OnSelect")
				fmt.Printf("event information: %#v\n", e)
				pt := image.Pt(int(e.X), int(e.Y))
				i := fr.IndexOf(pt)
				switch e.Button {
				case 3:
					h, _ := t.Pen[2].Addr()
					t.Open(h)
					t.Sweep(i)
					t.Commit()
					t.Next()
					moveMouse(t.Pen[0].Sp())
					fr.Mark()
				default:
					t.Sweep(i)
					t.Commit()
					fr.Mark()
				}
				debugln("leave f.Mouse.OnSelect")
			case frame.SnarfEvent:
				debugln("f.Mouse.OnSnarf")
				fmt.Println("Clip size is", len(Clip))
				if len(Clip) > 0{
					fmt.Println("t.Read(Clip)")
					t.Read(Clip)
					fmt.Println("t.Delete")
					t.Delete()
					t.P1 = t.P0
					t.Open(t.P0)
					fr.Mark()
				}
				debugln("leave f.Mouse.OnSnarf")
			case frame.InsertEvent:
				if fr.Menu.Visible(){
					continue
				}
				fmt.Println("f.Mouse.OnInsert")
				fmt.Println("write %s\n", Clip)
				t.Write(Clip[:])
				t.P0 -= len(Clip)
				t.Open(t.P1)
				fr.Mark()
				fmt.Println("leave f.Mouse.OnInsert")
			case key.Event:
					fr.Handle(e)
				win.Send(paint.Event{})
			case mouse.Event:
				fr.Handle(e)
				win.Send(paint.Event{})
			case size.Event:
				winSize = image.Pt(e.WidthPx, e.HeightPx)
				wg.Add(2)
				resized = true
				fr.Mark()
				go func(){ tx, _ = src.NewTexture(winSize); wg.Done();}()
				go func(){  buf, _ = src.NewBuffer(winSize); wg.Done();}()
				wg.Wait()
				fr.Resize(winSize)
			case paint.Event:
				if fr.Dirty() || true{
					 fr.Draw(false)
					draw.Draw(buf.RGBA(), buf.Bounds(), fr.RGBA(), image.ZP, draw.Src)
					tx.Upload(image.ZP, buf, buf.Bounds())
					win.Copy(buf.Bounds().Min, tx, tx.Bounds(), screen.Over, nil)
				} else if !focused || resized{
					fr.Draw(true)
				//draw.Draw(buf.RGBA(), buf.Bounds(), fr.RGBA(), image.ZP, draw.Src)
						var wg sync.WaitGroup
						pieces := 4
						x, y := 0, 0
						dx := buf.Bounds().Max.X/pieces
						dy := buf.Bounds().Max.Y
						wg.Add(pieces)
						for i := 0; i < pieces; i++{
							go func(r image.Rectangle){
			draw.Draw(buf.RGBA(), r, fr.RGBA(), r.Min, draw.Src); wg.Done()
					 		}(image.Rect(x,y,x+dx,dy))
					 		x += dx
					 	}
					 	wg.Wait()
						
						x, y = 0, 0
						wg.Add(pieces)
						for i := 0; i < pieces; i++{
							go func(r image.Rectangle){
					 			tx.Upload(r.Min, buf, r); wg.Done()
					 		}(image.Rect(x,y,x+dx,dy))
					 		x += dx
					 	}
					 	wg.Wait()
					 	//win.Copy(buf.Bounds().Min, tx, buf.Bounds(), screen.Over, nil)

						x, y = 0, 0
						wg.Add(pieces)
						for i := 0; i < pieces; i++{
							go func(r image.Rectangle){
								win.Copy(r.Min, tx, r, screen.Over, nil); wg.Done()
					 		}(image.Rect(x,y,x+dx,dy))
					 		x += dx
					 	}
					 	wg.Wait()					 	
					if resized {
						resized = false
					}
				}
				win.Publish()
			case lifecycle.Event:
				if e.To == lifecycle.StageDead {
					return
				}
				// NT doesn't repaint the window if another window covers it
				if e.Crosses(lifecycle.StageFocused) == lifecycle.CrossOff{
					focused = false
				} else if e.Crosses(lifecycle.StageFocused) == lifecycle.CrossOn{
					focused = true
				}
			case frame.MenuEvent:
				debugf("frame.MenuEvent %#v\n", e)
				var(
					cmd string
					args []string
				)
				s := e.Strings
				if len(s) == 0{
					continue
				}
				cmd = s[0]
				if len(s) > 1{
					args = s[1:]
				}
				
				if cmd == "Send" {
					log.Println(t.String())
					args = strings.Fields(t.String())
					if len(args) > 0{
						cmd = args[0]
						if len(args) > 1 {
							args = args[1:]
						}
					} 
				}
				
				c := exec.Command(cmd, args...)
				out, _ := c.StdoutPipe()
				go func(out io.ReadCloser){
					if err := c.Start(); err != nil{
						log.Printf("cmd %q: %s\n", cmd, err)
					}
					defer out.Close()
					var buf [8192]byte
					for{
						n, err := out.Read(buf[:])
						if n > 0{
							t.Write(buf[:n])
						}
						if err != nil{
							t.Write([]byte(fmt.Sprintln(err)))
							break
						}
					}
					c.Wait()
				}(out)
			}
		}
	})
}

var wg sync.WaitGroup

var timer *time.Timer

func init(){
	
}

/*
		measure := func(p []byte) int{
			x := font.MeasureBytes(fr.Font, p)
			return int(x >> 6)
		}
	boxes := frame.NewBoxes(measure)
	for i := 0; i < 5; i++{
		test := make([]byte, 20)
		boxes.WriteAt([]byte(fmt.Sprintf("%d", i)), int64(i))
		boxes.ReadAt(test, int64(i))
		fmt.Printf("%q\n",test)
	}
	for i := 0; i < 5; i++{
		test := make([]byte, 20)
		boxes.ReadAt(test, int64(i))
		fmt.Printf("%q\n",test)
	}
type Fd struct {
	inc  chan []byte
	outc chan []byte
	errc chan []byte
}

func Run(fd *Fd, s string, args ...string) {
	cmd := exec.Command(s, args...)
	in, _ := cmd.StdinPipe()
	out, _ := cmd.StdoutPipe()
	e, _ := cmd.StderrPipe()
	killc := make(chan bool)
	go func() {
		for {
			select{
			case <- killc:
				return
			case v := <- fd.inc:
				_, err := in.Write(v)
				if err != nil {
					return
				}
			}
		}
	}()
	go func() {
		p := make([]byte, 8192)
		for {
			select {
			case <-killc:
				return
			default:
				n, err := out.Read(p)
				if n > 0 {
					fd.outc <- append([]byte{}, p[:n]...)
				}
				if err != nil {
					return
				}
			}
		}
	}()
	go func() {
		p := make([]byte, 8192)
		for {
			select {
			case <-killc:
				return
			default:
				n, err := e.Read(p)
				if n > 0 {
					fd.errc <- append([]byte{}, p[:n]...)
				}
				if err != nil {
					return
				}
			}
		}
	}()
	
	err := cmd.Start()
	if err != nil {
		fd.errc <- []byte(fmt.Sprintf("%s\n", err))
	    close(killc)
		return
	}
	cmd.Wait()
	close(killc)
}

// in main body


		con := make(chan []byte)
		conin := make(chan []byte)

		go func(){
			for p := range con{
				tick.Write(p)
				win.Send(paint.Event{})			
			}
		}()
		
		go func() {
			for {
				select {
				case p := <-conin:
					s := strings.Fields(string(p))
					fmt.Println()
					fmt.Println("conin:", s)
					if len(s) == 1 && s[0] == "clear" || s[0] == "cls" {
						tick.P0 = 0
						tick.P1 = len(tick.Fr.Bytes())
						if tick.P1 >= 0{
							tick.Delete()
						}
					} else if len(s) > 0 {
						var args []string
						if len(s) > 1 {
							args = s[1:]
						}
						tick.P0 = tick.P1
						con <- []byte{'\n'}
						Run(&Fd{inc: conin, outc: con, errc: con}, s[0], args...)
					}
					con <- []byte{';'}
				}
			}
		}()
*/
