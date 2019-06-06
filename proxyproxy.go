package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/elazarl/goproxy"
)

type ProxyProxy struct {
	config *ProxyProxyConfig
	srv    *http.Server
}

func NewProxyProxy(c *ProxyProxyConfig) ProxyProxy {
	return ProxyProxy{config: c}
}

func (pp *ProxyProxy) Start(addr string) string {
	var out string
	proxy := goproxy.NewProxyHttpServer()

	if pp.config.RemoteProxy != "" {
		proxy.Tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(pp.config.RemoteProxy)
		}
		proxy.ConnectDial = proxy.NewConnectDialToProxy(pp.config.RemoteProxy)
		out = fmt.Sprintf("%sUsing remote proxy %s... ", out, pp.config.RemoteProxy)
	} else {
		out = fmt.Sprintf("%sUsing no remote proxy...", out)
	}

	proxy.Verbose = pp.config.Verbose

	pp.srv = &http.Server{
		Addr:    addr,
		Handler: proxy,
	}

	go func() {
		if err := pp.srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %s", err)
		}
	}()

	return out
}

func (pp *ProxyProxy) Stop() error {
	srv := pp.srv
	err := srv.Shutdown(context.Background())
	return err
}
