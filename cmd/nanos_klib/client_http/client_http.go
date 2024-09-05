package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
)

func main() {
	var (
		interfaceName = flag.String("interface", "wg2", "wg interface name")
		endpoint      = flag.String("endpoint", "10.0.2.2", "wg server endpoint")
		counter       = flag.Int("counter", 100, "amount of requests to send")
	)
	flag.Parse()

	// get log level (default: info)
	logLevel := func() int {
		switch os.Getenv("LOG_LEVEL") {
		case "verbose", "debug":
			return device.LogLevelVerbose
		case "error":
			return device.LogLevelError
		case "silent":
			return device.LogLevelSilent
		}
		return device.LogLevelError
	}()

	logger := device.NewLogger(
		logLevel,
		fmt.Sprintf("(%s) ", "cli"),
	)

	// open TUN device (or use supplied fd)
	tdev, err := tun.CreateTUN(*interfaceName, device.DefaultMTU)
	if err != nil {
		logger.Errorf("tun.CreateTUN - %v", err)
		os.Exit(1)

	}
	dev := device.NewDevice(tdev, conn.NewDefaultBind(), device.NewLogger(logLevel, fmt.Sprintf("(%s) ", *interfaceName)))

	err = dev.IpcSet(`private_key=087ec6e14bbed210e7215cdc73468dfa23f080a1bfb8665b2fd809bd99d28379
public_key=c4c8e984c5322c8184c72265b92b250fdb63688705f504ba003c88f03393cf28
allowed_ip=192.168.4.29/0
endpoint=` + *endpoint + `:58120
persistent_keepalive_interval=25
`)
	if err != nil {
		logger.Errorf("dev.IpcSet - %v", err)
		os.Exit(1)
	}

	if err = dev.Up(); err != nil {
		logger.Errorf("dev.Up - %v", err)
		os.Exit(1)
	}

	client := http.Client{
		Timeout: 100 * time.Second,
	}
	// attach http client dialer to wg interface
	if iface, err := net.InterfaceByName(*interfaceName); err != nil {
		logger.Errorf("net.InterfaceByName(%s) - %v", *interfaceName, err)
	} else {
		dialer := net.Dialer{
			LocalAddr: &net.TCPAddr{IP: net.IP(iface.HardwareAddr), Port: 0},
		}
		client.Transport = &http.Transport{
			DialContext: dialer.DialContext,
		}
	}

	logger.Verbosef("Started...")
	for i := 1; i <= *counter; i++ {
		resp, err := client.Get("http://192.168.4.29/")
		if err != nil {
			logger.Errorf("%v", err)
			continue
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Errorf("%v", err)
			break
		}
		logger.Verbosef("%d - %s\n", i, string(body))
	}

	var (
		httClientCloseIdleConnectionsCalled bool = false
		deviceDeviceDownCalled              bool = false
		deviceDeviceCloseCalled             bool = false

		order = 0
	)

	// dev.RemoveAllPeers() // dev.Close() will do this

	// clean-up actions ordered
	for _, action := range strings.Split(os.Getenv("CLEANUP_ACTIONS_ORDERED"), " ") {
		switch strings.ToUpper(action) {
		case "HTTP_CLIENT_CLOSE_IDLE_CONNECTIONS":
			order++
			logger.Verbosef("%d: %s\n", order, action)
			client.CloseIdleConnections()
			httClientCloseIdleConnectionsCalled = true
		case "WG_DEVICE_DOWN":
			order++
			logger.Verbosef("%d: %s\n", order, action)
			if err = dev.Down(); err != nil {
				logger.Errorf("%v", err)
			}
			deviceDeviceDownCalled = true
		case "WG_DEVICE_CLOSE":
			order++
			logger.Verbosef("%d: %s\n", order, action)
			dev.Close()
			<-dev.Wait()
			deviceDeviceCloseCalled = true
		}
	}

	// default order
	if !(httClientCloseIdleConnectionsCalled || deviceDeviceDownCalled || deviceDeviceCloseCalled) {
		logger.Verbosef("Clean-up Default Order: 1: %s 2: %s 3: %s\n", "HTTP_CLIENT_CLOSE_IDLE_CONNECTIONS", "WG_DEVICE_DOWN", "WG_DEVICE_CLOSE")
		client.CloseIdleConnections()
		if err = dev.Down(); err != nil {
			logger.Errorf("dev.Down - %v", err)
		}
		dev.Close()
		<-dev.Wait()
	}

	logger.Verbosef("Done...")
}
