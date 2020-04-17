test:
	composer update
	go test -v -race -cover

lint:
	go fmt ./...
	golint ./...
