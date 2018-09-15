package main

import (
	"flag"
	"net"

	als "github.com/envoyproxy/go-control-plane/envoy/service/accesslog/v2"
	"github.com/tommy351/envoy-control-plane-examples/util"
	"google.golang.org/grpc"
)

var (
	address = flag.String("address", ":4000", "Access log server address")
)

func main() {
	flag.Parse()
	ln, err := net.Listen("tcp", *address)
	log := util.Logger.WithField("address", *address)

	if err != nil {
		log.WithError(err).Fatalln("Failed to start TCP listener")
	}

	grpcServer := grpc.NewServer()
	logServer := &AccessLogServer{}
	als.RegisterAccessLogServiceServer(grpcServer, logServer)

	log.Infoln("Starting GRPC server")

	if err := grpcServer.Serve(ln); err != nil {
		log.WithError(err).Fatalln("Failed to start GRPC server")
	}
}
