#!/bin/bash

protoc -I=../protos --go_out=./build ../protos/join.proto
go build -o build/
