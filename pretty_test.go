package refaklet_test

import (
	"testing"

	"github.com/alingse/refaklet"
	"github.com/kr/pretty"
	"github.com/stretchr/testify/assert"
)

type DemoValue struct {
	A int
	B *int
	C string
	D []byte
}

func TestPretty(t *testing.T) {
	var b int = 100
	var demoValue = &DemoValue{
		A: 1,
		B: &b,
		C: "hello",
		D: []byte{1, 2, 3},
	}
	v := refaklet.ValueOf(demoValue)
	body := v.Format()
	assert.Equal(t, `&refaklet_test.DemoValue{
    A:  1,
    B:  &int(100),
    C:  "hello",
    D:  {0x1, 0x2, 0x3},
}`, body)

	body2 := pretty.Sprint(demoValue)
	assert.Equal(t, `&refaklet_test.DemoValue{
    A:  1,
    B:  &int(100),
    C:  "hello",
    D:  {0x1, 0x2, 0x3},
}`, body2)
}
