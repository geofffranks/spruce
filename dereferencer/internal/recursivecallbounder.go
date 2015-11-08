package internal

import "fmt"

const maxRecursiveCallLimit = 64

var ErrInfiniteRecursion = fmt.Errorf("possible infinite recursion detected in dereferencing")

type RecursiveCallBounder struct {
	ttl int
}

func (b *RecursiveCallBounder) Call(f func() (interface{}, error)) (interface{}, error) {
	if b.ttl -= 1; b.ttl == 0 {
		return "", ErrInfiniteRecursion
	}
	defer func() { b.ttl += 1 }()
	return f()
}

func (b *RecursiveCallBounder) Reset() {
	b.ttl = maxRecursiveCallLimit
}
