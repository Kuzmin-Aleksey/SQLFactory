package util

import (
	"io"
)

func InsertFunc[S ~[]E, E any](s S, v E, f func(E) bool) S {
	if len(s) == 0 {
		return []E{v}
	}
	for i := 0; i < len(s); i++ {
		if f(s[i]) {
			return append(s[:i], append([]E{v}, s[i:]...)...)
		}
	}

	return append(s, v)
}

func Ptr[T any](v T) *T {
	return &v
}

type MultiReadCloser struct {
	readers []io.ReadCloser
	io.Reader
}

func (r *MultiReadCloser) Close() (err error) {
	for _, reader := range r.readers {
		if reader != nil {
			if e := reader.Close(); e != nil {
				err = e
			}
		}
	}
	return
}

func NewMultiReadCloser(readers ...io.ReadCloser) *MultiReadCloser {
	simpleReaders := make([]io.Reader, len(readers))
	for i, reader := range readers {
		simpleReaders[i] = reader
	}

	return &MultiReadCloser{
		readers: readers,
		Reader:  io.MultiReader(simpleReaders...),
	}
}
