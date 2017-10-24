package netaddr

import (
	"math"
	"net"
)

func isZeros(p net.IP) bool {
	for _, b := range p {
		if b != 0 {
			return false
		}
	}
	return true
}

// IsIPv4 returns true if ip is IPv4 address.
func IsIPv4(ip net.IP) bool {
	return len(ip) == net.IPv4len ||
		isZeros(ip[0:10]) && ip[10] == 0xff && ip[11] == 0xff
}

func ipToI32(ip net.IP) int32 {
	ip = ip.To4()
	return int32(ip[0])<<24 | int32(ip[1])<<16 | int32(ip[2])<<8 | int32(ip[3])
}

func i32ToIP(a int32) net.IP {
	return net.IPv4(byte(a>>24), byte(a>>16), byte(a>>8), byte(a))
}

func ipToU64(ip net.IP) uint64 {
	return uint64(ip[0])<<56 | uint64(ip[1])<<48 | uint64(ip[2])<<40 |
		uint64(ip[3])<<32 | uint64(ip[4])<<24 | uint64(ip[5])<<16 |
		uint64(ip[6])<<8 | uint64(ip[7])
}

func u64ToIP(ip net.IP, a uint64) {
	ip[0] = byte(a >> 56)
	ip[1] = byte(a >> 48)
	ip[2] = byte(a >> 40)
	ip[3] = byte(a >> 32)
	ip[4] = byte(a >> 24)
	ip[5] = byte(a >> 16)
	ip[6] = byte(a >> 8)
	ip[7] = byte(a)
}

// IPAdd adds offset to ip
func IPAdd(ip net.IP, offset int) net.IP {
	if IsIPv4(ip) {
		a := int(ipToI32(ip[len(ip)-4:]))
		return i32ToIP(int32(a + offset))
	}
	a := ipToU64(ip[:net.IPv6len/2])
	b := ipToU64(ip[net.IPv6len/2:])
	o := uint64(offset)
	if math.MaxUint64-b < o {
		a++
	}
	b += o
	if offset < 0 {
		a += math.MaxUint64
	}
	ip = make(net.IP, net.IPv6len)
	u64ToIP(ip[:net.IPv6len/2], a)
	u64ToIP(ip[net.IPv6len/2:], b)
	return ip
}

// IPMod calculates ip % d
func IPMod(ip net.IP, d uint) uint {
	if IsIPv4(ip) {
		return uint(ipToI32(ip[len(ip)-4:])) % d
	}
	b := uint64(d)
	hi := ipToU64(ip[:net.IPv6len/2])
	lo := ipToU64(ip[net.IPv6len/2:])
	return uint(((hi%b)*((0-b)%b) + lo%b) % b)
}
