package xdatatypes

import (
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// JSONMap defines a map[string]string that behaves like a JSON object in the database.
// It implements driver.Valuer and sql.Scanner for database compatibility,
// and various GORM interfaces for optimized ORM behavior.
type JSONMap map[string]string

// Value converts the Go map into a JSON string for database storage.
// Part of the driver.Valuer interface.
func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	ba, err := m.MarshalJSON()
	return string(ba), err
}

// Scan parses the database value (string or bytes) back into the Go JSONMap.
// Part of the sql.Scanner interface.
func (m *JSONMap) Scan(val interface{}) error {
	if val == nil {
		*m = make(JSONMap)
		return nil
	}
	var ba []byte
	switch v := val.(type) {
	case []byte:
		ba = v
	case string:
		ba = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", val))
	}
	t := map[string]string{}
	rd := bytes.NewReader(ba)
	decoder := json.NewDecoder(rd)
	// UseNumber prevents large integers from being converted to scientific notation.
	decoder.UseNumber()
	err := decoder.Decode(&t)
	*m = t
	return err
}

// MarshalJSON converts the map to a JSON byte slice.
// This ensures standard JSON behavior during API serialization.
func (m JSONMap) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	t := (map[string]string)(m)
	return json.Marshal(t)
}

// UnmarshalJSON parses a JSON byte slice into the map.
func (m *JSONMap) UnmarshalJSON(b []byte) error {
	t := map[string]string{}
	err := json.Unmarshal(b, &t)
	*m = JSONMap(t)
	return err
}

// GormDataType returns the generic data type identifier for GORM.
func (m JSONMap) GormDataType() string {
	return "jsonmap"
}

// GormDBDataType tells GORM which native database column type to use based on the driver.
// This allows for cross-database compatibility (e.g., JSONB for Postgres, JSON for MySQL).
func (JSONMap) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "JSON"
	case "mysql":
		return "JSON"
	case "postgres":
		return "JSONB"
	case "sqlserver":
		return "NVARCHAR(MAX)"
	}
	return ""
}

// GormValue ensures that when inserting data, GORM uses the correct SQL expression.
// For standard MySQL, it casts the string to a JSON type to ensure validity.
func (jm JSONMap) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	data, _ := jm.MarshalJSON()
	switch db.Dialector.Name() {
	case "mysql":
		// MariaDB supports JSON as an alias for LONGTEXT, but standard MySQL
		// benefits from an explicit CAST to the JSON data type.
		if v, ok := db.Dialector.(*mysql.Dialector); ok && !strings.Contains(v.ServerVersion, "MariaDB") {
			return gorm.Expr("CAST(? AS JSON)", string(data))
		}
	}
	return gorm.Expr("?", string(data))
}
