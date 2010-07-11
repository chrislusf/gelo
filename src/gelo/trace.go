package gelo

import "fmt"
import "log"
import "sync"

type _trace byte

const (
	Alien_trace _trace = 1 << iota
	Runtime_trace
	System_trace
	Parser_trace
)

const All_traces = Alien_trace | Runtime_trace | System_trace | Parser_trace

//global tracer
var _tracer_mutex sync.RWMutex
var _the_tracer Port = Stderr
var _the_log *log.Logger
var _level _trace

func SetTracer(p Port) Port {
	old := _the_tracer
	_tracer_mutex.Lock()
	defer _tracer_mutex.Unlock()
	_the_tracer = p
	return old
}

func SetTracerLogger(l *log.Logger) *log.Logger {
	old := _the_log
	_tracer_mutex.Lock()
	defer _tracer_mutex.Unlock()
	_the_log = l
	return old
}

func TraceOn(lvl _trace) _trace {
	_tracer_mutex.Lock()
	defer _tracer_mutex.Unlock()
	_level |= lvl
	return _level
}

func TraceOff(lvl _trace) _trace {
	_tracer_mutex.Lock()
	defer _tracer_mutex.Unlock()
	_level &^= lvl
	return _level
}

func DEBUG(all ...interface{}) {
	//for internal debugging use only
	println(_format_trace("DEBUG::", all).CopyString())
}

func _format_trace(kind string, all []interface{}) *buffer {
	buf := newBuf(0)
	buf.WriteString(kind)
	for _, v := range all {
		buf.WriteString(" ")
		switch t := v.(type) {
		default:
			buf.WriteString(fmt.Sprint(v))
		case nil:
			buf.WriteString("NIL")
		case bool:
			if t {
				buf.WriteString("true")
			} else {
				buf.WriteString("false")
			}
		case *command:
			if t == nil {
				buf.WriteString("No-op")
			} else {
				buf.WriteString("\n")
				for s := t; s != nil; s = s.next {
					buf.WriteString("(v)->")
					buf.Write(_serialize_parse_tree(s.cmd))
					buf.WriteString("\n")
				}
				buf.WriteString("(0)\n")
			}
		case *sNode:
			buf.Write(_serialize_parse_tree(t))
		case Word:
			buf.WriteWord(t)
		case []byte:
			buf.Write(t)
		case string:
			buf.WriteString(t)
		}
	}
	return buf
}

func _tracer(req _trace, kind string, all []interface{}) {
	_tracer_mutex.RLock()
	defer _tracer_mutex.RUnlock()
	if _level&req != 0 {
		tr := _format_trace(kind, all).Symbol()
		_the_tracer.Send(tr)
		if _the_log != nil {
			_the_log.Log(tr)
		}
	}
}

func parse_trace(message ...interface{}) {
	_tracer(Parser_trace, "P:", message)
}

func sys_trace(message ...interface{}) {
	_tracer(System_trace, "S:", message)
}

func run_trace(message ...interface{}) {
	_tracer(Runtime_trace, "R:", message)
}

func alien_trace(message ...interface{}) {
	_tracer(Alien_trace, "X:", message)
}

func _serialize_parse_tree(s *sNode) []byte {
	buf := newBuf(0)
	for ; s != nil; s = s.next {
		buf.WriteString("[")
		if s == nil {
			buf.WriteString("NIL")
		} else {
			switch s.tag {
			case synLiteral:
				buf.WriteWord(s.val.(Word))
			case synIndirect:
				buf.WriteString("$<")
				if n, ok := s.val.(*sNode); ok {
					buf.Write(_serialize_parse_tree(n))
				} else {
					buf.WriteWord(s.val.(Word))
				}
				buf.WriteString(">")
			case synSplice:
				buf.WriteString("@<")
				if n, ok := s.val.(*sNode); ok {
					buf.Write(_serialize_parse_tree(n))
				} else {
					buf.WriteWord(s.val.(Word))
				}
				buf.WriteString(">")
			case synQuote:
				buf.WriteString("{")
				buf.WriteWord(s.val.(Quote).Ser())
				buf.WriteString("}")
			case synClause:
				buf.Write(_serialize_parse_tree(s.val.(*sNode)))
			}
		}
		buf.WriteString("]->")
	}
	buf.WriteString("0")
	return buf.Bytes()
}
