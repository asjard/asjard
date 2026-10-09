package runtime

import (
	"strings"
	"testing"

	"github.com/asjard/asjard/core/config"
	_ "github.com/asjard/asjard/pkg/config/mem"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	if err := config.Load(-1); err != nil {
		panic(err)
	}
	m.Run()
}

func TestGetAPP(t *testing.T) {
	assert.Equal(t, GetAPP().Instance.ID, app.Instance.ID)
	assert.Equal(t, GetAPP().Instance.ID, app.Instance.ID)
}

func TestResourceKey(t *testing.T) {
	app := GetAPP()
	datas := []struct {
		resource, delimiter, key, fullKey                                                                                                  string
		startWithDelimiter, endWithDelimiter, withoutApp, withoutRegion, withoutAz, withoutEnv, withoutService, withServiceId, withVersion bool
	}{
		{
			resource:  "test_resource_colon",
			delimiter: ":",
			key:       "test_key",
			fullKey: strings.Join([]string{
				app.App,
				app.Environment,
				app.Instance.Name,
				app.Region,
				app.AZ,
				"test_resource_colon",
				"test_key",
			}, ":"),
		},
		{
			resource:  "test_resource_slash",
			delimiter: "/",
			key:       "test_key",
			fullKey: strings.Join([]string{
				app.App,
				app.Environment,
				app.Instance.Name,
				app.Region,
				app.AZ,
				"test_resource_slash",
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
				app.Environment,
				app.Instance.Name,
				app.Region,
				app.AZ,
				"test_resource_startWithDelimiter",
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
				app.Environment,
				app.Instance.Name,
				app.Region,
				app.AZ,
				"test_resource_endWithDelimiter",
				"test_key",
				"",
			}, "/"),
		},
		{
			resource:   "test_resource_withoutApp",
			delimiter:  "/",
			key:        "test_key",
			withoutApp: true,
			fullKey: strings.Join([]string{
				app.Environment,
				app.Instance.Name,
				app.Region,
				app.AZ,
				"test_resource_withoutApp",
				"test_key",
			}, "/"),
		},
		{
			resource:      "test_resource_withoutRegion",
			delimiter:     "/",
			key:           "test_key",
			withoutRegion: true,
			fullKey: strings.Join([]string{
				app.App,
				app.Environment,
				app.Instance.Name,
				app.AZ,
				"test_resource_withoutRegion",
				"test_key",
			}, "/"),
		},
		{
			resource:  "test_resource_withoutAz",
			delimiter: "/",
			key:       "test_key",
			withoutAz: true,
			fullKey: strings.Join([]string{
				app.App,
				app.Environment,
				app.Instance.Name,
				app.Region,
				"test_resource_withoutAz",
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
				app.Instance.Name,
				app.Region,
				app.AZ,
				"test_resource_withoutEnv",
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
				app.Environment,
				app.Region,
				app.AZ,
				"test_resource_withoutService",
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
				app.Environment,
				app.Instance.ID,
				app.Region,
				app.AZ,
				"test_resource_withServiceId",
				"test_key",
			}, "/"),
		},
		{
			resource:    "test_resource_withoutVersion",
			delimiter:   "/",
			key:         "test_key",
			withVersion: true,
			fullKey: strings.Join([]string{
				app.App,
				app.Environment,
				app.Instance.Name,
				app.Instance.Version,
				app.Region,
				app.AZ,
				"test_resource_withoutVersion",
				"test_key",
			}, "/"),
		},
	}
	for _, data := range datas {
		fullKey := app.ResourceKey(data.resource, data.key,
			WithDelimiter(data.delimiter),
			WithStartWithDelimiter(data.startWithDelimiter),
			WithEndWithDelimiter(data.endWithDelimiter),
			WithoutApp(data.withoutApp),
			WithoutRegion(data.withoutRegion),
			WithoutAz(data.withoutAz),
			WithoutEnv(data.withoutEnv),
			WithoutService(data.withoutService),
			WithServiceId(data.withServiceId),
			WithVersion(data.withVersion))
		t.Log(fullKey)
		if data.fullKey != fullKey {
			t.Errorf("%s: not equal, want: %s, act: %s", data.resource, data.fullKey, fullKey)
			t.FailNow()
		}
	}
}

func BenchmarkResourceKey(b *testing.B) {
	var benchAPP = APP{
		App:         "usercenter",
		Environment: "prod",
		Region:      "cn-hangzhou",
		AZ:          "cn-hangzhou-h",
		Instance: Instance{
			ID:      "i-1234567890abcdef0",
			Name:    "usercenter-v2",
			Version: "20241118.1",
		},
	}

	var fullOpts = []Option{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = benchAPP.ResourceKey("cache", "user:12345", fullOpts...)
		}
	})
}
