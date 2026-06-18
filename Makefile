.PHONY: build test clean install

build:
	go build -o bin/stunnel ./cmd/stunnel/
	go build -o bin/signaling ./cmd/signaling/

test:
	go test ./... -v

clean:
	rm -rf bin/

install:
	go install ./cmd/stunnel/
