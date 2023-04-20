package refaklet

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"text/tabwriter"

	"github.com/kr/text"
	"github.com/rogpeppe/go-internal/fmtsort"
)

// code copy from  github.com/kr/pretty

type refakletValue struct {
	rv    reflect.Value
	quote bool
	ptr   bool
}

func ValueOf(v any) refakletValue {
	return refakletValue{
		rv:    reflect.ValueOf(v),
		quote: true,
	}
}

func (f refakletValue) Repr() string {
	buf := bytes.NewBuffer(nil)
	w := tabwriter.NewWriter(buf, 4, 4, 1, '\t', 0)
	p := &printer{
		tw:      w,
		Writer:  w,
		visited: make(map[visit]int),
		ref:     &f,
	}
	p.printValue(f.rv, true, f.quote)
	p.printHelper()
	w.Flush()
	return buf.String()
}

type printer struct {
	io.Writer
	tw      *tabwriter.Writer
	visited map[visit]int
	depth   int
	ref     *refakletValue
}

func (p *printer) indent() *printer {
	q := *p
	q.tw = tabwriter.NewWriter(p.Writer, 4, 4, 1, '\t', 0)
	q.Writer = text.NewIndentWriter(q.tw, []byte{'\t'})
	return &q
}

func (p *printer) printInline(v reflect.Value, x interface{}, showType bool) {
	if showType {
		io.WriteString(p, v.Type().String())
		fmt.Fprintf(p, "(%#v)", x)
	} else {
		fmt.Fprintf(p, "%#v", x)
	}
}

// printValue must keep track of already-printed pointer values to avoid
// infinite recursion.
type visit struct {
	v   uintptr
	typ reflect.Type
}

func (p *printer) catchPanic(v reflect.Value, method string) {
	if r := recover(); r != nil {
		if v.Kind() == reflect.Ptr && v.IsNil() {
			writeByte(p, '(')
			io.WriteString(p, v.Type().String())
			io.WriteString(p, ")(nil)")
			return
		}
		writeByte(p, '(')
		io.WriteString(p, v.Type().String())
		io.WriteString(p, ")(PANIC=calling method ")
		io.WriteString(p, strconv.Quote(method))
		io.WriteString(p, ": ")
		fmt.Fprint(p, r)
		writeByte(p, ')')
	}
}

