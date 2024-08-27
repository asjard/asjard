package runtime

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAPP(t *testing.T) {
	assert.Equal(t, GetAPP().Instance.ID, app.Instance.ID)
	assert.Equal(t, GetAPP().Instance.ID, app.Instance.ID)
}

func TestResourceKeyPrefix(t *testing.T) {
	app := GetAPP()
	datas := []struct {
		resource, delimiter, key, fullKey    string
		startWithDelimiter, endWithDelimiter bool
	}{
		{
			resource:  "caches",
			delimiter: ":",
			key:       "xxx",
		},
		{
			resource:           "caches",
			delimiter:          "/",
			key:                "xxx",
			startWithDelimiter: true,
			endWithDelimiter:   true,
		},
	}
	for _, data := range datas {
		fullKey := app.ResourceKey(data.resource, data.key, data.delimiter, data.startWithDelimiter, data.endWithDelimiter)
		if data.startWithDelimiter {
			if !strings.HasPrefix(fullKey, data.delimiter) {
				t.Error("not start with delimiter")
				t.FailNow()
			}
		} else if strings.HasPrefix(fullKey, data.delimiter) {
			t.Error("can not start with delimiter")
			t.FailNow()
		}
		if data.endWithDelimiter {
			if !strings.HasSuffix(fullKey, data.delimiter) {
				t.Error("not end with delimiter")
				t.FailNow()
			}
		} else if strings.HasSuffix(fullKey, data.delimiter) {
			t.Error("cant not end with delimiter")
			t.FailNow()
		}
	}
}
