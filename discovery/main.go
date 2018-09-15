package main

import (
	"flag"
	"net"

	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	xds "github.com/envoyproxy/go-control-plane/pkg/server"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	address      = flag.String("address", ":4000", "Discovery server address")
	node         = flag.String("envoy-node", "envoy", "Envoy node ID")
	listenerAddr = flag.String("listener", "0.0.0.0:10000", "Listener address")
)

func main() {
	flag.Parse()
	logger := logrus.StandardLogger()
	ln, err := net.Listen("tcp", *address)

	if err != nil {
		logger.WithError(err).WithField("address", *address).Fatalln("Failed to start TCP listener")
	}

	snapshotCache := cache.NewSnapshotCache(true, &NodeHash{}, logger)
	snapshotCache.SetSnapshot(*node, buildSnapshot())
	server := xds.NewServer(snapshotCache, &Callbacks{})
	grpcServer := grpc.NewServer()

	discovery.RegisterAggregatedDiscoveryServiceServer(grpcServer, server)
	api.RegisterEndpointDiscoveryServiceServer(grpcServer, server)
	api.RegisterClusterDiscoveryServiceServer(grpcServer, server)
	api.RegisterRouteDiscoveryServiceServer(grpcServer, server)
	api.RegisterListenerDiscoveryServiceServer(grpcServer, server)

	logger.WithField("address", ln.Addr()).Infoln("Starting GRPC server")

	if err := grpcServer.Serve(ln); err != nil {
		logger.WithError(err).WithField("address", ln.Addr()).Fatalln("Failed to start GRPC server")
	}
}
