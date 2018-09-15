package main

import (
	"flag"
	"net"

	auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v2alpha"
	"github.com/tommy351/envoy-control-plane-examples/util"
	"google.golang.org/grpc"
)

var (
	address  = flag.String("address", ":4000", "Authorization server address")
	username = flag.String("username", "envoy", "Username")
	password = flag.String("password", "envoy", "Password")
)

func main() {
	flag.Parse()
	ln, err := net.Listen("tcp", *address)
	log := util.Logger.WithField("address", *address)

	if err != nil {
		log.WithError(err).Fatalln("Failed to start TCP listener")
	}

	grpcServer := grpc.NewServer()
	authServer := &AuthorizationServer{
		Username: *username,
		Password: *password,
	}
	auth.RegisterAuthorizationServer(grpcServer, authServer)

	log.Infoln("Starting GRPC server")

	if err := grpcServer.Serve(ln); err != nil {
		log.WithError(err).Fatalln("Failed to start GRPC server")
	}
}
