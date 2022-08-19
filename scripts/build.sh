#!/bin/bash

apk add --no-cache git
go build -o bin/main main.go
sh ./scripts/prepare_env.sh
