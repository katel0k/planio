.PHONY: all server clean

all: server

protos=../protos/*.proto

server:
	protoc -I=../protos --go_out=./ $(protos) --experimental_allow_proto3_optional
	go build -o ./app

clean:
	rm app
	rm -rf protos/
