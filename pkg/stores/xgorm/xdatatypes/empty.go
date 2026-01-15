package xdatatypes

import (
	"database/sql/driver"
)

// Uint is a custom integer type that enforces a non-negative constraint.
// If a negative value is scanned from the database, it is automatically converted to 0.
type Uint int

// Value implements the driver.Valuer interface.
// It allows the Uint type to be converted into a format the database driver can handle (int64).
func (m Uint) Value() (driver.Value, error) {
	return int64(m), nil
}

// Scan implements the sql.Scanner interface.
// It is called when retrieving data from the database. It converts the database
// value into a Uint and ensures that if the input is negative, the result is 0.
func (m *Uint) Scan(val interface{}) error {
	// Handle NULL values in the database.
	if val == nil {
		*m = 0
		return nil
	}

	switch v := val.(type) {
	case int64:
		// Logic: clamp negative values to zero.
		if v < 0 {
			*m = 0
		} else {
			*m = Uint(v)
		}
	}
	return nil
}

// Ufloat is a custom float64 type that enforces a non-negative constraint.
// Similar to Uint, it prevents negative floating-point values from entering the application.
type Ufloat float64

// Value implements the driver.Valuer interface for database persistence.
func (m Ufloat) Value() (driver.Value, error) {
	return float64(m), nil
}

// Scan implements the sql.Scanner interface for data retrieval.
// It ensures that any negative float64 retrieved from the database is clamped to 0.0.
func (m *Ufloat) Scan(val interface{}) error {
	// Handle NULL values in the database.
	if val == nil {
		*m = 0
		return nil
	}

	switch v := val.(type) {
	case float64:
		// Logic: clamp negative values to zero.
		if v < 0 {
			*m = 0
		} else {
			*m = Ufloat(v)
		}
	}
	return nil
}
