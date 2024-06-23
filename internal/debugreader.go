package internal

import (
	"fmt"
	"io"
	"log/slog"
)

type DebugReader struct {
	Name string
	io.Reader
}

func WrapWithDebugReader(name string, reader io.Reader) io.Reader {
	teeReader := io.TeeReader(reader, &writePrintHook{
		name: name,
	})

	return &DebugReader{
		Name:   name,
		Reader: teeReader,
	}
}

type writePrintHook struct {
	name string
}

func (w *writePrintHook) Write(p []byte) (int, error) {
	n := len(p)
	slog.Debug("Write "+fmt.Sprint(n)+" bytes to "+w.name, "length", n)
	return n, nil
}
