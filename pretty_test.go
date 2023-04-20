package refaklet_test

import (
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

var demoValue *DemoValue

func ptr[T any](v T) *T { return &v }

func init() {
	demoValue = &DemoValue{
		A: 1,
		B: ptr(int(100)),
		C: "hello",
		D: []uint8{1, 2, 3},
		E: []interface{}{
			byte(10),
			int64(11),
		},
	}
}
func TestPretty(t *testing.T) {
	v := refaklet.ValueOf(demoValue)
	body := v.Format()
	assert.Equal(t,
		`&refaklet_test.DemoValue{
    A:  1,
    B:  ptr(int(100)),
    C:  "hello",
    D:  []uint8{0x1, 0x2, 0x3},
    E:  []interface {}{
        uint8(0xa),
        int64(11),
    },
}

func ptr[T any](v T) *T { return &v }`, body)

}
