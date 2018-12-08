# About

`rpc-go` is an external Nano RPC server written in Go. It receives JSON requests from clients
and forwards them to the node via its IPC mechanism, which supports domain sockets or tcp.

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

# Build

Clone the repository and build with the `go` tool.

```
git clone https://gitcom.com/nanocurrency/rpc-go
cd rpc-go

export GOPATH=`pwd`
cd src/cmd/rpc-go
go install
```

# Run
Configure config.json and run:

```
bin/rpc-go
```

# IDE notes for developers

If using Visual Studio Code, setting `go.inferGopath` to true is recommended. This will add the current workspace path to GOPATH.

A debugger is available via `go get -u github.com/derekparker/delve/cmd/dlv`