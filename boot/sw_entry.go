// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkgin

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/markbates/pkger"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-entry/entry"
	_ "github.com/rookie-ninja/rk-gin/boot/assets/sw/config"
	"github.com/rookie-ninja/rk-query"
	"github.com/swaggo/swag"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

var (
	swaggerJsonFiles     = make(map[string]string, 0)
	swConfigFileContents = ``
)

const (
	SwEntryType        = "GinSwEntry"
	SwEntryNameDefault = "GinSwDefault"
	SwEntryDescription = "Internal RK entry which implements swagger with Gin framework."
)

// Inner struct used while initializing swagger entry.
type swUrlConfig struct {
	Urls []*swUrl `json:"urls" yaml:"urls"`
}

// Inner struct used while initializing swagger entry.
type swUrl struct {
	Name string `json:"name" yaml:"name"`
	Url  string `json:"url" yaml:"url"`
}

// Bootstrap config of swagger.
// 1: Enabled: Enable swagger.
// 2: Path: Swagger path accessible from restful API.
// 3: JsonPath: The path of where swagger JSON file was located.
// 4: Headers: The headers that would added into each API response.
type BootConfigSw struct {
	Enabled  bool     `yaml:"enabled" yaml:"enabled"`
	Path     string   `yaml:"path" yaml:"path"`
	JsonPath string   `yaml:"jsonPath" yaml:"jsonPath"`
	Headers  []string `yaml:"headers" yaml:"headers"`
}

// SwEntry implements rkentry.Entry interface.
// 1: Path: Swagger path accessible from restful API.
// 2: JsonPath: The path of where swagger JSON file was located.
// 3: Headers: The headers that would added into each API response.
// 4: Port: The port where swagger would listen to.
// 5: EnableCommonService: Enable common service in swagger.
type SwEntry struct {
	EntryName           string                    `json:"entryName" yaml:"entryName"`
	EntryType           string                    `json:"entryType" yaml:"entryType"`
	EntryDescription    string                    `json:"entryDescription" yaml:"entryDescription"`
	EventLoggerEntry    *rkentry.EventLoggerEntry `json:"eventLoggerEntry" yaml:"eventLoggerEntry"`
	ZapLoggerEntry      *rkentry.ZapLoggerEntry   `json:"zapLoggerEntry" yaml:"zapLoggerEntry"`
	JsonPath            string                    `json:"jsonPath" yaml:"jsonPath"`
	Path                string                    `json:"path" yaml:"path"`
	Headers             map[string]string         `json:"headers" yaml:"headers"`
	Port                uint64                    `json:"port" yaml:"port"`
	EnableCommonService bool                      `json:"enableCommonService" yaml:"enableCommonService"`
}

// Swagger entry option.
type SwOption func(*SwEntry)

// Provide port.
func WithPortSw(port uint64) SwOption {
	return func(entry *SwEntry) {
		entry.Port = port
	}
}

// Provide name.
func WithNameSw(name string) SwOption {
	return func(entry *SwEntry) {
		entry.EntryName = name
	}
}

// Provide path.
func WithPathSw(path string) SwOption {
	return func(entry *SwEntry) {
		if len(path) < 1 {
			path = "sw"
		}
		entry.Path = path
	}
}

// Provide JsonPath.
func WithJsonPathSw(path string) SwOption {
	return func(entry *SwEntry) {
		entry.JsonPath = path
	}
}

// Provide headers.
func WithHeadersSw(headers map[string]string) SwOption {
	return func(entry *SwEntry) {
		entry.Headers = headers
	}
}

// Provide rkentry.ZapLoggerEntry.
func WithZapLoggerEntrySw(zapLoggerEntry *rkentry.ZapLoggerEntry) SwOption {
	return func(entry *SwEntry) {
		entry.ZapLoggerEntry = zapLoggerEntry
	}
}

// Provide rkentry.EventLoggerEntry.
func WithEventLoggerEntrySw(eventLoggerEntry *rkentry.EventLoggerEntry) SwOption {
	return func(entry *SwEntry) {
		entry.EventLoggerEntry = eventLoggerEntry
	}
}

