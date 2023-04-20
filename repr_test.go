package refaklet_test

import (
	"strings"
	"testing"

	"github.com/alingse/refaklet"
	"github.com/stretchr/testify/assert"
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

type DemoValue2 struct {
	D *DemoValue
	E []*DemoValue
	F []DemoValue
	G any
	H []any
	K map[string]any
}

var demoValue2 = &DemoValue2{
	D: demoValue,
	E: []*DemoValue{demoValue},
	F: []DemoValue{*demoValue},
	G: nil,
	H: []any{ptr(1), ptr("hello")},
	K: map[string]any{"demo": "x"},
}

var _ = &DemoValue2{
	D: &DemoValue{
		A: 1,
		B: ptr(int(100)),
		C: "hello",
		D: []byte{0x1, 0x2, 0x3},
		E: []any{
			uint8(0xa),
			int64(11),
		},
	},
	E: []*DemoValue{
		{
			A: 1,
			B: ptr(int(100)),
			C: "hello",
			D: []byte{0x1, 0x2, 0x3},
			E: []any{
				uint8(0xa),
				int64(11),
			},
		},
	},
	F: []DemoValue{
		{
			A: 1,
			B: ptr(int(100)),
			C: "hello",
			D: []byte{0x1, 0x2, 0x3},
			E: []any{
				uint8(0xa),
				int64(11),
			},
		},
	},
	G: nil,
	H: []any{
		ptr(int(1)),
		ptr("hello"),
	},
	K: map[string]any{
		"demo": "x",
	},
}

func TestPretty(t *testing.T) {
	v := refaklet.ValueOf(demoValue)
	body := v.Repr()
	assert.Equal(t,
		`&refaklet_test.DemoValue{
	A:	1,
	B:	ptr(int(100)),
	C:	"hello",
	D:	[]byte{0x1, 0x2, 0x3},
	E:	[]any{
		uint8(0xa),
		int64(11),
	},
}

func ptr[T any](v T) *T { return &v }`, body)
}

func TestPrettyComplex(t *testing.T) {
	v := refaklet.ValueOf(demoValue2)
	body := v.Repr()
	var repr = `
&refaklet_test.DemoValue2{
	D:	&refaklet_test.DemoValue{
		A:	1,
		B:	ptr(int(100)),
		C:	"hello",
		D:	[]byte{0x1, 0x2, 0x3},
		E:	[]any{
			uint8(0xa),
			int64(11),
		},
	},
	E:	[]*refaklet_test.DemoValue{
		&refaklet_test.DemoValue{
			A:	1,
			B:	ptr(int(100)),
			C:	"hello",
			D:	[]byte{0x1, 0x2, 0x3},
			E:	[]any{
				uint8(0xa),
				int64(11),
			},
		},
	},
	F:	[]refaklet_test.DemoValue{
		{
			A:	1,
			B:	ptr(int(100)),
			C:	"hello",
			D:	[]byte{0x1, 0x2, 0x3},
			E:	[]any{
				uint8(0xa),
				int64(11),
			},
		},
	},
	G:	nil,
	H:	[]any{
		ptr(int(1)),
		ptr("hello"),
	},
	K:	map[string]interface {}{
		"demo":	"x",
	},
}

func ptr[T any](v T) *T { return &v }`
	assert.Equal(t, strings.TrimSpace(repr), body)
}
