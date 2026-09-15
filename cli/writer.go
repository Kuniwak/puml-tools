package cli

import "io"

type NopWriteCloser struct {
	w io.Writer
}

func NewNopWriteCloser(w io.Writer) *NopWriteCloser {
	return &NopWriteCloser{w: w}
}

func (n *NopWriteCloser) Write(p []byte) (int, error) {
	return n.w.Write(p)
}

func (n *NopWriteCloser) Close() error {
	return nil
}
