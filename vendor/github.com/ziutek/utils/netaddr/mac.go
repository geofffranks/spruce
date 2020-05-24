package netaddr

import (
	"fmt"
	"strconv"
	"strings"
)

type MAC uint64

// ParseMAC returns 0 MAC if error
func ParseMAC(s string) MAC {
	s = strings.Map(
		func(r rune) rune {
			switch r {
			case '-', '.', ':':
				return -1
			}
			return r
		},
		s,
	)
	if len(s) != 12 {
		return 0
	}
	m, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		return 0
	}
	return MAC(m)
}

// String return string representation of m in form hh-hh-hh-hh-hh-hh, where h
// is hexadecimal digit: 0-9,a-f
func (m MAC) String() string {
	return fmt.Sprintf(
		"%02x-%02x-%02x-%02x-%02x-%02x",
		byte(m>>40), byte(m>>32), byte(m>>24), byte(m>>16), byte(m>>8), byte(m),
	)
}

// ColonString return string representation of m in form hh:hh:hh:hh:hh:hh,
// where h is hexadecimal digit: 0-9,a-f
func (m MAC) ColonString() string {
	return fmt.Sprintf(
		"%02x:%02x:%02x:%02x:%02x:%02x",
		byte(m>>40), byte(m>>32), byte(m>>24), byte(m>>16), byte(m>>8), byte(m),
	)
}

// PlainString return string representation of m in form hhhhhhhhhhhh, where h
// is hexadecimal digit: 0-9,a-f
func (m MAC) PlainString() string {
	return fmt.Sprintf("%012x", uint64(m))
}

// CiscoString return string representation of m in form hhhh.hhhh.hhhh,
// where h is hexadecimal digit: 0-9,a-f
func (m MAC) CiscoString() string {
	return fmt.Sprintf(
		"%02x%02x.%02x%02x.%02x%02x",
		byte(m>>40), byte(m>>32), byte(m>>24), byte(m>>16), byte(m>>8), byte(m),
	)
}
