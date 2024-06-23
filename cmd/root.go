package cmd

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"

	"github.com/denniskniep/http-connect-proxy-client/internal"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "http-connect-proxy-client",
	Short: "Proxy tcp connection via http2 connect server",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runRootCommand(); err != nil {
			return err
		}
		return nil
	},
}

var proxyServer string
var listenAddress string
var listenPort int
var debug bool

func Init() {
	rootCmd.Flags().StringVarP(&proxyServer, "proxyServer", "s", "", "Url of Http2 Connect Proxy Server: http://myServer:8080 (required)")
	_ = rootCmd.MarkFlagRequired("proxyServer")

	rootCmd.Flags().StringVarP(&listenAddress, "listenAddress", "l", "", "Listening Address (127.0.0.1 will be used if not specified)")
	_ = rootCmd.MarkFlagRequired("listenAddress")

	rootCmd.Flags().IntVarP(&listenPort, "listenPort", "p", 0, "Listening Port (A random highport will be used if not specified)")
	_ = rootCmd.MarkFlagRequired("listenPort")

	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging (false by default)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runRootCommand() error {
	logLevel := slog.LevelInfo
	addSource := false
	if debug {
		logLevel = slog.LevelDebug
		addSource = true
	}

	loggerOpts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: addSource,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, loggerOpts))
	slog.SetDefault(logger)

	url, err := url.Parse(proxyServer)
	if err != nil {
		slog.Error(err.Error(), slog.Any("error", err))
		os.Exit(1)
	}

	fwdListener := internal.NewForwardingListener(listenAddress, listenPort, &internal.Http2Forwarder{
		URL: url,
	})

	err = fwdListener.Start()
	if err != nil {
		slog.Error(err.Error(), slog.Any("error", err))
		os.Exit(1)
	}

	return nil
}
