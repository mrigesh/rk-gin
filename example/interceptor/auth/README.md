# Auth interceptor (middleware)
In this example, we will try to create gin server with auth interceptor enabled.

Auth interceptor will validate bellow authorizations.

| Type | Description | Example |
| ---- | ---- | ---- |
| Basic Auth | The client sends HTTP requests with the Authorization header that contains the word Basic, followed by a space and a base64-encoded(non-encrypted) string username: password. | Authorization: Basic AXVubzpwQDU1dzByYM== |
| Bearer Token | Commonly known as token authentication. It is an HTTP authentication scheme that involves security tokens called bearer tokens. | Authorization: Bearer [token] |
| API Key | An API key is a token that a client provides when making API calls. With API key auth, you send a key-value pair to the API in the request headers. | X-API-Key: abcdefgh123456789 | 

**Please make sure panic interceptor to be added at last in chain of interceptors.**

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Quick start](#quick-start)
  - [Code](#code)
- [Options](#options)
  - [Context Usage](#context-usage)
- [Example](#example)
  - [Start server](#start-server)
  - [Output unauthorized](#output-unauthorized)
  - [Output authorized](#output-authorized)
  - [Code](#code-1)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Quick start
Get rk-gin package from the remote repository.

```go
go get -u github.com/rookie-ninja/rk-gin
```

### Code
```go
import    "github.com/rookie-ninja/rk-gin/interceptor/auth"
```
```go
    // ********************************************
    // ********** Enable interceptors *************
    // ********************************************
    interceptors := []gin.HandlerFunc{
        rkginlog.Interceptor(),
        rkginauth.Interceptor(
            rkginauth.WithBasicAuth("", "rk-user:rk-pass"),
            rkginauth.WithBearerAuth("rk-token"),
            rkginauth.WithApiKeyAuth("rk-api-key"),
        ),
    }
```

## Options
Auth interceptor validate authorization for each request.

![server-arch](img/server-arch.png)

| Name | Default | Description |
| ---- | ---- | ---- |
| WithEntryNameAndType(entryName, entryType string) | entryName=gin, entryType=gin | entryName and entryType will be used to distinguish options if there are multiple interceptors in single process. |
| WithBasicAuth(realm string, cred ...string) | []string | Provide Basic auth credential with scheme of [user:pass]. Multiple credential are available for server. |
| WithBearerAuth(token ...string) | []string | Provide Bearer token. Multiple tokens are available for server. |
| WithApiKeyAuth(key ...string) | []string | Provide API key. Multiple keys are available for server. |

```go
    // ********************************************
    // ********** Enable interceptors *************
    // ********************************************
    interceptors := []gin.HandlerFunc{
        rkginlog.Interceptor(),
        rkginauth.Interceptor(
            rkginauth.WithBasicAuth("", "rk-user:rk-pass"),
            rkginauth.WithBearerAuth("rk-token"),
            rkginauth.WithApiKeyAuth("rk-api-key"),
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
In this example, we enable log interceptor either to monitor RPC.

### Start server
```shell script
$ go run greeter-server.go
```

### Output unauthorized
- Server side (event)
```shell script
------------------------------------------------------------------------
endTime=2021-06-24T20:16:30.011298+08:00
startTime=2021-06-24T20:16:30.011208+08:00
elapsedNano=90163
timezone=CST
ids={"eventId":"cc695229-0775-4caa-ac84-cdbb2eb11ca7"}
app={"appName":"rk","appVersion":"v0.0.0","entryName":"gin","entryType":"gin"}
env={"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"10.8.0.2","os":"darwin","realm":"*","region":"*"}
payloads={"apiMethod":"GET","apiPath":"/rk/example/greeter","apiProtocol":"HTTP/1.1","apiQuery":"name=rk-dev","userAgent":"curl/7.64.1"}
error={}
counters={}
pairs={}
timing={}
remoteAddr=localhost:65449
operation=/rk/example/greeter
resCode=401
eventStatus=Ended
EOE
```

- Client side 
```shell script
$ curl "localhost:8080/rk/example/greeter?name=rk-dev"
# Pretty print manually.
{
    "error":{
        "code":401,
        "status":"Unauthorized",
        "message":"Missing authorization",
        "details":[]
    }
}
```

### Output authorized
- Server side (event)
```shell script
------------------------------------------------------------------------
endTime=2021-06-24T20:46:29.987694+08:00
startTime=2021-06-24T20:46:29.987572+08:00
elapsedNano=121714
timezone=CST
ids={"eventId":"6250b92d-e54d-4250-bd7c-bc39047da471"}
app={"appName":"rk","appVersion":"v0.0.0","entryName":"gin","entryType":"gin"}
env={"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"10.8.0.2","os":"darwin","realm":"*","region":"*"}
payloads={"apiMethod":"GET","apiPath":"/rk/v1/greeter","apiProtocol":"HTTP/1.1","apiQuery":"name=rk-dev","userAgent":"curl/7.64.1"}
error={}
counters={}
pairs={}
timing={}
remoteAddr=localhost:55755
operation=/rk/v1/greeter
resCode=200
eventStatus=Ended
EOE
```

- Client side 
```shell script
# With basic auth
$ curl -u "rk-user:rk-pass" "localhost:8080/rk/v1/greeter?name=rk-dev"
{"Message":"Hello rk-dev!"}

# With bearer token
$ curl "localhost:8080/rk/v1/greeter?name=rk-dev" -H "Authorization: Bearer rk-token"
{"Message":"Hello rk-dev!"}

# With API key
$ curl "localhost:8080/rk/v1/greeter?name=rk-dev" -H "X-API-Key: rk-api-key"
{"Message":"Hello rk-dev!"}
```

### Code
- [greeter-server.go](greeter-server.go)