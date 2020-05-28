#!/usr/bin/env bash

go test -coverprofile=coverage.out ./...
gocov convert coverage.out > coverage.json
