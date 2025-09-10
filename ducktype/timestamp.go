package ducktype

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type Status byte

const (
	Undefined Status = iota
	Null
	Present
)

type InfinityModifier int8

const (
	Infinity         InfinityModifier = 1
	None             InfinityModifier = 0
	NegativeInfinity InfinityModifier = -Infinity
)

func (im InfinityModifier) String() string {
	switch im {
	case None:
		return "none"
	case Infinity:
		return "infinity"
	case NegativeInfinity:
		return "-infinity"
	default:
		return "invalid"
	}
}

type NullTime struct {
	Time             time.Time // Time must always be in UTC.
	Status           Status
	InfinityModifier InfinityModifier
}

func (dst *NullTime) Infinity() {
	dst.Status = Present
	dst.InfinityModifier = Infinity
}

func (dst *NullTime) NegativeInfinity() {
	dst.Status = Present
	dst.InfinityModifier = NegativeInfinity
}

// Scan implements the database/sql Scanner interface.
func (dst *NullTime) Scan(src interface{}) error {
	if src == nil {
		*dst = NullTime{Status: Null}
		return nil
	}

	switch src := src.(type) {
	case time.Time:
		// these two timestamps are what duckdb returns for pos/neg infinity.
		var (
			inf    = time.UnixMicro(9223372036854775807)
			neginf = time.UnixMicro(-9223372036854775807)
		)

		if src.Equal(inf) {
			*dst = NullTime{Status: Present, InfinityModifier: Infinity}
			return nil
		}

		if src.Equal(neginf) {
			*dst = NullTime{Status: Present, InfinityModifier: NegativeInfinity}
			return nil
		}

		*dst = NullTime{Time: src, Status: Present}
		return nil
	}

	return fmt.Errorf("cannot scan %T", src)
}

// Value implements the database/sql/driver Valuer interface.
func (src NullTime) Value() (driver.Value, error) {
	switch src.Status {
	case Present:
		if src.InfinityModifier != None {
			return src.InfinityModifier.String(), nil
		}
		return src.Time, nil
	case Null:
		return nil, nil
	default:
		return nil, fmt.Errorf("undefined timestamp value")
	}
}
