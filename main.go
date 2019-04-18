package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"

	"github.com/spf13/pflag"
)

var app App

type App struct {
	configPath string
	config     *Config
}

func init() {
	pflag.StringVarP(&app.configPath, "config", "c", "/etc/proxyproxy/config.yaml", "configuration file path")
}

func main() {
	pflag.Parse()

	var err error
	app.config, err = NewConfig(app.configPath)
	if err != nil {
		log.Fatal(err)
	}

	if app.config.MITMKey != "" && app.config.MITMCert != "" {
		log.Printf("Setting up MITM CA\n")
		err = setCA(app.config.MITMCert, app.config.MITMKey)
		if err != nil {
			log.Fatal(err)
		}
	}
	ips, err := getRelevantIPs(app.config.Interfaces)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("IPs detected %s.\n", ips)

	name, ppc := app.config.ProxyProxyConfigs.FindMatch(ips)
	log.Printf("Using profile '%s'", name)
	pp := ProxyProxy{
		Verbose: app.config.Verbose,
		MITM:    ppc.MITM,
		Proxy:   ppc.RemoteProxy,
	}

	out := pp.Start(app.config.ProxyAddress)
	log.Println(out)
	log.Printf("Started proxy with profile '%s'", name)

	server := NewServer(app.config.AdminAddress)
	server.Run()

	go func() {
		l, _ := ListenNetlink()

		for {
			msgs, err := l.ReadMsgs()
			if err != nil {
				log.Printf("Could not read netlink: %s\n", err.Error())
			}

			for _, m := range msgs {
				if IsNewAddr(&m) || IsDelAddr(&m) {
					log.Printf("Network configuration changed.\n")
					ips, err := getRelevantIPs(app.config.Interfaces)
					if err != nil {
						log.Fatal(err)
					}
					log.Printf("IPs detected %s.\n", ips)
					name, ppc := app.config.ProxyProxyConfigs.FindMatch(ips)
					log.Printf("Using profile '%s'", name)
					err = pp.Stop()
					if err != nil {
						log.Fatal(err)
					}

					pp = ProxyProxy{
						Verbose: app.config.Verbose,
						MITM:    ppc.MITM,
						Proxy:   ppc.RemoteProxy,
					}
					out := pp.Start(app.config.ProxyAddress)
					log.Println(out)
				}
			}
		}
	}()
	log.Println("Press Ctrl+C to end")
	waitForCtrlC()

	pp.Stop()
	server.Stop()
}

func getRelevantIPs(interfaces []string) ([]net.IP, error) {
	var ips []net.IP
	ifaces, err := net.Interfaces()
	if err != nil {
		return ips, fmt.Errorf("Could not read interfaces: %s\n", err.Error())
	}
	for _, i := range ifaces {
		for _, r := range interfaces {
			if r == i.Name {
				addrs, err := i.Addrs()
				if err != nil {
					return ips, fmt.Errorf("Could not read adresses of interface %s: %s\n", i.Name, err.Error())
				}
				for _, addr := range addrs {
					switch v := addr.(type) {
					case *net.IPNet:
						ips = append(ips, v.IP)
					case *net.IPAddr:
						ips = append(ips, v.IP)
					}
				}
			}
		}
	}
	return ips, nil
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
