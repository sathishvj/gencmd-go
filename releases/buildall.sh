#!/bin/bash

version="0.0.1"

rm -rf ./darwin-amd64
rm -rf ./darwin-arm64
rm -rf ./linux-amd64
rm -rf ./linux-arm64
rm -rf ./windows-amd64

GOOS=darwin GOARCH=amd64 go build -ldflags="-X main.Version=${version}" -o "darwin-amd64/gencmd" ..
GOOS=darwin GOARCH=arm64 go build -ldflags="-X main.Version=${version}" -o "darwin-arm64/gencmd" ..

GOOS=linux GOARCH=amd64 go build -ldflags="-X main.Version=${version}" -o "linux-amd64/gencmd" ..
GOOS=linux GOARCH=arm64 go build -ldflags="-X main.Version=${version}" -o "linux-arm64/gencmd" ..

GOOS=windows GOARCH=amd64 go build -ldflags="-X main.Version=${version}" -o "windows-amd64/gencmd.exe" ..
GOOS=windows GOARCH=arm64 go build -ldflags="-X main.Version=${version}" -o "windows-arm64/gencmd.exe" ..
