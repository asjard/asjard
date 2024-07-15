package config

import "testing"

func TestEventString(t *testing.T) {
	datas := []struct {
		tp  EventType
		str string
	}{
		{tp: EventTypeCreate, str: "Create"},
		{tp: EventTypeUpdate, str: "Update"},
		{tp: EventTypeDelete, str: "Delete"},
		{tp: EventType(11), str: "Unknown"},
	}
	for _, data := range datas {
		if out := data.tp.String(); out != data.str {
			t.Errorf("test %s fail, current: %s, want: %s", data.tp, out, data.str)
			t.FailNow()
		}
	}
}
