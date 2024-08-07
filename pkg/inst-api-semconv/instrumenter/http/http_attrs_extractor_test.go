package http

import (
	"context"
	"github.com/alibaba/opentelemetry-go-auto-instrumentation/pkg/inst-api-semconv/instrumenter/net"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"testing"
)

type httpServerAttrsGetter struct {
}

type httpClientAttrsGetter struct {
}

type networkAttrsGetter struct {
}

type urlAttrsGetter struct {
}

func (u urlAttrsGetter) GetUrlScheme(request testRequest) string {
	return "url-scheme"
}

func (u urlAttrsGetter) GetUrlPath(request testRequest) string {
	return "url-path"
}

func (u urlAttrsGetter) GetUrlQuery(request testRequest) string {
	return "url-query"
}

func (n networkAttrsGetter) GetNetworkType(request testRequest, response testResponse) string {
	return "network-type"
}

func (n networkAttrsGetter) GetNetworkTransport(request testRequest, response testResponse) string {
	return "network-transport"
}

func (n networkAttrsGetter) GetNetworkProtocolName(request testRequest, response testResponse) string {
	return "network-protocol-name"
}

func (n networkAttrsGetter) GetNetworkProtocolVersion(request testRequest, response testResponse) string {
	return "network-protocol-version"
}

func (n networkAttrsGetter) GetNetworkLocalInetAddress(request testRequest, response testResponse) string {
	return "network-local-inet-address"
}

func (n networkAttrsGetter) GetNetworkLocalPort(request testRequest, response testResponse) int {
	return 8080
}

func (n networkAttrsGetter) GetNetworkPeerInetAddress(request testRequest, response testResponse) string {
	return "network-peer-inet-address"
}

func (n networkAttrsGetter) GetNetworkPeerPort(request testRequest, response testResponse) int {
	return 8080
}

func (h httpClientAttrsGetter) GetRequestMethod(request testRequest) string {
	return "GET"
}

func (h httpClientAttrsGetter) GetHttpRequestHeader(request testRequest, name string) []string {
	return []string{"request-header"}
}

func (h httpClientAttrsGetter) GetHttpResponseStatusCode(request testRequest, response testResponse, err error) int {
	return 200
}

func (h httpClientAttrsGetter) GetHttpResponseHeader(request testRequest, response testResponse, name string) []string {
	return []string{"response-header"}
}

func (h httpClientAttrsGetter) GetErrorType(request testRequest, response testResponse, err error) string {
	return ""
}

func (h httpClientAttrsGetter) GetNetworkType(request testRequest, response testResponse) string {
	return "ipv4"
}

func (h httpClientAttrsGetter) GetNetworkTransport(request testRequest, response testResponse) string {
	return "TCP"
}

func (h httpClientAttrsGetter) GetNetworkProtocolName(request testRequest, response testResponse) string {
	return "HTTP"
}

func (h httpClientAttrsGetter) GetNetworkProtocolVersion(request testRequest, response testResponse) string {
	return "HTTP/1.1"
}

func (h httpClientAttrsGetter) GetNetworkLocalInetAddress(request testRequest, response testResponse) string {
	return "127.0.0.1"
}

func (h httpClientAttrsGetter) GetNetworkLocalPort(request testRequest, response testResponse) int {
	return 8080
}

func (h httpClientAttrsGetter) GetNetworkPeerInetAddress(request testRequest, response testResponse) string {
	return "127.0.0.1"
}

func (h httpClientAttrsGetter) GetNetworkPeerPort(request testRequest, response testResponse) int {
	return 8080
}

func (h httpClientAttrsGetter) GetUrlFull(request testRequest) string {
	return "url-full"
}

func (h httpClientAttrsGetter) GetServerAddress(request testRequest) string {
	return "server-address"
}

func (h httpServerAttrsGetter) GetRequestMethod(request testRequest) string {
	return "GET"
}

func (h httpServerAttrsGetter) GetHttpRequestHeader(request testRequest, name string) []string {
	return []string{"request-header"}
}

func (h httpServerAttrsGetter) GetHttpResponseStatusCode(request testRequest, response testResponse, err error) int {
	return 200
}

func (h httpServerAttrsGetter) GetHttpResponseHeader(request testRequest, response testResponse, name string) []string {
	return []string{"response-header"}
}

func (h httpServerAttrsGetter) GetErrorType(request testRequest, response testResponse, err error) string {
	return "error-type"
}

func (h httpServerAttrsGetter) GetUrlScheme(request testRequest) string {
	return "url-scheme"
}

func (h httpServerAttrsGetter) GetUrlPath(request testRequest) string {
	return "url-path"
}

