package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testBootstrap struct {
	executed   bool
	shuteddown bool
}

func (t *testBootstrap) Start() error {
	t.executed = true
	return nil
}

func (t *testBootstrap) Stop() {
	t.shuteddown = true
}

func TestBootstrap(t *testing.T) {
	b := &testBootstrap{}
	AddBootstrap(b)
	assert.Nil(t, Bootstrap())
	assert.Equal(t, true, b.executed)
	Shutdown()
	assert.Equal(t, true, b.shuteddown)

}
