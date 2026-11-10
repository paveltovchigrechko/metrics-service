package a

import (
	"time"

	"example.com/resettest/b"
	xb "example.com/resettest/x/b"
)

type ID int

// generate:reset
type ResetableStruct struct {
	i     int
	str   string
	strP  *string
	s     []int
	m     map[string]string
	child *ResetableStruct
}

// generate:reset
type Everything struct {
	*b.Inner
	_      int
	id     ID
	ok     bool
	f      float64
	at     time.Time
	inner  b.Inner
	manual b.Manual
	other  xb.Plain
	sp     *[]int
	pp     **int
	arr    [3]int
	iface  any
	fn     func()
	ch     chan int
	anon   struct{ A int }
	value  ResetableStruct
}

// The receiver of Box cannot be "b": that would shadow package b in the body.
//
// generate:reset
type Box[T any] struct {
	v     T
	list  []T
	plain b.Plain
}

type (
	// generate:reset
	Grouped struct{ n int }

	Unmarked struct{ n int }
)

// generate:reset
type NotAStruct int

// generate:reset

type SeparatedByBlankLine struct{ n int }
