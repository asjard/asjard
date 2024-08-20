package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testBootstrap struct {
	executed   bool
	shuteddown bool
}

func (t *testBootstrap) Bootstrap() error {
	t.executed = true
	return nil
}

func (t *testBootstrap) Shutdown() {
	t.shuteddown = true
}

func TestBootstrap(t *testing.T) {
	b := &testBootstrap{}
	AddBootstrap(b)
	assert.Nil(t, Start())
	assert.Equal(t, true, b.executed)
	Stop()
	assert.Equal(t, true, b.shuteddown)

}
