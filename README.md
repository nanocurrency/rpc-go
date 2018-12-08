# About

`rpc-go` is an external Nano RPC server written in Go. It receives JSON requests from clients and forwards them to the node via its IPC mechanism. The node currently supports domain sockets and TCP.

# Build

Clone the repository and run `make`

If you want to build with the go tool manually, see Makefile for relevant commands.

# Run
Change `config.json` as needed and run:

```
bin/rpc-go
```

# Configuration

Sample `config.json`:

```
{
	"port": 8080,
	"hostname": "localhost",
	"node": {
		"connection": "local:///tmp/nano",
		"poolsize": 10
	}
}
```

The port and hostname is where the *Web server* will listen for RPC POST requests.

The *node* configuration tells rpc-go where to find the Nano IPC server. The poolsize sets the maximum number of concurrent connections to the node. These are persistent connections to improve performance, but rpc-go will automatically reconnect if needed.

Replace the connection line with something like the following to use TCP instead of domain sockets:

```
    "connection": "tcp://localhost:7077"
```

The host and port must match the IPC configuration of the Nano node.

# IDE notes for developers

If using Visual Studio Code, setting `go.inferGopath` to true is recommended. This will add the current workspace path to GOPATH.

A debugger is available via `go get -u github.com/derekparker/delve/cmd/dlv`