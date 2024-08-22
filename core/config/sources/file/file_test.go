package file

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/asjard/asjard/core/config"
	_ "github.com/asjard/asjard/core/security"
	"github.com/asjard/asjard/utils"
	"github.com/stretchr/testify/assert"
)

func TestFile(t *testing.T) {
	tmpDir := t.TempDir()
	assert.Nil(t, os.Setenv(utils.CONF_DIR_ENV_NAME, tmpDir))
	testFile := "test_file.yaml"
	testKey := "test_key"
	testValue := "test_value"
	assert.Nil(t, os.WriteFile(filepath.Join(tmpDir, testFile),
		[]byte(fmt.Sprintf("%s: %s", testKey, testValue)), 0640))

	var m sync.RWMutex
	var eventKey string
	var eventValue any
	source, err := New(&config.SourceOptions{
		Callback: func(event *config.Event) {
			m.Lock()
			defer m.Unlock()
			eventKey = event.Key
			eventValue = event.Value.Value
		},
	})
	defer source.Disconnect()
	assert.Nil(t, err)

	t.Run("GetAll", func(t *testing.T) {
		configs := source.GetAll()
		value, ok := configs[testKey]
		assert.Equal(t, true, ok)
		assert.Equal(t, testValue, value.Value)
	})
	t.Run("EncryptFile", func(t *testing.T) {
		testFile := "encrypted_base64_file.yaml"
		testKey := "base64_key"
		testValue := "base64_value"

		content := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s: %s", testKey, testValue)))
		assert.Nil(t, os.WriteFile(filepath.Join(tmpDir, testFile), []byte(content), 0640))
		time.Sleep(50 * time.Millisecond)
		m.RLock()
		defer m.RUnlock()
		assert.Equal(t, testKey, eventKey)
		assert.Equal(t, testValue, eventValue)
	})
}
