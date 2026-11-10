package rest

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMapForm(t *testing.T) {
	type form struct {
		Name    string
		Count   int
		Enabled bool
		Ratio   float64
		When    time.Time `json:"When" time_format:"2006-01-02"`
		Tags    []string
	}
	var got form
	err := mapForm(&got, map[string][]string{
		"Name": {"test"}, "Count": {"3"}, "Enabled": {"true"}, "Ratio": {"1.5"},
		"When": {"2026-06-21"}, "Tags": {"one", "two"},
	})
	require.NoError(t, err)
	require.Equal(t, "test", got.Name)
	require.Equal(t, 3, got.Count)
	require.True(t, got.Enabled)
	require.Equal(t, 1.5, got.Ratio)
	require.Equal(t, []string{"one", "two"}, got.Tags)
	require.Equal(t, 2026, got.When.Year())

	require.Error(t, mapForm(&got, map[string][]string{"Count": {"not-an-int"}}))
	require.NoError(t, mapForm(&got, nil))
}

func TestCorsOriginValidation(t *testing.T) {
	conf := CorsConfig{AllowOrigins: []string{"https://example.com"}}
	require.True(t, corsIsOriginValid(conf, "https://example.com"))
	require.False(t, corsIsOriginValid(conf, "https://invalid.example"))
	require.True(t, corsIsOriginValid(CorsConfig{AllowOrigins: []string{"*"}, allowAllOrigins: true}, "https://anything.example"))
}
