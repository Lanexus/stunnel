.PHONY: build test clean install

build:
	go build -o bin/stunnel ./cmd/stunnel/
	go build -o bin/relay ./cmd/relay/

test:
	go test ./... -v

clean:
	rm -rf bin/

install:
	go install ./cmd/stunnel/
