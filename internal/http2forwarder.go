package internal

import (
	"crypto/tls"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/http2"
)

type Http2Forwarder struct {
	URL       *url.URL
	transport *http2.Transport
}

func (f *Http2Forwarder) Init() error {
	slog.Debug("Sending traffic through this proxy " + f.URL.Host)

	noLimit := 0 * time.Second
	f.transport = &http2.Transport{
		AllowHTTP: true,
		// maximum amount of time an idle (keep-alive) connection will remain idle before closing
		IdleConnTimeout: noLimit,
		// timeout after which a health check using ping is sent
		ReadIdleTimeout: 60 * time.Second,
		DialTLS:         f.dial,
	}

	err := f.send(nil, nil)
	if err != nil {
		return err
	}
	return nil
}

func (f *Http2Forwarder) Forward(params *ForwarderParams) {
	err := f.send(params.DownstreamIngressReader, params.DownstreamEgressWriter)
	if err != nil {
		slog.Error(err.Error(), slog.Any("error", err))
		params.DoneWithError <- err
		return
	}
	params.Done <- true
	return
}

func (f *Http2Forwarder) send(downstreamIngressReader io.ReadCloser, downstreamEgressWriter io.WriteCloser) error {
	req := &http.Request{
		Method: "CONNECT",
		URL:    f.URL,
		// DownstreamIngressReader --> Request.Body.Write()
		Body: downstreamIngressReader,
	}

	// Send connect request
	res, err := f.transport.RoundTrip(req)
	if err != nil {
		return errors.Wrap(err, "Error during sending http2 connect request")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return errors.New("Response Status was '" + res.Status + "', but expected 200!")
	}

	if downstreamEgressWriter != nil {
		// Response.Body.Read() --> DownstreamEgressWriter
		_, err = io.Copy(downstreamEgressWriter, res.Body)
		if err != nil {
			return errors.Wrap(err, "Error writing upstream response to downstream")
		}
	}
	return nil
}

func (f *Http2Forwarder) dial(netw, addr string, cfg *tls.Config) (net.Conn, error) {
	conn, err := net.Dial(netw, addr)
	if err != nil {
		return nil, err
	}
	return WrapWithDebugConnection("Upstream", conn), nil
}