// Provide enable common service option.
func WithEnableCommonServiceSw(enable bool) SwOption {
	return func(entry *SwEntry) {
		entry.EnableCommonService = enable
	}
}

// Create new swagger entry with options.
func NewSwEntry(opts ...SwOption) *SwEntry {
	entry := &SwEntry{
		EntryName:        SwEntryNameDefault,
		EntryType:        SwEntryType,
		EntryDescription: SwEntryDescription,
		ZapLoggerEntry:   rkentry.GlobalAppCtx.GetZapLoggerEntryDefault(),
		EventLoggerEntry: rkentry.GlobalAppCtx.GetEventLoggerEntryDefault(),
		Path:             "sw",
	}

	for i := range opts {
		opts[i](entry)
	}

	// Deal with Path
	// add "/" at start and end side if missing
	if !strings.HasPrefix(entry.Path, "/") {
		entry.Path = "/" + entry.Path
	}

	if !strings.HasSuffix(entry.Path, "/") {
		entry.Path = entry.Path + "/"
	}

	if len(entry.EntryName) < 1 {
		entry.EntryName = "ginSwEntry-" + strconv.FormatUint(entry.Port, 10)
	}

	// init swagger configs
	entry.initSwaggerConfig()

	return entry
}

// Bootstrap swagger entry.
func (entry *SwEntry) Bootstrap(context.Context) {
	// No op
	event := entry.EventLoggerEntry.GetEventHelper().Start(
		"bootstrap",
		rkquery.WithEntryName(entry.EntryName),
		rkquery.WithEntryType(entry.EntryType))

	entry.logBasicInfo(event)

	defer entry.EventLoggerEntry.GetEventHelper().Finish(event)

	entry.ZapLoggerEntry.GetLogger().Info("Bootstrapping SwEntry.", event.GetFields()...)
}

// Interrupt swagger entry.
func (entry *SwEntry) Interrupt(context.Context) {
	// No op
	event := entry.EventLoggerEntry.GetEventHelper().Start(
		"interrupt",
		rkquery.WithEntryName(entry.EntryName),
		rkquery.WithEntryType(entry.EntryType))

	entry.logBasicInfo(event)

	defer entry.EventLoggerEntry.GetEventHelper().Finish(event)

	entry.ZapLoggerEntry.GetLogger().Info("Interrupting SwEntry.", event.GetFields()...)
}

// Get name of entry.
func (entry *SwEntry) GetName() string {
	return entry.EntryName
}

// Get type of entry.
func (entry *SwEntry) GetType() string {
	return entry.EntryType
}

// Get description of entry
func (entry *SwEntry) GetDescription() string {
	return entry.EntryDescription
}

// Stringfy swagger entry
func (entry *SwEntry) String() string {
	bytes, _ := json.Marshal(entry)
	return string(bytes)
}

// Marshal entry
func (entry *SwEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"entryName":           entry.EntryName,
		"entryType":           entry.EntryType,
		"entryDescription":    entry.EntryDescription,
		"eventLoggerEntry":    entry.EventLoggerEntry.GetName(),
		"zapLoggerEntry":      entry.ZapLoggerEntry.GetName(),
		"jsonPath":            entry.JsonPath,
		"port":                entry.Port,
		"path":                entry.Path,
		"headers":             entry.Headers,
		"enableCommonService": entry.EnableCommonService,
	}

	return json.Marshal(&m)
}

// Unmarshal entry
func (entry *SwEntry) UnmarshalJSON([]byte) error {
	return nil
}

// Add basic fields into event
func (entry *SwEntry) logBasicInfo(event rkquery.Event) {
	event.AddFields(
		zap.String("entryName", entry.EntryName),
		zap.String("entryType", entry.EntryType),
		zap.String("jsonPath", entry.JsonPath),
		zap.String("path", entry.Path),
		zap.Uint64("port", entry.Port),
	)
}

func (entry *SwEntry) AssetsFileHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		w := ctx.Writer
		r := ctx.Request

		p := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/rk/v1"), "/")

		if file, err := pkger.Open(path.Join("/boot", p)); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		} else {
			http.ServeContent(w, r, path.Base(p), time.Now(), file)
		}
	}
}

