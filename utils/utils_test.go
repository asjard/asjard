package utils

import "testing"

func TestIntLen(t *testing.T) {
	datas := []struct {
		x uint32
		l uint32
	}{
		{x: 10, l: 2},
		{x: 0, l: 1},
		{x: 1, l: 1},
		{x: 10_0, l: 3},
		{x: 10_00, l: 4},
	}
	for _, data := range datas {
		out := Uint32Len(data.x)
		if out != data.l {
			t.Errorf("get %d len %d want %d", data.x, out, data.l)
			t.FailNow()
		}
	}
}