func (h httpServerAttrsGetter) GetUrlQuery(request testRequest) string {
	return "url-query"
}

func (h httpServerAttrsGetter) GetNetworkType(request testRequest, response testResponse) string {
	return "network-type"
}

func (h httpServerAttrsGetter) GetNetworkTransport(request testRequest, response testResponse) string {
	return "network-transport"
}

func (h httpServerAttrsGetter) GetNetworkProtocolName(request testRequest, response testResponse) string {
	return "network-protocol-name"
}

func (h httpServerAttrsGetter) GetNetworkProtocolVersion(request testRequest, response testResponse) string {
	return "network-protocol-version"
}

func (h httpServerAttrsGetter) GetNetworkLocalInetAddress(request testRequest, response testResponse) string {
	return "127.0.0.1"
}

func (h httpServerAttrsGetter) GetNetworkLocalPort(request testRequest, response testResponse) int {
	return 8080
}

func (h httpServerAttrsGetter) GetNetworkPeerInetAddress(request testRequest, response testResponse) string {
	return "127.0.0.1"
}

func (h httpServerAttrsGetter) GetNetworkPeerPort(request testRequest, response testResponse) int {
	return 8080
}

func (h httpServerAttrsGetter) GetHttpRoute(request testRequest) string {
	return "http-route"
}

func TestHttpClientExtractorStart(t *testing.T) {
	httpClientExtractor := HttpClientAttrsExtractor[testRequest, testResponse, httpClientAttrsGetter, networkAttrsGetter]{
		base:             HttpCommonAttrsExtractor[testRequest, testResponse, httpClientAttrsGetter, networkAttrsGetter]{},
		networkExtractor: net.NetworkAttrsExtractor[testRequest, testResponse, networkAttrsGetter]{},
	}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	attrs = httpClientExtractor.OnStart(attrs, parentContext, testRequest{})
	if attrs[0].Key != semconv.HTTPRequestMethodKey || attrs[0].Value.AsString() != "GET" {
		t.Fatalf("http method should be GET")
	}
	if attrs[1].Key != semconv.URLFullKey || attrs[1].Value.AsString() != "url-full" {
		t.Fatalf("urlfull should be url-full")
	}
}

func TestHttpClientExtractorEnd(t *testing.T) {
	httpClientExtractor := HttpClientAttrsExtractor[testRequest, testResponse, httpClientAttrsGetter, networkAttrsGetter]{
		base:             HttpCommonAttrsExtractor[testRequest, testResponse, httpClientAttrsGetter, networkAttrsGetter]{},
		networkExtractor: net.NetworkAttrsExtractor[testRequest, testResponse, networkAttrsGetter]{},
	}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	attrs = httpClientExtractor.OnEnd(attrs, parentContext, testRequest{}, testResponse{}, nil)
	if attrs[0].Key != semconv.HTTPResponseStatusCodeKey || attrs[0].Value.AsInt64() != 200 {
		t.Fatalf("status code should be 200")
	}
	if attrs[1].Key != semconv.NetworkProtocolNameKey || attrs[1].Value.AsString() != "network-protocol-name" {
		t.Fatalf("wrong network protocol name")
	}
	if attrs[2].Key != semconv.NetworkProtocolVersionKey || attrs[2].Value.AsString() != "network-protocol-version" {
		t.Fatalf("wrong network protocol version")
	}
	if attrs[3].Key != semconv.NetworkTransportKey || attrs[3].Value.AsString() != "network-transport" {
		t.Fatalf("wrong network transport")
	}
	if attrs[4].Key != semconv.NetworkTypeKey || attrs[4].Value.AsString() != "network-type" {
		t.Fatalf("wrong network type")
	}
	if attrs[5].Key != semconv.NetworkProtocolNameKey || attrs[5].Value.AsString() != "network-protocol-name" {
		t.Fatalf("wrong network protocol name")
	}
	if attrs[6].Key != semconv.NetworkProtocolVersionKey || attrs[6].Value.AsString() != "network-protocol-version" {
		t.Fatalf("wrong network protocol version")
	}
	if attrs[7].Key != semconv.NetworkLocalAddressKey || attrs[7].Value.AsString() != "network-local-inet-address" {
		t.Fatalf("wrong network protocol inet address")
	}
	if attrs[8].Key != semconv.NetworkPeerAddressKey || attrs[8].Value.AsString() != "network-peer-inet-address" {
		t.Fatalf("wrong network peer address")
	}
	if attrs[9].Key != semconv.NetworkLocalPortKey || attrs[9].Value.AsInt64() != 8080 {
		t.Fatalf("wrong network local port")
	}
	if attrs[10].Key != semconv.NetworkPeerPortKey || attrs[10].Value.AsInt64() != 8080 {
		t.Fatalf("wrong network peer port")
	}
}

