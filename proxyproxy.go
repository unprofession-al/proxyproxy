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
	MITM    bool
	Proxy   string
	Verbose bool
	srv     http.Server
}

func (pp ProxyProxy) Start(addr string) {
	proxy := goproxy.NewProxyHttpServer()

	if pp.Proxy != "" {
		proxy.Tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(pp.Proxy)
		}
		fmt.Println("using Proxy", pp.Proxy)
	}

	if pp.MITM {
		proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	}

	proxy.Verbose = pp.Verbose

	pp.srv = http.Server{
		Addr:    addr,
		Handler: proxy,
	}

	go func() {
		if err := pp.srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %s", err)
		}
	}()
}

func (pp ProxyProxy) Stop() error {
	fmt.Printf("%#v", pp.srv)
	err := pp.srv.Shutdown(context.TODO())
	return err
}
