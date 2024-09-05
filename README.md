# wireguard-netstack-nanos

https://github.com/nanovms/nanos/issues/2061

- `make help`

```sh
Available targets:

  all                    - Build all binaries

  netstack               - Build all netstack binaries
    netstack_client_http - Build the netstack client HTTP binary
    netstack_server_http - Build the netstack server HTTP binary

  klib                   - Build all klib binaries
    klib_client_http     - Build the klib client HTTP binary
    klib_server_http     - Build the klib server HTTP binary

  clean                  - Remove built binaries
  tidy                   - Clean up go.mod and go.sum
  lint                   - Run linter
```

- build the binaries (will be placed at `bin/` folder)

```sh
make
```

- `tree bin`

```sh
bin
├── klib_client_http
├── klib_server_http
├── netstack_client_http
└── netstack_server_http
```

- **start** *tun klib* based server

```sh
ops run -c config/klib_server_http.json bin/klib_server_http --smp=1
```

- **start** *tun klib* based client

```sh
ops run -c config/klib_client_http.json bin/klib_client_http --smp=1
```

- **stop** server

```sh
bash scripts/qemu_powerdown_server.sh
```

- **stop** client

```sh
bash scripts/qemu_powerdown_client.sh
```