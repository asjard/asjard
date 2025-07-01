package xdatatypes

import (
	"database/sql/driver"
)

// 当值小于0时返回0
type Uint int

func (m Uint) Value() (driver.Value, error) {
	return int64(m), nil
}

func (m *Uint) Scan(val interface{}) error {
	if val == nil {
		*m = 0
	}
	switch v := val.(type) {
	case int64:
		if v < 0 {
			*m = 0
		} else {
			*m = Uint(v)
		}
	}
	return nil
}

type Ufloat float64

func (m Ufloat) Value() (driver.Value, error) {
	return float64(m), nil
}

func (m *Ufloat) Scan(val interface{}) error {
	if val == nil {
		*m = 0
	}
	switch v := val.(type) {
	case float64:
		if v < 0 {
			*m = 0
		} else {
			*m = Ufloat(v)
		}
	}
	return nil
}
