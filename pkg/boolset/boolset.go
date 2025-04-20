package boolset

import (
	"database/sql/driver"
	"encoding/base64"
	"errors"
	"golang.org/x/exp/constraints"
)

var (
	ErrValueNotSupported = errors.New("value not supported")
)

type BooleanSet []byte

// FromString convert from base64 encoded boolset.
func FromString(data string) (*BooleanSet, error) {
	raw, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	b := BooleanSet(raw)
	return &b, nil
}

func (b *BooleanSet) UnmarshalBinary(data []byte) error {
	*b = make(BooleanSet, len(data))
	copy(*b, data)
	return nil
}

func (b *BooleanSet) MarshalBinary() (data []byte, err error) {
	return *b, nil
}

func (b *BooleanSet) String() (data string, err error) {
	raw, err := b.MarshalBinary()
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(raw), nil
}

func (b *BooleanSet) Enabled(flag int) bool {
	if flag >= len(*b)*8 {
		return false
	}

	return (*b)[flag/8]&(1<<uint(flag%8)) != 0
}

// Value implements the driver.Valuer method.
func (b *BooleanSet) Value() (driver.Value, error) {
	return b.MarshalBinary()
}

// Scan implements the sql.Scanner method.
func (b *BooleanSet) Scan(src any) error {
	srcByte, ok := src.([]byte)
	if !ok {
		return ErrValueNotSupported
	}
	return b.UnmarshalBinary(srcByte)
}

// Sets set BooleanSet values in batch.
func Sets[T constraints.Integer](val map[T]bool, bs *BooleanSet) {
	for flag, v := range val {
		Set(flag, v, bs)
	}
}

// Set sets a BooleanSet value.
func Set[T constraints.Integer](flag T, enabled bool, bs *BooleanSet) {
	if len(*bs) < int(flag/8)+1 {
		*bs = append(*bs, make([]byte, int(flag/8)+1-len(*bs))...)
	}

	if enabled {
		(*bs)[flag/8] |= 1 << uint(flag%8)
	} else {
		(*bs)[flag/8] &= ^(1 << uint(flag%8))
	}
}
