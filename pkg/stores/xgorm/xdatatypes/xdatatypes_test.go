package xdatatypes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type namedDialector struct {
	gorm.Dialector
	name string
}

func (d namedDialector) Name() string { return d.name }

func dbWithDialector(d gorm.Dialector) *gorm.DB {
	return &gorm.DB{Config: &gorm.Config{Dialector: d}}
}

func TestJSONMapRoundTrip(t *testing.T) {
	var nilMap JSONMap
	value, err := nilMap.Value()
	require.NoError(t, err)
	require.Nil(t, value)

	original := JSONMap{"one": "1", "two": "2"}
	value, err = original.Value()
	require.NoError(t, err)
	var decoded JSONMap
	require.NoError(t, decoded.Scan(value))
	require.Equal(t, original, decoded)
	require.NoError(t, decoded.Scan([]byte(`{"three":"3"}`)))
	require.Equal(t, "3", decoded["three"])
	require.NoError(t, decoded.Scan(nil))
	require.Empty(t, decoded)
	require.Error(t, decoded.Scan(123))
	require.Error(t, decoded.UnmarshalJSON([]byte(`{`)))
	require.Equal(t, "jsonmap", original.GormDataType())
}

func TestJSONMapGormTypesAndValues(t *testing.T) {
	for name, want := range map[string]string{
		"sqlite": "JSON", "mysql": "JSON", "postgres": "JSONB", "sqlserver": "NVARCHAR(MAX)", "unknown": "",
	} {
		t.Run(name, func(t *testing.T) {
			db := dbWithDialector(namedDialector{name: name})
			require.Equal(t, want, (JSONMap{}).GormDBDataType(db, nil))
			require.Equal(t, "?", (JSONMap{"a": "b"}).GormValue(context.Background(), db).SQL)
		})
	}
	mysqlDB := dbWithDialector(mysql.New(mysql.Config{ServerVersion: "8.0"}))
	require.Equal(t, "CAST(? AS JSON)", (JSONMap{"a": "b"}).GormValue(context.Background(), mysqlDB).SQL)
	mariaDB := dbWithDialector(mysql.New(mysql.Config{ServerVersion: "10.11-MariaDB"}))
	require.Equal(t, "?", (JSONMap{"a": "b"}).GormValue(context.Background(), mariaDB).SQL)
}

func TestUnsignedDatabaseTypes(t *testing.T) {
	for _, tc := range []struct {
		input int64
		want  Uint
	}{{5, 5}, {0, 0}, {-1, 0}} {
		var got Uint
		require.NoError(t, got.Scan(tc.input))
		require.Equal(t, tc.want, got)
	}
	var u Uint = 8
	require.NoError(t, u.Scan(nil))
	require.Zero(t, u)
	value, err := Uint(9).Value()
	require.NoError(t, err)
	require.Equal(t, int64(9), value)

	for _, tc := range []struct {
		input float64
		want  Ufloat
	}{{5.5, 5.5}, {0, 0}, {-1, 0}} {
		var got Ufloat
		require.NoError(t, got.Scan(tc.input))
		require.Equal(t, tc.want, got)
	}
	floatValue, err := Ufloat(3.5).Value()
	require.NoError(t, err)
	require.Equal(t, 3.5, floatValue)
}
