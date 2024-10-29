# gecho

[![Go Report Card](https://goreportcard.com/badge/github.com/pdrb/gecho)](https://goreportcard.com/report/github.com/pdrb/gecho)
[![CI](https://github.com/pdrb/gecho/actions/workflows/ci.yml/badge.svg)](https://github.com/pdrb/gecho/actions/workflows/ci.yml)
[![LICENSE](https://img.shields.io/github/license/pdrb/gecho)](https://github.com/pdrb/gecho/blob/main/LICENSE)

Simple http "echo" server written in Go using only the standard library.

The request will be parsed and the response will be a prettified (indented) json with the request data.

Duplicated headers and params are supported and will be comma separated in the response.

The origin IP will be extracted from headers (X-Real-IP, X-Forwarded-For, etc...) or directly from remote address.

## Install

Install compiling from source using Go:

```shell
go install github.com/pdrb/gecho@latest
```

## Usage

```text
Usage: gecho [options]

A simple http "echo" server written in Go

Options:
  -h, --help     Show this help message and exit
  -l, --listen   Listen address (default: ":8090")
  -t, --timeout  Server timeout in seconds (default: 60)
  -v, --version  Show version and exit

Example: gecho --listen 0.0.0.0:80
```

## Example

The following `curl`:

```shell
curl -X POST 'http://localhost:8090/headers?name=John&food=apple&food=banana&age=32' \
    -H 'H1: Header 1' \
    -H 'H1: Repeated Header 1' \
    -H "X-Auth: 1234" \
    -H 'Content-Type: application/json' \
    -d '{"foo": "bar", "foo": "baz"}'
```

Should return a response like:

```json
{
  "data": "{\"foo\": \"bar\", \"foo\": \"baz\"}",
  "headers": {
    "Accept": "*/*",
    "Content-Length": "28",
    "Content-Type": "application/json",
    "H1": "Header 1,Repeated Header 1",
    "User-Agent": "curl/7.81.0",
    "X-Auth": "1234"
  },
  "json": {
    "foo": "bar",
    "foo": "baz"
  },
  "method": "POST",
  "origin": "127.0.0.1",
  "params": {
    "age": "32",
    "food": "apple,banana",
    "name": "John"
  },
  "url": "http://localhost:8090/headers?name=John&food=apple&food=banana&age=32"
}
```
