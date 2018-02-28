# Url shortener service

It is just a simple implementation of url shortener like [Google URL Shortener](https://goo.gl/).
Uses SQLite as database storage.

## Build

```
go build shortener.go shortenerService.go
```

## Run

```
./shortener
```

Listens on 0.0.0.0:8080.
Max worker count is limited to 100.

## Usage

### Obtaining a short link

Request:
```
curl -X POST http://localhost:8080/set -H "Content-Type=text/plain" --header "Content-Type: text/plain" --data-binary 'http://www.github.com/'
```

Response:
```
Note: Unnecessary use of -X or --request, POST is already inferred.
*   Trying 127.0.0.1...
* TCP_NODELAY set
* Connected to localhost (127.0.0.1) port 8080 (#0)
> POST /set HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.55.1
> Accept: */*
> Content-Type: text/plain
> Content-Length: 22
> 
* upload completely sent off: 22 out of 22 bytes
< HTTP/1.1 200 OK
< Content-Type: text/plain
< Date: Wed, 28 Feb 2018 18:29:16 GMT
< Content-Length: 10
< 
* Connection #0 to host localhost left intact
rFU6tF488G
```

So, now **rFU6tF488G** is a short for [http://www.github.com/](http://www.github.com/).

## Getting redirect by a short url

Request:
```
curl http://localhost:8080/rFU6tF488G -v
```

Response:
```
*   Trying 127.0.0.1...
* TCP_NODELAY set
* Connected to localhost (127.0.0.1) port 8080 (#0)
> GET /rFU6tF488G HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.55.1
> Accept: */*
> 
< HTTP/1.1 301 Moved Permanently
< Content-Type: text/html; charset=utf-8
< Location: http://www.github.com/
< Date: Wed, 28 Feb 2018 18:37:38 GMT
< Content-Length: 57
< 
<a href="http://www.github.com/">Moved Permanently</a>.

* Connection #0 to host localhost left intact
```