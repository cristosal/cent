#!/bin/bash

set -a
source .env
templ generate ./templates && go run main.go     
