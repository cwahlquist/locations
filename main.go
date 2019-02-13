package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	pb "locations/api/go"

	h "locations/handler"
//	m "locations/model"
	s "locations/service"

	"google.golang.org/grpc"

	// needed for postman proxy
	_ "github.com/jnewmano/grpc-json-proxy/codec"
)

var (
	port = flag.Int("port", 31400, "The server grpc port")
)

func main() {
	// get env vars
	flag.Parse()

	// start listening tcp:host:port
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		panic(err)
	}

	// inject dependencies

	// initialize service layer
	srv := s.NewService()

	hnd := h.NewHandler(srv)

	// create grpc server and apply middleware
	grpcServer := grpc.NewServer()

	// register missions PB with grpcServer
	pb.RegisterLocationsServiceServer(grpcServer, hnd)

	// create http server
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/readiness", readinessHandler)
	http.HandleFunc("/", healthHandler)
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			// cannot panic, because this probably is an intentional close
			log.Printf("Httpserver: ListenAndServe() error: %s", err)
		}
	}()
	log.Printf("Service started on 0.0.0.0:%d", *port)

	// start gRPC server
	err = grpcServer.Serve(listen)
	if err != nil {
		panic("gRpc Server failed to start")
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Healthy")
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Ready")
}
