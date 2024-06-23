# Http Connect Proxy Client
Client to proxy tcp connection via http2 connect server

## Run 
```
hcpc -s http://127.0.0.1:4444 -l 127.0.0.1 -p 5555
```

## Command
```
Usage:
  hcpc [flags]

Flags:
  -d, --debug                  Enable debug logging (false by default)
  -h, --help                   help for hcpc
  -l, --listenAddress string   Listening Address (127.0.0.1 will be used if not specified)
  -p, --listenPort int         Listening Port (A random highport will be used if not specified)
  -s, --proxyServer string     Url of Http2 Connect Proxy Server: http://myServer:8080 (required)
```

## Install CLI
```
go install github.com/github.com/denniskniep/http-connect-proxy-client
```

## Development

### Run

```
go run .
```
