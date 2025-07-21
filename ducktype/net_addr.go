package ducktype

import (
	"database/sql/driver"
	"fmt"
	"net/netip"
)

// NullNetAddr represents a netip.Addr that may be null.
// The V field holds the netip.Addr value, and Valid indicates its validity.
type NullNetAddr struct {
	V     netip.Addr
	Valid bool
}

// Scan implements the sql.Scanner interface.
// It supports scanning a value from a database driver, including NULL,
// and can handle database types that represent an IP address as a string or a byte slice,
// including values from `net.IP`.
func (n *NullNetAddr) Scan(src any) error {
	if src == nil {
		n.V, n.Valid = netip.Addr{}, false
		return nil
	}
	n.Valid = true
	switch v := src.(type) {
	case []byte:
		addr, ok := netip.AddrFromSlice(v)
		if !ok {
			return fmt.Errorf("nullnetip: cannot scan []byte %q into netip.Addr", v)
		}
		n.V = addr
		return nil
	case string:
		addr, err := netip.ParseAddr(v)
		if err != nil {
			return fmt.Errorf("nullnetip: failed to parse string %q as netip.Addr: %w", err, v)
		}
		n.V = addr
		return nil
	default:
		n.Valid = false
		return fmt.Errorf("nullnetip: cannot scan type %T into NullNetIP", src)
	}
}

// Value implements the driver.Valuer interface.
// It returns nil if the value is not valid, otherwise it returns a byte slice
// representation of the netip.Addr for database storage.
func (n NullNetAddr) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.V.String(), nil
}
