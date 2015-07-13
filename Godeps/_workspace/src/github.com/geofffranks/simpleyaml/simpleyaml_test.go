package simpleyaml

import (
	"testing"
)

var data = []byte(`
name: smallfish
age: 99
bool: true
emails:
   - xxx@xx.com
   - yyy@yy.com
bb:
    cc:
        dd:
            - 111
            - 222
            - 333
        ee: aaa
`)

func TestBool(t *testing.T) {
	y, err := NewYaml(data)
	if err != nil {
		t.Fatal("init yaml failed")
	}
	v, err := y.Get("bool").Bool()
	if err != nil {
		t.Fatal("get yaml failed")
	}
	t.Log(v)
	if v != true {
		t.Fatal("match bool failed")
	}
}

func TestString(t *testing.T) {
	y, err := NewYaml(data)
	if err != nil {
		t.Fatal("init yaml failed")
	}
	v, err := y.Get("name").String()
	if err != nil {
		t.Fatal("get yaml failed")
	}
	t.Log(v)
	if v != "smallfish" {
		t.Fatal("match name failed")
	}
}

func TestInt(t *testing.T) {
	y, err := NewYaml(data)
	if err != nil {
		t.Fatal("init yaml failed")
	}
	v, err := y.Get("age").Int()
	if err != nil {
		t.Fatal("get yaml failed")
	}
	t.Log(v)
	if v != 99 {
		t.Fatal("match age failed")
	}
}

func TestGetIndex(t *testing.T) {
	y, err := NewYaml(data)
	if err != nil {
		t.Fatal("init yaml failed")
	}
	v, err := y.Get("bb").Get("cc").Get("dd").GetIndex(1).Int()
	t.Log(v)
	if err != nil {
		t.Fatal("match bb.cc.ee[1] failed")
	}
}

func TestString2(t *testing.T) {
	y, err := NewYaml(data)
	if err != nil {
		t.Fatal("init yaml failed")
	}
	v, err := y.Get("bb").Get("cc").Get("ee").String()
	t.Log(v)
	if err != nil {
		t.Fatal("match bb.cc.ee failed")
	}
	if v != "aaa" {
		t.Fatal("bb.cc.ee not equal bbb")
	}
}

func TestGetPath(t *testing.T) {
	y, err := NewYaml(data)
	if err != nil {
		t.Fatal("init yaml failed")
	}
	v, err := y.GetPath("bb", "cc", "ee").String()
	if err != nil {
		t.Fatal("get yaml failed")
	}
	t.Log(v)
	if v != "aaa" {
		t.Fatal("aa.bb.cc.ee not equal bbb")
	}
}

func TestArray(t *testing.T) {
	y, err := NewYaml(data)
	if err != nil {
		t.Fatal("init yaml failed")
	}
	v, err := y.Get("emails").Array()
	if err != nil {
		t.Fatal("get yaml failed")
	}
	t.Log(v)
	if len(v) != 2 {
		t.Fatal("emails length not equal 2")
	}
}
