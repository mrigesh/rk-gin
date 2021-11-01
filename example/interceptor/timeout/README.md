# Timeout interceptor
In this example, we will try to create gin server with timeout interceptor enabled.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Quick start](#quick-start)
  - [Code](#code)
- [Options](#options)
  - [Context Usage](#context-usage)
- [Example](#example)
    - [Start server](#start-server)
    - [Output](#output)
    - [Code](#code-1)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Quick start
Get rk-gin package from the remote repository.

```go
go get -u github.com/rookie-ninja/rk-gin
```

### Code
Add rkgintimeout.Interceptor() with option.

```go
import     "github.com/rookie-ninja/rk-gin/interceptor/timeout"
```
```go
    // ********************************************
    // ********** Enable interceptors *************
    // ********************************************
    interceptors := []gin.HandlerFunc{
        rkgintimeout.Interceptor(),
    }
```

## Options
| Name | Default | Description |
| ---- | ---- | ---- |
| WithEntryNameAndType(entryName, entryType string) | entryName=gin, entryType=gin | entryName and entryType will be used to distinguish options if there are multiple interceptors in single process. |
| WithTimeoutAndResp(time.Duration, gin.HandlerFunc) | 5*time.Second, response with http.StatusRequestTimeout | Set timeout interceptor with all routes. |
| WithTimeoutAndRespByPath(path string, time.Duration, gin.HandlerFunc) | "", 5*time.Second, response with http.StatusRequestTimeout | Set timeout interceptor with specified path. |

```go
	// ********************************************
	// ********** Enable interceptors *************
	// ********************************************
	interceptors := []gin.HandlerFunc{
		rkginpanic.Interceptor(),
		rkginlog.Interceptor(),
		rkgintimeout.Interceptor(
			// Entry name and entry type will be used for distinguishing interceptors. Recommended.
			// rkgintimeout.WithEntryNameAndType("greeter", "gin"),
			//
			// Provide timeout and response handler, a default one would be assigned with http.StatusRequestTimeout
			// This option impact all routes
			// rkgintimeout.WithTimeoutAndResp(time.Second, nil),
			//
			// Provide timeout and response handler by path, a default one would be assigned with http.StatusRequestTimeout
			// rkgintimeout.WithTimeoutAndRespByPath("/rk/v1/healthy", time.Second, nil),
		),
	}
```

### Context Usage
| Name | Functionality |
| ------ | ------ |
| rkginctx.GetLogger(*gin.Context) | Get logger generated by log interceptor. If there are X-Request-Id or X-Trace-Id as headers in incoming and outgoing metadata, then loggers will has requestId and traceId attached by default. |
| rkginctx.GetEvent(*gin.Context) | Get event generated by log interceptor. Event would be printed as soon as RPC finished. |
| rkginctx.GetIncomingHeaders(*gin.Context) | Get incoming header. |
| rkginctx.AddHeaderToClient(ctx, "k", "v") | Add k/v to headers which would be sent to client. This is append operation. |
| rkginctx.SetHeaderToClient(ctx, "k", "v") | Set k/v to headers which would be sent to client. |

## Example
In this example, we enable log and panic interceptor either to monitor RPC status.

#### Start server
```shell script
$ go run greeter-server.go
```

#### Output
- Response

```
$ curl "localhost:8080/rk/v1/greeter?name=rk-dev"
{
    "error":{
        "code":408,
        "status":"Request Timeout",
        "message":"Request timed out!",
        "details":[]
    }
}
```

- Server side (zap & event)

```shell script
2021-10-29T21:30:44.413+0800    INFO    timeout/greeter-server.go:95    Received request from client.
```

```shell script
------------------------------------------------------------------------
endTime=2021-10-29T21:30:49.416786+08:00
startTime=2021-10-29T21:30:44.413405+08:00
elapsedNano=5003376246
timezone=CST
ids={"eventId":"95a50e31-7a82-47ae-b0f4-8b3ad7a9b06d"}
app={"appName":"rk","appVersion":"","entryName":"gin","entryType":"gin"}
env={"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"10.8.0.2","os":"darwin","realm":"*","region":"*"}
payloads={"apiMethod":"GET","apiPath":"/rk/v1/greeter","apiProtocol":"HTTP/1.1","apiQuery":"name=rk-dev","userAgent":"curl/7.64.1"}
error={}
counters={"timeout":1}
pairs={}
timing={}
remoteAddr=127.0.0.1:57989
operation=/rk/v1/greeter
resCode=408
eventStatus=Ended
EOE
```

#### Code
- [greeter-server.go](greeter-server.go)