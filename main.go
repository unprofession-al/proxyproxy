package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"

	"github.com/elazarl/goproxy"
	"github.com/spf13/pflag"
)

var app App

type App struct {
	configPath string
	config     *Config
}

func init() {
	pflag.StringVarP(&app.configPath, "config", "c", "./config.yaml", "configuration file path")
}

func main() {
	pflag.Parse()

	var err error
	app.config, err = NewConfig(app.configPath)
	if err != nil {
		log.Fatal(err)
	}

	err = setCA(app.config.MITMCert, app.config.MITMKey)
	if err != nil {
		log.Fatal(err)
	}

	pp := &ProxyProxy{
		Verbose: app.config.Verbose,
		MITM:    false,
	}

	pp.Start(app.config.ProxyAddress)

	server := NewServer(app.config.AdminAddress)
	server.Run()

	l, _ := ListenNetlink()

	for {
		msgs, err := l.ReadMsgs()
		if err != nil {
			fmt.Println("Could not read netlink: %s", err)
		}

		for _, m := range msgs {
			if IsNewAddr(&m) || IsDelAddr(&m) {
				fmt.Println("change")
				addrs, err := net.InterfaceAddrs()
				if err != nil {
					os.Stderr.WriteString("Oops: " + err.Error() + "\n")
					os.Exit(1)
				}

				for _, a := range addrs {
					if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
						if ipnet.IP.To4() != nil {
							os.Stdout.WriteString(ipnet.IP.String() + "\n")
						}
					}
				}
			}
		}
	}

	log.Println("Press Ctrl+C to end")
	waitForCtrlC()

	pp.Stop()
	server.Stop()
}

func setCA(caCert, caKey string) error {
	goproxyCa, err := tls.LoadX509KeyPair(caCert, caKey)
	if err != nil {
		return err
	}
	if goproxyCa.Leaf, err = x509.ParseCertificate(goproxyCa.Certificate[0]); err != nil {
		return err
	}
	goproxy.GoproxyCa = goproxyCa
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	return nil
}

func waitForCtrlC() {
	var end_waiter sync.WaitGroup
	end_waiter.Add(1)
	var signal_channel chan os.Signal
	signal_channel = make(chan os.Signal, 1)
	signal.Notify(signal_channel, os.Interrupt)
	go func() {
		<-signal_channel
		end_waiter.Done()
	}()
	end_waiter.Wait()
}
