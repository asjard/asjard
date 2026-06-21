package validatepb

import (
	"os"
	"testing"

	"github.com/asjard/asjard/core/config"
	_ "github.com/asjard/asjard/pkg/config/mem"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	if err := config.Load(-1); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestSortValidation(t *testing.T) {
	require.NoError(t, isSortValid("", []string{"name"}))
	require.NoError(t, isSortValid("+name,-created_at", []string{"name", "created_at"}))
	require.Error(t, isSortValid("unknown", []string{"name"}))
	require.Equal(t, "child", ValidateFieldName("", "child"))
	require.Equal(t, "parent.child", ValidateFieldName("parent", "child"))
}

func TestCountryCodeValidation(t *testing.T) {
	type country struct {
		Alpha2  string `validate:"country_code=iso3166_1_alpha2"`
		Alpha3  string `validate:"country_code=iso3166_1_alpha3"`
		Numeric int    `validate:"country_code=iso3166_1_alpha_numeric"`
	}
	require.NoError(t, DefaultValidator.Struct(country{Alpha2: "CN", Alpha3: "CHN", Numeric: 156}))
	require.Error(t, DefaultValidator.Struct(country{Alpha2: "XX", Alpha3: "XXX", Numeric: 999}))

	type invalidMode struct {
		Code string `validate:"country_code=unknown"`
	}
	require.Error(t, DefaultValidator.Struct(invalidMode{Code: "CN"}))
}
