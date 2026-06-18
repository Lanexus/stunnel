.PHONY: build test clean install

build:
	go build -o bin/stunnel ./cmd/stunnel/

test:
	go test ./... -v

clean:
	rm -rf bin/

install:
	go install ./cmd/stunnel/
