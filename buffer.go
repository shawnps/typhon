package typhon

import (
	"bytes"
	"io"
	"sync"
)

type bufCloser struct {
	bytes.Buffer
}

func (b *bufCloser) Close() error {
	return nil // No-op
}

type streamer struct {
	pipeR *io.PipeReader
	pipeW *io.PipeWriter
}

// Streamer returns a reader/writer/closer that can be used to stream service responses. It does not necessarily
// perform internal buffering, so users should take care not to depend on such behaviour.
func Streamer() io.ReadWriteCloser {
	pipeR, pipeW := io.Pipe()
	return &streamer{
		pipeR: pipeR,
		pipeW: pipeW}
}

func (s *streamer) Read(p []byte) (int, error) {
	return s.pipeR.Read(p)
}

func (s *streamer) Write(p []byte) (int, error) {
	return s.pipeW.Write(p)
}

func (s *streamer) Close() error {
	return s.pipeW.Close()
}

// countingWriter is a writer which proxies writes to an underlying io.Writer, keeping track of how many bytes have
// been written in total
type countingWriter struct {
	n int
	io.Writer
}

func (w *countingWriter) Write(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	w.n += n
	return n, err
}

// doneReader is a wrapper around a ReadCloser which provides notification when the stream has been fully consumed
// (ie. when EOF is reached, or when the reader is closed.)
type doneReader struct {
	done     chan struct{}
	doneOnce sync.Once
	io.ReadCloser
}

func newDoneReader(r io.ReadCloser) *doneReader {
	return &doneReader{
		done:       make(chan struct{}),
		ReadCloser: r}
}

func (r *doneReader) Close() error {
	err := r.ReadCloser.Close()
	r.doneOnce.Do(func() { close(r.done) })
	return err
}

func (r *doneReader) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	if err == io.EOF {
		r.doneOnce.Do(func() { close(r.done) })
	}
	return n, err
}