package internal

import "io"

type Forwarder interface {
	// initalize from main goroutine
	Init() error

	// Forwards Data from DownstreamIngress to Upstream
	// Stream received Data from Upstream into DownstreamEgress
	// launched from a dedicated goroutine
	// method return via chan's
	Forward(params *ForwarderParams)
}

type ForwarderParams struct {
	// Forwarder Reads RequestData from DownstreamIngress
	DownstreamIngressReader io.ReadCloser

	// Forwarder Writes ResponseData to listener
	DownstreamEgressWriter io.WriteCloser

	Done          chan bool
	DoneWithError chan error
}
