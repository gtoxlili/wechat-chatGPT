package util

import (
	"context"
	"io"
	"sync"
)

var bytesPool = sync.Pool{New: func() any { return make([]byte, 0, 1024) }}

func PutBytes(b []byte) {
	// reset the slice
	b = b[:0]
	bytesPool.Put(b)
}

func ReadWithCtx(ctx context.Context, r io.Reader) ([]byte, error) {
	b := bytesPool.Get().([]byte)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			if len(b) == cap(b) {
				b = append(b, 0)[:len(b)]
			}
			n, err := r.Read(b[len(b):cap(b)])
			b = b[:len(b)+n]
			if err != nil {
				if err == io.EOF {
					return b, nil
				}
				return b, err
			}
		}
	}
}
