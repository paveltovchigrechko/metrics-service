package a

import (
	"testing"
	"time"

	"example.com/resettest/b"
	xb "example.com/resettest/x/b"
)

func TestResetableStruct(t *testing.T) {
	str := "x"
	child := &ResetableStruct{i: 7, str: "c"}
	rs := &ResetableStruct{
		i:     1,
		str:   "s",
		strP:  &str,
		s:     []int{1, 2, 3},
		m:     map[string]string{"k": "v"},
		child: child,
	}

	rs.Reset()

	if rs.i != 0 || rs.str != "" {
		t.Errorf("primitives not reset: %+v", rs)
	}
	if rs.strP != &str || str != "" {
		t.Errorf("pointee not reset: %q", str)
	}
	if len(rs.s) != 0 || cap(rs.s) != 3 {
		t.Errorf("slice: len %d cap %d, want len 0 cap 3", len(rs.s), cap(rs.s))
	}
	if rs.m == nil || len(rs.m) != 0 {
		t.Errorf("map not cleared: %v", rs.m)
	}
	if rs.child != child || child.i != 0 || child.str != "" {
		t.Errorf("child not reset: %+v", child)
	}
}

func TestResetHandlesNil(t *testing.T) {
	var rs *ResetableStruct
	rs.Reset()

	(&ResetableStruct{}).Reset()
	(&Everything{}).Reset()
}

func TestEverything(t *testing.T) {
	n := 5
	np := &n
	sl := []int{1, 2}
	e := &Everything{
		Inner:  &b.Inner{N: 1, Tags: []string{"a"}},
		id:     3,
		ok:     true,
		f:      1.5,
		at:     time.Now(),
		inner:  b.Inner{N: 2, Tags: []string{"a", "b"}},
		manual: b.Manual{Calls: 3},
		other:  xb.Plain{Y: 4},
		sp:     &sl,
		pp:     &np,
		arr:    [3]int{1, 2, 3},
		iface:  1,
		fn:     func() {},
		ch:     make(chan int),
		value:  ResetableStruct{i: 9},
	}
	e.anon.A = 4

	e.Reset()

	if e.Inner.N != 0 || len(e.Inner.Tags) != 0 {
		t.Errorf("embedded pointer not reset: %+v", e.Inner)
	}
	if e.id != 0 || e.ok || e.f != 0 {
		t.Errorf("primitives not reset: id %d ok %v f %v", e.id, e.ok, e.f)
	}
	if !e.at.IsZero() || e.other != (xb.Plain{}) || e.arr != [3]int{} || e.anon.A != 0 {
		t.Errorf("zero values not set: at %v other %v arr %v anon %v", e.at, e.other, e.arr, e.anon)
	}
	if e.inner.N != 0 || len(e.inner.Tags) != 0 || cap(e.inner.Tags) != 2 {
		t.Errorf("marked struct not reset with Reset: %+v", e.inner)
	}
	if e.manual.Calls != -1 {
		t.Errorf("own Reset not called: %+v", e.manual)
	}
	if e.sp != &sl || len(sl) != 0 || cap(sl) != 2 {
		t.Errorf("pointer to slice: len %d cap %d, want len 0 cap 2", len(sl), cap(sl))
	}
	if e.pp != &np || n != 0 {
		t.Errorf("pointer to pointer not reset: %d", n)
	}
	if e.iface != nil || e.fn != nil || e.ch != nil {
		t.Errorf("interface, func and chan not set to nil")
	}
	if e.value.i != 0 {
		t.Errorf("struct value not reset: %+v", e.value)
	}
}

func TestBox(t *testing.T) {
	box := &Box[string]{v: "v", list: []string{"a"}, plain: b.Plain{X: 1}}

	box.Reset()

	if box.v != "" || len(box.list) != 0 || cap(box.list) != 1 || box.plain != (b.Plain{}) {
		t.Errorf("box not reset: %+v", box)
	}
}

func TestGrouped(t *testing.T) {
	g := &Grouped{n: 1}
	g.Reset()
	if g.n != 0 {
		t.Errorf("grouped not reset: %+v", g)
	}
}