func TestHttpServerExtractorStart(t *testing.T) {
	httpServerExtractor := HttpServerAttrsExtractor[testRequest, testResponse, httpServerAttrsGetter, networkAttrsGetter, urlAttrsGetter]{
		base:             HttpCommonAttrsExtractor[testRequest, testResponse, httpServerAttrsGetter, networkAttrsGetter]{},
		networkExtractor: net.NetworkAttrsExtractor[testRequest, testResponse, networkAttrsGetter]{},
		urlExtractor:     net.UrlAttrsExtractor[testRequest, testResponse, urlAttrsGetter]{},
	}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	attrs = httpServerExtractor.OnStart(attrs, parentContext, testRequest{})
	if attrs[0].Key != semconv.HTTPRequestMethodKey || attrs[0].Value.AsString() != "GET" {
		t.Fatalf("http method should be GET")
	}
	if attrs[1].Key != semconv.URLSchemeKey || attrs[1].Value.AsString() != "url-scheme" {
		t.Fatalf("urlscheme should be url-scheme")
	}
	if attrs[2].Key != semconv.URLPathKey || attrs[2].Value.AsString() != "url-path" {
		t.Fatalf("urlpath should be url-path")
	}
	if attrs[3].Key != semconv.URLQueryKey || attrs[3].Value.AsString() != "url-query" {
		t.Fatalf("urlquery should be url-query")
	}
	if attrs[4].Key != semconv.HTTPRouteKey || attrs[4].Value.AsString() != "http-route" {
		t.Fatalf("httproute should be http-route")
	}
	if attrs[5].Key != semconv.UserAgentOriginalKey || attrs[5].Value.AsString() != "request-header" {
		t.Fatalf("user agent original should be request-header")
	}
}

func TestHttpServerExtractorEnd(t *testing.T) {
	httpServerExtractor := HttpServerAttrsExtractor[testRequest, testResponse, httpServerAttrsGetter, networkAttrsGetter, urlAttrsGetter]{
		base:             HttpCommonAttrsExtractor[testRequest, testResponse, httpServerAttrsGetter, networkAttrsGetter]{},
		networkExtractor: net.NetworkAttrsExtractor[testRequest, testResponse, networkAttrsGetter]{},
		urlExtractor:     net.UrlAttrsExtractor[testRequest, testResponse, urlAttrsGetter]{},
	}
	attrs := make([]attribute.KeyValue, 0)
	parentContext := context.Background()
	attrs = httpServerExtractor.OnEnd(attrs, parentContext, testRequest{}, testResponse{}, nil)
	if attrs[0].Key != semconv.HTTPResponseStatusCodeKey || attrs[0].Value.AsInt64() != 200 {
		t.Fatalf("status code should be 200")
	}
	if attrs[1].Key != semconv.NetworkProtocolNameKey || attrs[1].Value.AsString() != "network-protocol-name" {
		t.Fatalf("wrong network protocol name")
	}
	if attrs[2].Key != semconv.NetworkProtocolVersionKey || attrs[2].Value.AsString() != "network-protocol-version" {
		t.Fatalf("wrong network protocol version")
	}
	if attrs[3].Key != semconv.NetworkTransportKey || attrs[3].Value.AsString() != "network-transport" {
		t.Fatalf("wrong network transport")
	}
	if attrs[4].Key != semconv.NetworkTypeKey || attrs[4].Value.AsString() != "network-type" {
		t.Fatalf("wrong network type")
	}
	if attrs[5].Key != semconv.NetworkProtocolNameKey || attrs[5].Value.AsString() != "network-protocol-name" {
		t.Fatalf("wrong network protocol name")
	}
	if attrs[6].Key != semconv.NetworkProtocolVersionKey || attrs[6].Value.AsString() != "network-protocol-version" {
		t.Fatalf("wrong network protocol version")
	}
	if attrs[7].Key != semconv.NetworkLocalAddressKey || attrs[7].Value.AsString() != "network-local-inet-address" {
		t.Fatalf("wrong network protocol inet address")
	}
	if attrs[8].Key != semconv.NetworkPeerAddressKey || attrs[8].Value.AsString() != "network-peer-inet-address" {
		t.Fatalf("wrong network peer address")
	}
	if attrs[9].Key != semconv.NetworkLocalPortKey || attrs[9].Value.AsInt64() != 8080 {
		t.Fatalf("wrong network local port")
	}
	if attrs[10].Key != semconv.NetworkPeerPortKey || attrs[10].Value.AsInt64() != 8080 {
		t.Fatalf("wrong network peer port")
	}
}
