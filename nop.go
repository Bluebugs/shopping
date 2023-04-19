package main

import "io"

type nopWriteCloser struct {
	io.Writer
}

var _ io.WriteCloser = (*nopWriteCloser)(nil)

func (nopWriteCloser) Close() error { return nil }
