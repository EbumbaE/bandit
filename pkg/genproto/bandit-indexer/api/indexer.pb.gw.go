// Code generated by protoc-gen-grpc-gateway. DO NOT EDIT.
// source: bandit-indexer/api/indexer.proto

/*
Package banditindexer is a reverse proxy.

It translates gRPC into RESTful JSON APIs.
*/
package banditindexer

import (
	"context"
	"io"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Suppress "imported and not used" errors
var _ codes.Code
var _ io.Reader
var _ status.Status
var _ = runtime.String
var _ = utilities.NewDoubleArray
var _ = metadata.Join

func request_BanditIndexerService_GetRuleScores_0(ctx context.Context, marshaler runtime.Marshaler, client BanditIndexerServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq GetRuleScoresRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["id"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "id")
	}

	protoReq.Id, err = runtime.String(val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "id", err)
	}

	msg, err := client.GetRuleScores(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_BanditIndexerService_GetRuleScores_0(ctx context.Context, marshaler runtime.Marshaler, server BanditIndexerServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq GetRuleScoresRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["id"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "id")
	}

	protoReq.Id, err = runtime.String(val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "id", err)
	}

	msg, err := server.GetRuleScores(ctx, &protoReq)
	return msg, metadata, err

}

// RegisterBanditIndexerServiceHandlerServer registers the http handlers for service BanditIndexerService to "mux".
// UnaryRPC     :call BanditIndexerServiceServer directly.
// StreamingRPC :currently unsupported pending https://github.com/grpc/grpc-go/issues/906.
// Note that using this registration option will cause many gRPC library features to stop working. Consider using RegisterBanditIndexerServiceHandlerFromEndpoint instead.
func RegisterBanditIndexerServiceHandlerServer(ctx context.Context, mux *runtime.ServeMux, server BanditIndexerServiceServer) error {

	mux.Handle("GET", pattern_BanditIndexerService_GetRuleScores_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		var err error
		var annotatedContext context.Context
		annotatedContext, err = runtime.AnnotateIncomingContext(ctx, mux, req, "/bandit.services.banditindexer.BanditIndexerService/GetRuleScores", runtime.WithHTTPPathPattern("/v1/indexer/rule/{id}"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_BanditIndexerService_GetRuleScores_0(annotatedContext, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
		if err != nil {
			runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_BanditIndexerService_GetRuleScores_0(annotatedContext, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

// RegisterBanditIndexerServiceHandlerFromEndpoint is same as RegisterBanditIndexerServiceHandler but
// automatically dials to "endpoint" and closes the connection when "ctx" gets done.
func RegisterBanditIndexerServiceHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
	conn, err := grpc.DialContext(ctx, endpoint, opts...)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()

	return RegisterBanditIndexerServiceHandler(ctx, mux, conn)
}

// RegisterBanditIndexerServiceHandler registers the http handlers for service BanditIndexerService to "mux".
// The handlers forward requests to the grpc endpoint over "conn".
func RegisterBanditIndexerServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return RegisterBanditIndexerServiceHandlerClient(ctx, mux, NewBanditIndexerServiceClient(conn))
}

// RegisterBanditIndexerServiceHandlerClient registers the http handlers for service BanditIndexerService
// to "mux". The handlers forward requests to the grpc endpoint over the given implementation of "BanditIndexerServiceClient".
// Note: the gRPC framework executes interceptors within the gRPC handler. If the passed in "BanditIndexerServiceClient"
// doesn't go through the normal gRPC flow (creating a gRPC client etc.) then it will be up to the passed in
// "BanditIndexerServiceClient" to call the correct interceptors.
func RegisterBanditIndexerServiceHandlerClient(ctx context.Context, mux *runtime.ServeMux, client BanditIndexerServiceClient) error {

	mux.Handle("GET", pattern_BanditIndexerService_GetRuleScores_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		var err error
		var annotatedContext context.Context
		annotatedContext, err = runtime.AnnotateContext(ctx, mux, req, "/bandit.services.banditindexer.BanditIndexerService/GetRuleScores", runtime.WithHTTPPathPattern("/v1/indexer/rule/{id}"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_BanditIndexerService_GetRuleScores_0(annotatedContext, inboundMarshaler, client, req, pathParams)
		annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
		if err != nil {
			runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_BanditIndexerService_GetRuleScores_0(annotatedContext, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

var (
	pattern_BanditIndexerService_GetRuleScores_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 1, 0, 4, 1, 5, 3}, []string{"v1", "indexer", "rule", "id"}, ""))
)

var (
	forward_BanditIndexerService_GetRuleScores_0 = runtime.ForwardResponseMessage
)
