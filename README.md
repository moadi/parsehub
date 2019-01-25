# Introduction

This small project implements the interview question for ParseHub using the go language.

# Building

Use the provided Makefile to build the "proxy" binary. It fetches the dependency from github (gorilla/mux)
and executes go build to build the binary. It specifies the GOPATH currently since I'm not sure which version
of go will be used for testing.

All my testing has been done on OSX 10.11

Sample output:

```
Mohammads-MacBook-Pro:parsehub adi$ go version
go version go1.11.4 darwin/amd64
Mohammads-MacBook-Pro:parsehub adi$ 
Mohammads-MacBook-Pro:parsehub adi$ ls -l
total 24
-rw-r--r--  1 adi  staff   279 Jan 25 00:06 Makefile
-rw-r--r--  1 adi  staff   287 Jan 25 00:09 README.md
-rw-r--r--  1 adi  staff  3007 Jan 24 23:50 proxy.go
Mohammads-MacBook-Pro:parsehub adi$ 
Mohammads-MacBook-Pro:parsehub adi$ 
Mohammads-MacBook-Pro:parsehub adi$ make
GOPATH=/Users/adi/workspace/interview/parsehub go get -u github.com/gorilla/mux
GOPATH=/Users/adi/workspace/interview/parsehub go build proxy.go
Mohammads-MacBook-Pro:parsehub adi$ ls -l
total 13600
-rw-r--r--  1 adi  staff      279 Jan 25 00:06 Makefile
-rw-r--r--  1 adi  staff      287 Jan 25 00:09 README.md
drwxr-xr-x  3 adi  staff      102 Jan 25 00:09 pkg
-rwxr-xr-x  1 adi  staff  6948644 Jan 25 00:09 proxy
-rw-r--r--  1 adi  staff     3007 Jan 24 23:50 proxy.go
drwxr-xr-x  3 adi  staff      102 Jan 25 00:09 src
Mohammads-MacBook-Pro:parsehub adi$
```

