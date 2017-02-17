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

	"github.com/as/frame"
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

func main() {
	fmt.Print()
	driver.Main(func(src screen.Screen) {
		var focused bool
		win, _ := src.NewWindow(&screen.NewWindowOptions{winSize.X, winSize.Y})
		tx, _ := src.NewTexture(winSize)
		buf, _ := src.NewBuffer(winSize)
		fr := frame.New(
				image.Pt(25, 25), 
				image.Pt(winSize.X, winSize.Y),
				win,
				&frame.Option{
					Colors: *frame.DefaultColors,
					Font: frame.ParseDefaultFont(20),
					Wrap: 40,
		})
		t := &frame.Tick{
			Fr: fr, 
			Select: frame.Select{
				Img: image.NewRGBA(fr.Bounds()),
			},
		}
		fr.Tick = t
		Clip = make([]byte, t.Size()+1)
		t.Write([]byte(""))
		go func(){
		for _, y := range []string{"super", "mink", "super", "mink", "super", "minky"}{
			v := []byte(y)
			t.Write(v)
			time.Sleep(time.Second/2)
		}
		}()
		for {
			switch e := win.NextEvent().(type) {
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
				
				/*
				log.Printf("frame.MenuEvent %#v\n", e)
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
				c := exec.Command(cmd, args...)
				out, _ := c.StdoutPipe()
				go func(out io.ReadCloser){
					defer out.Close()
					var buf [8192]byte
					for{
						n, err := out.Read(buf[:])
						if n > 0{
							t.Write(buf[:])
						}
						if err != nil{
							t.Write([]byte(fmt.Sprintln(err)))
							break
						}
					}
				}(out)
				if err := c.Start(); err != nil{
					log.Printf("cmd %q: %s\n", cmd, err)
				}
				c.Wait()
				*/
			case frame.SelectEvent:
				debugln("f.Mouse.OnSelect")
				pt := image.Pt(int(e.X), int(e.Y))
				t.Close()
				t.P0 = fr.IndexOf(pt)
				t.P1 = t.P0
				t.SelectAt(t.P0)
				fr.Mark()
				debugln("leave f.Mouse.OnSelect")
			case frame.SweepEvent:
				if fr.Menu.Visible(){
					continue
				}
				debugln("f.Mouse.OnSweep")
				pt := image.Pt(int(e.X), int(e.Y))
				t.P1 = fr.IndexOf(pt)
				fmt.Println("p0/p1", t.P0, t.P1)
				fr.Mark()
				debugln("leave f.Mouse.OnSweep")
			case frame.SnarfEvent:
				debugln("f.Mouse.OnSnarf")
				fmt.Println("Clip size is", len(Clip))
				t.Read(Clip[:])
				t.Delete()
				t.P1 = t.P0
				t.SelectAt(t.P0)
				fr.Mark()
				debugln("leave f.Mouse.OnSnarf")
			case frame.InsertEvent:
				if fr.Menu.Visible(){
					continue
				}
				fmt.Println("f.Mouse.OnInsert")
				fmt.Println("write %s\n", Clip)
				t.Write(Clip[:])
				t.P0 -= len(Clip)
				t.SelectAt(t.P1)
				fr.Mark()
				fmt.Println("leave f.Mouse.OnInsert")
			case key.Event:
				if e.Code == key.CodeReturnEnter {
					t.Write([]byte{'\n'})
				} else {
					fr.Handle(e)
				}
				win.Send(paint.Event{})
			case mouse.Event:
				fr.Handle(e)
				win.Send(paint.Event{})
			case size.Event, paint.Event:
				if fr.Dirty() || true{
					fr.Draw()
					draw.Draw(buf.RGBA(), buf.Bounds(), fr.RGBA(), image.ZP, draw.Src)
					tx.Upload(image.ZP, buf, buf.Bounds())
					win.Copy(buf.Bounds().Min, tx, tx.Bounds(), screen.Over, nil)
				} else if  !focused{
					win.Copy(buf.Bounds().Min, tx, tx.Bounds(), screen.Over, nil)
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
			}
		}
	})
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
