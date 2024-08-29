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

func TestResourceKey(t *testing.T) {
	app := GetAPP()
	datas := []struct {
		resource, delimiter, key, fullKey                                                              string
		startWithDelimiter, endWithDelimiter, withoutRegion, withoutEnv, withoutService, withServiceId bool
	}{
		{
			resource:  "test_resource_colon",
			delimiter: ":",
			key:       "test_key",
			fullKey: strings.Join([]string{
				app.App,
				"test_resource_colon",
				app.Environment,
				app.Instance.Name,
				app.Region,
				app.AZ,
				"test_key",
			}, ":"),
		},
		{
			resource:  "test_resource_slash",
			delimiter: "/",
			key:       "test_key",
			fullKey: strings.Join([]string{
				app.App,
				"test_resource_slash",
				app.Environment,
				app.Instance.Name,
				app.Region,
				app.AZ,
				"test_key",
			}, "/"),
		},
		{
			resource:           "test_resource_startWithDelimiter",
			delimiter:          "/",
			key:                "test_key",
			startWithDelimiter: true,
			fullKey: strings.Join([]string{
				"",
				app.App,
				"test_resource_startWithDelimiter",
				app.Environment,
				app.Instance.Name,
				app.Region,
				app.AZ,
				"test_key",
			}, "/"),
		},
		{
			resource:         "test_resource_endWithDelimiter",
			delimiter:        "/",
			key:              "test_key",
			endWithDelimiter: true,
			fullKey: strings.Join([]string{
				app.App,
				"test_resource_endWithDelimiter",
				app.Environment,
				app.Instance.Name,
				app.Region,
				app.AZ,
				"test_key",
				"",
			}, "/"),
		},
		{
			resource:      "test_resource_withoutRegion",
			delimiter:     "/",
			key:           "test_key",
			withoutRegion: true,
			fullKey: strings.Join([]string{
				app.App,
				"test_resource_withoutRegion",
				app.Environment,
				app.Instance.Name,
				"test_key",
			}, "/"),
		},
		{
			resource:   "test_resource_withoutEnv",
			delimiter:  "/",
			key:        "test_key",
			withoutEnv: true,
			fullKey: strings.Join([]string{
				app.App,
				"test_resource_withoutEnv",
				app.Instance.Name,
				app.Region,
				app.AZ,
				"test_key",
			}, "/"),
		},
		{
			resource:       "test_resource_withoutService",
			delimiter:      "/",
			key:            "test_key",
			withoutService: true,
			fullKey: strings.Join([]string{
				app.App,
				"test_resource_withoutService",
				app.Environment,
				app.Region,
				app.AZ,
				"test_key",
			}, "/"),
		},
		{
			resource:      "test_resource_withServiceId",
			delimiter:     "/",
			key:           "test_key",
			withServiceId: true,
			fullKey: strings.Join([]string{
				app.App,
				"test_resource_withServiceId",
				app.Environment,
				app.Instance.ID,
				app.Region,
				app.AZ,
				"test_key",
			}, "/"),
		},
	}
	for _, data := range datas {
		fullKey := app.ResourceKey(data.resource, data.key,
			WithDelimiter(data.delimiter),
			WithStartWithDelimiter(data.startWithDelimiter),
			WithEndWithDelimiter(data.endWithDelimiter),
			WithoutRegion(data.withoutRegion),
			WithoutEnv(data.withoutEnv),
			WithoutService(data.withoutService),
			WithServiceId(data.withServiceId))
		t.Log(fullKey)
		if data.fullKey != fullKey {
			t.Errorf("%s: not equal, want: %s, act: %s", data.resource, data.fullKey, fullKey)
			t.FailNow()
		}
	}
}
