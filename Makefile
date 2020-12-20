test:
	go test -v -race -cover -tags=debug ./pkg/frame
	go test -v -race -cover -tags=debug ./pkg/pipe
	go test -v -race -cover -tags=debug ./pkg/rpc
	go test -v -race -cover -tags=debug ./pkg/socket
