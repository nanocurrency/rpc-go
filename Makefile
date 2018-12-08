export GOPATH=${CURDIR}
all: build
build:
	@cd src/cmd/rpc-go && go install && cd ${GOPATH} && echo Run "bin/rpc-go" to start the RPC server
clean:
	rm bin/rpc-go
