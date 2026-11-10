package pool_test

import (
	"fmt"

	"github.com/paveltovchigrechko/metrics-service/internal/pool"
)

// Buffer stands for a struct marked with "// generate:reset"; its Reset method
// is what cmd/reset would generate.
type Buffer struct {
	Data []byte
}

func (b *Buffer) Reset() {
	if b == nil {
		return
	}
	b.Data = b.Data[:0]
}

func Example() {
	buffers := pool.New(func() *Buffer {
		return &Buffer{Data: make([]byte, 0, 1024)}
	})

	buf := buffers.Get()
	buf.Data = append(buf.Data, "hello"...)
	fmt.Println(string(buf.Data))

	// Put resets the buffer; the next Get may return it with its capacity kept.
	buffers.Put(buf)
	fmt.Println(len(buf.Data), cap(buf.Data))

	// Output:
	// hello
	// 0 1024
}