func (entry *SwEntry) ConfigFileHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		w := ctx.Writer
		r := ctx.Request

		p := strings.TrimSuffix(r.URL.Path, "/")

		w.Header().Set("cache-control", "no-cache")

		for k, v := range entry.Headers {
			w.Header().Set(k, v)
		}

		switch p {
		case "/sw":
			if file, err := pkger.Open(path.Join("/boot/assets/sw/index.html")); err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			} else {
				http.ServeContent(w, r, "index.html", time.Now(), file)
			}
		case "/sw/swagger-config.json":
			http.ServeContent(w, r, "swagger-config.json", time.Now(), strings.NewReader(swConfigFileContents))
		default:
			p = strings.TrimPrefix(p, "/sw/")
			value, ok := swaggerJsonFiles[p]

			if ok {
				http.ServeContent(w, r, p, time.Now(), strings.NewReader(value))
			} else {
				http.NotFound(w, r)
			}
		}
	}
}

// Init swagger config.
// This function do the things bellow:
// 1: List swagger files from entry.JSONPath.
// 2: Read user swagger json files and deduplicate.
// 3: Assign swagger contents into swaggerConfigJson variable
func (entry *SwEntry) initSwaggerConfig() {
	swaggerUrlConfig := &swUrlConfig{
		Urls: make([]*swUrl, 0),
	}

	// 1: Add user API swagger JSON
	entry.listFilesWithSuffix(swaggerUrlConfig)

	// 2: Add rk common APIs
	if entry.EnableCommonService {
		key := entry.EntryName + "-rk-common.swagger.json"
		// add common service json file
		swaggerJsonFiles[key], _ = swag.ReadDoc()
		swaggerUrlConfig.Urls = append(swaggerUrlConfig.Urls, &swUrl{
			Name: key,
			Url:  path.Join(entry.Path, key),
		})
	}

	// 3: Marshal to swagger-config.json and write to pkger
	bytes, err := json.Marshal(swaggerUrlConfig)
	if err != nil {
		entry.ZapLoggerEntry.GetLogger().Error("Failed to unmarshal swagger-config.json",
			zap.Error(err))
		rkcommon.ShutdownWithError(err)
	}

	swConfigFileContents = string(bytes)
}

// List files with .json suffix and store them into swaggerJsonFiles variable.
func (entry *SwEntry) listFilesWithSuffix(urlConfig *swUrlConfig) {
	jsonPath := entry.JsonPath
	suffix := ".json"
	// re-path it with working directory if not absolute path
	if !path.IsAbs(entry.JsonPath) {
		wd, err := os.Getwd()
		if err != nil {
			entry.ZapLoggerEntry.GetLogger().Info("Failed to get working directory",
				zap.String("error", err.Error()))
			rkcommon.ShutdownWithError(err)
		}
		jsonPath = path.Join(wd, jsonPath)
	}

	files, err := ioutil.ReadDir(jsonPath)
	if err != nil {
		entry.ZapLoggerEntry.GetLogger().Error("Failed to list files with suffix",
			zap.String("path", jsonPath),
			zap.String("suffix", suffix),
			zap.String("error", err.Error()))
		rkcommon.ShutdownWithError(err)
	}

	for i := range files {
		file := files[i]
		if !file.IsDir() && strings.HasSuffix(file.Name(), suffix) {
			bytes, err := ioutil.ReadFile(path.Join(jsonPath, file.Name()))
			key := entry.EntryName + "-" + file.Name()

			if err != nil {
				entry.ZapLoggerEntry.GetLogger().Info("Failed to read file with suffix",
					zap.String("path", path.Join(jsonPath, key)),
					zap.String("suffix", suffix),
					zap.String("error", err.Error()))
				rkcommon.ShutdownWithError(err)
			}

			swaggerJsonFiles[key] = string(bytes)

			urlConfig.Urls = append(urlConfig.Urls, &swUrl{
				Name: key,
				Url:  path.Join(entry.Path, key),
			})
		}
	}
}