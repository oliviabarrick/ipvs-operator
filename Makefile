.PHONY: format test build

format:
	gofmt -w pkg/ cmd/

test:
	go test ./pkg/...

build: format test
	CGO_ENABLED=0 GOOS=linux go build -ldflags '-extldflags "-static"' ./cmd/ipvs-operator
