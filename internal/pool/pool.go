// Package pool provides a typed pool of reusable objects that are reset
// before they are reused.
//
// The objects must have a Reset method, such as the one cmd/reset generates
// for structs marked with a "// generate:reset" comment.
package pool

import (
	"reflect"
	"sync"
)

// Resetter is an object that can restore its initial state for reuse.
type Resetter interface {
	Reset()
}

// Pool is a set of reusable objects of type T, usually a pointer to a struct.
// It is safe for concurrent use.
//
// Pool is built on sync.Pool: the garbage collector may drop pooled objects at
// any time, so Get creates a new object when the pool is empty.
type Pool[T Resetter] struct {
	pool sync.Pool
}

// New returns an empty Pool. newFn creates an object when the pool has none to
// give out. If newFn is nil, Get returns the zero value of T in that case.
func New[T Resetter](newFn func() T) *Pool[T] {
	p := &Pool[T]{}
	if newFn != nil {
		p.pool.New = func() any { return newFn() }
	}

	return p
}

// Get takes an object from the pool, or creates a new one if the pool is
// empty. Objects from the pool were reset by Put.
func (p *Pool[T]) Get() T {
	v, ok := p.pool.Get().(T)
	if !ok {
		var zero T
		return zero
	}

	return v
}

// Put resets v and returns it to the pool. The caller must not use v after
// Put. A nil v is ignored, so Get never returns a nil from the pool.
func (p *Pool[T]) Put(v T) {
	if isNil(v) {
		return
	}

	v.Reset()
	p.pool.Put(v)
}

// isNil reports whether v is nil. T is usually a pointer, and a nil pointer
// stored in an interface is not equal to nil, so a plain comparison is not
// enough.
func isNil[T any](v T) bool {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return true
	}

	switch rv.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Slice, reflect.Interface, reflect.Func, reflect.Chan:
		return rv.IsNil()
	default:
		return false
	}
}
