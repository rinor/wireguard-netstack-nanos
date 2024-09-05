.PHONY: all clean tidy lint help

.DEFAULT_GOAL := all

GO_TAGS := netgo,osusergo
LD_FLAGS := -s -w
GO_BUILD := go build -trimpath -tags "${GO_TAGS}" -ldflags '$(LD_FLAGS)'


all: netstack klib

# netstack builds
netstack: netstack_client_http netstack_server_http

netstack_client_http: tidy
	@$(GO_BUILD) -o bin/netstack_client_http cmd/netstack/client_http/client_http.go

netstack_server_http: tidy
	@$(GO_BUILD) -o bin/netstack_server_http cmd/netstack/server_http/server_http.go

# nanos tun klib builds
klib: klib_client_http klib_server_http

klib_client_http: tidy
	@$(GO_BUILD) -o bin/klib_client_http cmd/nanos_klib/client_http/client_http.go

klib_server_http: tidy
	@$(GO_BUILD) -o bin/klib_server_http cmd/nanos_klib/server_http/server_http.go

# chores
clean:
	@$(RM) bin/*_http

tidy:
	@go mod tidy

lint:
	@golangci-lint run ./...

help:
	@echo "Available targets:"
	@echo ""
	@echo "  all                    - Build all binaries"
	@echo ""
	@echo "  netstack               - Build all netstack binaries"
	@echo "    netstack_client_http - Build the netstack client HTTP binary"
	@echo "    netstack_server_http - Build the netstack server HTTP binary"
	@echo ""
	@echo "  klib                   - Build all klib binaries"
	@echo "    klib_client_http     - Build the klib client HTTP binary"
	@echo "    klib_server_http     - Build the klib server HTTP binary"
	@echo ""
	@echo "  clean                  - Remove built binaries"
	@echo "  tidy                   - Clean up go.mod and go.sum"
	@echo "  lint                   - Run linter"