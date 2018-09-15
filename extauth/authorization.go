package main

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v2alpha"
	"github.com/envoyproxy/go-control-plane/envoy/type"
	"github.com/gogo/googleapis/google/rpc"
)

const basicAuthPrefix = "Basic "

type AuthorizationServer struct {
	Username string
	Password string
}

func (a *AuthorizationServer) Check(ctx context.Context, req *auth.CheckRequest) (*auth.CheckResponse, error) {
	authHeader, ok := req.Attributes.Request.Http.Headers["authorization"]

	if ok && strings.HasPrefix(authHeader, basicAuthPrefix) {
		payload, err := base64.StdEncoding.DecodeString(authHeader[len(basicAuthPrefix):])

		if err != nil {
			return nil, err
		}

		parts := strings.SplitN(string(payload), ":", 2)

		if len(parts) != 2 || parts[0] != a.Username || parts[1] != a.Password {
			return &auth.CheckResponse{
				Status: &rpc.Status{
					Code: int32(rpc.UNAUTHENTICATED),
				},
				HttpResponse: &auth.CheckResponse_DeniedResponse{
					DeniedResponse: &auth.DeniedHttpResponse{
						Status: &envoy_type.HttpStatus{
							Code: envoy_type.StatusCode_Unauthorized,
						},
						Body: "Go away",
					},
				},
			}, nil
		}
	}

	return &auth.CheckResponse{
		Status: &rpc.Status{
			Code: int32(rpc.OK),
		},
		HttpResponse: &auth.CheckResponse_OkResponse{
			OkResponse: &auth.OkHttpResponse{
				Headers: []*core.HeaderValueOption{
					{
						Header: &core.HeaderValue{
							Key:   "x-ext-auth",
							Value: "passed",
						},
					},
				},
			},
		},
	}, nil
}
