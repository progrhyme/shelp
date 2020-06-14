#!/usr/bin/env bash

export SHELP_DEBUG=1
exec go run -tags debug *.go "$@"
