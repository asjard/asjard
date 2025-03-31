package etcd

import (
	"testing"

	"github.com/asjard/asjard/core/config"
)

func TestMain(m *testing.M) {
	if err := config.Load(2); err != nil {
		panic(err)
	}
	m.Run()
}

func TestPrefix(t *testing.T) {
	source := &Etcd{
		conf: &defaultConfig,
	}
	t.Log(source.prefix())
	t.Log(source.prefixs())
}
