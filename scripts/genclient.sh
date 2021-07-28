#!/bin/sh


service=$(basename `pwd`)

cd ../proto

protoc --rpcxclient_out=plugins=rpcx:. "$service".proto
protoc --go_out=plugins=rpcx:../"$service" "$service".proto