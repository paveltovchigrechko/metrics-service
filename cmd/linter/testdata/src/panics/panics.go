package panics

import "errors"

func withString() {
	panic("boom") // want `usage of panic function`
}

func withError() {
	panic(errors.New("boom")) // want `usage of panic function`
}

func inClosure() {
	f := func() {
		panic("boom") // want `usage of panic function`
	}
	f()
}

func recovers() {
	defer func() { _ = recover() }()
}

type thrower struct{}

func (thrower) panic(string) {}

func method() {
	thrower{}.panic("not the builtin")
}

func shadowed() {
	panic := func(string) {}
	panic("not the builtin")
}
