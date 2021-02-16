go-gelf - GELF Library and Writer for Go
========================================

[GELF] (Graylog Extended Log Format) is an application-level logging
protocol that avoids many of the shortcomings of [syslog]. While it
can be run over any stream or datagram transport protocol, it has
special support ([chunking]) to allow long messages to be split over
multiple datagrams.

History and Overview
------
This repo was duplicated from https://github.com/Graylog2/go-gelf/tree/v2 (not forked for originally private repo).

It has specific customizations for Jobot company needs but is not company confidential and currently public as such and for easier installs.

Primary customizations:

- A [standardized logging format](gelf.go) so data can be extracted from logged message and input better into Seq
  - [Additional fields are added](https://github.com/teamjobot/go-gelf/blob/main/message.go#L176) that can be filtered and searched separately from log msg
- Support for log levels - source package logged everything as Info
- Support for Go function name
- Support for log message id
- Support for https://github.com/op/go-logging (source is go log pkg only)
  - Excluding it from call stack for log location
  - Handling file and line number appropriately

This library helps strike a good middle ground between the two extremes of:
1. Having to replace a log library and use it everywhere OR
2. Just sending stdout to a GELF endpoint and not be able to add fields or customize message format

By only requiring creating a GELF writer and adding as a backend, apps can quickly send logs that have common
helpful fields and formatted in a way that makes them easier to view.

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
	settings := gelf.Settings {
        Address: address,
		AppName: stringp("my-app"), 
		Env: stringp("uat"), 
		Version: &Version}
	var logger, err = gelf.NewUDPWriter(settings)

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
