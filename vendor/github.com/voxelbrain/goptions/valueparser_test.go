package goptions

import (
	"net"
	"net/url"
	"os"
	"testing"
	"time"
)

func TestParse_File(t *testing.T) {
	var args []string
	var err error
	var fs *FlagSet
	var options struct {
		Output *os.File `goptions:"-o, create, trunc, wronly"`
	}

	args = []string{"-o", "testfile"}
	fs = NewFlagSet("goptions", &options)
	err = fs.Parse(args)
	if err != nil {
		t.Fatalf("Parsing failed: %s", err)
	}
	if !(options.Output != nil) {
		t.Fatalf("Unexpected value: %#v", options)
	}
	options.Output.Close()
	os.Remove("testfile")
}

func TestParse_TCPAddr(t *testing.T) {
	var args []string
	var err error
	var fs *FlagSet
	var options struct {
		Server *net.TCPAddr `goptions:"-a"`
	}

	args = []string{"-a", "192.168.0.100:8080"}
	fs = NewFlagSet("goptions", &options)
	err = fs.Parse(args)
	if err != nil {
		t.Fatalf("Parsing failed: %s", err)
	}
	if !(options.Server.IP.String() == "192.168.0.100" &&
		options.Server.Port == 8080) {
		t.Fatalf("Unexpected value: %#v", options)
	}
}

func TestParse_URL(t *testing.T) {
	var args []string
	var err error
	var fs *FlagSet
	var options struct {
		Server *url.URL `goptions:"-a"`
	}

	args = []string{"-a", "http://www.google.com"}
	fs = NewFlagSet("goptions", &options)
	err = fs.Parse(args)
	if err != nil {
		t.Fatalf("Parsing failed: %s", err)
	}
	if !(options.Server.Scheme == "http" &&
		options.Server.Host == "www.google.com") {
		t.Fatalf("Unexpected value: %#v", options.Server)
	}
}

func TestParse_Duration(t *testing.T) {
	var args []string
	var err error
	var fs *FlagSet
	var options struct {
		Cache time.Duration `goptions:"-d"`
	}

	args = []string{"-d", "1h45m"}
	fs = NewFlagSet("goptions", &options)
	err = fs.Parse(args)
	if err != nil {
		t.Fatalf("Parsing failed: %s", err)
	}
	if !(int64(options.Cache) != (1*60+45)*60*1e12) {
		t.Fatalf("Unexpected value: %#v", options.Cache)
	}
}
