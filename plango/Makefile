.PHONY: all server mock_client

all: server mock_client

protos=../protos/*.proto

server:
	protoc -I=../protos --go_out=./server/build $(protos) --experimental_allow_proto3_optional
	go build -C server -o ./build/

mock_client:
	protoc -I=../protos --go_out=./mock_client/build $(protos) --experimental_allow_proto3_optional
	go build -C mock_client -o ./build/
