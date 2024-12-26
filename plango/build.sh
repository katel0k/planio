#!/bin/bash

protoc -I=../protos --go_out=./build ../protos/join.proto ../protos/msg.proto
go build -o build/
