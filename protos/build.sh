#!/bin/bash

protoc -I=. --go_out="$1"/build ./join.proto ./msg.proto ./plan.proto --experimental_allow_proto3_optional
