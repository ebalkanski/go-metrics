#!/bin/bash

go build -buildvcs=false -o /server ./cmd/...
/server