func (p *printer) printValue(v reflect.Value, showType, quote bool) {
	if p.depth > 10 {
		io.WriteString(p, "!%v(DEPTH EXCEEDED)")
		return
	}

	if v.IsValid() && v.CanInterface() {
		i := v.Interface()
		if goStringer, ok := i.(fmt.GoStringer); ok {
			defer p.catchPanic(v, "GoString")
			io.WriteString(p, goStringer.GoString())
			return
		}
	}

	switch v.Kind() {
	case reflect.Bool:
		p.printInline(v, v.Bool(), showType)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		p.printInline(v, v.Int(), showType)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		p.printInline(v, v.Uint(), showType)
	case reflect.Float32, reflect.Float64:
		p.printInline(v, v.Float(), showType)
	case reflect.Complex64, reflect.Complex128:
		fmt.Fprintf(p, "%#v", v.Complex())
	case reflect.String:
		p.fmtString(v.String(), quote)
	case reflect.Map:
		t := v.Type()
		if showType {
			io.WriteString(p, t.String())
		}
		writeByte(p, '{')
		if nonzero(v) {
			expand := !canInline(v.Type())
			pp := p
			if expand {
				writeByte(p, '\n')
				pp = p.indent()
			}
			sm := fmtsort.Sort(v)
			for i := 0; i < v.Len(); i++ {
				k := sm.Key[i]
				mv := sm.Value[i]
				pp.printValue(k, false, true)
				writeByte(pp, ':')
				if expand {
					writeByte(pp, '\t')
				}
				showTypeInStruct := t.Elem().Kind() == reflect.Interface
				pp.printValue(mv, showTypeInStruct, true)
				if expand {
					io.WriteString(pp, ",\n")
				} else if i < v.Len()-1 {
					io.WriteString(pp, ", ")
				}
			}
			if expand {
				pp.tw.Flush()
			}
		}
		writeByte(p, '}')
	case reflect.Struct:
		t := v.Type()
		if v.CanAddr() {
			addr := v.UnsafeAddr()
			vis := visit{addr, t}
			if vd, ok := p.visited[vis]; ok && vd < p.depth {
				p.fmtString(t.String()+"{(CYCLIC REFERENCE)}", false)
				break // don't print v again
			}
			p.visited[vis] = p.depth
		}

		if showType {
			io.WriteString(p, t.String())
		}
		writeByte(p, '{')
		if nonzero(v) {
			expand := !canInline(v.Type())
			pp := p
			if expand {
				writeByte(p, '\n')
				pp = p.indent()
			}
			for i := 0; i < v.NumField(); i++ {
				showTypeInStruct := true
				if f := t.Field(i); f.Name != "" {
					io.WriteString(pp, f.Name)
					writeByte(pp, ':')
					if expand {
						writeByte(pp, '\t')
					}
					showTypeInStruct = labelType(f.Type)
				}
				pp.printValue(getField(v, i), showTypeInStruct, true)
				if expand {
					io.WriteString(pp, ",\n")
				} else if i < v.NumField()-1 {
					io.WriteString(pp, ", ")
				}
			}
			if expand {
				pp.tw.Flush()
			}
		}
		writeByte(p, '}')
	case reflect.Interface:
		switch e := v.Elem(); {
		case e.Kind() == reflect.Invalid:
			io.WriteString(p, "nil")
		case e.IsValid():
			pp := *p
			pp.depth++
			pp.printValue(e, showType, true)
		default:
			io.WriteString(p, v.Type().String())
			io.WriteString(p, "(nil)")
		}
	case reflect.Array, reflect.Slice:
		t := v.Type()
		if showType {
			io.WriteString(p, typeString(t))
		}
		if v.Kind() == reflect.Slice && v.IsNil() && showType {
			io.WriteString(p, "(nil)")
			break
		}
		if v.Kind() == reflect.Slice && v.IsNil() {
			io.WriteString(p, "nil")
			break
		}
		writeByte(p, '{')
		expand := !canInline(v.Type())
		pp := p
		if expand {
			writeByte(p, '\n')
			pp = p.indent()
		}
		for i := 0; i < v.Len(); i++ {
			showTypeInSlice := t.Elem().Kind() == reflect.Interface
			pp.printValue(v.Index(i), showTypeInSlice, true)
			if expand {
				io.WriteString(pp, ",\n")
			} else if i < v.Len()-1 {
				io.WriteString(pp, ", ")
			}
		}
		if expand {
			pp.tw.Flush()
		}
		writeByte(p, '}')
	case reflect.Ptr:
		e := v.Elem()
		if !e.IsValid() {
			writeByte(p, '(')
			io.WriteString(p, v.Type().String())
			io.WriteString(p, ")(nil)")
		} else {
			pp := *p
			pp.depth++

			if isBasicType(e) {
				io.WriteString(pp, "ptr(")
				pp.printValue(e, true, true)
				writeByte(pp, ')')
				p.ref.ptr = true
			} else {
				writeByte(p, '&')
				pp.printValue(e, true, true)
			}
		}
	case reflect.Chan:
		x := v.Pointer()
		if showType {
			writeByte(p, '(')
			io.WriteString(p, v.Type().String())
			fmt.Fprintf(p, ")(%#v)", x)
		} else {
			fmt.Fprintf(p, "%#v", x)
		}
	case reflect.Func:
		io.WriteString(p, v.Type().String())
		io.WriteString(p, " {...}")
	case reflect.UnsafePointer:
		p.printInline(v, v.Pointer(), showType)
	case reflect.Invalid:
		io.WriteString(p, "nil")
	}
}

func (p *printer) printHelper() {
	if p.ref.ptr {
		io.WriteString(p, "\n\nfunc ptr[T any](v T) *T { return &v }")
	}
}

func canInline(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Map:
		return !canExpand(t.Elem())
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if canExpand(t.Field(i).Type) {
				return false
			}
		}
		return true
	case reflect.Interface:
		return false
	case reflect.Array, reflect.Slice:
		return !canExpand(t.Elem())
	case reflect.Ptr:
		return false
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return false
	}
	return true
}

func canExpand(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Map, reflect.Struct,
		reflect.Interface, reflect.Array, reflect.Slice,
		reflect.Ptr:
		return true
	}
	return false
}

func labelType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Interface, reflect.Struct:
		return true
	case reflect.Slice, reflect.Map:
		return true
	}
	return false
}

func (p *printer) fmtString(s string, quote bool) {
	if quote {
		s = strconv.Quote(s)
	}
	io.WriteString(p, s)
}

func writeByte(w io.Writer, b byte) {
	w.Write([]byte{b})
}

func getField(v reflect.Value, i int) reflect.Value {
	val := v.Field(i)
	if val.Kind() == reflect.Interface && !val.IsNil() {
		val = val.Elem()
	}
	return val
}

func isBasicType(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool, reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return true
	default:
		return false
	}
}

func typeString(t reflect.Type) string {
	s := t.String()
	switch s {
	case "uint8":
		return "byte"
	case "[]uint8":
		return "[]byte"
	case "interface {}":
		return "any"
	case "[]interface {}":
		return "[]any"
	default:
		return s
	}
}
