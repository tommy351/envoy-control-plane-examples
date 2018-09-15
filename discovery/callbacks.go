package main

import api "github.com/envoyproxy/go-control-plane/envoy/api/v2"

type Callbacks struct{}

func (*Callbacks) OnFetchRequest(*api.DiscoveryRequest) {}

func (*Callbacks) OnFetchResponse(*api.DiscoveryRequest, *api.DiscoveryResponse) {}

func (*Callbacks) OnStreamClosed(int64) {}

func (*Callbacks) OnStreamOpen(int64, string) {}

func (*Callbacks) OnStreamRequest(int64, *api.DiscoveryRequest) {}

func (*Callbacks) OnStreamResponse(int64, *api.DiscoveryRequest, *api.DiscoveryResponse) {}
