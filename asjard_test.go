package asjard

import (
	"syscall"
	"testing"
	"time"

	"github.com/asjard/asjard/core/server"
	"github.com/stretchr/testify/assert"
)

type testServer struct{}

func newTestServer(options *server.ServerOptions) (server.Server, error) {
	return &testServer{}, nil
}
func (testServer) AddHandler(handler any) error {
	return nil
}

func (testServer) Start(startErr chan error) error {
	return nil
}

func (testServer) Stop() {}
func (testServer) Protocol() string {
	return "test"
}
func (testServer) ListenAddresses() server.AddressConfig {
	return server.AddressConfig{}
}
func (testServer) Enabled() bool {
	return true
}

type testAPI struct{}

func TestAddHandler(t *testing.T) {
	server.AddServer("test", newTestServer)
	t.Run("AddHandler", func(t *testing.T) {
		server := New()
		var errChan = make(chan error)
		assert.Nil(t, server.AddHandler(&testAPI{}, "test"))
		go func() {
			if err := server.Start(); err != nil {
				errChan <- err
			}
		}()
		select {
		case err := <-errChan:
			t.Error(err)
			t.FailNow()
		case <-time.After(5 * time.Second):
			syscall.Kill(syscall.Getpid(), syscall.SIGHUP)
			break
		}

	})
}
