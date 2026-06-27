.PHONY: build test lint docker run clean

BINARY := api-server

build:
	go build -ldflags="-s -w" -o $(BINARY) ./cmd/

test:
	go test ./... -v -count=1

lint:
	go vet ./...

docker:
	docker build -t rustdesk-api-server .

run: build
	./$(BINARY) serve

clean:
	rm -f $(BINARY)
	rm -rf data/
