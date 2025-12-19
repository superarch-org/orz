#!/bin/sh
set -e

GOOS=linux GOARCH=amd64 go build -o out/linux main.go
GOOS=netbsd GOARCH=amd64 go build -o out/netbsd main.go
GOOS=freebsd GOARCH=amd64 go build -o out/freebsd main.go

scp out/linux vds:/srv/http/superarch.org/orz/linux
scp out/netbsd vds:/srv/http/superarch.org/orz/netbsd
scp out/freebsd vds:/srv/http/superarch.org/orz/freebsd

echo "done"
