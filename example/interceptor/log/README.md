# Log interceptor
In this example, we will try to create gin server with log interceptor enabled.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Quick start](#quick-start)
  - [Code](#code)
- [Options](#options)
  - [Encoding](#encoding)
  - [OutputPath](#outputpath)
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
```go
import     "github.com/rookie-ninja/rk-gin/interceptor/log/zap"
```

```go
    interceptors := []gin.HandlerFunc{
        rkginlog.Interceptor(),
    }
```

## Options
Log interceptor will init rkquery.Event, zap.Logger and entryName which will be injected into request context before user function.
As soon as user function returns, interceptor will write the event into files.

![arch](img/arch.png)

| Name | Default | Description |
| ---- | ---- | ---- |
| WithEntryNameAndType(entryName, entryType string) | entryName=grpc, entryType=grpc | entryName and entryType will be used to distinguish options if there are multiple interceptors in single process. |
| WithZapLoggerEntry(zapLoggerEntry *rkentry.ZapLoggerEntry) | [rkentry.GlobalAppCtx.GetZapLoggerEntryDefault()](https://github.com/rookie-ninja/rk-entry/blob/master/entry/context.go) | Zap logger would print to stdout with console encoding type. |
| WithEventLoggerEntry(eventLoggerEntry *rkentry.EventLoggerEntry) | [rkentry.GlobalAppCtx.GetEventLoggerEntryDefault()](https://github.com/rookie-ninja/rk-entry/blob/master/entry/context.go) | Event logger would print to stdout with console encoding type. |
| WithZapLoggerEncoding(ec int) | rkginlog.ENCODING_CONSOLE | rkginlog.ENCODING_CONSOLE and rkginlog.ENCODING_JSON are available options. |
| WithZapLoggerOutputPaths(path ...string) | stdout | Both absolute path and relative path is acceptable. Current working directory would be used if path is relative. |
| WithEventLoggerEncoding(ec int) | rkginlog.ENCODING_CONSOLE | rkginlog.ENCODING_CONSOLE and rkginlog.ENCODING_JSON are available options. |
| WithEventLoggerOutputPaths(path ...string) | stdout | Both absolute path and relative path is acceptable. Current working directory would be used if path is relative. |

```go
    // ********************************************
    // ********** Enable interceptors *************
    // ********************************************
    interceptors := []gin.HandlerFunc{
        rkginlog.Interceptor(
            // Entry name and entry type will be used for distinguishing interceptors. Recommended.
            // rkginlog.WithEntryNameAndType("greeter", "grpc"),
            //
            // Zap logger would be logged as JSON format.
            // rkginlog.WithZapLoggerEncoding(rkgrpclog.ENCODING_JSON),
            //
            // Event logger would be logged as JSON format.
            // rkginlog.WithEventLoggerEncoding(rkgrpclog.ENCODING_JSON),
            //
            // Zap logger would be logged to specified path.
            // rkginlog.WithZapLoggerOutputPaths("logs/server-zap.log"),
            //
            // Event logger would be logged to specified path.
            // rkginlog.WithEventLoggerOutputPaths("logs/server-event.log"),
        ),
    }
```

### Encoding
- CONSOLE
No options needs to be provided. 
```shell script
2021-06-24T21:01:24.633+0800    INFO    log/greeter-server.go:84        Received request from client.
```
```shell script
------------------------------------------------------------------------
endTime=2021-06-24T21:01:24.634+08:00
startTime=2021-06-24T21:01:24.633836+08:00
elapsedNano=164521
timezone=CST
ids={"eventId":"5cdcb51d-4e92-4d9e-afce-381bea1bf6de"}
app={"appName":"rk","appVersion":"v0.0.0","entryName":"gin","entryType":"gin"}
env={"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"10.8.0.2","os":"darwin","realm":"*","region":"*"}
payloads={"apiMethod":"GET","apiPath":"/rk/v1/greeter","apiProtocol":"HTTP/1.1","apiQuery":"name=rk-dev","userAgent":"curl/7.64.1"}
error={}
counters={}
pairs={}
timing={}
remoteAddr=localhost:59410
operation=/rk/v1/greeter
resCode=200
eventStatus=Ended
EOE
```

- JSON
```go
    // ********************************************
    // ********** Enable interceptors *************
    // ********************************************
    interceptors := []gin.HandlerFunc{
        rkginlog.Interceptor(
            // Zap logger would be logged as JSON format.
            rkginlog.WithZapLoggerEncoding(rkginlog.ENCODING_JSON),
            //
            // Event logger would be logged as JSON format.
            rkginlog.WithEventLoggerEncoding(rkginlog.ENCODING_JSON),
        ),
    }
```
```json
{"level":"INFO","ts":"2021-06-24T21:17:14.995+0800","msg":"Received request from client."}
```
```json
{"endTime": "2021-06-24T21:17:14.995+0800", "startTime": "2021-06-24T21:17:14.995+0800", "elapsedNano": 148030, "timezone": "CST", "ids": {"eventId":"03e71ee0-428f-4830-85b6-5ce56108907e"}, "app": {"appName":"rk","appVersion":"v0.0.0","entryName":"gin","entryType":"gin"}, "env": {"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"10.8.0.2","os":"darwin","realm":"*","region":"*"}, "payloads": {"apiMethod":"GET","apiPath":"/rk/v1/greeter","apiProtocol":"HTTP/1.1","apiQuery":"name=rk-dev","userAgent":"curl/7.64.1"}, "error": {}, "counters": {}, "pairs": {}, "timing": {}, "remoteAddr": "localhost:63320", "operation": "/rk/v1/greeter", "eventStatus": "Ended", "resCode": "200"}
```

### OutputPath
- Stdout
No options needs to be provided. 

- Files
```go
    // ********************************************
    // ********** Enable interceptors *************
    // ********************************************
    interceptors := []gin.HandlerFunc{
        rkginlog.Interceptor(
            // Event logger would be logged as JSON format.
            // rkginlog.WithEventLoggerEncoding(rkginlog.ENCODING_JSON),
            //
            // Zap logger would be logged to specified path.
            rkginlog.WithZapLoggerOutputPaths("logs/server-zap.log"),
            //
            // Event logger would be logged to specified path.
            rkginlog.WithEventLoggerOutputPaths("logs/server-event.log"),
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
In this example, we enable log interceptor.

#### Start server
```shell script
$ go run greeter-server.go
```

#### Output
- Server side (zap & event)
```shell script
2021-06-24T21:23:06.100+0800    INFO    log/greeter-server.go:84        Received request from client.
```
```shell script
------------------------------------------------------------------------
endTime=2021-06-24T21:23:06.100956+08:00
startTime=2021-06-24T21:23:06.10083+08:00
elapsedNano=125493
timezone=CST
ids={"eventId":"b3dd6eb6-316a-4f58-b8b0-3429ef46a2ea"}
app={"appName":"rk","appVersion":"v0.0.0","entryName":"gin","entryType":"gin"}
env={"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"10.8.0.2","os":"darwin","realm":"*","region":"*"}
payloads={"apiMethod":"GET","apiPath":"/rk/v1/greeter","apiProtocol":"HTTP/1.1","apiQuery":"name=rk-dev","userAgent":"curl/7.64.1"}
error={}
counters={}
pairs={}
timing={}
remoteAddr=localhost:64769
operation=/rk/v1/greeter
resCode=200
eventStatus=Ended
EOE
```

- Client side
```shell script
$ curl "localhost:8080/rk/v1/greeter?name=rk-dev"
{"Message":"Hello rk-dev!"}
```

### Code
- [greeter-server.go](greeter-server.go)