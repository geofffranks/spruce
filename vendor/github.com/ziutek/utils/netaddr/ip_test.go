package netaddr

import (
	"net"
	"testing"
)

type exampleIPAdd struct {
	ip     string
	offset int
	exp    string
}

const ipv6max = "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"

var testIPAdd = []exampleIPAdd{
	{"0.0.0.0", 1, "0.0.0.1"},
	{"0.0.0.0", -1, "255.255.255.255"},
	{"255.255.255.255", 1, "0.0.0.0"},
	{"255.255.255.255", -1, "255.255.255.254"},
	{"0.0.0.0", 256, "0.0.1.0"},
	{"0.0.0.0", -256, "255.255.255.0"},
	{"255.255.255.255", 256, "0.0.0.255"},
	{"255.255.255.255", -256, "255.255.254.255"},

	{"::", 1, "::1"},
	{"::", -1, ipv6max},
	{ipv6max, 1, "::"},
	{ipv6max, -1, "ffff:ffff:ffff:ffff:ffff:ffff:ffff:fffe"},
	{"::ffff:ffff:ffff:0000", 0x10000, "::1:0000:0000:0000:0000"},
	{"::ffff:ffff:ffff:0000", -0x10000, "::ffff:ffff:fffe:0000"},
}

func TestIPAdd(t *testing.T) {
	for _, e := range testIPAdd {
		a := net.ParseIP(e.ip)
		b := IPAdd(a, e.offset)
		if !b.Equal(net.ParseIP(e.exp)) {
			t.Errorf("IPAdd(%s, %d)=%s != %s", e.ip, e.offset, b, e.exp)
		}
	}
}

type exampleIPMod struct {
	ip  string
	d   uint
	exp uint
}

var testIPMod = []exampleIPMod{
	{"192.168.1.246", 100, 22},
	{"192.168.200.1", 10000, 6721},

	{"1234::1222:aaaa", 12652, 4262},
	{"1234::1222:aaaa", 1e4, 9162},
	{"1234::1222:aaaa", 1e5, 9162},
	{"1234::1222:aaaa", 1e6, 909162},
	{"1234:1111:1::1222:aaaa", 123456789, 12319868},
	{"1234:1111:1::1222:aaaa", 1e7, 3633322},
	{"1234:1111:1::1222:aaaa", 1e8, 63633322},
}

func TestIPMod(t *testing.T) {
	for _, e := range testIPMod {
		a := net.ParseIP(e.ip)
		b := IPMod(a, e.d)
		if b != e.exp {
			t.Errorf("IPMod(%s, %d)=%d != %d", e.ip, e.d, b, e.exp)
		}
	}
}
