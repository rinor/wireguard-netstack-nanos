package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
)

func main() {
	var (
		interfaceName = flag.String("interface", "wg2", "wg interface name")
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
		fmt.Sprintf("(%s) ", "srv"),
	)

	// open TUN device (or use supplied fd)
	tdev, err := tun.CreateTUN(*interfaceName, device.DefaultMTU)
	if err != nil {
		logger.Errorf("tun.CreateTUN - %v", err)
		os.Exit(1)
	}

	dev := device.NewDevice(tdev, conn.NewDefaultBind(), device.NewLogger(logLevel, fmt.Sprintf("(%s) ", *interfaceName)))

	err = dev.IpcSet(`private_key=003ed5d73b55806c30de3f8a7bdab38af13539220533055e635690b8b87ad641
listen_port=58120
public_key=f928d4f6c1b86c12f2562c10b07c555c5c57fd00f59e90c8d8d88767271cbf7c
allowed_ip=192.168.4.28/32
persistent_keepalive_interval=0
`)
	if err != nil {
		logger.Errorf("device.Device.IpcSet - %v", err)
		os.Exit(1)
	}

	if err = dev.Up(); err != nil {
		logger.Errorf("device.Device.Up - %v", err)
		os.Exit(1)
	}

	listener, err := net.Listen("tcp4", "192.168.4.29:80")
	if err != nil {
		logger.Errorf("net.Listener.Listen - %v", err)
		os.Exit(1)
	}

	server := &http.Server{}

	errs := make(chan error)
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	go func() {
		http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
			logger.Verbosef("> %s - %s - %s", request.RemoteAddr, request.URL.String(), request.UserAgent())
			_, _ = io.WriteString(writer, "Hello from userspace klib TCP!")
		})
		err := server.Serve(listener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		errs <- err
		close(errs)
	}()

	select {
	case sig := <-term:
		logger.Verbosef("Signal - %v", sig)
	case err = <-errs:
		if err != nil {
			logger.Errorf("http.Server.Serve - %v", err)
		}
	case dw := <-dev.Wait():
		logger.Verbosef("device.Device.Wait - %v", dw)
	}

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err = server.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("http.Server.Shutdown - %v", err)
	}

	if err = <-errs; err != nil {
		logger.Errorf("http.Server.Serve - %v", err)
	}

	// dev.RemoveAllPeers() // dev.Close() will do this

	if err = dev.Down(); err != nil {
		logger.Errorf("device.Device.Down - %v", err)
	}

	dev.Close()
	<-dev.Wait()

	logger.Verbosef("Done...")
}
