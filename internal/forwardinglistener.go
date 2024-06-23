package internal

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"

	"github.com/pkg/errors"
)

type forwardingListener struct {
	address   string
	port      int
	forwarder Forwarder
}

func NewForwardingListener(address string, port int, forwarder Forwarder) *forwardingListener {
	return &forwardingListener{address: address, port: port, forwarder: forwarder}
}

func (f *forwardingListener) Start() error {
	err := f.forwarder.Init()
	if err != nil {
		return err
	}

	slog.Debug("Start listening on " + f.address + ":" + fmt.Sprint(f.port))
	ln, err := net.Listen("tcp", f.address+":"+fmt.Sprint(f.port))
	if err != nil {
		return err
	}
	slog.Info("Listening on " + ln.Addr().String())

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		slog.Info("Client connected: " + conn.RemoteAddr().String())

		go f.handleConnection(conn)
	}
}

func (f *forwardingListener) handleConnection(conn net.Conn) {
	listenerConn := WrapWithDebugConnection("Downstream", conn)
	defer listenerConn.Close()

	done := make(chan bool)
	doneWithError := make(chan error)

	downstreamIngress := createPipe()
	defer downstreamIngress.close()

	downstreamEgress := createPipe()
	defer downstreamEgress.close()

	params := &ForwarderParams{
		DownstreamIngressReader: downstreamIngress.reader,
		DownstreamEgressWriter:  downstreamEgress.writer,
		Done:                    done,
		DoneWithError:           doneWithError,
	}

	// DownstreamIngress --> UpstreamRequest
	// UpstreamResponse --> DownstreamEgress
	go f.forwarder.Forward(params)

	// listenerConn.Read() --> DownstreamIngress
	go streamFromListenerAsDownstreamIngress(listenerConn, downstreamIngress.writer, done, doneWithError)

	// DownstreamEgress --> listenerConn.Write()
	go streamFromForwarderAsDownstreamEgress(listenerConn, downstreamEgress.reader, done, doneWithError)

	select {
	case <-done:
	case <-doneWithError:
		if conn, ok := listenerConn.(*net.TCPConn); ok {
			conn.SetLinger(0)
		}
	}
}

func streamFromListenerAsDownstreamIngress(listenerConn net.Conn, downstreamIngressWriter *io.PipeWriter, done chan bool, doneWithError chan error) {
	defer func() {
		done <- true
	}()

	// listenerConn.Read() --> downstreamIngressWriter
	_, err := io.Copy(downstreamIngressWriter, listenerConn)
	if err != nil {
		err = errors.Wrap(err, "Error reading from downstream")
		slog.Error(err.Error(), slog.Any("error", err))
		doneWithError <- err
	}
}

func streamFromForwarderAsDownstreamEgress(listenerConn net.Conn, downstreamEgressReader *io.PipeReader, done chan bool, doneWithError chan error) {
	defer func() {
		done <- true
	}()

	// downstreamEgressReader --> listenerConn.Write()
	_, err := io.Copy(listenerConn, downstreamEgressReader)
	if err != nil {
		err = errors.Wrap(err, "Error writing to downstream")
		slog.Error(err.Error(), slog.Any("error", err))
		doneWithError <- err
	}
}

type pipe struct {
	reader *io.PipeReader
	writer *io.PipeWriter
}

func createPipe() *pipe {
	r, w := io.Pipe()
	return &pipe{
		reader: r,
		writer: w,
	}
}

func (p *pipe) close() {
	p.reader.Close()
	p.writer.Close()
}
