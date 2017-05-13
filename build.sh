#!/bin/bash

export GOPATH=$(pwd)

go build amspackager.go
go build ams.go

sudo cp amspackager /usr/local/bin/
sudo cp ams /usr/local/bin/

