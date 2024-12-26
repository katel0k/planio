#!/bin/bash

protos_dir=../../protos

protoc -I=$protos_dir --go_out=./build $protos_dir/join.proto $protos_dir/msg.proto
go build -o build/
