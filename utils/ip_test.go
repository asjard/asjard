package utils

import (
	"net"
	"testing"
)

func TestGetListenAddress(t *testing.T) {
	datas := []struct {
		hostPort string
		output   string
		ok       bool
	}{
		{hostPort: "0.0.0.0:8080", output: LocalIPv4() + ":8080", ok: true},
		{hostPort: "127.0.0.1:8080", output: "127.0.0.1:8080", ok: true},
		{hostPort: "127.0.0.1:", output: "127.0.0.1:", ok: true},
		// 没有IPv6不会有预期结果
		{hostPort: "[::]:8080", output: "[" + LocalIPv6() + "]" + ":8080", ok: LocalIPv6() != ""},
		{hostPort: "[::]:", output: "[" + LocalIPv6() + "]" + ":", ok: LocalIPv6() != ""},
		// invlaid host port
		{hostPort: "0.0.0.0", ok: false},
		{hostPort: "::", ok: false},
		// invalid ip address
		{hostPort: "0.0.0:8080", ok: false},
		{hostPort: ":8080", ok: false},
		{hostPort: "[]:8080", ok: false},
	}
	for _, data := range datas {
		output, err := GetListenAddress(data.hostPort)
		if ((err == nil) != data.ok) || (data.ok && output != data.output) {
			t.Log(output, err)
			t.Errorf("test %s fail, current: %s, want: %s", data.hostPort, output, data.output)
			t.FailNow()
		}
	}
}

func TestIsIPv6(t *testing.T) {
	datas := []struct {
		input  string
		output bool
	}{
		{input: "", output: false},
		{input: "abc", output: false},
		{input: "127.0.0.1", output: false},
		{input: "0000:0000:0000:0000:0000:0000:0000:000", output: true},
		{input: "::", output: true},
		{input: "0000:0000:0000:0000:0000:0000:0000:0000", output: true},
		{input: "fe80::c706:e006:d53e:f9fb", output: true},
		{input: "fe80::10.25.21.2", output: true},
		{input: "10.25.21.2", output: false},
	}
	for _, data := range datas {
		if output := IsIPv6(net.ParseIP(data.input)); output != data.output {
			t.Errorf("test %s fail, current: %v, want: %v", data.input, output, data.output)
		}
	}
}

func TestLocalIPv6(t *testing.T) {
	LocalIPv6()
}

func TestLocalIPv4(t *testing.T) {
	LocalIPv4()
}
