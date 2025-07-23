package model

import (
	"database/sql"
	"encoding/json"
	"strconv"
)

// NullString is a wrapper around sql.NullString
// swagger:ignore
type NullString struct {
	sql.NullString
}

// MarshalJSON method is called by json.Marshal,
// whenever it is of type NullString
func (x NullString) MarshalJSON() ([]byte, error) {
	if !x.Valid {
		// return []byte("null"), nil
		return []byte(`""`), nil
	}
	return json.Marshal(x.String)
}

func (x *NullString) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		x.Valid = false
		return nil
	}
	x.Valid = true
	err := json.Unmarshal(b, &x.String)
	return err
}

// NullInt is a wrapper around sql.NullInt32
// swagger:ignore
type NullInt struct {
	sql.NullInt32
}

// MarshalJSON method is called by json.Marshal,
// whenever it is of type NullInt
func (x NullInt) MarshalJSON() ([]byte, error) {
	if !x.Valid {
		// return []byte("null"), nil
		return []byte(`0`), nil
	}
	return json.Marshal(x.Int32)
}

func (x *NullInt) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		x.Valid = false
		return nil
	}
	x.Valid = true
	err := json.Unmarshal(b, &x.Int32)
	return err
}

// NullFloat is a wrapper around sql.NullFloat64
// swagger:ignore
type NullFloat struct {
	sql.NullFloat64
}

// MarshalJSON method is called by json.Marshal,
// whenever it is of type NullFloat
func (x NullFloat) MarshalJSON() ([]byte, error) {
	if !x.Valid {
		// return []byte("null"), nil
		return []byte(`0`), nil
	}
	return json.Marshal(x.Float64)
}

func (x *NullFloat) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		x.Valid = false
		return nil
	}
	x.Valid = true
	err := json.Unmarshal(b, &x.Float64)
	return err
}

// NullFloat2d is a wrapper around sql.NullFloat64
// swagger:ignore
type NullFloat2d struct {
	sql.NullFloat64
}

// MarshalJSON method is called by json.Marshal,
// whenever it is of type NullFloat2d
func (x NullFloat2d) MarshalJSON() ([]byte, error) {
	if !x.Valid {
		// return []byte("null"), nil
		return []byte(`0`), nil
	}
	if float64(x.Float64) == float64(int(x.Float64)) {
		return []byte(strconv.FormatFloat(float64(x.Float64), 'f', 0, 32)), nil
	}
	return []byte(strconv.FormatFloat(float64(x.Float64), 'f', 2, 32)), nil
}

func (x *NullFloat2d) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		x.Valid = false
		return nil
	}
	x.Valid = true
	err := json.Unmarshal(b, &x.Float64)
	return err
}

// NullBool is a wrapper around sql.NullBool
// swagger:ignore
type NullBool struct {
	sql.NullBool
}

// MarshalJSON method is called by json.Marshal,
// whenever it is of type NullBool
func (x NullBool) MarshalJSON() ([]byte, error) {
	if !x.Valid {
		// return []byte("null"), nil
		return []byte(`false`), nil
	}
	return json.Marshal(x.Bool)
}

func (x *NullBool) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		x.Valid = false
		return nil
	}
	x.Valid = true
	err := json.Unmarshal(b, &x.Bool)
	return err
}

// swagger:model
type PriceFloat float64

// swagger:ignore
type IntArrayAsString string

// MarshalJSON method is called by json.Marshal,
// whenever it is of type IntArrayAsString
func (x IntArrayAsString) MarshalJSON() ([]byte, error) {
	return []byte("[" + x + "]"), nil
}

func (x *IntArrayAsString) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*x = ""
	}
	b = b[1 : len(b)-1]
	*x = IntArrayAsString(b)
	return nil
}

func (f PriceFloat) MarshalJSON() ([]byte, error) {
	if float64(f) == float64(int(f)) {
		return []byte(strconv.FormatFloat(float64(f), 'f', 0, 32)), nil
	}
	return []byte(strconv.FormatFloat(float64(f), 'f', 2, 32)), nil
}
