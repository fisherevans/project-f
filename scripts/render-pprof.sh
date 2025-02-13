#!/bin/bash
go tool pprof -lines -unit=MB -http=:8080 http://localhost:6060/debug/pprof/heap
