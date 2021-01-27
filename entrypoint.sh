#!/bin/sh
# Github actions supplies inputs prefixed with "INPUT_" as envvars, so setting these in a format for the aws CLI to use.
export AWS_ACCESS_KEY_ID=${INPUT_AWS_ACCESS_KEY_ID}
export AWS_SECRET_ACCESS_KEY=${INPUT_AWS_SECRET_ACCESS_KEY}
go run /main.go