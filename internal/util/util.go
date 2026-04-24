package util

import (
	"io"
)

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

func TrimJson(s []byte) []byte {
	idx1 := 0

	for i, b := range s {
		if b == '{' {
			idx1 = i
		}
	}

	idx2 := len(s) - 1
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '}' {
			idx2 = i
		}
	}

	return s[idx1 : idx2+1]
}
