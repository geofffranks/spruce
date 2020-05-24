package ciede2000

import (
	"image/color"
	"testing"
)

func TestDiff(t *testing.T) {
	c1 := &color.RGBA{200, 255, 0, 255}
	c2 := &color.RGBA{50, 60, 255, 255}
	value := Diff(c1, c2)
	expected := 101.84978806864332
	if value != expected {
		t.Fatalf("Expected %v, but %v:", expected, value)
	}
}
