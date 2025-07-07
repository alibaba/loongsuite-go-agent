// Copyright (c) 2024 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rpc

import (
	"context"
	"errors"
	"github.com/alibaba/opentelemetry-go-auto-instrumentation/pkg/inst-api/utils"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"testing"
)

type testRequest struct {
}

type testResponse struct {
}

type rpcAttrsGetter struct {
}

func (h rpcAttrsGetter) GetSystem(request testRequest) string {
	return "system"
}

func (h rpcAttrsGetter) GetService(request testRequest) string {
	return "service"
}

func (h rpcAttrsGetter) GetMethod(request testRequest) string {
	return "method"
}

func (h rpcAttrsGetter) GetServerAddress(request testRequest) string {
	return "serverAddress"
}

// gRPC attributes getter for testing gRPC-specific functionality
type grpcAttrsGetter struct {
}

func (h grpcAttrsGetter) GetSystem(request testRequest) string {
	return "grpc"
}

func (h grpcAttrsGetter) GetService(request testRequest) string {
	return "service"
}

func (h grpcAttrsGetter) GetMethod(request testRequest) string {
	return "method"
}

func (h grpcAttrsGetter) GetServerAddress(request testRequest) string {
	return "serverAddress"
}

func TestClientGetSpanKey(t *testing.T) {
	rpcExtractor := &ClientRpcAttrsExtractor[testRequest, any, rpcAttrsGetter]{}
	if rpcExtractor.GetSpanKey() != utils.RPC_CLIENT_KEY {
		log.Fatal("Should have returned RPC_CLIENT_KEY")
	}
}

func TestServerGetSpanKey(t *testing.T) {
	rpcExtractor := &ServerRpcAttrsExtractor[testRequest, any, rpcAttrsGetter]{}
	if rpcExtractor.GetSpanKey() != utils.RPC_SERVER_KEY {
		log.Fatal("Should have returned RPC_SERVER_KEY")
	}
}

func TestRpcClientExtractorStart(t *testing.T) {
	rpcExtractor := ClientRpcAttrsExtractor[testRequest, testResponse, rpcAttrsGetter]{}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	attrs, _ = rpcExtractor.OnStart(attrs, parentContext, testRequest{})
	if attrs[0].Key != semconv.RPCSystemKey || attrs[0].Value.AsString() != "system" {
		log.Fatal("rpc system should be system")
	}
	if attrs[1].Key != semconv.RPCServiceKey || attrs[1].Value.AsString() != "service" {
		log.Fatal("rpc service should be service")
	}
	if attrs[2].Key != semconv.RPCMethodKey || attrs[2].Value.AsString() != "method" {
		log.Fatal("rpc method should be method")
	}
}

func TestRpcClientExtractorEnd(t *testing.T) {
	rpcExtractor := ClientRpcAttrsExtractor[testRequest, testResponse, rpcAttrsGetter]{}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	attrs, _ = rpcExtractor.OnEnd(attrs, parentContext, testRequest{}, testResponse{}, nil)
	if len(attrs) != 0 {
		log.Fatal("attrs should be empty")
	}
}

func TestRpcServerExtractorStart(t *testing.T) {
	rpcExtractor := ServerRpcAttrsExtractor[testRequest, testResponse, rpcAttrsGetter]{}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	attrs, _ = rpcExtractor.OnStart(attrs, parentContext, testRequest{})
	if attrs[0].Key != semconv.RPCSystemKey || attrs[0].Value.AsString() != "system" {
		log.Fatal("rpc system should be system")
	}
	if attrs[1].Key != semconv.RPCServiceKey || attrs[1].Value.AsString() != "service" {
		log.Fatal("rpc service should be service")
	}
	if attrs[2].Key != semconv.RPCMethodKey || attrs[2].Value.AsString() != "method" {
		log.Fatal("rpc method should be method")
	}
}

