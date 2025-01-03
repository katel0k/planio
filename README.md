# Planio

## Planbook

That is main database of the application, powered by postgresql. Launching it only requires you to build image from dockerfile and start corresponding container

## Plango

That is server of the application written in go. Compiling is done via `make server`. Running is `./server/build/server`. Mock client is just a client for testing, also written in go

## Planer

That is web frontend for the application written in typescript and powered by React. Compiling it is done via `npm i; npx webpack`. Resulting html is in planer/dist/index.html

### protobufjs-loader

While working with this specific structure of my project, I've stumbled upon an issue. I wanted to keep protos directory as clean as possible, ideally with only .proto files even after compilations. That was not possible until I've altered some settings in this package. While my pr is on review, that is going to remain a submodule

## Protos

Those services are communicating using protobufs described in the protos directory. CLI applications are using binary representation, and web is sending jsons

## Future plans

Ideally in the future plango will be a complete server in go, capable of communicating both with web and a CLI written in rust. Everything in here should be capable of being launched on a remote server, which is going to be achieved via containerization. Just for fun I'm also going to try achieveing hot replacements of at least web applications and maybe of some parts of the server
