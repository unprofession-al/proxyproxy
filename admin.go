package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	yaml "gopkg.in/yaml.v2"
)

type Server struct {
	listener string
	handler  http.Handler
	srv      http.Server
}

func NewServer(listener string) Server {
	s := Server{
		listener: listener,
	}

	r := mux.NewRouter().StrictSlash(true)

	r.HandleFunc("/config", s.ConfigHandler).Methods("GET")

	s.handler = alice.New().Then(r)
	return s
}

func (s Server) Run() {
	s.srv = http.Server{
		Addr:    s.listener,
		Handler: s.handler,
	}

	go func() {
		if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %s", err)
		}
	}()
}

func (s Server) Stop() error {
	fmt.Printf("%#v", s.srv)
	err := s.srv.Shutdown(context.TODO())
	return err
}

func (s Server) ConfigHandler(res http.ResponseWriter, req *http.Request) {
	s.respond(res, req, http.StatusOK, app.config)
}

func (s Server) respond(res http.ResponseWriter, req *http.Request, code int, data interface{}) {
	if code != http.StatusOK {
		fmt.Println(data)
	}
	var err error
	var errMesg []byte
	var out []byte

	f := "json"
	format := req.URL.Query()["f"]
	if len(format) > 0 {
		f = format[0]
	}

	if f == "yaml" {
		res.Header().Set("Content-Type", "text/yaml; charset=utf-8")
		out, err = yaml.Marshal(data)
		errMesg = []byte("--- error: failed while rendering data to yaml")
	} else {
		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		out, err = json.Marshal(data)
		errMesg = []byte("{ 'error': 'failed while rendering data to json' }")
	}

	if err != nil {
		out = errMesg
		code = http.StatusInternalServerError
	}
	res.WriteHeader(code)
	res.Write(out)
}
