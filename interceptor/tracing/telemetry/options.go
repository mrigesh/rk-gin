// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkgintrace is a middleware of gin framework for recording tracing
package rkgintrace

import (
	"context"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-gin/interceptor"
	"github.com/rookie-ninja/rk-logger"
	"go.opentelemetry.io/contrib"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	"os"
	"path"
	"strings"
)

// NoopExporter noop
type NoopExporter struct{}

// ExportSpans handles export of SpanSnapshots by dropping them.
func (nsb *NoopExporter) ExportSpans(context.Context, []*sdktrace.SpanSnapshot) error { return nil }

// Shutdown stops the exporter by doing nothing.
func (nsb *NoopExporter) Shutdown(context.Context) error { return nil }

// CreateNoopExporter create a noop exporter
func CreateNoopExporter() sdktrace.SpanExporter {
	return &NoopExporter{}
}

// CreateFileExporter create a file exporter whose default output is stdout.
func CreateFileExporter(outputPath string, opts ...stdout.Option) sdktrace.SpanExporter {
	if opts == nil {
		opts = make([]stdout.Option, 0)
	}

	if outputPath == "" {
		outputPath = "stdout"
	}

	if outputPath == "stdout" {
		opts = append(opts,
			stdout.WithPrettyPrint(),
			stdout.WithoutMetricExport())
	} else {
		// init lumberjack logger
		writer := rklogger.NewLumberjackConfigDefault()
		if !path.IsAbs(outputPath) {
			wd, _ := os.Getwd()
			outputPath = path.Join(wd, outputPath)
		}

		writer.Filename = outputPath

		opts = append(opts, stdout.WithWriter(writer), stdout.WithoutMetricExport())
	}

	exporter, _ := stdout.NewExporter(opts...)

	return exporter
}

// CreateJaegerExporter in beta stage
// TODO: Wait for opentelemetry update version of jeager exporter. Current exporter is not compatible with jaeger agent.
func CreateJaegerExporter(endpoint, username, password string) sdktrace.SpanExporter {
	if len(endpoint) < 1 {
		endpoint = "http://localhost:14268"
	}

	if !strings.HasPrefix(endpoint, "http://") {
		endpoint = "http://" + endpoint
	}

	exporter, err := jaeger.NewRawExporter(
		jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(endpoint+"/api/traces"),
			jaeger.WithUsername(username),
			jaeger.WithPassword(password)),
	)

	if err != nil {
		rkcommon.ShutdownWithError(err)
	}

	return exporter
}

// Interceptor would distinguish logs set based on.
var optionsMap = make(map[string]*optionSet)

// Create an optionSet with rpc type.
func newOptionSet(opts ...Option) *optionSet {
	set := &optionSet{
		EntryName: rkgininter.RpcEntryNameValue,
		EntryType: rkgininter.RpcEntryTypeValue,
	}

	for i := range opts {
		opts[i](set)
	}

	if set.Exporter == nil {
		set.Exporter = CreateNoopExporter()
	}

	if set.Processor == nil {
		set.Processor = sdktrace.NewBatchSpanProcessor(set.Exporter)
	}

	if set.Provider == nil {
		set.Provider = sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithSpanProcessor(set.Processor),
			sdktrace.WithResource(
				sdkresource.NewWithAttributes(
					attribute.String("service.name", rkentry.GlobalAppCtx.GetAppInfoEntry().AppName),
					attribute.String("service.version", rkentry.GlobalAppCtx.GetAppInfoEntry().Version),
					attribute.String("service.entryName", set.EntryName),
					attribute.String("service.entryType", set.EntryType),
				)),
		)
	}

	set.Tracer = set.Provider.Tracer(set.EntryName, oteltrace.WithInstrumentationVersion(contrib.SemVersion()))

	if set.Propagator == nil {
		set.Propagator = propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{})
	}

	if _, ok := optionsMap[set.EntryName]; !ok {
		optionsMap[set.EntryName] = set
	}

	return set
}

// Options which is used while initializing logging interceptor
type optionSet struct {
	EntryName  string
	EntryType  string
	Exporter   sdktrace.SpanExporter
	Processor  sdktrace.SpanProcessor
	Provider   *sdktrace.TracerProvider
	Propagator propagation.TextMapPropagator
	Tracer     oteltrace.Tracer
}

// Option is used while creating middleware as param
type Option func(*optionSet)

// Provide sdktrace.SpanExporter.
func WithExporter(exporter sdktrace.SpanExporter) Option {
	return func(opt *optionSet) {
		if exporter != nil {
			opt.Exporter = exporter
		}
	}
}

// WithSpanProcessor provide sdktrace.SpanProcessor.
func WithSpanProcessor(processor sdktrace.SpanProcessor) Option {
	return func(opt *optionSet) {
		if processor != nil {
			opt.Processor = processor
		}
	}
}

// WithTracerProvider provide *sdktrace.TracerProvider.
func WithTracerProvider(provider *sdktrace.TracerProvider) Option {
	return func(opt *optionSet) {
		if provider != nil {
			opt.Provider = provider
		}
	}
}

// WithPropagator provide propagation.TextMapPropagator.
func WithPropagator(propagator propagation.TextMapPropagator) Option {
	return func(opt *optionSet) {
		if propagator != nil {
			opt.Propagator = propagator
		}
	}
}

// WithEntryNameAndType provide entry name and entry type.
func WithEntryNameAndType(entryName, entryType string) Option {
	return func(opt *optionSet) {
		opt.EntryName = entryName
		opt.EntryType = entryType
	}
}

// ShutdownExporters shutdown all exporters.
func ShutdownExporters() {
	for _, v := range optionsMap {
		v.Exporter.Shutdown(context.Background())
	}
}
