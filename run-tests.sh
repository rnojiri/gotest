#!/bin/bash
go test -v -p 1 -count 1 -timeout 120s github.com/rnojiri/gotest/http/
go test -v -p 1 -count 1 -timeout 120s github.com/rnojiri/gotest/tcpudp/
