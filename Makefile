.PHONY: test
test:
	go test -v -race -cover -tags=debug ./internal
	go test -v -race -cover -tags=debug ./pkg/frame
	go test -v -race -cover -tags=debug ./pkg/pipe
	go test -v -race -cover -tags=debug ./pkg/rpc
	go test -v -race -cover -tags=debug ./pkg/socket

.PHONY: fuzz
fuzz:
	go test -fuzz=FuzzReadFrame -fuzztime=10s ./pkg/frame
	go test -fuzz=FuzzReadOptions -fuzztime=10s ./pkg/frame
	go test -fuzz=FuzzVerifyCRC -fuzztime=10s ./pkg/frame
	go test -fuzz=FuzzFrameRoundTrip -fuzztime=10s ./pkg/frame
	go test -fuzz=FuzzReceiveFrame -fuzztime=10s ./internal
