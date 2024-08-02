package etcd

import (
	"testing"

	"github.com/asjard/asjard/core/runtime"
)

func TestPrefix(t *testing.T) {
	source := &Etcd{
		conf: &defaultConfig,
		app:  runtime.GetAPP(),
	}
	t.Log(source.prefix())
	t.Log(source.globalPrefix())
	t.Log(source.runtimePrefix())
	t.Log(source.prefixs())
}
