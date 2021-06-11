package main

import (
	"context"
	"net"
	"net/http"
	"net/http/pprof"
	"net/rpc"
	"time"

	goridgeRpc "github.com/spiral/goridge/v3/pkg/rpc"
	"github.com/spiral/goridge/v3/test"
)

func main() {
	s := NewServer()
	go func() {
		_ = s.Start("localhost:6061")
	}()
	defer func() {
		_ = s.Stop(context.Background())
	}()

	time.Sleep(time.Second * 5)
	// create an pprof server
	server()
	time.Sleep(time.Second * 1)

	client()
}

// testService sample
type testService struct{}

func (s *testService) ProtoMessage(payload *test.Payload, item *test.Item) error {
	(*item).Key = payload.Items[0].Key
	return nil
}

func client() {
	err := rpc.RegisterName("testbench", new(testService))
	if err != nil {
		panic(err)
	}

	conn, err := net.Dial("tcp", "127.0.0.1:18321")
	if err != nil {
		panic(err)
	}

	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
	defer func() {
		err := client.Close()
		if err != nil {
			panic(err)
		}
	}()

	tt := time.Now().String()

	keysP := &test.Payload{
		Storage: "memory-rr",
		Items: []*test.Item{
			{
				Key:     "a",
				Value:   "hhhhhhhhhhhhhhhhheeeeeeeeeeeeeeeeeeeeeeeeeeeelllllllllllllllllllllllllllllllllloooooooooooooooooooooooooooooo",
				Timeout: tt,
			},
			{
				Key:     "b",
				Value:   "hhhhhhhhhhhhhhhhheeeeeeeeeeeeeeeeeeeeeeeeeeeelllllllllllllllllllllllllllllllllloooooooooooooooooooooooooooooo",
				Timeout: tt,
			},
			{
				Key:     "c",
				Value:   "hhhhhhhhhhhhhhhhheeeeeeeeeeeeeeeeeeeeeeeeeeeelllllllllllllllllllllllllllllllllloooooooooooooooooooooooooooooo",
				Timeout: tt,
			},
		},
	}

	item := &test.Item{}
	for i := 0; i < 1000000; i++ {
		err = client.Call("testbench.ProtoMessage", keysP, item)
		if err != nil {
			panic(err)
		}
	}
}

func server() {
	ln, err := net.Listen("tcp", "127.0.0.1:18321")
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			conn, err2 := ln.Accept()
			if err2 != nil {
				panic(err2)
			}
			rpc.ServeCodec(goridgeRpc.NewCodec(conn))
		}
	}()
}

// Server is a HTTP server for debugging.
type Server struct {
	srv *http.Server
}

// NewServer creates new HTTP server for debugging.
func NewServer() Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return Server{srv: &http.Server{Handler: mux}}
}

// Start debug server.
func (s *Server) Start(addr string) error {
	s.srv.Addr = addr

	return s.srv.ListenAndServe()
}

// Stop debug server.
func (s *Server) Stop(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
