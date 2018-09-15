package main

import "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"

type NodeHash struct{}

func (*NodeHash) ID(node *core.Node) string {
	if node == nil {
		return "unknown"
	}

	return node.Id
}
