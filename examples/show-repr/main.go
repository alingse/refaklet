package main

import (
	"fmt"

	"github.com/alingse/refaklet"
)

type DemoValue struct {
	A int
	B *int
	C string
	D []byte
	E []any
}

var demoValue = &DemoValue{
	A: 1,
	B: ptr(int(100)),
	C: "hello",
	D: []byte{1, 2, 3},
	E: []interface{}{
		byte(10),
		int64(11),
	},
}

func ptr[T any](v T) *T { return &v }

var repr_ = `&main.DemoValue{
	A:	1,
	B:	ptr(int(100)),
	C:	"hello",
	D:	[]byte{0x1, 0x2, 0x3},
	E:	[]any{
		uint8(0xa),
		int64(11),
	},
}

func ptr[T any](v T) *T { return &v }`

func main() {
	v := refaklet.ValueOf(demoValue)
	repr := v.Repr()
	fmt.Println(repr == repr_)
}