func TestRpcServerExtractorEnd(t *testing.T) {
	rpcExtractor := ServerRpcAttrsExtractor[testRequest, testResponse, rpcAttrsGetter]{}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	attrs, _ = rpcExtractor.OnEnd(attrs, parentContext, testRequest{}, testResponse{}, nil)
	if len(attrs) != 0 {
		log.Fatal("attrs should be empty")
	}
}

// Test gRPC status code extraction for successful requests
func TestGrpcClientExtractorEndSuccess(t *testing.T) {
	rpcExtractor := ClientRpcAttrsExtractor[testRequest, testResponse, grpcAttrsGetter]{}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	
	// Test successful gRPC call (no error)
	attrs, _ = rpcExtractor.OnEnd(attrs, parentContext, testRequest{}, testResponse{}, nil)
	
	// Should have one attribute for gRPC status code
	if len(attrs) != 1 {
		log.Fatal("Expected 1 attribute for gRPC status code")
	}
	
	// Check that the status code attribute is present and set to 0 (OK)
	if attrs[0].Key != semconv.RPCGRPCStatusCodeKey {
		log.Fatal("Expected RPCGRPCStatusCodeKey")
	}
	
	if attrs[0].Value.AsInt64() != 0 {
		log.Fatal("Expected status code 0 (OK)")
	}
}

// Test gRPC status code extraction for error requests
func TestGrpcClientExtractorEndError(t *testing.T) {
	rpcExtractor := ClientRpcAttrsExtractor[testRequest, testResponse, grpcAttrsGetter]{}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	
	// Create a gRPC status error
	grpcErr := status.Error(codes.NotFound, "resource not found")
	
	// Test gRPC call with error
	attrs, _ = rpcExtractor.OnEnd(attrs, parentContext, testRequest{}, testResponse{}, grpcErr)
	
	// Should have one attribute for gRPC status code
	if len(attrs) != 1 {
		log.Fatal("Expected 1 attribute for gRPC status code")
	}
	
	// Check that the status code attribute is present and set to 5 (NotFound)
	if attrs[0].Key != semconv.RPCGRPCStatusCodeKey {
		log.Fatal("Expected RPCGRPCStatusCodeKey")
	}
	
	if attrs[0].Value.AsInt64() != int64(codes.NotFound) {
		log.Fatal("Expected status code for NotFound")
	}
}

// Test gRPC status code extraction for non-gRPC status errors
func TestGrpcClientExtractorEndNonGrpcError(t *testing.T) {
	rpcExtractor := ClientRpcAttrsExtractor[testRequest, testResponse, grpcAttrsGetter]{}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	
	// Create a non-gRPC error
	regularErr := errors.New("regular error")
	
	// Test gRPC call with non-gRPC error
	attrs, _ = rpcExtractor.OnEnd(attrs, parentContext, testRequest{}, testResponse{}, regularErr)
	
	// Should have no attributes since the error is not a gRPC status error
	if len(attrs) != 0 {
		log.Fatal("Expected 0 attributes for non-gRPC error")
	}
}

// Test gRPC server extractor for successful requests
func TestGrpcServerExtractorEndSuccess(t *testing.T) {
	rpcExtractor := ServerRpcAttrsExtractor[testRequest, testResponse, grpcAttrsGetter]{}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	
	// Test successful gRPC call (no error)
	attrs, _ = rpcExtractor.OnEnd(attrs, parentContext, testRequest{}, testResponse{}, nil)
	
	// Should have one attribute for gRPC status code
	if len(attrs) != 1 {
		log.Fatal("Expected 1 attribute for gRPC status code")
	}
	
	// Check that the status code attribute is present and set to 0 (OK)
	if attrs[0].Key != semconv.RPCGRPCStatusCodeKey {
		log.Fatal("Expected RPCGRPCStatusCodeKey")
	}
	
	if attrs[0].Value.AsInt64() != 0 {
		log.Fatal("Expected status code 0 (OK)")
	}
}

