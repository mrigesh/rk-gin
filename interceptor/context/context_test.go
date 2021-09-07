// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkginctx

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-gin/interceptor"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"net/url"
	"os"
	"testing"
)

func NewMockResponseWriter() *MockResponseWriter {
	return &MockResponseWriter{
		data:   make([]byte, 0),
		header: http.Header{},
	}
}

type MockResponseWriter struct {
	data       []byte
	statusCode int
	header     http.Header
}

func (m *MockResponseWriter) Header() http.Header {
	return m.header
}

func (m *MockResponseWriter) Write(bytes []byte) (int, error) {
	m.data = bytes
	return len(bytes), nil
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}

func TestGetIncomingHeaders(t *testing.T) {
	header := http.Header{}
	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "ut-path",
		},
		Header: header,
	}

	assert.Equal(t, header, GetIncomingHeaders(ctx))
}

func TestAddHeaderToClient(t *testing.T) {
	defer assertNotPanic(t)

	// With nil context
	AddHeaderToClient(nil, "", "")

	// With nil writer
	ctx := &gin.Context{}
	AddHeaderToClient(ctx, "", "")

	// Happy case
	ctx, _ = gin.CreateTestContext(NewMockResponseWriter())
	AddHeaderToClient(ctx, "key", "value")
	assert.Equal(t, "value", ctx.Writer.Header().Get("key"))
}

func TestSetHeaderToClient(t *testing.T) {
	defer assertNotPanic(t)

	// With nil context
	SetHeaderToClient(nil, "", "")

	// With nil writer
	ctx := &gin.Context{}
	SetHeaderToClient(ctx, "", "")

	// Happy case
	ctx, _ = gin.CreateTestContext(NewMockResponseWriter())
	SetHeaderToClient(ctx, "key", "value")
	assert.Equal(t, "value", ctx.Writer.Header().Get("key"))
}

func TestGetEvent(t *testing.T) {
	// With nil context
	assert.Equal(t, noopEvent, GetEvent(nil))

	// With no event in context
	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	assert.Equal(t, noopEvent, GetEvent(ctx))

	// Happy case
	event := rkquery.NewEventFactory().CreateEventNoop()
	ctx.Set(rkgininter.RpcEventKey, event)
	assert.Equal(t, event, GetEvent(ctx))

}

func TestGetLogger(t *testing.T) {
	// With nil context
	assert.Equal(t, rklogger.NoopLogger, GetLogger(nil))

	// With no logger in context
	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	assert.Equal(t, rklogger.NoopLogger, GetLogger(ctx))

	// Happy case
	// Add request id and trace id
	ctx.Writer.Header().Set(RequestIdKey, "ut-request-id")
	ctx.Writer.Header().Set(TraceIdKey, "ut-trace-id")
	ctx.Set(rkgininter.RpcLoggerKey, rklogger.NoopLogger)

	assert.Equal(t, rklogger.NoopLogger, GetLogger(ctx))
}

func TestGetRequestId(t *testing.T) {
	// With nil context
	assert.Empty(t, GetRequestId(nil))

	// With no requestId in context
	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	assert.Empty(t, GetRequestId(ctx))

	// Happy case
	ctx.Writer.Header().Set(RequestIdKey, "ut-request-id")
	assert.Equal(t, "ut-request-id", GetRequestId(ctx))
}

func TestGetTraceId(t *testing.T) {
	// With nil context
	assert.Empty(t, GetTraceId(nil))

	// With no traceId in context
	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	assert.Empty(t, GetTraceId(ctx))

	// Happy case
	ctx.Writer.Header().Set(TraceIdKey, "ut-trace-id")
	assert.Equal(t, "ut-trace-id", GetTraceId(ctx))
}

func TestGetEntryName(t *testing.T) {
	// With nil context
	assert.Empty(t, GetEntryName(nil))

	// With no entry name in context
	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	assert.Empty(t, GetEntryName(ctx))

	// Happy case
	ctx.Set(rkgininter.RpcEntryNameKey, "ut-entry-name")
	assert.Equal(t, "ut-entry-name", GetEntryName(ctx))
}

func TestGetTraceSpan(t *testing.T) {
	// With nil context
	assert.NotNil(t, GetTraceSpan(nil))

	// With no span in context
	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	assert.NotNil(t, GetTraceSpan(ctx))

	// Happy case
	_, span := noopTracerProvider.Tracer("ut-trace").Start(ctx, "noop-span")
	ctx.Set(rkgininter.RpcSpanKey, span)
	assert.Equal(t, span, GetTraceSpan(ctx))
}

func TestGetTracer(t *testing.T) {
	// With nil context
	assert.NotNil(t, GetTracer(nil))

	// With no tracer in context
	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	assert.NotNil(t, GetTracer(ctx))

	// Happy case
	tracer := noopTracerProvider.Tracer("ut-trace")
	ctx.Set(rkgininter.RpcTracerKey, tracer)
	assert.Equal(t, tracer, GetTracer(ctx))
}

func TestGetTracerProvider(t *testing.T) {
	// With nil context
	assert.NotNil(t, GetTracerProvider(nil))

	// With no tracer provider in context
	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	assert.NotNil(t, GetTracerProvider(ctx))

	// Happy case
	provider := trace.NewNoopTracerProvider()
	ctx.Set(rkgininter.RpcTracerProviderKey, provider)
	assert.Equal(t, provider, GetTracerProvider(ctx))
}

func TestGetTracerPropagator(t *testing.T) {
	// With nil context
	assert.Nil(t, GetTracerPropagator(nil))

	// With no tracer propagator in context
	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	assert.Nil(t, GetTracerPropagator(ctx))

	// Happy case
	prop := propagation.NewCompositeTextMapPropagator()
	ctx.Set(rkgininter.RpcPropagatorKey, prop)
	assert.Equal(t, prop, GetTracerPropagator(ctx))
}

func TestInjectSpanToHttpRequest(t *testing.T) {
	assertNotPanic(t)

	// With nil context and request
	InjectSpanToHttpRequest(nil, nil)

	// Happy case
	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	prop := propagation.NewCompositeTextMapPropagator()
	ctx.Set(rkgininter.RpcPropagatorKey, prop)
	InjectSpanToHttpRequest(ctx, &http.Request{
		Header: http.Header{},
	})
}

func TestNewTraceSpan(t *testing.T) {
	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{}
	assert.NotNil(t, NewTraceSpan(ctx, "ut-span"))
}

func TestEndTraceSpan(t *testing.T) {
	assertNotPanic(t)

	// With success
	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	span := GetTraceSpan(ctx)
	EndTraceSpan(ctx, span, true)

	// With failure
	ctx, _ = gin.CreateTestContext(NewMockResponseWriter())
	span = GetTraceSpan(ctx)
	EndTraceSpan(ctx, span, false)
}

func assertNotPanic(t *testing.T) {
	if r := recover(); r != nil {
		// Expect panic to be called with non nil error
		assert.True(t, false)
	} else {
		// This should never be called in case of a bug
		assert.True(t, true)
	}
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.ReleaseMode)
	code := m.Run()
	os.Exit(code)
}
