#!/usr/bin/env bash
rm -rf build
cd executor_go
env GOOS=linux GOARCH=amd64 go build -o ../build/executor_go_test main/executor_go.go &&
cd ../build &&
tar -czf executor_go_test.tar.gz executor_go_test