// Test gRPC server extractor for error requests
func TestGrpcServerExtractorEndError(t *testing.T) {
	rpcExtractor := ServerRpcAttrsExtractor[testRequest, testResponse, grpcAttrsGetter]{}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	
	// Create a gRPC status error
	grpcErr := status.Error(codes.InvalidArgument, "invalid argument")
	
	// Test gRPC call with error
	attrs, _ = rpcExtractor.OnEnd(attrs, parentContext, testRequest{}, testResponse{}, grpcErr)
	
	// Should have one attribute for gRPC status code
	if len(attrs) != 1 {
		log.Fatal("Expected 1 attribute for gRPC status code")
	}
	
	// Check that the status code attribute is present and set to 3 (InvalidArgument)
	if attrs[0].Key != semconv.RPCGRPCStatusCodeKey {
		log.Fatal("Expected RPCGRPCStatusCodeKey")
	}
	
	if attrs[0].Value.AsInt64() != int64(codes.InvalidArgument) {
		log.Fatal("Expected status code for InvalidArgument")
	}
}

// Test non-gRPC system should not have status code attribute
func TestNonGrpcSystemNoStatusCode(t *testing.T) {
	// Use regular rpcAttrsGetter which returns "system" instead of "grpc"
	rpcExtractor := ClientRpcAttrsExtractor[testRequest, testResponse, rpcAttrsGetter]{}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	
	// Create a gRPC status error (but system is not "grpc")
	grpcErr := status.Error(codes.Internal, "internal error")
	
	// Test non-gRPC call with gRPC error
	attrs, _ = rpcExtractor.OnEnd(attrs, parentContext, testRequest{}, testResponse{}, grpcErr)
	
	// Should have no attributes since the system is not "grpc"
	if len(attrs) != 0 {
		log.Fatal("Expected 0 attributes for non-gRPC system")
	}
}

// Test various gRPC status codes
func TestGrpcStatusCodes(t *testing.T) {
	rpcExtractor := ClientRpcAttrsExtractor[testRequest, testResponse, grpcAttrsGetter]{}
	parentContext := context.Background()
	
	testCases := []struct {
		name       string
		code       codes.Code
		message    string
		expected   int64
	}{
		{"OK", codes.OK, "success", 0},
		{"Cancelled", codes.Canceled, "cancelled", 1},
		{"Unknown", codes.Unknown, "unknown error", 2},
		{"InvalidArgument", codes.InvalidArgument, "invalid argument", 3},
		{"DeadlineExceeded", codes.DeadlineExceeded, "deadline exceeded", 4},
		{"NotFound", codes.NotFound, "not found", 5},
		{"AlreadyExists", codes.AlreadyExists, "already exists", 6},
		{"PermissionDenied", codes.PermissionDenied, "permission denied", 7},
		{"ResourceExhausted", codes.ResourceExhausted, "resource exhausted", 8},
		{"FailedPrecondition", codes.FailedPrecondition, "failed precondition", 9},
		{"Aborted", codes.Aborted, "aborted", 10},
		{"OutOfRange", codes.OutOfRange, "out of range", 11},
		{"Unimplemented", codes.Unimplemented, "unimplemented", 12},
		{"Internal", codes.Internal, "internal error", 13},
		{"Unavailable", codes.Unavailable, "unavailable", 14},
		{"DataLoss", codes.DataLoss, "data loss", 15},
		{"Unauthenticated", codes.Unauthenticated, "unauthenticated", 16},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			attrs := make([]attribute.KeyValue, 0)
			
			var err error
			if tc.code == codes.OK {
				err = nil
			} else {
				err = status.Error(tc.code, tc.message)
			}
			
			attrs, _ = rpcExtractor.OnEnd(attrs, parentContext, testRequest{}, testResponse{}, err)
			
			// Should have one attribute for gRPC status code
			if len(attrs) != 1 {
				log.Fatal("Expected 1 attribute for gRPC status code")
			}
			
			// Check that the status code matches expected value
			if attrs[0].Key != semconv.RPCGRPCStatusCodeKey {
				log.Fatal("Expected RPCGRPCStatusCodeKey")
			}
			
			if attrs[0].Value.AsInt64() != tc.expected {
				log.Fatal("Status code mismatch")
			}
		})
	}
}
