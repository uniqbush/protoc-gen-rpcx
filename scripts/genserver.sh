#!/bin/sh


service=$(basename `pwd`)

cd ../proto

protoc --rpcxserver_out=plugins=rpcx:. "$service".proto