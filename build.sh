#!/bin/bash

export GOPATH=$(pwd)

go build amspackager.go $1
go build ams.go $1

