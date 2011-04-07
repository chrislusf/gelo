package main

import (
	"os"
	"math"
	"time"

	"image"
	"exp/draw"
	"exp/draw/x11"

	"gelo"
	"gelo/commands"
	"gelo/extensions"
)

func check(failmsg string, e os.Error) {
	if e != nil {
		println(failmsg)
		println(e.String())
		os.Exit(1)
	}
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func line(img draw.Image, x0, y0, x1, y1 int) {
	//generalized Bresenham's algorithm from wikipedia
	steep := abs(y1-y0) > abs(x1-x0)
	if steep {
		x0, y0 = y0, x0
		x1, y1 = y1, x1
	}
	if x0 > x1 {
		x0, x1 = x1, x0
		y0, y1 = y1, y0
	}
	dx, dy := x1-x0, abs(y1-y0)
	err, derr := 0.0, float64(dy)/float64(dx)
	y, ystep := y0, 1
	if y1 <= y0 {
		ystep = -1
	}
	B := img.Bounds()
	for x := x0; x <= x1; x++ {
		if x >= B.Min.X && x < B.Max.X && y >= B.Min.Y && y < B.Max.Y {
			if steep {
				img.Set(y, x, image.Black)
			} else {
				img.Set(x, y, image.Black)
			}
		}
		err += derr
		if err >= .5 {
			y += ystep
			err -= 1.0
		}
	}
}

const (
	reset byte = iota
	clear
	up
	down
	rotate
	forward
	getx
	gety
	getang
	ispenup
)

type gcom struct {
	name  byte
	value float64
}

var gchan chan *gcom

func graphics_server(ctx draw.Window) {
	image := ctx.Screen()
	//we foolishly assume that the window will never be resized
	var (
		w   int = image.Bounds().Max.X
		h   = image.Bounds().Max.Y
		x0  = w / 2
		y0  = h / 2
		x   = x0
		y   = y0
		ang float64 = 0
		pen bool    = true
	)
	for {
		switch cmd := <-gchan; cmd.name {
		case reset:
			x, y, ang = x0, y0, 0
		case clear:
			image = clear_screen(ctx)
		case up:
			pen = false
		case down:
			pen = true
		case rotate:
			ang = math.Fmod(cmd.value+ang, 360)
		case forward:
			sx, sy := x, y
			sin, cos := math.Sincos(ang * math.Pi / 180)
			x += int(cmd.value * cos)
			y += int(cmd.value * sin)
			if pen {
				line(image, sx, sy, x, y)
			}
		case getx:
			gchan <- &gcom{getx, float64(x - x0)}
		case gety:
			gchan <- &gcom{gety, float64(y - y0)}
		case getang:
			gchan <- &gcom{getang, ang}
		case ispenup:
			var up float64
			if !pen {
				up = 1
			}
			gchan <- &gcom{ispenup, up}
		}
	}
}

func clear_screen(ctx draw.Window) draw.Image {
	s := ctx.Screen()
	B := s.Bounds().Max
	draw.Draw(s, image.Rect(0, 0, B.X, B.Y), image.White, image.ZP)
	ctx.FlushImage()
	return s
}


func flusher(ctx draw.Window) {
	t := time.NewTicker(1e9 / 50)
	for {
		<-t.C
		ctx.FlushImage()
	}
}

//machines to build the required commands

func Nullary(name byte) func(*gelo.VM, *gelo.List, uint) gelo.Word {
	cmd := &gcom{name, 0}
	return func(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
		if ac != 0 {
			gelo.ArgumentError(vm, "turtle", "", args)
		}
		gchan <- cmd
		return gelo.Null
	}
}

func Unary(name byte) func(*gelo.VM, *gelo.List, uint) gelo.Word {
	return func(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
		if ac != 1 {
			gelo.ArgumentError(vm, "turtle", "number", args)
		}
		n := vm.API.NumberOrElse(args.Value)
		gchan <- &gcom{name, n.Real()}
		return n
	}
}

func Get(which byte) func(*gelo.VM, *gelo.List, uint) gelo.Word {
	cmd := &gcom{which, 0}
	return func(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
		if ac != 0 {
			gelo.ArgumentError(vm, "turtle", "", args)
		}
		gchan <- cmd
		return gelo.NewNumber((<-gchan).value)
	}
}

func main() {
	if len(os.Args) < 2 {
		println("No script specified")
		os.Exit(1)
	}

	file, err := os.Open(os.Args[1], os.O_RDONLY, 0664)
	check("Could not open script", err)

	vm := gelo.NewVM(extensions.Stdio)
	defer vm.Destroy()

	gelo.SetTracer(extensions.Stderr)

	vm.RegisterBundle(gelo.Core)
	vm.RegisterBundles(commands.All)

	prelude, err := os.Open("prelude.gel", os.O_RDONLY, 0664)
	defer prelude.Close()
	check("Could not open prelude", err)

	_, err = vm.Run(prelude, nil)
	check("Could not load prelude", err)

	context, err := x11.NewWindow()
	check("Could not create window", err)

	vm.Register("W", int(context.Screen().Bounds().Max.X/2))
	vm.Register("H", int(context.Screen().Bounds().Max.Y/2))
	vm.Register("reset", Nullary(reset))
	vm.Register("clear", Nullary(clear))
	vm.Register("up", Nullary(up))
	vm.Register("down", Nullary(down))
	vm.Register("rotate", Unary(rotate))
	vm.Register("forward", Unary(forward))
	vm.Register("get-x", Get(getx))
	vm.Register("get-y", Get(gety))
	vm.Register("angle", Get(getang))
	vm.Register("pen-up?",
		func(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
			if ac != 0 {
				gelo.ArgumentError(vm, "turtle", "", args)
			}
			gchan <- &gcom{ispenup, 0}
			v := (<-gchan).value
			return gelo.ToBool(v == 1)
		},
	)

	turtle_prelude, err := os.Open("turtle.prelude.gel", os.O_RDONLY, 0664)
	defer turtle_prelude.Close()
	check("Could not open turtle prelude", err)

	_, err = vm.Run(turtle_prelude, nil)
	check("Could not load turtle prelude", err)

	clear_screen(context)

	gchan = make(chan *gcom)

	go flusher(context)
	go graphics_server(context)

	_, err = vm.Run(file, os.Args[2:])
	check("===ERROR===", err)

	E := context.EventChan()
	for {
		e := <-E
		//for some reason I always get a KeyEvent with this mysterious
		//magic number
		if k, ok := e.(draw.KeyEvent); ok && k.Key != -65293 {
			break
		}
	}
}
