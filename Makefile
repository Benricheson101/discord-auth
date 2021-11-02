SERVER := ./server.go
SERVER_BINARY := server

PKG := $(shell find . -type f -name *.go)

$(SERVER_BINARY): $(SERVER) $(PKG)
	go build -o $(SERVER_BINARY) $(SERVER)

build: $(SERVER_BINARY)

clean:
	@go clean
	@rm -f $(SERVER_BINARY)

.PHONY: dev
dev:
	go run $(SERVER)