NOTE: Due to the behavior of the default HTTP mux in the go standard which doesn't accept paths with
double slashes and results in a 301, I'm using the gorilla/mux package along with its SkipClean
option (http://www.gorillatoolkit.org/pkg/mux#Router.SkipClean).

Using the default standard library router results in the following:

```
Mohammads-MacBook-Pro:parsehub adi$ curl "http://localhost:8080/proxy/b/c"
received/proxy/b/c
Mohammads-MacBook-Pro:parsehub adi$ curl "http://localhost:8080/proxy//a"
<a href="/proxy/a">Moved Permanently</a>.

Mohammads-MacBook-Pro:parsehub adi$ curl "http://localhost:8080/proxy/http://"
<a href="/proxy/http:/">Moved Permanently</a>.
```

# Running

The binary can be executed directly from the command line. The default port the server is launched on
is 8000, and the server also has a timeout associated with the requests it proxies which defaults to 5s.

The PORT and TIMEOUT env vars can be used to override the defaults. The TIMEOUT env var needs to be specified
as a string that can be parsed by golang's ```time.ParseDuration``` method, so it needs to be of the form
`5s` or `500ms` etc.

```
Mohammads-MacBook-Pro:parsehub adi$ ./proxy 
2019/01/25 00:17:54 Server port = 8000
2019/01/25 00:17:54 Server timeout = 5s
2019/01/25 00:17:54 Starting HTTP server
^C2019/01/25 00:17:56 Stopping HTTP Server
Mohammads-MacBook-Pro:parsehub adi$ 
Mohammads-MacBook-Pro:parsehub adi$ 
Mohammads-MacBook-Pro:parsehub adi$ PORT=8080 TIMEOUT=3s ./proxy 
2019/01/25 00:17:59 Server port = 8080
2019/01/25 00:17:59 Server timeout = 3s
2019/01/25 00:17:59 Starting HTTP server
^C2019/01/25 00:18:00 Stopping HTTP Server
Mohammads-MacBook-Pro:parsehub adi$
```

# Sample Request Output

```
Mohammads-MacBook-Pro:parsehub adi$ curl http://localhost:8000/proxy/http://httpbin.org/get
{
  "args": {}, 
  "headers": {
    "Accept-Encoding": "gzip", 
    "Connection": "close", 
    "Host": "httpbin.org", 
    "User-Agent": "curl/7.43.0"
  }, 
  "origin": "99.229.201.24", 
  "url": "http://httpbin.org/get"
}
Mohammads-MacBook-Pro:parsehub adi$ 
Mohammads-MacBook-Pro:parsehub adi$ 
Mohammads-MacBook-Pro:parsehub adi$ curl -X POST -d asdf=blah http://localhost:8000/proxy/http://httpbin.org/post
{
  "args": {}, 
  "data": "asdf=blah", 
  "files": {}, 
  "form": {}, 
  "headers": {
    "Accept-Encoding": "gzip", 
    "Connection": "close", 
    "Host": "httpbin.org", 
    "Transfer-Encoding": "chunked", 
    "User-Agent": "curl/7.43.0"
  }, 
  "json": null, 
  "origin": "99.229.201.24", 
  "url": "http://httpbin.org/post"
}
Mohammads-MacBook-Pro:parsehub adi$ 
Mohammads-MacBook-Pro:parsehub adi$ 
Mohammads-MacBook-Pro:parsehub adi$ curl -v http://localhost:8000/proxy/http://httpbin.org/delay/8
*   Trying ::1...
* Connected to localhost (::1) port 8000 (#0)
> GET /proxy/http://httpbin.org/delay/8 HTTP/1.1
> Host: localhost:8000
> User-Agent: curl/7.43.0
> Accept: */*
> 
< HTTP/1.1 504 Gateway Timeout
< Date: Fri, 25 Jan 2019 05:21:50 GMT
< Content-Length: 57
< Content-Type: text/plain; charset=utf-8
< 
* Connection #0 to host localhost left intact
Get http://httpbin.org/delay/8: context deadline exceededMohammads-MacBook-Pro:parsehub adi$ 
Mohammads-MacBook-Pro:parsehub adi$ 
```

The server logs all requests it receives to stdout:

```
2019/01/25 00:21:13 Server port = 8000
2019/01/25 00:21:13 Server timeout = 5s
2019/01/25 00:21:13 Starting HTTP server
2019/01/25 00:21:21 Received GET request for http://httpbin.org/get
2019/01/25 00:21:34 Received POST request for http://httpbin.org/post
2019/01/25 00:21:45 Received GET request for http://httpbin.org/delay/8
```

Requests to any other endpoints besides /proxy/ result in a 404

```
Mohammads-MacBook-Pro:parsehub adi$ curl http://localhost:8000/
404 page not found
Mohammads-MacBook-Pro:parsehub adi$ 
Mohammads-MacBook-Pro:parsehub adi$ curl http://localhost:8000/proxy
404 page not found
Mohammads-MacBook-Pro:parsehub adi$ 
Mohammads-MacBook-Pro:parsehub adi$ curl http://localhost:8000/test
404 page not found
Mohammads-MacBook-Pro:parsehub adi$
```

Requests to /proxy/ without a specified url results in a HTTP 400 (Bad Request):

```
Mohammads-MacBook-Pro:parsehub adi$ curl -v http://localhost:8000/proxy/
*   Trying ::1...
* Connected to localhost (::1) port 8000 (#0)
> GET /proxy/ HTTP/1.1
> Host: localhost:8000
> User-Agent: curl/7.43.0
> Accept: */*
> 
< HTTP/1.1 400 Bad Request
< Date: Fri, 25 Jan 2019 05:46:57 GMT
< Content-Length: 24
< Content-Type: text/plain; charset=utf-8
< 
* Connection #0 to host localhost left intact
No url specified in pathMohammads-MacBook-Pro:parsehub adi$
```

NOTE: We could make /proxy also result in the same behavior as above, but I'm using the default behavior
of the gorilla/mux router for trailing slashes (http://www.gorillatoolkit.org/pkg/mux#Router.StrictSlash).
