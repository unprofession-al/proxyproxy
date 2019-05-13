package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/elazarl/goproxy"
)

type ProxyProxy struct {
	MITM        bool         `yaml:"mitm" json:"mitm"`
	RemoteProxy string       `yaml:"remote_proxy" json:"remote_proxy"`
	Verbose     bool         `yaml:"verbose" json:"verbose"`
	srv         *http.Server `yaml:"srv" json:"srv"`
}

func (pp *ProxyProxy) Start(addr string) string {
	var out string
	proxy := goproxy.NewProxyHttpServer()

	if pp.RemoteProxy != "" {
		proxy.Tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(pp.RemoteProxy)
		}
		proxy.ConnectDial = proxy.NewConnectDialToProxy(pp.RemoteProxy)
		out = fmt.Sprintf("%sUsing remote proxy %s | ", out, pp.RemoteProxy)
	} else {
		out = fmt.Sprintf("%sUsing no remote proxy | ", out)
	}

	if pp.MITM {
		proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
		out = fmt.Sprintf("%sMITM on | ", out)
	}

	proxy.Verbose = pp.Verbose

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
