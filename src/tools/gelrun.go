package main

import (
	"os"
	"flag"
	"log"
	"io"
	"gelo"
	"gelo/commands"
	"gelo/extensions"
)

type LiterateReader struct {
	code, nl, first bool
	start           int
	src             io.Reader
	err             os.Error
	scratch         []byte
}

//      A proxy reader that filters its source to support literate programming
// in the same limited manner as the Haskell .lhs format. Namely, every line of
// text in the source reader is discarded unless it's first character is the
// literal, >, which is also dropped.
func NewLiterateReader(src io.Reader) *LiterateReader {
	return &LiterateReader{nl: true, first: true, src: src,
		scratch: make([]byte, 128)}
}

func (lr *LiterateReader) Read(p []byte) (n int, err os.Error) {
	//input has been exhausted
	if lr.scratch == nil {
		return 0, lr.err
	}

	//requesting more input than scratch can hold
	if lr.err != nil && len(lr.scratch) < len(p) {
		newscr := make([]byte, len(lr.scratch), len(p))
		copy(newscr, lr.scratch)
		lr.scratch = newscr
	}

outer:
	for {
		//need to fill scratch
		if lr.first || lr.start == len(lr.scratch) {
			m := 0
			if m, lr.err = lr.src.Read(lr.scratch); m == 0 {
				//been sucked dry
				lr.scratch = nil
				break
			}
			//in case m < len(lr.scratch)
			lr.scratch = lr.scratch[:m]
			lr.start, lr.first = 0, false
		}
		//push out what we can from scratch
		for _, c := range lr.scratch[lr.start:] {
			if n == len(p)-1 {
				//filled p
				break outer
			}
			lr.start++
			//last was nl, select mode
			if lr.nl {
				lr.code, lr.nl = c == '>', false
			} else {
				if lr.code {
					//write
					p[n] = c
					n++
				}
				lr.nl = c == '\n'
			}
		}
	}

	return n, lr.err
}

var trace = flag.Bool("trace", false, "turn on all traces")
var logit = flag.Bool("log", false, "log traces (does not activate traces)")
var lit = flag.Bool("literate", false, "force reading in literate mode")
var no_prelude = flag.Bool("no-prelude", false, "do not load prelude.gel")

func check(failmsg string, e os.Error) {
	if e != nil {
		println(failmsg)
		println(e.String())
		os.Exit(1)
	}
}

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		println("No input file to process")
		os.Exit(1)
	}

	file_name := flag.Arg(0)

	vm := gelo.NewVM(extensions.Stdio)
	defer vm.Destroy()

	gelo.SetTracer(extensions.Stderr)

	vm.RegisterBundle(gelo.Core)
	vm.RegisterBundles(commands.All)

	if !*no_prelude {
		prelude, err := os.Open("prelude.gel", os.O_RDONLY, 0664)
		defer prelude.Close()
		check("Could not open prelude.gel", err)

		_, err = vm.Run(prelude, nil)
		check("Could not load prelude", err)
	}

	file, err := os.Open(file_name, os.O_RDONLY, 0664)
	defer file.Close()
	check("Could not open: "+file_name, err)
	reader := io.Reader(file)

	if *lit || file_name[len(file_name)-3:] == "lit" {
		reader = NewLiterateReader(reader)
		t := make([]byte, 64)
		for {
			n, err := reader.Read(t)
			println(string(t), "n:", n, "err", err == nil)
		}
	}

	if *logit {
		out, err := os.Open(flag.Arg(0)+".log", os.O_WRONLY|os.O_CREATE,
			0664)
		defer out.Close()
		check("Could not create log file", err)
		gelo.SetTracerLogger(log.New(out, nil, "", log.Lok))
	}

	if *trace {
		gelo.TraceOn(gelo.All_traces)
	}

	ret, err := vm.Run(reader, flag.Args()[1:])
	check("===PROGRAM=ERROR===", err)
	vm.API.Trace("The ultimate result of the program was", ret)
}
