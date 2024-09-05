module github.com/rinor/wireguard-netstack-nanos

go 1.23.0

require golang.zx2c4.com/wireguard v0.0.0-20231211153847-12269c276173

require (
	github.com/google/btree v1.0.1 // indirect
	golang.org/x/crypto v0.13.0 // indirect
	golang.org/x/net v0.15.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
	golang.org/x/time v0.0.0-20220210224613-90d013bbcef8 // indirect
	golang.zx2c4.com/wintun v0.0.0-20230126152724-0fa3db229ce2 // indirect
	gvisor.dev/gvisor v0.0.0-20230927004350-cbd86285d259 // indirect
)

replace golang.zx2c4.com/wireguard => github.com/rinor/wireguard-go v0.0.0-20240902093400-e917c9f2e2f3
