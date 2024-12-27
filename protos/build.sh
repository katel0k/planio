#!/bin/bash

protoc -I=. --go_out="$1"/build ./join.proto ./msg.proto
