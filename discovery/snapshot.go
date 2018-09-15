package main

import (
	"net"
	"strconv"
	"time"

	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	extauth "github.com/envoyproxy/go-control-plane/envoy/config/filter/http/ext_authz/v2alpha"
	hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/envoyproxy/go-control-plane/pkg/util"
	"github.com/gogo/protobuf/types"
)

func buildSnapshot() cache.Snapshot {
	var endpoints, clusters, routes, listeners []cache.Resource

	clusters = append(clusters, &api.Cluster{
		Name:           "httpbin",
		ConnectTimeout: time.Second,
		LbPolicy:       api.Cluster_ROUND_ROBIN,
		Type:           api.Cluster_LOGICAL_DNS,
		LoadAssignment: &api.ClusterLoadAssignment{
			ClusterName: "httpbin",
			Endpoints: []endpoint.LocalityLbEndpoints{
				{
					LbEndpoints: []endpoint.LbEndpoint{
						{
							Endpoint: &endpoint.Endpoint{
								Address: buildAddress("httpbin.org:80"),
							},
						},
					},
				},
			},
		},
	})

	routeConf := &api.RouteConfiguration{
		Name: "route",
		VirtualHosts: []route.VirtualHost{
			{
				Name:    "httpbin",
				Domains: []string{"*"},
				Routes: []route.Route{
					{
						Match: route.RouteMatch{
							PathSpecifier: &route.RouteMatch_Prefix{
								Prefix: "/",
							},
						},
						Action: &route.Route_Route{
							Route: &route.RouteAction{
								HostRewriteSpecifier: &route.RouteAction_AutoHostRewrite{
									AutoHostRewrite: &types.BoolValue{Value: true},
								},
								ClusterSpecifier: &route.RouteAction_Cluster{
									Cluster: "httpbin",
								},
							},
						},
					},
				},
			},
		},
	}

	routes = append(routes, routeConf)

	authService := &api.Cluster{
		Name:                 "extauth",
		Type:                 api.Cluster_STRICT_DNS,
		LbPolicy:             api.Cluster_ROUND_ROBIN,
		Http2ProtocolOptions: &core.Http2ProtocolOptions{},
		ConnectTimeout:       time.Second,
		LoadAssignment: &api.ClusterLoadAssignment{
			ClusterName: "extauth",
			Endpoints: []endpoint.LocalityLbEndpoints{
				{
					LbEndpoints: []endpoint.LbEndpoint{
						{
							Endpoint: &endpoint.Endpoint{
								Address: buildAddress(*extauthAddr),
							},
						},
					},
				},
			},
		},
	}

	clusters = append(clusters, authService)

	extAuthConf, err := util.MessageToStruct(&extauth.ExtAuthz{
		Services: &extauth.ExtAuthz_GrpcService{
			GrpcService: &core.GrpcService{
				TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &core.GrpcService_EnvoyGrpc{
						ClusterName: authService.Name,
					},
				},
			},
		},
	})

	if err != nil {
		panic(err)
	}

	hcmConfig, err := util.MessageToStruct(&hcm.HttpConnectionManager{
		CodecType:  hcm.AUTO,
		StatPrefix: "http",
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				RouteConfigName: routeConf.Name,
				ConfigSource: core.ConfigSource{
					ConfigSourceSpecifier: &core.ConfigSource_Ads{
						Ads: &core.AggregatedConfigSource{},
					},
				},
			},
		},
		HttpFilters: []*hcm.HttpFilter{
			{
				Name:   util.ExternalAuthorization,
				Config: extAuthConf,
			},
			{
				Name: util.Router,
			},
		},
	})

	if err != nil {
		panic(err)
	}

	listeners = append(listeners, &api.Listener{
		Name:    "main",
		Address: *buildAddress(*listenerAddr),
		FilterChains: []listener.FilterChain{
			{
				Filters: []listener.Filter{
					{
						Name:   util.HTTPConnectionManager,
						Config: hcmConfig,
					},
				},
			},
		},
	})

	snapshot := cache.NewSnapshot("1", endpoints, clusters, routes, listeners)

	if err := snapshot.Consistent(); err != nil {
		panic(err)
	}

	return snapshot
}

func buildAddress(addr string) *core.Address {
	host, port, err := net.SplitHostPort(addr)

	if err != nil {
		panic(err)
	}

	portNum, err := strconv.ParseUint(port, 10, 32)

	if err != nil {
		panic(err)
	}

	return &core.Address{
		Address: &core.Address_SocketAddress{
			SocketAddress: &core.SocketAddress{
				Address: host,
				PortSpecifier: &core.SocketAddress_PortValue{
					PortValue: uint32(portNum),
				},
			},
		},
	}
}
