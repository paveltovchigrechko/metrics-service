package pool

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// item has a Reset method like the ones cmd/reset generates.
type item struct {
	n      int
	tags   []string
	resets int
}

func (i *item) Reset() {
	if i == nil {
		return
	}
	i.n = 0
	i.tags = i.tags[:0]
	i.resets++
}

func TestGet_CreatesObjectWhenEmpty(t *testing.T) {
	var created atomic.Int32
	p := New(func() *item {
		created.Add(1)
		return &item{n: 42}
	})

	got := p.Get()

	require.NotNil(t, got)
	assert.Equal(t, 42, got.n, "a new object comes from the constructor as is")
	assert.Equal(t, int32(1), created.Load())
}

func TestGet_WithoutConstructorReturnsZero(t *testing.T) {
	p := New[*item](nil)

	assert.Nil(t, p.Get())
}

func TestPut_ResetsObject(t *testing.T) {
	p := New(func() *item { return &item{} })
	it := &item{n: 7, tags: make([]string, 2, 8)}

	p.Put(it)

	assert.Equal(t, 0, it.n)
	assert.Empty(t, it.tags)
	assert.Equal(t, 8, cap(it.tags), "Reset keeps the slice capacity")
	assert.Equal(t, 1, it.resets)
}

func TestPut_IgnoresNil(t *testing.T) {
	p := New(func() *item { return &item{} })

	p.Put(nil)

	assert.NotNil(t, p.Get(), "a nil must never be handed out")
}

func TestGet_ReusesPutObjects(t *testing.T) {
	p := New(func() *item { return &item{} })

	// sync.Pool may drop objects, and it does so on purpose under the race
	// detector, so one Put and Get is not guaranteed to reuse the object.
	reused := false
	for range 100 {
		it := &item{n: 1}
		p.Put(it)

		if got := p.Get(); got == it {
			reused = true
			assert.Equal(t, 0, got.n, "a reused object is reset")
			break
		}
	}

	assert.True(t, reused, "the pool never returned a put object")
}

func TestPool_Concurrent(t *testing.T) {
	p := New(func() *item { return &item{} })

	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for i := range 1000 {
				it := p.Get()
				if it.n != 0 || len(it.tags) != 0 {
					t.Errorf("got an object that was not reset: %+v", it)
					return
				}
				it.n = i
				it.tags = append(it.tags, "tag")
				p.Put(it)
			}
		})
	}
	wg.Wait()
}

func TestIsNil(t *testing.T) {
	var nilItem *item
	var nilResetter Resetter

	assert.True(t, isNil(nilItem))
	assert.True(t, isNil(nilResetter))
	assert.False(t, isNil(&item{}))
	assert.False(t, isNil(Resetter(&item{})))
	assert.False(t, isNil(value{}))
}

// value has a value receiver, so T is not a pointer.
type value struct{}

func (value) Reset() {}

func BenchmarkPool(b *testing.B) {
	p := New(func() *item { return &item{tags: make([]string, 0, 16)} })

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			it := p.Get()
			it.tags = append(it.tags, "a", "b")
			p.Put(it)
		}
	})
}
