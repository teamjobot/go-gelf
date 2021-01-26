go-gelf - GELF Library and Writer for Go
========================================

[GELF] (Graylog Extended Log Format) is an application-level logging
protocol that avoids many of the shortcomings of [syslog]. While it
can be run over any stream or datagram transport protocol, it has
special support ([chunking]) to allow long messages to be split over
multiple datagrams.

History
------
This repo was duplicated from https://github.com/Graylog2/go-gelf/tree/v2.

It has specific changes to work better in the Jobot environment:

- A standardized logging format so data can be extracted from logged message and input better into Seq
- Support for log levels - source package logged everything as Info
- Support for Go function name
- Support for log message id
- Support for https://github.com/op/go-logging used by several Jobot apps
- Support for file and line number - source only worked with standard go log pkg

Versions
--------

v1.0.0
------

This implementation currently supports UDP and TCP as a transport
protocol. TLS is unsupported.

The library provides an API that applications can use to log messages
directly to a Graylog server and an `io.Writer` that can be used to
redirect the standard library's log messages (`os.Stdout`) to a
Graylog server.

[GELF]: http://docs.graylog.org/en/2.2/pages/gelf.html
[syslog]: https://tools.ietf.org/html/rfc5424
[chunking]: http://docs.graylog.org/en/2.2/pages/gelf.html#chunked-gelf


Installing
----------
	cd $GOPATH
    go get github.com/teamjobot/go-gelf

Usage
-----

The easiest way to integrate graylog logging into your go app is by
having your `main` function (or even `init`) call `log.SetOutput()`.
By using an `io.MultiWriter`, we can log to both stdout and graylog -
giving us both centralized and local logs.  (Redundancy is nice).

```golang
package main

import (
	"github.com/op/go-logging"
	"github.com/teamjobot/go-gelf"
	"os"
)

func main() {
	var graylogAddr string

	flag.StringVar(&graylogAddr, "graylog", "", "graylog server addr")
	flag.Parse()

	seqBackend := graylogAddr(graylogAddr)
	fileBackend := newFileBackend("app-name.log")
	stdOutBackend := newStdOutBackend()

	var level = logging.INFO
	setLevel(level, seqBackend, fileBackend, outBackend)

	logging.SetBackend(seqBackend, fileBackend, outBackend)
}

func newGraylogBackend(address string) logging.LeveledBackend {
	// or NewTCPWriter(address)
	var logger, err = NewUDPWriter(address, os.Getenv("jax_environment"))

	if err != nil {
		log.Fatalf("Failed to create UDP writer: %s", err)
	}

	backend := logging.NewLogBackend(logger, "", 0)
	return logging.AddModuleLevel(logging.NewBackendFormatter(backend, gelfFormat()))
}

func gelfFormat() logging.Formatter {
	return logging.MustStringFormatter(gelf.LogFormat)
}

func setLevel(level logging.Level, backends ...logging.LeveledBackend) {
	for _, backend := range backends {
		backend.SetLevel(level, "")
	}
}

```
The above program can be invoked as:

    go run test.go -graylog=localhost:12201

When using UDP messages may be dropped or re-ordered. However, Graylog
server availability will not impact application performance; there is
a small, fixed overhead per log call regardless of whether the target
server is reachable or not.
