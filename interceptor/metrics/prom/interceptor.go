// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkginmetrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-gin/interceptor/basic"
	"github.com/rookie-ninja/rk-prom"
	"strconv"
	"strings"
	"time"
)

var (
	DefaultLabelKeys = []string{
		"entryName",
		"entryType",
		"realm",
		"region",
		"az",
		"domain",
		"instance",
		"appVersion",
		"appName",
		"restMethod",
		"restPath",
		"type",
		"resCode",
	}
)

const (
	ElapsedNano = "elapsedNano"
	Errors      = "errors"
	ResCode     = "resCode"
	unknown     = "unknown"
)

// Create a new prometheus metrics intercepter with options.
func MetricsPromInterceptor(opts ...Option) gin.HandlerFunc {
	set := &optionSet{
		EntryName:  rkginbasic.RkEntryNameValue,
		EntryType:  rkginbasic.RkEntryTypeValue,
		Registerer: prometheus.DefaultRegisterer,
	}

	for i := range opts {
		opts[i](set)
	}

	if len(set.EntryName) > 0 && set.Registerer != nil {
		set.MetricsSet = rkprom.NewMetricsSet(
			rkentry.GlobalAppCtx.GetAppInfoEntry().AppName,
			set.EntryName,
			set.Registerer)
	} else {
		set.EntryName = rkginbasic.RkEntryNameValue
		set.Registerer = prometheus.DefaultRegisterer
		set.MetricsSet = rkprom.NewMetricsSet(
			rkentry.GlobalAppCtx.GetAppInfoEntry().AppName,
			set.EntryName,
			set.Registerer)
	}

	if _, ok := optionsMap[set.EntryName]; !ok {
		optionsMap[set.EntryName] = set
		// init server and client metrics
		initMetrics(set)
	}

	return func(ctx *gin.Context) {
		// start timer
		startTime := time.Now()

		ctx.Next()

		// end timer
		elapsed := time.Now().Sub(startTime)

		// ignoring /rk/v1/assets, /rk/v1/tv and /sw/ path while logging since these are internal APIs.
		if !strings.HasPrefix(ctx.Request.RequestURI, "/rk/v1/assets") &&
			!strings.HasPrefix(ctx.Request.RequestURI, "/rk/v1/tv") &&
			!strings.HasPrefix(ctx.Request.RequestURI, "/sw/") {
			if durationMetrics := GetServerDurationMetrics(ctx); durationMetrics != nil {
				durationMetrics.Observe(float64(elapsed.Nanoseconds()))
			}
			if len(ctx.Errors) > 0 {
				if errorMetrics := GetServerErrorMetrics(ctx); errorMetrics != nil {
					errorMetrics.Inc()
				}
			}
			if resCodeMetrics := GetServerResCodeMetrics(ctx); resCodeMetrics != nil {
				resCodeMetrics.Inc()
			}
		}
	}
}

// Register bellow metrics into metrics set.
// 1: Request elapsed time with summary.
// 2: Error count with counter.
// 3: ResCode count with counter.
func initMetrics(opts *optionSet) {
	opts.MetricsSet.RegisterSummary(ElapsedNano, rkprom.SummaryObjectives, DefaultLabelKeys...)
	opts.MetricsSet.RegisterCounter(Errors, DefaultLabelKeys...)
	opts.MetricsSet.RegisterCounter(ResCode, DefaultLabelKeys...)
}

// Server request elapsed metrics.
func GetServerDurationMetrics(ctx *gin.Context) prometheus.Observer {
	if metricsSet := GetServerMetricsSet(ctx); metricsSet != nil {
		return metricsSet.GetSummaryWithValues(ElapsedNano, getValues(ctx)...)
	}

	return nil
}

// Server error metrics.
func GetServerErrorMetrics(ctx *gin.Context) prometheus.Counter {
	if ctx == nil {
		return nil
	}

	if metricsSet := GetServerMetricsSet(ctx); metricsSet != nil {
		return metricsSet.GetCounterWithValues(Errors, getValues(ctx)...)
	}

	return nil
}

// Server response code metrics.
func GetServerResCodeMetrics(ctx *gin.Context) prometheus.Counter {
	if ctx == nil {
		return nil
	}

	if metricsSet := GetServerMetricsSet(ctx); metricsSet != nil {
		return metricsSet.GetCounterWithValues(ResCode, getValues(ctx)...)
	}

	return nil
}

// Server metrics set.
func GetServerMetricsSet(ctx *gin.Context) *rkprom.MetricsSet {
	if set := GetOptionSet(ctx); set != nil {
		return set.MetricsSet
	}

	return nil
}

// List all server metrics set associate with GinEntry.
func ListServerMetricsSets() []*rkprom.MetricsSet {
	res := make([]*rkprom.MetricsSet, 0)
	for _, v := range optionsMap {
		res = append(res, v.MetricsSet)
	}

	return res
}

// metrics set already set into context
func getValues(ctx *gin.Context) []string {
	entryName, entryType, method, path, resCode := unknown, unknown, unknown, unknown, unknown
	if ctx != nil && ctx.Request != nil {
		method = ctx.Request.Method
		if ctx.Request.URL != nil {
			path = ctx.Request.URL.Path
		}

		if ctx.Writer != nil {
			resCode = strconv.Itoa(ctx.Writer.Status())
		}
	}

	if set := GetOptionSet(ctx); set != nil {
		entryName = set.EntryName
		entryType = set.EntryType
	}

	values := []string{
		entryName,
		entryType,
		rkginbasic.Realm.String,
		rkginbasic.Region.String,
		rkginbasic.AZ.String,
		rkginbasic.Domain.String,
		rkginbasic.LocalHostname.String,
		rkentry.GlobalAppCtx.GetAppInfoEntry().Version,
		rkentry.GlobalAppCtx.GetAppInfoEntry().AppName,
		method,
		path,
		"ginServer",
		resCode,
	}

	return values
}

// Internal use only.
func clearAllMetrics() {
	for _, v := range optionsMap {
		v.MetricsSet.UnRegisterSummary(ElapsedNano)
		v.MetricsSet.UnRegisterCounter(Errors)
		v.MetricsSet.UnRegisterCounter(ResCode)
	}

	optionsMap = make(map[string]*optionSet)
}

// Global map stores metrics sets
// Interceptor would distinguish metrics set based on
var optionsMap = make(map[string]*optionSet)

// options which is used while initializing logging interceptor
type optionSet struct {
	EntryName  string
	EntryType  string
	Registerer prometheus.Registerer
	MetricsSet *rkprom.MetricsSet
}

type Option func(*optionSet)

// Provide entry name and entry type.
func WithEntryNameAndType(entryName, entryType string) Option {
	return func(opt *optionSet) {
		if len(entryName) > 0 {
			opt.EntryName = entryName
		}

		if len(entryType) > 0 {
			opt.EntryType = entryType
		}
	}
}

// Provide prometheus.Registerer.
func WithRegisterer(registerer prometheus.Registerer) Option {
	return func(opt *optionSet) {
		if registerer != nil {
			opt.Registerer = registerer
		}
	}
}

// Get optionSet with gin.Context.
func GetOptionSet(ctx *gin.Context) *optionSet {
	if ctx == nil {
		return nil
	}

	entryName := ctx.GetString(rkginbasic.RkEntryNameKey)
	return optionsMap[entryName]
}
