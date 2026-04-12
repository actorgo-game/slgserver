#!/bin/bash

go mod tidy
go build -o bin/nodes nodes/main.go
